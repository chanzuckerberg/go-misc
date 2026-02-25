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
