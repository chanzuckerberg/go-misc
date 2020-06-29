package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"
)

// Keyring implements the storage interface for the cache
// Using zalango/go-keyring because it doesn't rely on CGO
//  at the cost of less flexibility.
//  We can re-evaluate as needed and update this struct
type Keyring struct {
	key string

	mu sync.Mutex
}

// NewKeyring returns a new keyring
func NewKeyring(clientID string, issuerURL string) *Keyring {
	return &Keyring{key: fmt.Sprintf("%s %s %s", storageVersion, issuerURL, clientID)}
}

// Read will read from the keyring
func (k *Keyring) Read(ctx context.Context) (*string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	val, err := keyring.Get(service, k.key)
	// Make this more idiomatic
	if err == keyring.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not read from keyring")
	}
	return &val, nil
}

// Set sets a value to the keyring
func (k *Keyring) Set(ctx context.Context, value string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	err := keyring.Set(service, k.key, value)
	if err == keyring.ErrNotFound {
		return nil
	}
	return errors.Wrap(err, "could not set value to keyring")
}

// Delete will delete a value from the keyring
func (k *Keyring) Delete(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	err := keyring.Delete(service, k.key)
	if err == keyring.ErrNotFound {
		return nil
	}
	return errors.Wrap(err, "could not delete from keyring")
}

func (k *Keyring) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
