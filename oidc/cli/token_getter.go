package cli

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
)

const (
	lockFilePath = "/tmp/aws-oidc.lock"
)

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

	log.Info("GetToken started",
		"client_id", clientID,
		"issuer_url", issuerURL,
		"num_client_options", len(clientOptions),
	)

	log.Debug("Creating pid lock",
		"lock_file_path", lockFilePath,
	)
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		log.Error("Failed to create pid lock",
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
		log.Error("Failed to create OIDC client",
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
		log.Error("Failed to get storage backend",
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
		log.Error("Failed to read token from cache",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		log.Error("Received nil token from OIDC-IDP",
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}

	log.Info("GetToken completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
		"has_refresh_token", token.Token.RefreshToken != "",
	)
	return token, nil
}
