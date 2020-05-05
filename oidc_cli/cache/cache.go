package cache

import (
	"context"

	client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Cache to cache credentials
type Cache struct {
	storage storage.Storage

	refreshToken func(context.Context, *client.Token) (*client.Token, error)
}

// Cache returns a new cache
func NewCache(
	storage storage.Storage,
	refreshToken func(context.Context, *client.Token) (*client.Token, error),
) *Cache {
	return &Cache{
		storage:      storage,
		refreshToken: refreshToken,
	}
}

// Read will attempt to read a token from the cache
//      if not present or expired, will refresh
func (c *Cache) Read(ctx context.Context) (*client.Token, error) {
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
