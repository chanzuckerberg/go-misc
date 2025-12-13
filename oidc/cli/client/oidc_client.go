package client

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"

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
	Authenticate(context.Context, *OIDCClient) (*Token, error)
}

type OIDCClient struct {
	authenticator
	*oauth2.Config
	*oidc.IDTokenVerifier
	Scopes []string
}

type OIDCClientOption func(context.Context, *OIDCClient) error

func WithDeviceGrantAuthenticator(a *DeviceGrantAuthenticator) OIDCClientOption {
	return func(ctx context.Context, c *OIDCClient) error {
		c.authenticator = a
		return nil
	}
}

func WithAuthzGrantAuthenticator(a *AuthorizationGrantConfig, authenticatorOptions ...AuthorizationGrantAuthenticatorOption) OIDCClientOption {
	return func(ctx context.Context, c *OIDCClient) error {
		authenticator, err := NewAuthorizationGrantAuthenticator(ctx, a, c.Config, authenticatorOptions...)
		if err != nil {
			return fmt.Errorf("creating default authenticator: %w", err)
		}
		c.authenticator = authenticator
		return nil
	}
}

func WithScopes(scopes []string) OIDCClientOption {
	return func(ctx context.Context, c *OIDCClient) error {
		c.Scopes = scopes
		return nil
	}
}

func NewOIDCClient(ctx context.Context, clientID, issuerURL string, clientOptions ...OIDCClientOption) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	oidcClient := &OIDCClient{
		Scopes: DefaultScopes,
	}
	for _, clientOption := range clientOptions {
		err = clientOption(ctx, oidcClient)
		if err != nil {
			return nil, err
		}
	}

	oidcClient.Config = &oauth2.Config{
		ClientID: clientID,
		Endpoint: provider.Endpoint(),
		Scopes:   oidcClient.Scopes,
	}

	oidcConfig := &oidc.Config{
		ClientID:             clientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	oidcClient.IDTokenVerifier = provider.Verifier(oidcConfig)

	if oidcClient.authenticator == nil {
		// this binds to a port, so only do it at the end once we know they didn't set up an
		// authenticator already
		err = WithAuthzGrantAuthenticator(DefaultAuthorizationGrantConfig)(ctx, oidcClient)
		if err != nil {
			return nil, err
		}
	}

	return oidcClient, nil
}

func (c *OIDCClient) ParseAsIDToken(ctx context.Context, oauth2Token *oauth2.Token) (*Claims, *oidc.IDToken, string, error) {
	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, "", fmt.Errorf("no id_token found in oauth2 token")
	}

	idToken, err := c.Verify(ctx, unverifiedIDToken)
	if err != nil {
		return nil, nil, "", fmt.Errorf("verifying ID token: %w", err)
	}
	verifiedIDToken := unverifiedIDToken // now is verified

	claims := &Claims{}
	err = idToken.Claims(claims)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unmarshalling claims: %w", err)
	}

	return claims, idToken, verifiedIDToken, nil
}

// RefreshToken will fetch a new token
func (c *OIDCClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	slog.Debug("refreshing token", "scopes", c.Scopes, "oldToken", oldToken)

	newToken, err := c.refreshToken(ctx, oldToken)
	// if we could refresh successfully, do so.
	// otherwise try a new token
	if err == nil {
		return newToken, nil
	}

	slog.Debug("failed to refresh token, requesting new one", "error", err)
	return c.authenticator.Authenticate(ctx, c)
}

func (c *OIDCClient) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil {
		slog.Debug("nil refresh token, skipping refresh flow")
		return nil, fmt.Errorf("cannot refresh nil token")
	}

	slog.Debug("refresh token found, attempting refresh flow")
	newOauth2Token, err := c.TokenSource(ctx, token.Token).Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	claims, _, verifiedIDToken, err := c.ParseAsIDToken(ctx, newOauth2Token)
	if err != nil {
		return nil, err
	}

	slog.Debug("successfully refreshed token", "expiry", newOauth2Token.Expiry)
	return &Token{
		Version: token.Version,
		IDToken: verifiedIDToken,
		Claims:  *claims,
		Token:   newOauth2Token,
	}, nil

}

func bytesAreEqual(this []byte, that []byte) bool {
	return subtle.ConstantTimeCompare(this, that) == 1
}
