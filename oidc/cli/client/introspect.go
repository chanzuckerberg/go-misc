package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LookupRefreshExpiry discovers the issuer's introspection endpoint and
// returns the expiry time for the given refresh token.
func LookupRefreshExpiry(ctx context.Context, clientID, issuerURL, refreshToken string) (time.Time, error) {
	introspectURL, err := discoverIntrospectionEndpoint(ctx, issuerURL)
	if err != nil {
		return time.Time{}, fmt.Errorf("discovering introspection endpoint: %w", err)
	}
	return introspectTokenExpiry(ctx, introspectURL, clientID, refreshToken)
}

// discoverIntrospectionEndpoint fetches the OIDC discovery document and
// returns the introspection_endpoint URL.
func discoverIntrospectionEndpoint(ctx context.Context, issuerURL string) (string, error) {
	wellKnown := strings.TrimSuffix(issuerURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return "", fmt.Errorf("creating discovery request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discovery endpoint returned %s", resp.Status)
	}

	var doc struct {
		IntrospectionEndpoint string `json:"introspection_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decoding discovery document: %w", err)
	}

	if doc.IntrospectionEndpoint == "" {
		return "", fmt.Errorf("no introspection_endpoint in discovery document")
	}

	return doc.IntrospectionEndpoint, nil
}

// introspectTokenExpiry calls the OAuth 2.0 introspection endpoint (RFC 7662)
// and returns the token's expiry time.
func introspectTokenExpiry(ctx context.Context, introspectURL, clientID, token string) (time.Time, error) {
	form := url.Values{
		"token":           {token},
		"token_type_hint": {"refresh_token"},
		"client_id":       {clientID},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(form.Encode()))
	if err != nil {
		return time.Time{}, fmt.Errorf("creating introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("calling introspection endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("introspection endpoint returned %s", resp.Status)
	}

	var result struct {
		Active bool  `json:"active"`
		Exp    int64 `json:"exp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, fmt.Errorf("decoding introspection response: %w", err)
	}

	if !result.Active {
		return time.Time{}, fmt.Errorf("refresh token is no longer active")
	}

	if result.Exp == 0 {
		return time.Time{}, fmt.Errorf("no exp in introspection response")
	}

	return time.Unix(result.Exp, 0), nil
}
