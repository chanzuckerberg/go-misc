package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
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
		return fmt.Errorf("toPort %d must be >= fromPort %d", c.ToPort, c.FromPort)
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

		result: make(chan *Token, 1),
		err:    make(chan error, 1),
	}

	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("could not validate new server: %w", err)
	}

	err = s.bind(c)
	if err != nil {
		return nil, fmt.Errorf("could not open a port: %w", err)
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
	return fmt.Errorf("failed to bind to any port in range %d-%d: %w", c.FromPort, c.ToPort, result)
}

// GetBoundPort returns the port we bound to
func (s *server) GetBoundPort() int {
	return s.port
}

// Start will start a webserver to capture oidc response
func (s *server) Start(ctx context.Context, oidcClient *AuthorizationGrantClient, oauthMaterial *oauthMaterial) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := oidcClient.ValidateState(
			oauthMaterial.StateBytes,
			[]byte(req.URL.Query().Get("state")))
		if err != nil {
			http.Error(w, "state did not match", http.StatusBadRequest)
			s.err <- fmt.Errorf("state did not match: %w", err)
			return
		}

		oauth2Token, err := oidcClient.Exchange(ctx, req.URL.Query().Get("code"), oauthMaterial.CodeVerifier)
		if err != nil {
			errMsg := "failed to exchange token"
			http.Error(w, errMsg, http.StatusInternalServerError)
			s.err <- fmt.Errorf("%s: %w", errMsg, err)
			return
		}

		claims, idToken, verifiedIDToken, err := oidcClient.idTokenFromOauth2Token(ctx, oauth2Token, oauthMaterial.NonceBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- fmt.Errorf("could not verify ID token: %w", err)
			return
		}

		_, err = w.Write([]byte(oidcClient.customMessages[oidcStatusSuccess]))
		if err != nil {
			s.err <- err
			return
		}

		slog.Debug("server responded with success message; consuming token")
		s.result <- &Token{
			Expiry:       idToken.Expiry,
			IDToken:      verifiedIDToken,
			AccessToken:  oauth2Token.AccessToken,
			RefreshToken: oauth2Token.RefreshToken,
			Claims:       *claims,
		}
		slog.Debug("token consumed")
	})

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		err := s.server.Serve(*s.listener)
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			panic(err)
		}
	}()
}

// Wait waits for the oauth2 payload
func (s *server) Wait(ctx context.Context) (*Token, error) {
	// nolint:errcheck
	defer s.server.Shutdown(ctx)

	select {
	case err := <-s.err:
		return nil, fmt.Errorf("server.Wait failed: %w", err)
	case token := <-s.result:
		return token, nil
	case <-time.After(s.config.Timeout):
		return nil, fmt.Errorf("timed out waiting for oauth callback")
	}
}
