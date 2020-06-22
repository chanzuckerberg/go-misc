package client

import (
	"bytes"
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Client is an oauth client
type Client struct {
	provider    *oidc.Provider
	oauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	server      *server
}

// Config is required to config a client
type Config struct {
	ClientID  string
	IssuerURL string

	ServerConfig *ServerConfig
}

// NewClient returns a new client
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not create oidc provider")
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

	return &Client{
		provider:    provider,
		verifier:    verifier,
		oauthConfig: oauthConfig,

		server: server,
	}, nil
}

func (c *Client) idTokenFromOauth2Token(
	ctx context.Context,
	oauth2Token *oauth2.Token,
	ourNonce []byte,
) (*Claims, *oidc.IDToken, string, error) {
	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, "", errors.New("no id_token found in oauth2 token")
	}

	idToken, err := c.Verify(ctx, ourNonce, unverifiedIDToken)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "could not verify id token")
	}

	verifiedIDToken := unverifiedIDToken // now is verified
	claims := &Claims{}

	err = idToken.Claims(claims)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "could not verify claims")
	}
	return claims, idToken, verifiedIDToken, nil
}

// RefreshToken will fetch a new token
func (c *Client) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	newToken, err := c.refreshToken(ctx, oldToken)
	// if we could refresh successfully, do so.
	// otherwise try a new token
	if err == nil {
		return newToken, nil
	}
	logrus.WithError(err).Error("failed to refresh token, requesting new one")

	return c.Authenticate(ctx)
}

func (c *Client) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil {
		logrus.Debug("nil refresh token, skipping refresh flow")
		return nil, errors.New("cannot refresh nil token")
	}
	logrus.Debug("refresh token found, attempting refresh flow")

	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	tokenSource := c.oauthConfig.TokenSource(ctx, oauthToken)

	newOauth2Token, err := tokenSource.Token()
	if err != nil {
		return nil, errors.Wrap(err, "could not refresh token")
	}

	// We don't have a nonce in this flow since we're refreshing our refresh token -- auth already happened
	claims, idToken, verifiedIDToken, err := c.idTokenFromOauth2Token(ctx, newOauth2Token, []byte{})
	if err != nil {
		return nil, err
	}

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
func (c *Client) GetAuthCodeURL(oauthMaterial *oauthMaterial) string {
	return c.oauthConfig.AuthCodeURL(
		oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", oauthMaterial.Nonce),
	)
}

// ValidateState validates the state from the authorize request
func (c *Client) ValidateState(ourState []byte, otherState []byte) error {
	if !c.bytesAreEqual(ourState, otherState) {
		return errors.New("invalid state")
	}
	return nil
}

// Exchange will exchange a token
func (c *Client) Exchange(ctx context.Context, code string, codeVerifier string) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_id", c.oauthConfig.ClientID),
	)
	return token, errors.Wrap(err, "failed to exchange oauth token")
}

func (c *Client) bytesAreEqual(this []byte, that []byte) bool {
	return 1 == subtle.ConstantTimeCompare(this, that)
}

// Verify verifies an oidc id token
func (c *Client) Verify(ctx context.Context, ourNonce []byte, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "could not verify id token")
	}

	if !c.bytesAreEqual([]byte(idToken.Nonce), ourNonce) {
		return nil, errors.Errorf("nonce does not match")
	}

	return idToken, nil
}

// Authenticate will authenticate authenticate with the idp
func (c *Client) Authenticate(ctx context.Context) (*Token, error) {
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
		return nil, errors.Wrap(err, "could not open browser")
	}
	token, err := c.server.Wait(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Successfully authenticated!\n")
	return token, nil
}
