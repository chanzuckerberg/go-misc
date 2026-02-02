package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
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
	log := slog.Default()
	startTime := time.Now()

	log.Debug("GetToken started",
		"client_id", clientID,
		"issuer_url", issuerURL,
		"num_client_options", len(clientOptions),
	)

	lockFilePath, err := DefaultLockFilePath()
	if err != nil {
		log.Error("GetToken: getting default lock file path",
			"error", err,
		)
		return nil, fmt.Errorf("getting lock file path: %w", err)
	}
	log.Debug("Creating pid lock",
		"lock_file_path", lockFilePath,
	)
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		log.Error("GetToken: creating pid lock",
			"error", err,
			"lock_file_path", lockFilePath,
		)
		return nil, fmt.Errorf("creating lock: %w", err)
	}
	log.Debug("Pid lock created successfully")

	log.Debug("Creating OIDC client",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)
	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, clientOptions...)
	if err != nil {
		log.Error("GetToken: creating OIDC client",
			"error", err,
			"client_id", clientID,
			"issuer_url", issuerURL,
		)
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}
	log.Debug("OIDC client created successfully")

	log.Debug("Getting storage backend",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)
	storageBackend, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		log.Error("GetToken: getting storage backend",
			"error", err,
			"client_id", clientID,
			"issuer_url", issuerURL,
		)
		return nil, err
	}
	log.Debug("Storage backend obtained successfully",
		"storage_type", fmt.Sprintf("%T", storageBackend),
	)

	log.Debug("Creating cache with storage and refresh function")
	tokenCache := cache.NewCache(storageBackend, oidcClient.RefreshToken, fileLock)
	log.Debug("Cache created successfully")

	log.Debug("Reading token from cache")
	token, err := tokenCache.Read(ctx)
	if err != nil {
		log.Error("GetToken: reading token from cache",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		log.Error("GetToken: received nil token from OIDC-IDP",
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	log.Debug("GetToken completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
		"has_refresh_token", token.Token.RefreshToken != "",
	)
	return token, nil
}

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(
	ctx context.Context,
	clientID string,
	issuerURL string,
	clientOptions ...client.OIDCClientOption,
) (*client.Token, error) {
	log := slog.Default()
	startTime := time.Now()

	log.Debug("GetToken started",
		"client_id", clientID,
		"issuer_url", issuerURL,
		"num_client_options", len(clientOptions),
	)

	lockFilePath, err := DefaultLockFilePath()
	if err != nil {
		log.Error("GetToken: getting default lock file path",
			"error", err,
		)
		return nil, fmt.Errorf("getting lock file path: %w", err)
	}
	log.Debug("Creating pid lock",
		"lock_file_path", lockFilePath,
	)
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		log.Error("GetToken: creating pid lock",
			"error", err,
			"lock_file_path", lockFilePath,
		)
		return nil, fmt.Errorf("creating lock: %w", err)
	}
	log.Debug("Pid lock created successfully")

	log.Debug("Creating OIDC client",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)
	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, clientOptions...)
	if err != nil {
		log.Error("GetToken: creating OIDC client",
			"error", err,
			"client_id", clientID,
			"issuer_url", issuerURL,
		)
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}
	log.Debug("OIDC client created successfully")

	log.Debug("Getting storage backend",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)
	storageBackend, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		log.Error("GetToken: getting storage backend",
			"error", err,
			"client_id", clientID,
			"issuer_url", issuerURL,
		)
		return nil, err
	}
	log.Debug("Storage backend obtained successfully",
		"storage_type", fmt.Sprintf("%T", storageBackend),
	)

	log.Debug("Creating cache with storage and refresh function")
	tokenCache := cache.NewCache(storageBackend, oidcClient.RefreshToken, fileLock)
	log.Debug("Cache created successfully")

	log.Debug("Refreshing token from cache")
	token, err := tokenCache.Refresh(ctx)
	if err != nil {
		log.Error("GetToken: refreshing token",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		log.Error("GetToken: received nil token from OIDC-IDP",
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	log.Debug("GetToken completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
		"has_refresh_token", token.Token.RefreshToken != "",
	)
	return token, nil
}
