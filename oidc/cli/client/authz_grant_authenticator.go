package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type oidcStatus string

const (
	defaultSuccessMessage            = "Signed in successfully! You can now return to CLI."
	oidcStatusSuccess     oidcStatus = "success"
)

var DefaultAuthorizationGrantConfig *AuthorizationGrantConfig = &AuthorizationGrantConfig{
	ServerConfig: &ServerConfig{
		FromPort: 49152,
		ToPort:   49152 + 63,
		Timeout:  30 * time.Second,
		CustomMessages: map[oidcStatus]string{
			oidcStatusSuccess: defaultSuccessMessage,
		},
	},
}

// AuthorizationGrantConfig is required to config a client
type AuthorizationGrantConfig struct {
	ServerConfig *ServerConfig
}

// AuthorizationGrantAuthenticator is an oauth client
type AuthorizationGrantAuthenticator struct {
	server *server
}

type AuthorizationGrantAuthenticatorOption func(*AuthorizationGrantAuthenticator)

func WithSuccessMessage(successMsg string) AuthorizationGrantAuthenticatorOption {
	return func(a *AuthorizationGrantAuthenticator) {
		a.server.config.CustomMessages[oidcStatusSuccess] = successMsg
	}
}

// NewAuthorizationGrantAuthenticator returns a new client
func NewAuthorizationGrantAuthenticator(
	ctx context.Context,
	config *AuthorizationGrantConfig,
	oauth2Config *oauth2.Config,
	authenticatorOptions ...AuthorizationGrantAuthenticatorOption,
) (*AuthorizationGrantAuthenticator, error) {
	log := logging.Get()
	log.Debug("NewAuthorizationGrantAuthenticator: creating authenticator",
		"port_range_from", config.ServerConfig.FromPort,
		"port_range_to", config.ServerConfig.ToPort,
		"timeout", config.ServerConfig.Timeout,
	)

	server, err := newServer(config.ServerConfig)
	if err != nil {
		log.Error("NewAuthorizationGrantAuthenticator: failed to create server",
			"error", err,
		)
		return nil, err
	}
	log.Debug("NewAuthorizationGrantAuthenticator: server created")

	authenticator := &AuthorizationGrantAuthenticator{
		server: server,
	}

	log.Debug("NewAuthorizationGrantAuthenticator: applying options",
		"num_options", len(authenticatorOptions),
	)
	for _, opt := range authenticatorOptions {
		opt(authenticator)
	}

	log.Debug("NewAuthorizationGrantAuthenticator: authenticator created successfully")
	return authenticator, nil
}

// GetBoundPort returns the port we bound to
func (c *AuthorizationGrantAuthenticator) GetBoundPort() int {
	return c.server.port
}

// GetAuthCodeURL gets the url to the oauth2 consent page
func (c *AuthorizationGrantAuthenticator) GetAuthCodeURL(oauthMaterial *oauthMaterial, client *OIDCClient) string {
	log := logging.Get()
	log.Debug("GetAuthCodeURL: generating authorization URL",
		"client_id", client.ClientID,
		"scopes", client.Scopes,
	)

	url := client.AuthCodeURL(
		oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", oauthMaterial.Nonce),
	)

	log.Debug("GetAuthCodeURL: authorization URL generated",
		"url_length", len(url),
	)
	return url
}

// Authenticate will authenticate authenticate with the idp
func (c *AuthorizationGrantAuthenticator) Authenticate(ctx context.Context, client *OIDCClient) (*Token, error) {
	log := logging.Get()
	startTime := time.Now()

	log.Info("Authenticate: starting interactive authentication flow")

	log.Debug("Authenticate: binding to local port")
	err := c.server.Bind()
	if err != nil {
		log.Error("Authenticate: failed to bind to port",
			"error", err,
		)
		return nil, fmt.Errorf("binding to port: %w", err)
	}
	boundPort := c.GetBoundPort()
	log.Debug("Authenticate: bound to port",
		"port", boundPort,
	)

	client.RedirectURL = fmt.Sprintf("http://localhost:%d", boundPort)
	log.Debug("Authenticate: set redirect URL",
		"redirect_url", client.RedirectURL,
	)

	log.Debug("Authenticate: generating OAuth material (state, nonce, PKCE)")
	oauthMaterial, err := newOauthMaterial()
	if err != nil {
		log.Error("Authenticate: failed to generate OAuth material",
			"error", err,
		)
		return nil, err
	}
	log.Debug("Authenticate: OAuth material generated")

	log.Debug("Authenticate: starting local callback server")
	c.server.Start(ctx, client, oauthMaterial)
	log.Debug("Authenticate: local callback server started")

	fmt.Fprintf(os.Stderr, "Opening browser in order to authenticate with Okta, hold on a brief second...\n")
	log.Info("Authenticate: waiting before opening browser",
		"wait_duration_seconds", 2,
	)
	time.Sleep(2 * time.Second)

	// intercept these outputs, send them back on error
	browserStdOut := bytes.NewBuffer(nil)
	browserStdErr := bytes.NewBuffer(nil)
	browser.Stdout = browserStdOut
	browser.Stderr = browserStdErr

	authURL := c.GetAuthCodeURL(oauthMaterial, client)
	log.Info("Authenticate: opening browser for authentication",
		"port", boundPort,
	)
	err = browser.OpenURL(authURL)
	if err != nil {
		log.Error("Authenticate: failed to open browser",
			"error", err,
			"browser_stdout", browserStdOut.String(),
			"browser_stderr", browserStdErr.String(),
		)
		// if we error out, send back stdout, stderr
		io.Copy(os.Stdout, browserStdOut) //nolint:errcheck
		io.Copy(os.Stderr, browserStdErr) //nolint:errcheck
		return nil, fmt.Errorf("opening browser: %w", err)
	}
	log.Debug("Authenticate: browser opened successfully")

	log.Debug("Authenticate: waiting for OAuth callback")
	token, err := c.server.Wait(ctx)
	if err != nil {
		log.Error("Authenticate: failed waiting for callback",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	log.Info("Authenticate: interactive authentication completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
	)
	fmt.Fprintf(os.Stderr, "Successfully authenticated!\n")
	return token, nil
}
