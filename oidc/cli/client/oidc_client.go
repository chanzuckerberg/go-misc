package client

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	// maxRefreshRetries is the number of times to retry token refresh when IDP doesn't return an ID token
	maxRefreshRetries = 5
	// refreshRetryDelay is the delay between refresh retries
	refreshRetryDelay = 15 * time.Second
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
	log *slog.Logger
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
		c.Config.Scopes = scopes
		return nil
	}
}

func NewOIDCClient(ctx context.Context, clientID, issuerURL string, clientOptions ...OIDCClientOption) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID:             clientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	oidcClient := &OIDCClient{
		Config: &oauth2.Config{
			ClientID: clientID,
			Endpoint: provider.Endpoint(),
			Scopes:   DefaultScopes,
		},
		IDTokenVerifier: provider.Verifier(oidcConfig),
		log:             logging.FromContext(ctx),
	}

	for _, clientOption := range clientOptions {
		err = clientOption(ctx, oidcClient)
		if err != nil {
			return nil, err
		}
	}

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
	idTokenStr, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		// id_token is optional in some flows per OIDC spec
		c.log.Warn("ParseAsIDToken: id_token not found in token response")
		return nil, nil, "", nil
	}

	idToken, err := c.Verify(ctx, idTokenStr)
	if err != nil {
		return nil, nil, "", fmt.Errorf("verifying ID token: %w", err)
	}

	claims := &Claims{}
	err = idToken.Claims(claims)
	if err != nil {
		return nil, nil, "", fmt.Errorf("unmarshalling claims: %w", err)
	}

	return claims, idToken, idTokenStr, nil
}

// RefreshToken will fetch a new token
func (c *OIDCClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	// Try refresh_token grant first
	newToken, err := c.refreshToken(ctx, oldToken)
	if err == nil {
		c.log.Debug("OIDCClient.RefreshToken: refreshed via refresh_token grant",
			"new_expiry", newToken.Token.Expiry,
			"email", newToken.Claims.Email,
		)
		return newToken, nil
	}

	// Fall back to interactive authentication
	c.log.Debug("OIDCClient.RefreshToken: refresh_token grant failed, falling back to interactive auth",
		"reason", err.Error(),
	)
	return c.authenticator.Authenticate(ctx, c)
}

func (c *OIDCClient) refreshToken(ctx context.Context, existingToken *Token) (*Token, error) {
	if existingToken == nil {
		return nil, fmt.Errorf("cannot refresh nil token")
	}

	if existingToken.Token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	c.log.Debug("OIDCClient.refreshToken: attempting token refresh",
		"current_expiry", existingToken.Token.Expiry,
		"email", existingToken.Claims.Email,
		"preferred_username", existingToken.Claims.PreferredUsername,
	)

	// Retry loop: sometimes the IDP doesn't return an ID token on the first attempt
	var (
		newOauth2Token *oauth2.Token
		err            error
	)
	for attempt := 1; attempt <= maxRefreshRetries; attempt++ {
		newOauth2Token, err = c.TokenSource(ctx, existingToken.Token).Token()
		if err != nil {
			// This is expected if refresh token is expired - not an error severity
			c.log.Warn("OIDCClient.refreshToken: error refreshing token",
				"attempt", attempt,
				"error", err.Error(),
			)
			return nil, fmt.Errorf("refreshing token: %w", err)
		}

		c.log.Debug("OIDCClient.refreshToken: received new oauth2 token",
			"attempt", attempt,
			"new_expiry", newOauth2Token.Expiry,
			"has_new_refresh_token", newOauth2Token.RefreshToken != "",
		)

		claims, verifiedIDToken, verifiedIDTokenStr, err := c.ParseAsIDToken(ctx, newOauth2Token)
		if err != nil {
			return nil, err
		}

		// If we got an ID token, we're done
		if verifiedIDToken != nil {
			c.log.Debug("OIDCClient.refreshToken: received new ID token",
				"attempt", attempt,
				"new_expiry", newOauth2Token.Expiry,
				"has_new_refresh_token", newOauth2Token.RefreshToken != "",
				"email", claims.Email,
				"preferred_username", claims.PreferredUsername,
			)
			return &Token{
				Version: existingToken.Version,
				IDToken: verifiedIDTokenStr,
				Claims:  *claims,
				Token:   newOauth2Token,
			}, nil
		}

		select {
		case <-time.After(refreshRetryDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	c.log.Debug("OIDCClient.refreshToken: IDP did not return new ID token after retries, reusing existing",
		"attempts", maxRefreshRetries,
	)
	// After all retries, if we still don't have an ID token, use the existing one
	// Sometimes, the IDP won't send a new ID token, per the spec. It's optional.
	// For example, if you're attempting a refresh flow in Okta, but your ID token is
	// not expired, it won't send a new ID token. Additionally, if you are using
	// one of the default applications in Okta and your web session expires, the
	// refresh flow won't send a new ID token.
	return &Token{
		Version: existingToken.Version,
		IDToken: existingToken.IDToken,
		Claims:  existingToken.Claims,
		Token:   newOauth2Token,
	}, nil
}

func bytesAreEqual(this []byte, that []byte) bool {
	return subtle.ConstantTimeCompare(this, that) == 1
}
