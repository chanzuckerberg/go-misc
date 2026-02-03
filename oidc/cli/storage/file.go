package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
)

type File struct {
	key string
	dir string
	log *slog.Logger

	mu sync.Mutex
}

func NewFile(ctx context.Context, dir string, clientID string, issuerURL string) *File {
	key := generateKey(dir, clientID, issuerURL)
	log := logging.FromContext(ctx)
	log.Debug("File storage initialized",
		"cache_dir", dir,
		"cache_file", key,
		"client_id", clientID,
	)
	return &File{
		dir: dir,
		key: key,
		log: log,
	}
}

func generateKey(dir string, clientID string, issuerURL string) string {
	k := fmt.Sprintf("%s %s %s", storageVersion, clientID, issuerURL)
	h := sha256.Sum256([]byte(k))
	return path.Join(dir, hex.EncodeToString(h[:]))
}

func (f *File) Read(_ context.Context) (*string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	contents, err := os.ReadFile(f.key)
	if os.IsNotExist(err) {
		f.log.Debug("File.Read: cache file does not exist", "path", f.key)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	f.log.Debug("File.Read: loaded from cache file", "path", f.key, "size_bytes", len(contents))
	stringContents := string(contents)
	return &stringContents, nil
}

func (f *File) Set(_ context.Context, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := os.MkdirAll(f.dir, 0700)
	if err != nil {
		return fmt.Errorf("could not create cache dir %s: %w", f.dir, err)
	}

	err = os.WriteFile(f.key, []byte(value), 0600)
	if err != nil {
		return fmt.Errorf("could not set value to file: %w", err)
	}

	f.log.Debug("File.Set: saved to cache file", "path", f.key, "size_bytes", len(value))
	return nil
}

func (f *File) Delete(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := os.Stat(f.key)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(f.key)
	if err != nil {
		return fmt.Errorf("could not delete from file: %w", err)
	}

	f.log.Debug("File.Delete: removed cache file", "path", f.key)
	return nil
}

func (f *File) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
