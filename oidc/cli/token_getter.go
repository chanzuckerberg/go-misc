package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
)

// DefaultLockFilePath returns the default path for the OIDC lock file.
// The lock file is stored in ~/.oidc_locks/oidc.lock
func DefaultLockFilePath() (string, error) {
	const (
		lockDir  = ".oidc_locks"
		lockFile = "oidc.lock"
	)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}
	return filepath.Join(home, lockDir, lockFile), nil
}

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(
	ctx context.Context,
	clientID string,
	issuerURL string,
	clientOptions ...client.OIDCClientOption,
) (*client.Token, error) {
	ctx, log := logging.WithSessionID(ctx)
	startTime := time.Now()

	log.Debug("GetToken: started",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	lockFilePath, err := DefaultLockFilePath()
	if err != nil {
		return nil, fmt.Errorf("getting lock file path: %w", err)
	}

	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}

	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}

	storageBackend, err := storage.GetOIDC(ctx, clientID, issuerURL)
	if err != nil {
		return nil, err
	}

	tokenCache := cache.NewCache(ctx, storageBackend, oidcClient.RefreshToken, fileLock)

	token, err := tokenCache.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	log.Debug("GetToken: completed",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
	)
	return token, nil
}

// RefreshToken refreshes an oidc token.
// Helpful for debugging refresh token flows. Like GetToken
// but only calls the refresh token flow.
func RefreshToken(
	ctx context.Context,
	clientID string,
	issuerURL string,
	clientOptions ...client.OIDCClientOption,
) (*client.Token, error) {
	ctx, log := logging.WithSessionID(ctx)
	startTime := time.Now()

	log.Debug("RefreshToken: started",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	lockFilePath, err := DefaultLockFilePath()
	if err != nil {
		return nil, fmt.Errorf("getting lock file path: %w", err)
	}

	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}

	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}

	storageBackend, err := storage.GetOIDC(ctx, clientID, issuerURL)
	if err != nil {
		return nil, err
	}

	tokenCache := cache.NewCache(ctx, storageBackend, oidcClient.RefreshToken, fileLock)

	token, err := tokenCache.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	log.Debug("RefreshToken: completed",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
	)
	return token, nil
}
