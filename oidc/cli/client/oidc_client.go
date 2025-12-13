package client

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var DefaultScopes = []string{
	oidc.ScopeOpenID,
	oidc.ScopeOfflineAccess,
	"email",
	"groups",
}

type authenticator interface {
	Authenticate(context.Context, *oauth2.Config) (*Token, error)
}

type OIDClient struct {
	provider      *oidc.Provider
	authenticator authenticator
	oauthConfig   *oauth2.Config
	verifier      *oidc.IDTokenVerifier
}

type OIDCClientConfig struct {
	ClientID  string
	IssuerURL string
	Scopes    []string
}

type OIDCClientOption func(*OIDClient)

func WithDeviceGrantAuthenticator(a *DeviceGrantAuthenticator) OIDCClientOption {
	return func(c *OIDClient) {
		c.authenticator = a
	}
}

func WithAuthzGrantAuthenticator(a *AuthorizationGrantAuthenticator) OIDCClientOption {
	return func(c *OIDClient) {
		c.authenticator = a
	}
}

func NewOIDCClient(ctx context.Context, config *OIDCClientConfig, clientOptions ...OIDCClientOption) (*OIDClient, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	scopes := DefaultScopes
	if len(config.Scopes) != 0 {
		scopes = config.Scopes
	}
	oauthConfig := &oauth2.Config{
		ClientID: config.ClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   scopes,
	}

	oidcConfig := &oidc.Config{
		ClientID:             config.ClientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	verifier := provider.Verifier(oidcConfig)

	defaultAuthenticator, err := NewAuthorizationGrantAuthenticator(
		ctx,
		&AuthorizationGrantConfig{
			ClientID: config.ClientID,
			Provider: provider,
			Verifier: verifier,
			Scopes:   scopes,
			ServerConfig: &ServerConfig{
				// TODO (el): Make these configurable?
				FromPort: 49152,
				ToPort:   49152 + 63,
				Timeout:  30 * time.Second,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating default authenticator: %w", err)
	}

	oidcClient := &OIDClient{
		provider:      provider,
		verifier:      verifier,
		oauthConfig:   oauthConfig,
		authenticator: defaultAuthenticator,
	}

	for _, clientOption := range clientOptions {
		clientOption(oidcClient)
	}

	return oidcClient, nil
}

func idTokenFromOauth2Token(
	ctx context.Context,
	oauth2Token *oauth2.Token,
	verifier *oidc.IDTokenVerifier,
) (*Claims, *oidc.IDToken, string, error) {
	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, "", fmt.Errorf("no id_token found in oauth2 token")
	}

	idToken, err := verifier.Verify(ctx, unverifiedIDToken)
	if err != nil {
		return nil, nil, "", fmt.Errorf("could not verify id token: %w", err)
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
func (c *OIDClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	slog.Debug(fmt.Sprintf("refresh scopes: %#v", c.oauthConfig.Scopes))

	newToken, err := c.refreshToken(ctx, oldToken)
	// if we could refresh successfully, do so.
	// otherwise try a new token
	if err == nil {
		return newToken, nil
	}
	slog.Debug("failed to refresh token, requesting new one", "error", err)

	return c.authenticator.Authenticate(ctx, c.oauthConfig)
}

func (c *OIDClient) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil {
		slog.Debug("nil refresh token, skipping refresh flow")
		return nil, fmt.Errorf("cannot refresh nil token")
	}

	slog.Debug("refresh token found, attempting refresh flow")
	newOauth2Token, err := c.oauthConfig.TokenSource(ctx, token.Token).Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	claims, _, verifiedIDToken, err := idTokenFromOauth2Token(ctx, newOauth2Token, c.verifier)
	if err != nil {
		return nil, err
	}
	slog.Debug("refresh successful")

	return &Token{
		Version: token.Version,
		IDToken: verifiedIDToken,
		Claims:  *claims,
		Token:   newOauth2Token,
	}, nil

}

// Verify verifies an oidc id token
func (c *OIDClient) Verify(ctx context.Context, ourNonce []byte, rawIDToken string) (*oidc.IDToken, error) {
	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("could not verify id token: %w", err)
	}

	if !bytesAreEqual([]byte(idToken.Nonce), ourNonce) {
		return nil, fmt.Errorf("nonce does not match")
	}

	return idToken, nil
}

func bytesAreEqual(this []byte, that []byte) bool {
	return subtle.ConstantTimeCompare(this, that) == 1
}
