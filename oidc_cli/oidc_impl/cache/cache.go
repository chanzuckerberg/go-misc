package cache

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Cache to cache credentials
type Cache struct {
	storage storage.Storage
	lock    *pidlock.Lock

	refreshToken func(context.Context, *client.Token) (*client.Token, error)
}

// Cache returns a new cache
func NewCache(
	storage storage.Storage,
	refreshToken func(context.Context, *client.Token) (*client.Token, error),
	lock *pidlock.Lock,
) *Cache {
	return &Cache{
		storage:      storage,
		refreshToken: refreshToken,
		lock:         lock,
	}
}

// Read will attempt to read a token from the cache
//
//	if not present or expired, will refresh
func (c *Cache) Read(ctx context.Context) (*client.Token, error) {
	cachedToken, err := c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	if cachedToken == nil {
		fmt.Println("no cached Token")
	}
	// if we have a valid token, use it
	if cachedToken.IsFresh() {
		return cachedToken, nil
	}

	// otherwise, try refreshing
	return c.refresh(ctx)
}

func (c *Cache) refresh(ctx context.Context) (*client.Token, error) {
	err := c.lock.Lock()
	if err != nil {
		return nil, err
	}
	defer c.lock.Unlock() //nolint:errcheck

	// acquire lock, try reading from cache again just in case
	// someone else got here first
	cachedToken, err := c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	// if we have a valid token, use it
	if cachedToken.IsFresh() {
		return cachedToken, nil
	}

	// ok, at this point we have the lock and there are no good tokens around
	// fetch a new one and save it
	token, err := c.refreshToken(ctx, cachedToken)
	if err != nil {
		return nil, err
	}

	// check the new token is good to use
	if !token.IsFresh() {
		return nil, errors.New("invalid token fetched")
	}

	// marshal token with options
	strToken, err := token.Marshal(c.storage.MarshalOpts()...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall token")
	}
	fmt.Println("about to put strToken to storage: ", strToken)
	// save token to storage
	err = c.storage.Set(ctx, strToken)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to cache the strToken")
	}

	return token, nil
}

// reads token from storage, potentially returning a nil/expired token
// users must call IsFresh to check token validty
func (c *Cache) readFromStorage(ctx context.Context) (*client.Token, error) {
	cached, err := c.storage.Read(ctx)
	if err != nil {
		return nil, err
	}
	cachedToken, err := client.TokenFromString(cached)
	if err != nil {
		logrus.WithError(err).Debug("error fetching stored token")
		err = c.storage.Delete(ctx) // can't read it, so attempt to purge it
		if err != nil {
			logrus.WithError(err).Debug("error clearing token from storage")
		}
	}
	return cachedToken, nil
}
