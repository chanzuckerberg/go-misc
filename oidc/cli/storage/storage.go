package storage

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/osutil"
	"github.com/mitchellh/go-homedir"
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
	isDesktop := osutil.IsDesktopEnvironment()

	// If WSL we use a file storage which does not cache refreshTokens
	//    we do this because WSL doesn't have a graphical interface
	//    and therefore limits how we can interact with a keyring (such as gnome-keyring).
	// To limit the risks of having a long-lived refresh token around,
	//    we disable this part of the flow for WSL. This could change in the future
	//    when we find a better way to work with a WSL secure storage.
	if isWSL || !isDesktop {
		return getFileStorage(clientID, issuerURL)
	}

	return getFileStorage(clientID, issuerURL)
}

func getFileStorage(clientID string, issuerURL string) (Storage, error) {
	dir, err := homedir.Expand(defaultFileStorageDir)
	if err != nil {
		return nil, fmt.Errorf("could not expand path: %w", err)
	}

	return NewFile(dir, clientID, issuerURL), nil
}
