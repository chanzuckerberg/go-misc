package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
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
//      if not present or expired, will refresh
func (c *Cache) Read(ctx context.Context) (*client.Token, error) {
	// If we have a fresh token, use it
	cachedToken, err := c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	if cachedToken != nil {
		return cachedToken, nil
	}

	// otherwise, try refreshing
	return c.refresh(ctx, cachedToken)
}

func (c *Cache) refresh(ctx context.Context, cachedToken *client.Token) (*client.Token, error) {
	f, err := os.OpenFile("/tmp/oidc-lock-time", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	start := time.Now()
	err = c.lock.Lock()
	if err != nil {
		return nil, err
	}

	defer func() {
		c.lock.Unlock()                                                      //nolint: errcheck
		f.WriteString(fmt.Sprintf("%d\n", time.Since(start).Milliseconds())) //nolint: errcheck
	}()

	// acquire lock, try reading from cache again just in case
	// someone else got here first
	cachedToken, err = c.readFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	if cachedToken != nil {
		return cachedToken, nil
	}

	// ok, at this point we have the lock and there are no good tokens around
	// fetch a new one and save it
	token, err := c.refreshToken(ctx, cachedToken)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, errors.New("nil token returned")
	}

	strToken, err := token.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshall token")
	}

	err = c.storage.Set(ctx, strToken)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to cache the strToken")
	}

	return token, nil
}

func (c *Cache) readFromStorage(ctx context.Context) (*client.Token, error) {
	cached, err := c.storage.Read(ctx)
	if err != nil {
		return nil, err
	}
	cachedToken, err := client.TokenFromString(cached)
	if err != nil {
		logrus.Debugf("error fetching stored token: %s", err)
		err = c.storage.Delete(ctx) // can't read it, so attempt to purge it
		if err != nil {
			logrus.Debugf("error clearing token from storage: %s", err)
		}
	}
	if cachedToken.IsFresh() {
		return cachedToken, nil
	}
	return nil, nil
}
