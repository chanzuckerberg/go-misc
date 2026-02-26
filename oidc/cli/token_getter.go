package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
)

type getTokenConfig struct {
	localCacheDir string
	fileOptions   []storage.FileOption
	clientOptions []client.OIDCClientOption
}

// GetTokenOption configures GetToken behavior.
type GetTokenOption func(*getTokenConfig)

// WithLocalCacheDir stores a per-hostname cache file on node-local disk,
// bootstrapped from the default (e.g. NFS) cache on first access.
// This avoids cross-host NFS lock contention.
func WithLocalCacheDir(dir string) GetTokenOption {
	return func(c *getTokenConfig) {
		c.localCacheDir = dir
		c.fileOptions = append(c.fileOptions, storage.WithLocalCacheDir(dir))
	}
}

// WithClientOptions appends OIDCClientOptions to the underlying OIDC client.
func WithClientOptions(opts ...client.OIDCClientOption) GetTokenOption {
	return func(c *getTokenConfig) {
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
	var cfg getTokenConfig
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

	if cfg.localCacheDir != "" {
		rootStorage, rootErr := storage.GetOIDC(ctx, clientID, issuerURL)
		if rootErr != nil {
			logger.Warn("GetToken: failed to get root storage backend, skipping sync with root", "error", rootErr)
		} else {
			trySyncFromRootIfNewer(ctx, rootStorage, storageBackend)
		}
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
	var cfg getTokenConfig
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

	if cachedToken.AccessToken == "" {
		return fmt.Errorf("no valid token in cache")
	}
	if !cachedToken.Valid() {
		return fmt.Errorf("cached access token expired %s ago", time.Since(cachedToken.Expiry).Truncate(time.Second))
	}

	return nil
}

// CheckRefreshTokenTTL returns the duration until the cached refresh token
// expires using the RefreshTokenExpiry stored in the serialized token.
// Returns an error if no refresh token expiry is stored in the cache.
func CheckRefreshTokenTTL(
	ctx context.Context,
	clientID string,
	issuerURL string,
	opts ...GetTokenOption,
) (time.Duration, error) {
	var cfg getTokenConfig
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

	if cachedToken.RefreshTokenExpiry.IsZero() {
		return 0, fmt.Errorf("no refresh token expiry stored in cache")
	}

	ttl := time.Until(cachedToken.RefreshTokenExpiry)
	if ttl < 0 {
		return 0, nil
	}

	logger.Debug("CheckRefreshTokenTTL: computed TTL",
		"refresh_token_expiry", cachedToken.RefreshTokenExpiry,
		"ttl", ttl,
	)
	return ttl, nil
}

// trySyncFromRootIfNewer compares the refresh token expiry stored in the root
// (NFS) cache with the local node cache. If root's refresh token expiry is
// later, the root's raw data is copied into local storage so the subsequent
// cache read uses the fresher refresh token.
func trySyncFromRootIfNewer(ctx context.Context, rootStorage, localStorage storage.Storage) {
	logger := logging.FromContext(ctx)

	rootCache := cache.NewCache(ctx, rootStorage, nil, nil)
	rootToken, err := rootCache.DecodeFromStorage(ctx)
	if err != nil || rootToken.RefreshTokenExpiry.IsZero() {
		return
	}

	localCache := cache.NewCache(ctx, localStorage, nil, nil)
	localToken, err := localCache.DecodeFromStorage(ctx)
	if err != nil {
		return
	}

	if !rootToken.RefreshTokenExpiry.After(localToken.RefreshTokenExpiry) {
		return
	}

	logger.Debug("syncFromRootIfNewer: root has newer refresh token expiry, syncing to local",
		"root_expiry", rootToken.RefreshTokenExpiry,
		"local_expiry", localToken.RefreshTokenExpiry,
	)

	rootRaw, err := rootStorage.Read(ctx)
	if err != nil {
		logger.Warn("trySyncFromRootIfNewer: failed to read root storage", "error", err)
		return
	}
	if rootRaw == nil {
		logger.Warn("trySyncFromRootIfNewer: root storage is empty, skipping sync")
		return
	}

	err = localStorage.Set(ctx, *rootRaw)
	if err != nil {
		logger.Warn("trySyncFromRootIfNewer: failed to update local storage from root", "error", err)
	}
}
