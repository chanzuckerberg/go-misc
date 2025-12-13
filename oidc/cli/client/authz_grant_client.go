package client

import (
	"bytes"
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// AuthorizationGrantClient is an oauth client
type AuthorizationGrantClient struct {
	provider    *oidc.Provider
	OauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	server      *server

	// Extra configuration options
	customMessages map[oidcStatus]string
}

// Config is required to config a client
type Config struct {
	ClientID  string
	IssuerURL string

	ServerConfig *ServerConfig
}

// NewAuthorizationGrantClient returns a new client
func NewAuthorizationGrantClient(ctx context.Context, config *Config, clientOptions ...Option) (*AuthorizationGrantClient, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	server, err := newServer(config.ServerConfig)
	if err != nil {
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:    config.ClientID,
		RedirectURL: fmt.Sprintf("http://localhost:%d", server.GetBoundPort()),
		Endpoint:    provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			oidc.ScopeOfflineAccess,
			"email",
			"groups",
		},
	}

	oidcConfig := &oidc.Config{
		ClientID:             config.ClientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	verifier := provider.Verifier(oidcConfig)

	clientConfig := &AuthorizationGrantClient{
		provider:    provider,
		verifier:    verifier,
		OauthConfig: oauthConfig,

		server: server,
		customMessages: map[oidcStatus]string{
			oidcStatusSuccess: defaultSuccessMessage,
		},
	}

	for _, clientOption := range clientOptions {
		clientOption(clientConfig)
	}

	return clientConfig, nil
}

func (c *AuthorizationGrantClient) idTokenFromOauth2Token(
	ctx context.Context,
	oauth2Token *oauth2.Token,
	ourNonce []byte,
) (*Claims, *oidc.IDToken, string, error) {
	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, "", fmt.Errorf("no id_token found in oauth2 token")
	}

	idToken, err := c.Verify(ctx, ourNonce, unverifiedIDToken)
	if err != nil {
		return nil, nil, "", fmt.Errorf("verifying id token: %w", err)
	}

	verifiedIDToken := unverifiedIDToken // now is verified
	claims := &Claims{}

	err = idToken.Claims(claims)
	if err != nil {
		return nil, nil, "", fmt.Errorf("verifying claims: %w", err)
	}
	return claims, idToken, verifiedIDToken, nil
}

// RefreshToken will fetch a new token
func (c *AuthorizationGrantClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	slog.Debug(fmt.Sprintf("refresh scopes: %#v", c.OauthConfig.Scopes))

	newToken, err := c.refreshToken(ctx, oldToken)
	// if we could refresh successfully, do so.
	// otherwise try a new token
	if err == nil {
		return newToken, nil
	}
	slog.Debug("failed to refresh token, requesting new one", "error", err)

	return c.Authenticate(ctx)
}

func (c *AuthorizationGrantClient) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil {
		slog.Debug("nil refresh token, skipping refresh flow")
		return nil, fmt.Errorf("cannot refresh nil token")
	}
	slog.Debug("refresh token found, attempting refresh flow")

	oauthToken := &oauth2.Token{
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	tokenSource := c.OauthConfig.TokenSource(ctx, oauthToken)

	newOauth2Token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	// We don't have a nonce in this flow since we're refreshing
	//    our refresh token -- auth already happened
	zeroNonce := []byte{}
	claims, idToken, verifiedIDToken, err := c.idTokenFromOauth2Token(
		ctx,
		newOauth2Token,
		zeroNonce)
	if err != nil {
		return nil, err
	}
	slog.Debug("refresh successful")

	return &Token{
		Version: token.Version,
		Expiry:  idToken.Expiry,

		IDToken:      verifiedIDToken,
		AccessToken:  newOauth2Token.AccessToken,
		RefreshToken: newOauth2Token.RefreshToken,
		Claims:       *claims,
	}, nil

}

// GetAuthCodeURL gets the url to the oauth2 consent page
func (c *AuthorizationGrantClient) GetAuthCodeURL(oauthMaterial *oauthMaterial) string {
	return c.OauthConfig.AuthCodeURL(
		oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", oauthMaterial.Nonce),
	)
}

// ValidateState validates the state from the authorize request
func (c *AuthorizationGrantClient) ValidateState(ourState []byte, otherState []byte) error {
	if !c.bytesAreEqual(ourState, otherState) {
		return fmt.Errorf("invalid state")
	}
	return nil
}

// Exchange will exchange a token
func (c *AuthorizationGrantClient) Exchange(ctx context.Context, code string, codeVerifier string) (*oauth2.Token, error) {
	token, err := c.OauthConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_id", c.OauthConfig.ClientID),
	)
	return token, fmt.Errorf("exchanging oauth token: %w", err)
}

func (c *AuthorizationGrantClient) bytesAreEqual(this []byte, that []byte) bool {
	return subtle.ConstantTimeCompare(this, that) == 1
}

// Verify verifies an oidc id token
func (c *AuthorizationGrantClient) Verify(ctx context.Context, ourNonce []byte, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("could not verify id token: %w", err)
	}

	if !c.bytesAreEqual([]byte(idToken.Nonce), ourNonce) {
		return nil, fmt.Errorf("nonce does not match")
	}

	return idToken, nil
}

// Authenticate will authenticate authenticate with the idp
func (c *AuthorizationGrantClient) Authenticate(ctx context.Context) (*Token, error) {
	oauthMaterial, err := newOauthMaterial()
	if err != nil {
		return nil, err
	}

	c.server.Start(ctx, c, oauthMaterial)
	fmt.Fprintf(os.Stderr, "Opening browser in order to authenticate with Okta, hold on a brief second...\n")
	time.Sleep(2 * time.Second)

	// intercept these outputs, send them back on error
	browserStdOut := bytes.NewBuffer(nil)
	browserStdErr := bytes.NewBuffer(nil)
	browser.Stdout = browserStdOut
	browser.Stderr = browserStdErr

	err = browser.OpenURL(c.GetAuthCodeURL(oauthMaterial))
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
