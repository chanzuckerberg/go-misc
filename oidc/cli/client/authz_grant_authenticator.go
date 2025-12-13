package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

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
	server, err := newServer(config.ServerConfig)
	if err != nil {
		return nil, err
	}

	authenticator := &AuthorizationGrantAuthenticator{
		server: server,
	}

	for _, opt := range authenticatorOptions {
		opt(authenticator)
	}

	oauth2Config.RedirectURL = fmt.Sprintf("http://localhost:%d", authenticator.GetBoundPort())

	return authenticator, nil
}

// GetBoundPort returns the port we bound to
func (c *AuthorizationGrantAuthenticator) GetBoundPort() int {
	return c.server.port
}

// GetAuthCodeURL gets the url to the oauth2 consent page
func (c *AuthorizationGrantAuthenticator) GetAuthCodeURL(oauthMaterial *oauthMaterial, client *OIDCClient) string {
	return client.AuthCodeURL(
		oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", oauthMaterial.Nonce),
	)
}

// Authenticate will authenticate authenticate with the idp
func (c *AuthorizationGrantAuthenticator) Authenticate(ctx context.Context, client *OIDCClient) (*Token, error) {
	oauthMaterial, err := newOauthMaterial()
	if err != nil {
		return nil, err
	}

	c.server.Start(ctx, client, oauthMaterial)
	fmt.Fprintf(os.Stderr, "Opening browser in order to authenticate with Okta, hold on a brief second...\n")
	time.Sleep(2 * time.Second)

	// intercept these outputs, send them back on error
	browserStdOut := bytes.NewBuffer(nil)
	browserStdErr := bytes.NewBuffer(nil)
	browser.Stdout = browserStdOut
	browser.Stderr = browserStdErr

	err = browser.OpenURL(c.GetAuthCodeURL(oauthMaterial, client))
	if err != nil {
		// if we error out, send back stdout, stderr
		io.Copy(os.Stdout, browserStdOut) //nolint:errcheck
		io.Copy(os.Stderr, browserStdErr) //nolint:errcheck
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	token, err := c.server.Wait(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Successfully authenticated!\n")
	return token, nil
}
