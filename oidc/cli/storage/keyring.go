package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/zalando/go-keyring"
)

const (
	service = "aws-oidc"
)

// Keyring implements the storage interface for the cache
// Using zalango/go-keyring because it doesn't rely on CGO
//
//	at the cost of less flexibility.
//	We can re-evaluate as needed and update this struct
type Keyring struct {
	key string
	log *slog.Logger

	mu sync.Mutex
}

// NewKeyring returns a new keyring
func NewKeyring(ctx context.Context, clientID string, issuerURL string) *Keyring {
	key := fmt.Sprintf("%s %s %s", storageVersion, issuerURL, clientID)
	log := logging.FromContext(ctx)
	log.Debug("Keyring storage initialized",
		"service", service,
		"key", key,
		"client_id", clientID,
	)
	return &Keyring{
		key: key,
		log: log,
	}
}

// Read will read from the keyring
func (k *Keyring) Read(ctx context.Context) (*string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	val, err := keyring.Get(service, k.key)
	// Make this more idiomatic
	if err == keyring.ErrNotFound {
		k.log.Debug("Keyring.Read: key not found in keyring", "key", k.key)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading from keyring: %w", err)
	}

	k.log.Debug("Keyring.Read: loaded from keyring", "key", k.key, "size_bytes", len(val))
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
	if err != nil {
		return fmt.Errorf("setting value to keyring: %w", err)
	}

	k.log.Debug("Keyring.Set: saved to keyring", "key", k.key, "size_bytes", len(value))
	return nil
}

// Delete will delete a value from the keyring
func (k *Keyring) Delete(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	err := keyring.Delete(service, k.key)
	if err == keyring.ErrNotFound {
		return nil
	}

	if err != nil {
		return fmt.Errorf("could not delete from keyring: %w", err)
	}

	k.log.Debug("Keyring.Delete: removed from keyring", "key", k.key)
	return nil
}

func (k *Keyring) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
