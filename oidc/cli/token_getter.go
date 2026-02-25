package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
)

var (
	// ErrTokenNotFound indicates no usable token was found in the cache.
	ErrTokenNotFound = errors.New("no valid token in cache")

	// ErrTokenExpired indicates the cached access token has expired.
	ErrTokenExpired = errors.New("cached access token is expired")
)

type _getTokenConfig struct {
	localCacheDir string
	fileOptions   []storage.FileOption
	clientOptions []client.OIDCClientOption
}

// GetTokenOption configures GetToken behavior.
type GetTokenOption func(*_getTokenConfig)

// WithLocalCacheDir stores a per-hostname cache file on node-local disk,
// bootstrapped from the default (e.g. NFS) cache on first access.
// This avoids cross-host NFS lock contention.
func WithLocalCacheDir(dir string) GetTokenOption {
	return func(c *_getTokenConfig) {
		c.localCacheDir = dir
		c.fileOptions = append(c.fileOptions, storage.WithLocalCacheDir(dir))
	}
}

// WithClientOption appends an OIDCClientOption to the underlying OIDC client.
func WithClientOptions(opts ...client.OIDCClientOption) GetTokenOption {
	return func(c *_getTokenConfig) {
		c.clientOptions = append(c.clientOptions, opts...)
	}
}

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(
	ctx context.Context,
	clientID string,
	issuerURL string,
	opts ...GetTokenOption,
) (*client.Token, error) {
	var cfg _getTokenConfig
	for _, o := range opts {
		o(&cfg)
	}

	ctx, logger := logging.NewLogger(ctx)
	startTime := time.Now()

	logger.Debug("GetToken: started",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, cfg.clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}

	storageBackend, err := storage.GetOIDC(ctx, clientID, issuerURL, cfg.fileOptions...)
	if err != nil {
		return nil, fmt.Errorf("getting storage backend: %w", err)
	}

	lockPath, err := lockFilePath(clientID, issuerURL, cfg.localCacheDir)
	if err != nil {
		return nil, fmt.Errorf("getting lock file path: %w", err)
	}
	fileLock, err := pidlock.NewLock(lockPath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}
	logger.Debug("GetToken: created refresh lock",
		"lock_path", lockPath,
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	tokenCache := cache.NewCache(ctx, storageBackend, oidcClient.RefreshToken, fileLock)

	token, err := tokenCache.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	logger.Debug("GetToken: completed",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
	)
	return token, nil
}

// lockFilePath returns a deterministic lock file path derived from clientID
// and issuerURL. When localCacheDir is set the lock lives there (local disk);
// otherwise it falls back to the default storage dir (~/.cache/oidc-cli).
func lockFilePath(clientID, issuerURL, localCacheDir string) (string, error) {
	dir := localCacheDir
	if dir == "" {
		d, err := storage.DefaultStorageDir()
		if err != nil {
			return "", fmt.Errorf("determining lock dir: %w", err)
		}
		dir = d
	}

	return fmt.Sprintf("%s.lock", storage.GenerateKey(dir, clientID, issuerURL)), nil
}

// CheckTokenIsValid reads the cached OIDC token and returns nil if it is present
// and valid. Returns ErrTokenNotFound or ErrTokenExpired otherwise.
// It never triggers a refresh flow.
func CheckTokenIsValid(
	ctx context.Context,
	clientID string,
	issuerURL string,
	opts ...GetTokenOption,
) error {
	var cfg _getTokenConfig
	for _, o := range opts {
		o(&cfg)
	}

	ctx, logger := logging.NewLogger(ctx)
	logger.Debug("CheckTokenIsValid: started",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	storageBackend, err := storage.GetOIDC(ctx, clientID, issuerURL, cfg.fileOptions...)
	if err != nil {
		return fmt.Errorf("getting storage backend: %w", err)
	}

	tokenCache := cache.NewCache(ctx, storageBackend, nil, nil)
	cachedToken, err := tokenCache.DecodeFromStorage(ctx)
	if err != nil {
		return fmt.Errorf("decoding cached token: %w", err)
	}

	if !cachedToken.Valid() {
		return fmt.Errorf("cached token is invalid")
	}

	return nil
}

// CheckRefreshTokenTTL returns the duration until the cached refresh token
// expires, determined by calling the issuer's introspection endpoint.
// Returns an error if no refresh token is cached or introspection fails.
func CheckRefreshTokenTTL(
	ctx context.Context,
	clientID string,
	issuerURL string,
	opts ...GetTokenOption,
) (time.Duration, error) {
	var cfg _getTokenConfig
	for _, o := range opts {
		o(&cfg)
	}

	ctx, logger := logging.NewLogger(ctx)
	logger.Debug("CheckRefreshTokenTTL: started",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	storageBackend, err := storage.GetOIDC(ctx, clientID, issuerURL, cfg.fileOptions...)
	if err != nil {
		return 0, fmt.Errorf("getting storage backend: %w", err)
	}

	tokenCache := cache.NewCache(ctx, storageBackend, nil, nil)
	cachedToken, err := tokenCache.DecodeFromStorage(ctx)
	if err != nil {
		return 0, fmt.Errorf("decoding cached token: %w", err)
	}

	if !cachedToken.Valid() {
		return 0, fmt.Errorf("cached token is invalid")
	}

	if cachedToken.RefreshToken == "" {
		return 0, fmt.Errorf("no refresh token in cache")
	}

	introspectURL, err := discoverIntrospectionEndpoint(ctx, issuerURL)
	if err != nil {
		return 0, fmt.Errorf("discovering introspection endpoint: %w", err)
	}

	expiry, err := introspectTokenExpiry(ctx, introspectURL, clientID, cachedToken.RefreshToken)
	if err != nil {
		return 0, fmt.Errorf("introspecting token expiry: %w", err)
	}

	ttl := time.Until(expiry)
	if ttl < 0 {
		return 0, nil
	}

	logger.Debug("CheckRefreshTokenTTL: computed TTL",
		"refresh_token_expiry", expiry,
		"ttl", ttl,
	)
	return ttl, nil
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
