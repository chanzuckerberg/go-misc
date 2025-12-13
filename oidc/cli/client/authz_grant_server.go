package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"
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

func (s *server) Exchange(ctx context.Context, config *oauth2.Config, code, codeVerifier string) (*oauth2.Token, error) {
	token, err := config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_id", config.ClientID),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging oauth token: %w", err)
	}

	return token, nil
}

// Start will start a webserver to capture oidc response
func (s *server) Start(ctx context.Context, authenticator *AuthorizationGrantAuthenticator, config *oauth2.Config, oauthMaterial *oauthMaterial) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if !bytesAreEqual(oauthMaterial.StateBytes, []byte(req.URL.Query().Get("state"))) {
			http.Error(w, "state did not match", http.StatusBadRequest)
			s.err <- fmt.Errorf("state did not match")
			return
		}

		oauth2Token, err := s.Exchange(ctx, config, req.URL.Query().Get("code"), oauthMaterial.CodeVerifier)
		if err != nil {
			errMsg := "failed to exchange token"
			http.Error(w, errMsg, http.StatusInternalServerError)
			s.err <- fmt.Errorf("%s: %w", errMsg, err)
			return
		}

		claims, idToken, verifiedIDToken, err := idTokenFromOauth2Token(ctx, oauth2Token, authenticator.verifier)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- fmt.Errorf("could not verify ID token: %w", err)
			return
		}

		if !bytesAreEqual([]byte(idToken.Nonce), oauthMaterial.NonceBytes) {
			s.err <- fmt.Errorf("nonce does not match")
			return
		}

		_, err = w.Write([]byte(authenticator.customMessages[oidcStatusSuccess]))
		if err != nil {
			s.err <- err
			return
		}

		slog.Debug("server responded with success message; consuming token")
		s.result <- &Token{
			IDToken: verifiedIDToken,
			Claims:  *claims,
			Token:   oauth2Token,
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
