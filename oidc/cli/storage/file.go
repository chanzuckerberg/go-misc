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
	"github.com/pkg/errors"
)

type File struct {
	key string
	dir string

	mu sync.Mutex
}

func NewFile(dir string, clientID string, issuerURL string) *File {
	return &File{
		dir: dir,
		key: generateKey(dir, clientID, issuerURL),
	}
}

func generateKey(dir string, clientID string, issuerURL string) string {
	k := fmt.Sprintf("%s %s %s", storageVersion, clientID, issuerURL)
	h := sha256.Sum256([]byte(k))

	return path.Join(dir, hex.EncodeToString(h[:]))
}

func (f *File) Read(ctx context.Context) (*string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	contents, err := os.ReadFile(f.key)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not read from file")
	}

	stringContents := string(contents)
	return &stringContents, nil
}

func (f *File) Set(ctx context.Context, value string) error {
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
	return nil
}

func (f *File) Delete(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// check if the key exists first
	_, err := os.Stat(f.key)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(f.key)
	if err != nil {
		return fmt.Errorf("could not delete from file: %w", err)
	}
	return nil
}

func (f *File) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
