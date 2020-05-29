package client

import (
	"context"
	"crypto/subtle"
	"fmt"

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

	oauthMaterial *oauthMaterial
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

	oauthMaterial, err := newOauthMaterial()
	if err != nil {
		return nil, err
	}

	oidcConfig := &oidc.Config{
		ClientID:             config.ClientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	verifier := provider.Verifier(oidcConfig)

	return &Client{
		provider:      provider,
		verifier:      verifier,
		oauthConfig:   oauthConfig,
		oauthMaterial: oauthMaterial,

		server: server,
	}, nil
}

func (c *Client) idTokenFromOauth2Token(
	ctx context.Context,
	oauth2Token *oauth2.Token,
) (*Claims, *oidc.IDToken, string, error) {
	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, "", fmt.Errorf("no id_token found in oauth2 token")
	}

	idToken, err := c.Verify(ctx, unverifiedIDToken)
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not verify id token")
	}

	verifiedIDToken := unverifiedIDToken // now is verified
	claims := &Claims{}

	err = idToken.Claims(claims)
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not verify claims")
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
	logrus.Warnf("failed to refresh token %s", err)

	return c.Authenticate(ctx)
}

func (c *Client) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil {
		logrus.Info("nil refresh token")
		return nil, errors.New("cannot refresh nil token")
	}
	logrus.Info("attempting refresh token")

	oauthToken := &oauth2.Token{}
	tokenSource := c.oauthConfig.TokenSource(ctx, oauthToken)

	newOauth2Token, err := tokenSource.Token()
	if err != nil {
		return nil, errors.Wrap(err, "could not refresh token")
	}

	claims, idToken, verifiedIDToken, err := c.idTokenFromOauth2Token(ctx, newOauth2Token)
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
func (c *Client) GetAuthCodeURL() string {
	return c.oauthConfig.AuthCodeURL(
		c.oauthMaterial.State,
		oauth2.SetAuthURLParam("grant_type", "refresh_token"),
		oauth2.SetAuthURLParam("code_challenge", c.oauthMaterial.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", c.oauthMaterial.Nonce),
	)
}

// ValidateState validates the state from the authorize request
func (c *Client) ValidateState(otherState string) error {
	if !c.bytesAreEqual(c.oauthMaterial.StateBytes, []byte(otherState)) {
		return errors.New("invalid state")
	}
	return nil
}

// Exchange will exchange a token
func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("code_verifier", c.oauthMaterial.CodeVerifier),
		oauth2.SetAuthURLParam("client_id", c.oauthConfig.ClientID),
	)
	return token, errors.Wrap(err, "failed to exchange oauth token")
}

func (c *Client) bytesAreEqual(this []byte, that []byte) bool {
	return 1 == subtle.ConstantTimeCompare(this, that)
}

// Verify verifies an oidc id token
func (c *Client) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "could not verify id token")
	}
	if !c.bytesAreEqual([]byte(idToken.Nonce), c.oauthMaterial.NonceBytes) {
		return nil, errors.Errorf("nonce does not match")
	}
	return idToken, nil
}

// Authenticate will authenticate authenticate with the idp
func (c *Client) Authenticate(ctx context.Context) (*Token, error) {
	c.server.Start(ctx, c)

	err := browser.OpenURL(c.GetAuthCodeURL())
	if err != nil {
		return nil, errors.Wrap(err, "could not open browser")
	}

	return c.server.Wait(ctx)
}
