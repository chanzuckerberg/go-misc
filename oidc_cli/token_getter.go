package oidc

import (
	"context"

	"github.com/chanzuckerberg/go-misc/oidc_cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/pkg/errors"
)

const (
	lockFilePath = "/tmp/aws-oidc.lock"
)

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(ctx context.Context, clientID string, issuerURL string, serverConfig *client.ServerConfig) (*client.Token, error) {
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create lock")
	}

	conf := &client.Config{
		ClientID:     clientID,
		IssuerURL:    issuerURL,
		ServerConfig: serverConfig,
	}

	c, err := client.NewClient(ctx, conf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create client")
	}

	storage := storage.NewKeyring(clientID, issuerURL)
	cache := cache.NewCache(storage, c.RefreshToken, fileLock)

	token, err := cache.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to extract token from client")
	}
	if token == nil {
		return nil, errors.New("nil token from OIDC-IDP")
	}
	return token, nil
}
