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
	Authenticate(context.Context, *OIDCClient) (*Token, error)
}

type OIDCClient struct {
	authenticator
	*oauth2.Config
	*oidc.IDTokenVerifier
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
	log := slog.Default()
	startTime := time.Now()

	log.Debug("NewOIDCClient: creating OIDC client",
		"client_id", clientID,
		"issuer_url", issuerURL,
		"num_options", len(clientOptions),
	)

	log.Debug("NewOIDCClient: creating OIDC provider")
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		log.Error("NewOIDCClient: creating OIDC provider",
			"error", err,
			"issuer_url", issuerURL,
		)
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}
	log.Debug("NewOIDCClient: OIDC provider created successfully",
		"endpoint_auth_url", provider.Endpoint().AuthURL,
		"endpoint_token_url", provider.Endpoint().TokenURL,
	)

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
	}

	log.Debug("NewOIDCClient: applying client options",
		"scopes", DefaultScopes,
	)
	for i, clientOption := range clientOptions {
		log.Debug("NewOIDCClient: applying client option",
			"option_index", i,
		)
		err = clientOption(ctx, oidcClient)
		if err != nil {
			log.Error("NewOIDCClient: applying client option",
				"error", err,
				"option_index", i,
			)
			return nil, err
		}
	}

	if oidcClient.authenticator == nil {
		// this binds to a port, so only do it at the end once we know they didn't set up an
		// authenticator already
		log.Debug("NewOIDCClient: no authenticator set, creating default AuthzGrantAuthenticator")
		err = WithAuthzGrantAuthenticator(DefaultAuthorizationGrantConfig)(ctx, oidcClient)
		if err != nil {
			log.Error("NewOIDCClient: creating default authenticator",
				"error", err,
			)
			return nil, err
		}
		log.Debug("NewOIDCClient: default authenticator created")
	}

	log.Debug("NewOIDCClient: OIDC client created successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"scopes", oidcClient.Scopes,
	)
	return oidcClient, nil
}

func (c *OIDCClient) ParseAsIDToken(ctx context.Context, oauth2Token *oauth2.Token) (*Claims, *oidc.IDToken, string, error) {
	log := slog.Default()
	log.Debug("ParseAsIDToken: extracting id_token from oauth2 token")

	unverifiedIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Error("ParseAsIDToken: extracting id_token from oauth2 token")
		return nil, nil, "", fmt.Errorf("no id_token found in oauth2 token")
	}
	log.Debug("ParseAsIDToken: id_token extracted",
		"token_length", len(unverifiedIDToken),
	)

	log.Debug("ParseAsIDToken: verifying ID token")
	idToken, err := c.Verify(ctx, unverifiedIDToken)
	if err != nil {
		log.Error("ParseAsIDToken: verifying ID token",
			"error", err,
		)
		return nil, nil, "", fmt.Errorf("verifying ID token: %w", err)
	}
	verifiedIDToken := unverifiedIDToken // now is verified
	log.Debug("ParseAsIDToken: ID token verified successfully",
		"issuer", idToken.Issuer,
		"subject", idToken.Subject,
		"expiry", idToken.Expiry,
	)

	log.Debug("ParseAsIDToken: extracting claims from ID token")
	claims := &Claims{}
	err = idToken.Claims(claims)
	if err != nil {
		log.Error("ParseAsIDToken: unmarshalling claims",
			"error", err,
		)
		return nil, nil, "", fmt.Errorf("unmarshalling claims: %w", err)
	}
	log.Debug("ParseAsIDToken: claims extracted successfully",
		"email", claims.Email,
	)

	return claims, idToken, verifiedIDToken, nil
}

// RefreshToken will fetch a new token
func (c *OIDCClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	log := slog.Default()
	startTime := time.Now()

	log.Debug("RefreshToken: attempting to refresh token",
		"scopes", c.Scopes,
		"has_old_token", oldToken != nil,
	)
	if oldToken != nil {
		log.Debug("RefreshToken: old token details",
			"old_token_expiry", oldToken.Token.Expiry,
			"has_refresh_token", oldToken.Token.RefreshToken != "",
		)
	}

	newToken, err := c.refreshToken(ctx, oldToken)
	// if we could refresh successfully, do so.
	// otherwise try a new token
	if err == nil {
		log.Debug("RefreshToken: token refreshed successfully via refresh_token grant",
			"elapsed_ms", time.Since(startTime).Milliseconds(),
			"new_expiry", newToken.Token.Expiry,
		)
		return newToken, nil
	}

	log.Debug("RefreshToken: refresh_token grant failed, falling back to authentication",
		"error", err,
	)
	log.Debug("RefreshToken: initiating interactive authentication")
	token, err := c.authenticator.Authenticate(ctx, c)
	if err != nil {
		log.Error("RefreshToken: authenticating",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	log.Debug("RefreshToken: interactive authentication succeeded",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"new_expiry", token.Token.Expiry,
	)
	return token, nil
}

func (c *OIDCClient) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	log := slog.Default()

	if token == nil {
		log.Debug("refreshToken: nil token provided, skipping refresh flow")
		return nil, fmt.Errorf("cannot refresh nil token")
	}

	if token.Token.RefreshToken == "" {
		log.Debug("refreshToken: no refresh_token available, skipping refresh flow")
		return nil, fmt.Errorf("no refresh token available")
	}

	log.Debug("refreshToken: refresh token found, attempting refresh flow",
		"current_expiry", token.Token.Expiry,
	)

	log.Debug("refreshToken: requesting new token from token endpoint")
	newOauth2Token, err := c.TokenSource(ctx, token.Token).Token()
	if err != nil {
		log.Error("refreshToken: refreshing token",
			"error", err,
		)
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	log.Debug("refreshToken: received new oauth2 token",
		"new_expiry", newOauth2Token.Expiry,
		"has_new_refresh_token", newOauth2Token.RefreshToken != "",
	)

	log.Debug("refreshToken: parsing new token as ID token")
	claims, _, verifiedIDToken, err := c.ParseAsIDToken(ctx, newOauth2Token)
	if err != nil {
		log.Error("refreshToken: parsing new token",
			"error", err,
		)
		return nil, err
	}

	log.Debug("refreshToken: token refresh completed successfully",
		"new_expiry", newOauth2Token.Expiry,
	)
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
