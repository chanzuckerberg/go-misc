package oidc

import (
	"context"
	"time"

	"github.com/chanzuckerberg/ago-misc/pkg/oidc_cli/util"
	client "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/chanzuckerberg/go-misc/pkg/oidc_cli/cache"
	"github.com/chanzuckerberg/go-misc/pkg/oidc_cli/storage"
	"github.com/pkg/errors"
)

const (
	lockFilePath = "/tmp/aws-oidc.lock"
)

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(ctx context.Context, clientID string, issuerURL string) (*client.Token, error) {
	fileLock, err := util.NewLock(lockFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create lock")
	}

	err = fileLock.Lock()
	if err != nil {
		return nil, errors.Wrap(err, "unable to lock")
	}
	defer fileLock.Unlock() //nolint: errcheck

	conf := &client.Config{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		ServerConfig: &client.ServerConfig{
			// TODO (el): Make these configurable
			FromPort: 49152,
			ToPort:   49152 + 63,
			Timeout:  30 * time.Second,
		},
	}

	c, err := client.NewClient(ctx, conf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create client")
	}

	storage := storage.NewKeyring(clientID, issuerURL)
	cache := cache.NewCache(storage, c.RefreshToken)

	token, err := cache.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to extract token from client")
	}
	if token == nil {
		return nil, errors.New("nil token from OIDC-IDP")
	}
	return token, nil
}
