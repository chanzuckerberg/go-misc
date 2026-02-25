package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/chanzuckerberg/go-misc/osutil"
)

const (
	storageVersion = "v0"
)

func getDefaultStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "oidc-cli"), nil
}

// Storage represents a storage backend for a cache
type Storage interface {
	Read(context.Context) (*string, error)
	Set(ctx context.Context, value string) error
	Delete(context.Context) error

	MarshalOpts() []client.MarshalOpts

	// ActivePath returns the filesystem path used for reads/writes.
	// Used to derive a co-located lock file path.
	ActivePath() string
}

func GetOIDC(ctx context.Context, clientID string, issuerURL string, fileOpts ...FileOption) (Storage, error) {
	log := logging.FromContext(ctx)

	isWSL, err := osutil.IsWSL()
	if err != nil {
		return nil, err
	}
	isDesktop := osutil.IsDesktopEnvironment()

	log.Debug("GetOIDC: detected environment",
		"is_wsl", isWSL,
		"is_desktop", isDesktop,
		"client_id", clientID,
	)

	// If WSL we use a file storage which does not cache refreshTokens
	//    we do this because WSL doesn't have a graphical interface
	//    and therefore limits how we can interact with a keyring (such as gnome-keyring).
	// To limit the risks of having a long-lived refresh token around,
	//    we disable this part of the flow for WSL. This could change in the future
	//    when we find a better way to work with a WSL secure storage.
	if isWSL || !isDesktop {
		dir, err := getDefaultStorageDir()
		if err != nil {
			return nil, err
		}
		log.Debug("GetOIDC: using file storage backend", "dir", dir)
		return NewFile(ctx, dir, clientID, issuerURL, fileOpts...)
	}

	log.Debug("GetOIDC: using keyring storage backend")
	return NewKeyring(ctx, clientID, issuerURL), nil
}
