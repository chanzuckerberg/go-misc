package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
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

	// Endpoints
	deviceAuthEndpoint string
	tokenEndpoint      string
}

// deviceAuthResponse represents the response from the device authorization endpoint
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
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

// tokenErrorResponse represents an error response from the token endpoint
type tokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
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

	// Fallback to constructing the device auth endpoint if not in discovery
	deviceAuthEndpoint := providerClaims.DeviceAuthEndpoint
	if deviceAuthEndpoint == "" {
		// Okta typically uses this pattern for the device authorization endpoint
		deviceAuthEndpoint = strings.TrimSuffix(config.IssuerURL, "/") + "/v1/device/authorize"
	}

	tokenEndpoint := providerClaims.TokenEndpoint
	if tokenEndpoint == "" {
		tokenEndpoint = provider.Endpoint().TokenURL
	}

	oidcConfig := &oidc.Config{
		ClientID:             config.ClientID,
		SupportedSigningAlgs: []string{"RS256"},
	}
	verifier := provider.Verifier(oidcConfig)

	scopes := config.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email", "groups"}
	}

	return &DeviceGrantClient{
		provider:           provider,
		verifier:           verifier,
		httpClient:         http.DefaultClient,
		clientID:           config.ClientID,
		issuerURL:          config.IssuerURL,
		scopes:             scopes,
		deviceAuthEndpoint: deviceAuthEndpoint,
		tokenEndpoint:      tokenEndpoint,
	}, nil
}

// Authenticate initiates the device authorization flow and waits for user authentication
func (c *DeviceGrantClient) Authenticate(ctx context.Context) (*Token, error) {
	// Step 1: Request device code
	deviceAuth, err := c.requestDeviceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}

	// Step 2: Display the code to the user
	c.displayUserCode(deviceAuth)

	// Step 3: Poll for token
	token, err := c.pollForToken(ctx, deviceAuth)
	if err != nil {
		return nil, fmt.Errorf("polling for token: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Successfully authenticated!\n")
	return token, nil
}

// requestDeviceCode requests a device code from the authorization server
func (c *DeviceGrantClient) requestDeviceCode(ctx context.Context) (*deviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("scope", strings.Join(c.scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.deviceAuthEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating device authorization request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making device authorization request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp deviceAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &authResp, nil
}

// displayUserCode displays the user code and verification URL to the user
func (c *DeviceGrantClient) displayUserCode(deviceAuth *deviceAuthResponse) {
	data := &deviceAuthTemplateData{
		VerificationURI:  deviceAuth.VerificationURI,
		UserCode:         deviceAuth.UserCode,
		ExpiresInMinutes: deviceAuth.ExpiresIn / 60,
	}
	_ = renderDeviceAuthTemplate(os.Stderr, data)
}

// pollForToken polls the token endpoint until the user completes authentication
func (c *DeviceGrantClient) pollForToken(ctx context.Context, deviceAuth *deviceAuthResponse) (*Token, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second // Default polling interval
	}

	expiresAt := time.Now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(expiresAt) {
				return nil, fmt.Errorf("device code expired")
			}

			token, err := c.requestToken(ctx, deviceAuth.DeviceCode)
			if err != nil {
				var pollErr *pollError
				if errors.As(err, &pollErr) {
					switch pollErr.errorCode {
					case "authorization_pending":
						// User hasn't completed auth yet, keep polling
						fmt.Fprintf(os.Stderr, ".")
						continue
					case "slow_down":
						// Increase polling interval
						interval += 5 * time.Second
						ticker.Reset(interval)
						logrus.Debug("slowing down polling interval")
						continue
					case "expired_token":
						return nil, fmt.Errorf("device code expired")
					case "access_denied":
						return nil, fmt.Errorf("access denied by user")
					}
				}
				return nil, err
			}

			return token, nil
		}
	}
}

// pollError represents an expected error during polling
type pollError struct {
	errorCode   string
	description string
}

func (e *pollError) Error() string {
	return fmt.Sprintf("%s: %s", e.errorCode, e.description)
}

// requestToken requests a token using the device code
func (c *DeviceGrantClient) requestToken(ctx context.Context, deviceCode string) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Check for error responses (these come with 400 status for pending auth)
	if resp.StatusCode == http.StatusBadRequest {
		var errResp tokenErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, &pollError{
				errorCode:   errResp.Error,
				description: errResp.ErrorDescription,
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
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

	return &Token{
		Version:      tokenVersion,
		Expiry:       idToken.Expiry,
		IDToken:      tokenResp.IDToken,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Claims:       *claims,
	}, nil
}

// RefreshToken will fetch a new token using the refresh token
func (c *DeviceGrantClient) RefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	if oldToken == nil {
		logrus.Debug("nil refresh token, skipping refresh flow")
		return nil, fmt.Errorf("cannot refresh nil token")
	}
	if oldToken.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	logrus.Debug("refresh token found, attempting refresh flow")

	newToken, err := c.doRefreshToken(ctx, oldToken)
	if err == nil {
		return newToken, nil
	}

	logrus.WithError(err).Debug("failed to refresh token, requesting new one")
	return c.Authenticate(ctx)
}

// doRefreshToken performs the actual token refresh
func (c *DeviceGrantClient) doRefreshToken(ctx context.Context, oldToken *Token) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", oldToken.RefreshToken)
	data.Set("scope", strings.Join(c.scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenEndpoint, strings.NewReader(data.Encode()))
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

	logrus.Debug("refresh successful")

	return &Token{
		Version:      oldToken.Version,
		Expiry:       idToken.Expiry,
		IDToken:      tokenResp.IDToken,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Claims:       *claims,
	}, nil
}
