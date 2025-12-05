package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// DeviceGrantConfig is required to configure a DeviceGrantClient
type DeviceGrantConfig struct {
	ClientID  string
	IssuerURL string
	Scopes    []string
}

// DeviceGrantClient implements the OAuth 2.0 Device Authorization Grant flow
type DeviceGrantClient struct {
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	httpClient *http.Client

	clientID  string
	issuerURL string
	scopes    []string

	oauthConfig *oauth2.Config
}

// tokenResponse represents the response from the token endpoint
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// NewDeviceGrantClient creates a new DeviceGrantClient
func NewDeviceGrantClient(ctx context.Context, config *DeviceGrantConfig) (*DeviceGrantClient, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	// Get provider claims to discover endpoints
	var providerClaims struct {
		DeviceAuthEndpoint string `json:"device_authorization_endpoint"`
		TokenEndpoint      string `json:"token_endpoint"`
	}
	if err := provider.Claims(&providerClaims); err != nil {
		return nil, fmt.Errorf("getting provider claims: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID:             config.ClientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	oauthConfig := &oauth2.Config{
		ClientID: config.ClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   config.Scopes,
	}
	verifier := provider.Verifier(oidcConfig)

	scopes := config.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email", "groups"}
	}

	return &DeviceGrantClient{
		provider:    provider,
		verifier:    verifier,
		httpClient:  http.DefaultClient,
		clientID:    config.ClientID,
		issuerURL:   config.IssuerURL,
		scopes:      scopes,
		oauthConfig: oauthConfig,
	}, nil
}

// Authenticate initiates the device authorization flow and waits for user authentication
func (c *DeviceGrantClient) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	response, err := c.oauthConfig.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}

	err = c.displayUserCode(response)
	if err != nil {
		return nil, err
	}

	token, err := c.oauthConfig.DeviceAccessToken(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("requesting access token: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Successfully authenticated!\n")
	return token, nil
}

// displayUserCode displays the user code and verification URL to the user
func (c *DeviceGrantClient) displayUserCode(deviceAuth *oauth2.DeviceAuthResponse) error {
	data := &deviceAuthTemplateData{
		VerificationURI:  deviceAuth.VerificationURI,
		UserCode:         deviceAuth.UserCode,
		ExpiresInMinutes: int(time.Until(deviceAuth.Expiry).Minutes()),
	}
	err := renderDeviceAuthTemplate(os.Stderr, data)
	if err != nil {
		return fmt.Errorf("rendering device auth template: %w", err)
	}

	return nil
}

func oath2TokenToToken(oauth2Token *oauth2.Token) *Token {
	if oauth2Token == nil {
		return nil
	}

	return &Token{
		AccessToken:  oauth2Token.AccessToken,
		RefreshToken: oauth2Token.RefreshToken,
		Expiry:       oauth2Token.Expiry,
		IDToken:      oauth2Token.Extra("id_token").(string),
	}
}

// RefreshToken will fetch a new token using the refresh token
func (c *DeviceGrantClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	newToken, err := c.doRefreshToken(ctx, oldToken)
	if err == nil {
		return newToken, nil
	}

	slog.Debug("failed to refresh token, requesting new one", "error", err)
	token, err := c.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	return oath2TokenToToken(token), nil
}

// doRefreshToken performs the actual token refresh
func (c *DeviceGrantClient) doRefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	if oldToken == nil {
		slog.Debug("nil refresh token, skipping refresh flow")
		return nil, errors.New("cannot refresh nil token")
	}
	slog.Debug("refresh token found, attempting refresh flow")
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", oldToken.RefreshToken)
	data.Set("scope", strings.Join(c.scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.oauthConfig.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making refresh token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	// Verify and parse the ID token
	idToken, err := c.verifier.Verify(ctx, tokenResp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("verifying ID token: %w", err)
	}

	claims := &Claims{}
	if err := idToken.Claims(claims); err != nil {
		return nil, fmt.Errorf("parsing claims: %w", err)
	}

	slog.Debug("refresh successful")

	return &Token{
		Version:      oldToken.Version,
		Expiry:       idToken.Expiry,
		IDToken:      tokenResp.IDToken,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Claims:       *claims,
	}, nil
}
