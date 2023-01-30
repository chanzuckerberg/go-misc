package storage

import (
	"context"

	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/client"
	"github.com/chanzuckerberg/go-misc/osutil"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	service               = "aws-oidc"
	defaultFileStorageDir = "~/.cache/oidc-cli"
	storageVersion        = "v0"
)

// Storage represents a storage backend for a cache
type Storage interface {
	Read(context.Context) (*string, error)
	Set(ctx context.Context, value string) error
	Delete(context.Context) error

	MarshalOpts() []client.MarshalOpts
}

func GetOIDC(clientID string, issuerURL string) (Storage, error) {
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

	return NewKeyring(clientID, issuerURL), nil
}

func getFileStorage(clientID string, issuerURL string) (Storage, error) {
	dir, err := homedir.Expand(defaultFileStorageDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not expand path")
	}

	return NewFile(dir, clientID, issuerURL), nil
}
