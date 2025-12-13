package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type oidcStatus string

const (
	defaultSuccessMessage            = "Signed in successfully! You can now return to CLI."
	oidcStatusSuccess     oidcStatus = "success"
)

// AuthorizationGrantConfig is required to config a client
type AuthorizationGrantConfig struct {
	ClientID     string
	Verifier     *oidc.IDTokenVerifier
	Provider     *oidc.Provider
	Scopes       []string
	ServerConfig *ServerConfig
}

// AuthorizationGrantAuthenticator is an oauth client
type AuthorizationGrantAuthenticator struct {
	verifier *oidc.IDTokenVerifier
	server   *server

	// Extra configuration options
	customMessages map[oidcStatus]string
}

type AuthorizationGrantAuthenticatorOption func(*AuthorizationGrantAuthenticator)

var SetSuccessMessage = func(successMessage string) AuthorizationGrantAuthenticatorOption {
	return func(c *AuthorizationGrantAuthenticator) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}

// NewAuthorizationGrantAuthenticator returns a new client
func NewAuthorizationGrantAuthenticator(
	ctx context.Context,
	config *AuthorizationGrantConfig,
	authenticatorOptions ...AuthorizationGrantAuthenticatorOption,
) (*AuthorizationGrantAuthenticator, error) {
	server, err := newServer(config.ServerConfig)
	if err != nil {
		return nil, err
	}

	authenticator := &AuthorizationGrantAuthenticator{
		server: server,
		customMessages: map[oidcStatus]string{
			oidcStatusSuccess: defaultSuccessMessage,
		},
		verifier: config.Verifier,
	}

	for _, clientOption := range authenticatorOptions {
		clientOption(authenticator)
	}

	return authenticator, nil
}

// GetAuthCodeURL gets the url to the oauth2 consent page
func (c *AuthorizationGrantAuthenticator) GetAuthCodeURL(oauthMaterial *oauthMaterial, config *oauth2.Config) string {
	return config.AuthCodeURL(
		oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", oauthMaterial.Nonce),
	)
}

// Verify verifies an oidc id token
func (c *AuthorizationGrantAuthenticator) Verify(ctx context.Context, ourNonce []byte, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("could not verify id token: %w", err)
	}

	if !bytesAreEqual([]byte(idToken.Nonce), ourNonce) {
		return nil, fmt.Errorf("nonce does not match")
	}

	return idToken, nil
}

// Authenticate will authenticate authenticate with the idp
func (c *AuthorizationGrantAuthenticator) Authenticate(ctx context.Context, config *oauth2.Config) (*Token, error) {
	oauthMaterial, err := newOauthMaterial()
	if err != nil {
		return nil, err
	}

	c.server.Start(ctx, c, config, oauthMaterial)
	fmt.Fprintf(os.Stderr, "Opening browser in order to authenticate with Okta, hold on a brief second...\n")
	time.Sleep(2 * time.Second)

	// intercept these outputs, send them back on error
	browserStdOut := bytes.NewBuffer(nil)
	browserStdErr := bytes.NewBuffer(nil)
	browser.Stdout = browserStdOut
	browser.Stderr = browserStdErr

	err = browser.OpenURL(c.GetAuthCodeURL(oauthMaterial, config))
	if err != nil {
		// if we error out, send back stdout, stderr
		io.Copy(os.Stdout, browserStdOut) //nolint:errcheck
		io.Copy(os.Stderr, browserStdErr) //nolint:errcheck
		return nil, fmt.Errorf("could not open browser: %w", err)
	}

	token, err := c.server.Wait(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Successfully authenticated!\n")
	return token, nil
}
