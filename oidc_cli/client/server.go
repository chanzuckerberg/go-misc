package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
	"html/template"
	"path"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ServerConfig is a server config
type ServerConfig struct {
	FromPort int
	ToPort   int
	Timeout  time.Duration
}

// Validate validates the config
func (c *ServerConfig) Validate() error {
	if c.ToPort < c.FromPort {
		return errors.Errorf("toPort %d must be >= fromPort %d", c.ToPort, c.FromPort)
	}
	return nil
}

// Server is a server on localhost to capture oauth redirects
type server struct {
	config *ServerConfig

	listener *net.Listener
	port     int
	server   *http.Server

	result chan *Token
	err    chan error
}

// newServer returns a new server
func newServer(c *ServerConfig) (*server, error) {
	s := &server{
		config: c,

		result: make(chan *Token),
		err:    make(chan error),
	}

	err := c.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "could not validate new server")
	}

	err = s.bind(c)
	if err != nil {
		return nil, errors.Wrap(err, "could not open a port")
	}

	return s, nil
}

// bind will attempt to open a socket
// on a port in the range FromPort to ToPort
func (s *server) bind(c *ServerConfig) error {
	var result *multierror.Error
	for port := c.FromPort; port <= c.ToPort; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		// If we manage to bind, then great use it
		if err == nil {
			s.listener = &l
			s.port = port
			return nil
		}
		result = multierror.Append(result, err)
	}

	// at this point we failed to bind to all ports, error out
	return errors.Wrapf(
		result,
		"failed to bind to any port in range %d-%d", c.FromPort, c.ToPort)
}

// GetBoundPort returns the port we bound to
func (s *server) GetBoundPort() int {
	return s.port
}

// Start will start a webserver to capture oidc response
func (s *server) Start(ctx context.Context, oidcClient *Client, oauthMaterial *oauthMaterial) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := oidcClient.ValidateState(
			oauthMaterial.StateBytes,
			[]byte(req.URL.Query().Get("state")))
		if err != nil {
			http.Error(w, "state did not match", http.StatusBadRequest)
			s.err <- errors.Wrap(err, "state did not match")
			return
		}

		oauth2Token, err := oidcClient.Exchange(ctx, req.URL.Query().Get("code"), oauthMaterial.CodeVerifier)
		if err != nil {
			errMsg := "failed to exchange token"
			http.Error(w, errMsg, http.StatusInternalServerError)
			s.err <- errors.Wrap(err, errMsg)
			return
		}

		claims, idToken, verifiedIDToken, err := oidcClient.idTokenFromOauth2Token(ctx, oauth2Token, oauthMaterial.NonceBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- errors.Wrap(err, "could not verify ID token")
			return
		}
		fp := path.Join("templates", "success.html")
		tmpl, err := template.ParseFiles(fp)
		if err != nil {
		    http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- errors.Wrap(err, "could not render html template")
		    return
		}
		if err := tmpl.Execute(w, s.result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		s.result <- &Token{
			Expiry:       idToken.Expiry,
			IDToken:      verifiedIDToken,
			AccessToken:  oauth2Token.AccessToken,
			RefreshToken: oauth2Token.RefreshToken,
			Claims:       *claims,
		}
	})

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		err := s.server.Serve(*s.listener)
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("server error: %s", err.Error())
		}
	}()
}

// Wait waits for the oauth2 payload
func (s *server) Wait(ctx context.Context) (*Token, error) {
	// nolint:errcheck
	defer s.server.Shutdown(ctx)

	select {
	case err := <-s.err:
		return nil, errors.Wrap(err, "server.Wait failed")
	case token := <-s.result:
		return token, nil
	case <-time.After(s.config.Timeout):
		return nil, errors.New("timed out waiting for oauth callback")
	}
}
