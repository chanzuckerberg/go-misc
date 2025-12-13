package cli

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/storage"
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
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}
	oidcClient, err := client.NewOIDCClient(ctx, clientID, issuerURL, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating oidc client: %w", err)
	}

	storage, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		return nil, err
	}

	cache := cache.NewCache(storage, oidcClient.RefreshToken, fileLock)

	token, err := cache.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}
	return token, nil
}
