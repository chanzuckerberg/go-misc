package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/osutil"
)

const (
	service        = "aws-oidc"
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
}

func GetOIDC(clientID string, issuerURL string) (Storage, error) {
	log := slog.Default()
	log.Debug("GetOIDC: determining storage backend",
		"client_id", clientID,
		"issuer_url", issuerURL,
	)

	isWSL, err := osutil.IsWSL()
	if err != nil {
		log.Error("GetOIDC: checking WSL status",
			"error", err,
		)
		return nil, err
	}
	isDesktop := osutil.IsDesktopEnvironment()

	log.Debug("GetOIDC: environment detection complete",
		"is_wsl", isWSL,
		"is_desktop", isDesktop,
	)

	// If WSL we use a file storage which does not cache refreshTokens
	//    we do this because WSL doesn't have a graphical interface
	//    and therefore limits how we can interact with a keyring (such as gnome-keyring).
	// To limit the risks of having a long-lived refresh token around,
	//    we disable this part of the flow for WSL. This could change in the future
	//    when we find a better way to work with a WSL secure storage.
	if isWSL || !isDesktop {
		log.Debug("GetOIDC: using file storage (WSL or non-desktop environment)")
		return getFileStorage(clientID, issuerURL)
	}

	log.Debug("GetOIDC: using key ring storage (desktop environment)")
	return NewKeyring(clientID, issuerURL), nil
}

func getFileStorage(clientID string, issuerURL string) (Storage, error) {
	log := slog.Default()

	dir, err := getDefaultStorageDir()
	if err != nil {
		log.Error("getFileStorage: expanding storage directory path",
			"error", err,
		)
		return nil, err
	}

	log.Debug("getFileStorage: creating file storage",
		"directory", dir,
		"client_id", clientID,
		"issuer_url", issuerURL,
	)
	return NewFile(dir, clientID, issuerURL), nil
}
