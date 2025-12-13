package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"golang.org/x/oauth2"
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
	clientOptions ...client.Option,
) (*client.Token, error) {
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}

	conf := &client.Config{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		ServerConfig: &client.ServerConfig{
			// TODO (el): Make these configurable?
			FromPort: 49152,
			ToPort:   49152 + 63,
			Timeout:  30 * time.Second,
		},
	}

	c, err := client.NewAuthorizationGrantClient(ctx, conf, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	storage, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		return nil, err
	}

	cache := cache.NewCache(storage, c.RefreshToken, fileLock)

	oauth2.ReuseTokenSource()
	token, err := cache.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}
	return token, nil
}

func GetDeviceGrantToken(
	ctx context.Context,
	clientID string,
	issuerURL string,
	scopes []string,
) (*client.Token, error) {
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("creating lock: %w", err)
	}

	conf := &client.DeviceGrantConfig{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		Scopes:    scopes,
	}

	c, err := client.NewDeviceGrantClient(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	storage, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		return nil, err
	}

	cache := cache.NewCache(storage, c.RefreshToken, fileLock)

	token, err := cache.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("extracting token from client: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("nil token from OIDC-IDP")
	}
	return token, nil
}
