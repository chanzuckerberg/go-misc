package client

import (
	"context"
	"fmt"
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

	listener *net.Listener
	port     int
	server   *http.Server

	result chan *Token
	err    chan error
}

// newServer returns a new server
func newServer(c *ServerConfig) (*server, error) {
	log := logging.Get()
	log.Debug("newServer: creating callback server",
		"port_range_from", c.FromPort,
		"port_range_to", c.ToPort,
		"timeout", c.Timeout,
	)

	s := &server{
		config: c,

		result: make(chan *Token, 1),
		err:    make(chan error, 1),
	}

	err := c.Validate()
	if err != nil {
		log.Error("newServer: server config validation failed",
			"error", err,
		)
		return nil, fmt.Errorf("could not validate new server: %w", err)
	}

	log.Debug("newServer: server created successfully")
	return s, nil
}

// Bind will attempt to open a socket
// on a port in the range FromPort to ToPort
func (s *server) Bind() error {
	log := logging.Get()
	log.Debug("server.Bind: attempting to bind to a port",
		"port_range_from", s.config.FromPort,
		"port_range_to", s.config.ToPort,
	)

	var result *multierror.Error

	for port := s.config.FromPort; port <= s.config.ToPort; port++ {
		log.Debug("server.Bind: trying port",
			"port", port,
		)
		l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		// If we manage to bind, then great use it
		if err == nil {
			s.listener = &l
			s.port = port
			log.Info("server.Bind: successfully bound to port",
				"port", port,
			)
			return nil
		}

		log.Debug("server.Bind: failed to bind to port",
			"port", port,
			"error", err,
		)
		result = multierror.Append(result, err)
	}

	// at this point we failed to bind to all ports, error out
	log.Error("server.Bind: failed to bind to any port in range",
		"port_range_from", s.config.FromPort,
		"port_range_to", s.config.ToPort,
		"errors", result.Error(),
	)
	return fmt.Errorf("binding to any port in range %d-%d: %w", s.config.FromPort, s.config.ToPort, result)
}

func (s *server) Exchange(ctx context.Context, client *OIDCClient, code, codeVerifier string) (*oauth2.Token, error) {
	log := logging.Get()
	startTime := time.Now()

	log.Debug("server.Exchange: exchanging authorization code for token",
		"client_id", client.ClientID,
		"code_length", len(code),
	)

	token, err := client.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_id", client.ClientID),
	)
	if err != nil {
		log.Error("server.Exchange: failed to exchange authorization code",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("exchanging oauth token: %w", err)
	}

	log.Debug("server.Exchange: token exchange successful",
		"token_expiry", token.Expiry,
		"has_refresh_token", token.RefreshToken != "",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
	)
	return token, nil
}

// Start will start a webserver to capture oidc response
func (s *server) Start(
	ctx context.Context,
	client *OIDCClient,
	oauthMaterial *oauthMaterial,
) {
	log := logging.Get()
	log.Debug("server.Start: setting up HTTP handler for OAuth callback",
		"port", s.port,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		requestStartTime := time.Now()
		log.Info("server.Start: received OAuth callback request",
			"method", req.Method,
			"url", req.URL.String(),
			"remote_addr", req.RemoteAddr,
		)

		log.Debug("server.Start: validating state parameter")
		if !bytesAreEqual(oauthMaterial.StateBytes, []byte(req.URL.Query().Get("state"))) {
			log.Error("server.Start: state parameter mismatch",
				"expected_length", len(oauthMaterial.StateBytes),
				"received_length", len(req.URL.Query().Get("state")),
			)
			http.Error(w, "state did not match", http.StatusBadRequest)
			s.err <- fmt.Errorf("state did not match")
			return
		}
		log.Debug("server.Start: state parameter validated")

		log.Debug("server.Start: exchanging authorization code for token")
		oauth2Token, err := s.Exchange(ctx, client, req.URL.Query().Get("code"), oauthMaterial.CodeVerifier)
		if err != nil {
			errMsg := "failed to exchange token"
			log.Error("server.Start: token exchange failed",
				"error", err,
			)
			http.Error(w, errMsg, http.StatusInternalServerError)
			s.err <- fmt.Errorf("%s: %w", errMsg, err)
			return
		}
		log.Debug("server.Start: token exchange successful")

		log.Debug("server.Start: parsing OAuth token as ID token")
		claims, idToken, verifiedIDToken, err := client.ParseAsIDToken(ctx, oauth2Token)
		if err != nil {
			log.Error("server.Start: failed to parse/verify ID token",
				"error", err,
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			s.err <- fmt.Errorf("could not verify ID token: %w", err)
			return
		}
		log.Debug("server.Start: ID token parsed successfully")

		log.Debug("server.Start: validating nonce")
		if !bytesAreEqual([]byte(idToken.Nonce), oauthMaterial.NonceBytes) {
			log.Error("server.Start: nonce mismatch")
			s.err <- fmt.Errorf("nonce does not match")
			return
		}
		log.Debug("server.Start: nonce validated")

		_, err = w.Write([]byte(s.config.CustomMessages[oidcStatusSuccess]))
		if err != nil {
			log.Error("server.Start: failed to write success response",
				"error", err,
			)
			s.err <- err
			return
		}

		log.Info("server.Start: OAuth callback completed successfully, sending token",
			"elapsed_ms", time.Since(requestStartTime).Milliseconds(),
			"token_expiry", oauth2Token.Expiry,
		)
		s.result <- &Token{
			IDToken: verifiedIDToken,
			Claims:  *claims,
			Token:   oauth2Token,
		}
		log.Debug("server.Start: token sent to result channel")
	})

	s.server = &http.Server{
		Handler: mux,
	}

	log.Debug("server.Start: starting HTTP server goroutine")
	go func() {
		log.Debug("server.Start: HTTP server listening",
			"port", s.port,
		)
		err := s.server.Serve(*s.listener)
		if err != nil && err != http.ErrServerClosed {
			log.Error("server.Start: HTTP server error",
				"error", err,
			)
			panic(err)
		}
		log.Debug("server.Start: HTTP server stopped")
	}()
}

// Wait waits for the oauth2 payload
func (s *server) Wait(ctx context.Context) (*Token, error) {
	log := logging.Get()
	startTime := time.Now()

	log.Debug("server.Wait: waiting for OAuth callback",
		"timeout", s.config.Timeout,
	)

	// nolint:errcheck
	defer func() {
		log.Debug("server.Wait: shutting down HTTP server")
		s.server.Shutdown(ctx)
	}()

	select {
	case err := <-s.err:
		log.Error("server.Wait: received error from callback handler",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("server.Wait failed: %w", err)
	case token := <-s.result:
		log.Info("server.Wait: received token from callback",
			"elapsed_ms", time.Since(startTime).Milliseconds(),
			"token_expiry", token.Token.Expiry,
		)
		return token, nil
	case <-time.After(s.config.Timeout):
		log.Error("server.Wait: timed out waiting for OAuth callback",
			"timeout", s.config.Timeout,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("timed out waiting for oauth callback")
	}
}
