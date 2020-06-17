package oidc

import (
	"context"
	"os"
	"time"

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
func GetToken(ctx context.Context, clientID string, issuerURL string) (*client.Token, error) {
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create lock")
	}

	f, err := os.OpenFile("/tmp/oidc-lock", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.WriteString("text to append\n"); err != nil {
		return nil, err
	}

	start := time.Now()

	err = fileLock.Lock()
	if err != nil {
		return nil, errors.Wrap(err, "unable to lock")
	}
	defer func() {
		fileLock.Unlock()                         //nolint: errcheck
		f.WriteString(time.Since(start).String()) //nolint: errcheck
	}()

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
