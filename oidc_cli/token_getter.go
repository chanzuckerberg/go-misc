package oidc

import (
	"context"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc_cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
	"github.com/chanzuckerberg/go-misc/osutil"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	lockFilePath          = "/tmp/aws-oidc.lock"
	defaultFileStorageDir = "~/.oidc-cli"
)

// GetToken gets an oidc token.
// It handles caching with a default cache and keyring storage.
func GetToken(ctx context.Context, clientID string, issuerURL string) (*client.Token, error) {
	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create lock")
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

	c, err := client.NewClient(ctx, conf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create client")
	}

	storage, err := getStorage(clientID, issuerURL)
	if err != nil {
		return nil, err
	}

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

func getStorage(clientID string, issuerURL string) (storage.Storage, error) {
	isWSL, err := osutil.IsWSL()
	if err != nil {
		return nil, err
	}

	// If WSL we use a file storage which does not cache refreshTokens
	//    we do this because WSL doesn't have a graphical interface
	//    and therefore limits how we can interact with a keyring (such as gnome-keyring).
	// To limit the risks of having a long-lived refresh token around,
	//    we disable this part of the flow for WSL. This could change in the future
	//    when we find a better way to work with a WSL secure storage.
	if isWSL {
		return getFileStorage(clientID, issuerURL)
	}

	return storage.NewKeyring(clientID, issuerURL), nil
}

func getFileStorage(clientID string, issuerURL string) (storage.Storage, error) {
	dir, err := homedir.Expand(defaultFileStorageDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not expand path")
	}

	return storage.NewFile(dir, clientID, issuerURL), nil
}
