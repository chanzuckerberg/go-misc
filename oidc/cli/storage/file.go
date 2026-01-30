package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/pkg/errors"
)

type File struct {
	key string
	dir string

	mu sync.Mutex
}

func NewFile(dir string, clientID string, issuerURL string) *File {
	log := logging.Get()
	key := generateKey(dir, clientID, issuerURL)
	log.Debug("NewFile: creating file storage",
		"directory", dir,
		"client_id", clientID,
		"issuer_url", issuerURL,
		"key_path", key,
	)
	return &File{
		dir: dir,
		key: key,
	}
}

func generateKey(dir string, clientID string, issuerURL string) string {
	log := logging.Get()
	log.Debug("generateKey: generating storage key",
		"directory", dir,
		"client_id", clientID,
		"issuer_url", issuerURL,
		"storage_version", storageVersion,
	)

	k := fmt.Sprintf("%s %s %s", storageVersion, clientID, issuerURL)
	h := sha256.Sum256([]byte(k))
	key := path.Join(dir, hex.EncodeToString(h[:]))

	log.Debug("generateKey: key generated",
		"key_path", key,
		"hash", hex.EncodeToString(h[:]),
	)
	return key
}

func (f *File) Read(ctx context.Context) (*string, error) {
	log := logging.Get()
	log.Debug("File.Read: acquiring mutex lock",
		"key_path", f.key,
	)
	f.mu.Lock()
	defer f.mu.Unlock()

	log.Debug("File.Read: reading file",
		"key_path", f.key,
	)
	contents, err := os.ReadFile(f.key)
	if os.IsNotExist(err) {
		log.Debug("File.Read: file does not exist (cache miss)",
			"key_path", f.key,
		)
		return nil, nil
	}
	if err != nil {
		log.Error("File.Read: failed to read file",
			"error", err,
			"key_path", f.key,
		)
		return nil, errors.Wrap(err, "could not read from file")
	}

	log.Debug("File.Read: file read successfully",
		"key_path", f.key,
		"content_length", len(contents),
	)
	stringContents := string(contents)
	return &stringContents, nil
}

func (f *File) Set(ctx context.Context, value string) error {
	log := logging.Get()
	log.Debug("File.Set: acquiring mutex lock",
		"key_path", f.key,
	)
	f.mu.Lock()
	defer f.mu.Unlock()

	log.Debug("File.Set: ensuring cache directory exists",
		"directory", f.dir,
	)
	err := os.MkdirAll(f.dir, 0700)
	if err != nil {
		log.Error("File.Set: failed to create cache directory",
			"error", err,
			"directory", f.dir,
		)
		return fmt.Errorf("could not create cache dir %s: %w", f.dir, err)
	}

	log.Debug("File.Set: writing value to file",
		"key_path", f.key,
		"value_length", len(value),
	)
	err = os.WriteFile(f.key, []byte(value), 0600)
	if err != nil {
		log.Error("File.Set: failed to write file",
			"error", err,
			"key_path", f.key,
		)
		return fmt.Errorf("could not set value to file: %w", err)
	}

	log.Debug("File.Set: file written successfully",
		"key_path", f.key,
	)
	return nil
}

func (f *File) Delete(ctx context.Context) error {
	log := logging.Get()
	log.Debug("File.Delete: acquiring mutex lock",
		"key_path", f.key,
	)
	f.mu.Lock()
	defer f.mu.Unlock()

	// check if the key exists first
	log.Debug("File.Delete: checking if file exists",
		"key_path", f.key,
	)
	_, err := os.Stat(f.key)
	if os.IsNotExist(err) {
		log.Debug("File.Delete: file does not exist, nothing to delete",
			"key_path", f.key,
		)
		return nil
	}

	log.Debug("File.Delete: removing file",
		"key_path", f.key,
	)
	err = os.Remove(f.key)
	if err != nil {
		log.Error("File.Delete: failed to remove file",
			"error", err,
			"key_path", f.key,
		)
		return fmt.Errorf("could not delete from file: %w", err)
	}

	log.Debug("File.Delete: file removed successfully",
		"key_path", f.key,
	)
	return nil
}

func (f *File) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
