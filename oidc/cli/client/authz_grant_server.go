package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"
)

// ServerConfig is a server config
type ServerConfig struct {
	FromPort       int
	ToPort         int
	Timeout        time.Duration
	CustomMessages map[oidcStatus]string
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
	log    *slog.Logger

	listener *net.Listener
	port     int
	server   *http.Server

	result chan *Token
	err    chan error
}

// newServer returns a new server
func newServer(ctx context.Context, c *ServerConfig) (*server, error) {
	s := &server{
		config: c,
		log:    logging.FromContext(ctx),
		result: make(chan *Token, 1),
		err:    make(chan error, 1),
	}

	err := c.Validate()
	if err != nil {
		return nil, fmt.Errorf("could not validate new server: %w", err)
	}

	return s, nil
}

// Bind will attempt to open a socket
// on a port in the range FromPort to ToPort
func (s *server) Bind() error {
	var result *multierror.Error

	for port := s.config.FromPort; port <= s.config.ToPort; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		// If we manage to bind, then great use it
		if err == nil {
			s.listener = &l
			s.port = port
			s.log.Debug("server.Bind: bound to port", "port", port)
			return nil
		}
		result = multierror.Append(result, err)
	}

	// at this point we failed to bind to all ports, error out
	return fmt.Errorf("binding to any port in range %d-%d: %w", s.config.FromPort, s.config.ToPort, result)
}

func (s *server) Exchange(ctx context.Context, client *OIDCClient, code, codeVerifier string) (*oauth2.Token, error) {
	token, err := client.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_id", client.ClientID),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging oauth token: %w", err)
	}

	return token, nil
}

// Start will start a webserver to capture oidc response
func (s *server) Start(
	ctx context.Context,
	client *OIDCClient,
	oauthMaterial *oauthMaterial,
) {
	s.log.Debug("server.Start: starting OAuth callback server", "port", s.port)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		s.log.Debug("server.Start: received OAuth callback",
			"method", req.Method,
			"path", req.URL.Path,
			"has_state", req.URL.Query().Get("state") != "",
		)

		if !bytesAreEqual(oauthMaterial.StateBytes, []byte(req.URL.Query().Get("state"))) {
			s.log.Debug("server.Start: state parameter mismatch")
			http.Error(w, "state did not match", http.StatusBadRequest)
			s.err <- fmt.Errorf("state did not match")
			return
		}

		s.log.Debug("server.Start: exchanging authorization code for token")
		oauth2Token, err := s.Exchange(ctx, client, req.URL.Query().Get("code"), oauthMaterial.CodeVerifier)
		if err != nil {
			http.Error(w, "failed to exchange token", http.StatusInternalServerError)
			s.err <- fmt.Errorf("failed to exchange token: %w", err)
			return
		}

		s.log.Debug("server.Start: token exchange successful",
			"token_expiry", oauth2Token.Expiry,
			"has_refresh_token", oauth2Token.RefreshToken != "",
		)

		claims, verifiedIDToken, verifiedIDStr, err := client.ParseAsIDToken(ctx, oauth2Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- fmt.Errorf("could not verify ID token: %w", err)
			return
		}

		if verifiedIDToken == nil {
			s.log.Warn("server.Start: ID token not found")
			s.err <- fmt.Errorf("ID token not found")
			return
		}

		if !bytesAreEqual([]byte(verifiedIDToken.Nonce), oauthMaterial.NonceBytes) {
			s.log.Debug("server.Start: nonce mismatch")
			s.err <- fmt.Errorf("nonce does not match")
			return
		}

		_, err = w.Write([]byte(s.config.CustomMessages[oidcStatusSuccess]))
		if err != nil {
			s.err <- err
			return
		}

		s.log.Debug("server.Start: OAuth flow completed successfully",
			"email", claims.Email,
			"token_expiry", oauth2Token.Expiry,
		)

		s.result <- &Token{
			IDToken: verifiedIDStr,
			Claims:  *claims,
			Token:   oauth2Token,
		}
	})

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		s.log.Debug("server.Start: HTTP server listening", "port", s.port)
		err := s.server.Serve(*s.listener)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
		s.log.Debug("server.Start: HTTP server stopped")
	}()
}

// Wait waits for the oauth2 payload
func (s *server) Wait(ctx context.Context) (*Token, error) {
	s.log.Debug("server.Wait: waiting for OAuth callback", "timeout", s.config.Timeout)

	// nolint:errcheck
	defer s.server.Shutdown(ctx)

	select {
	case err := <-s.err:
		return nil, fmt.Errorf("server.Wait failed: %w", err)
	case token := <-s.result:
		s.log.Debug("server.Wait: received token", "email", token.Claims.Email)
		return token, nil
	case <-time.After(s.config.Timeout):
		s.log.Debug("server.Wait: timeout waiting for callback", "timeout", s.config.Timeout)
		return nil, fmt.Errorf("timed out waiting for oauth callback")
	}
}
