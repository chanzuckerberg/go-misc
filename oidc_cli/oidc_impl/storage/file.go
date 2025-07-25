package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/chanzuckerberg/go-misc/oidc_cli/v3/oidc_impl/client"
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

	contents, err := ioutil.ReadFile(f.key)
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
		return errors.Wrapf(err, "could not create cache dir %s", f.dir)
	}

	err = ioutil.WriteFile(f.key, []byte(value), 0600)
	return errors.Wrap(err, "could not set value to file")
}

func (f *File) Delete(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := os.Remove(f.key)
	return errors.Wrap(err, "could not delete from file")
}

// Note: We disable the refresh flow because we are storing
//
//	      credentials in plaintext.
//				 We therefore ensure that any plaintext credentials that hit disk
//	      have a well-defined ttl.
func (f *File) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{client.MarshalOptNoRefresh}
}
