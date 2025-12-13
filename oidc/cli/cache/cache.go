package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"
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
		return nil, fmt.Errorf("invalid token fetched")
	}

	// marshal token with options
	strToken, err := token.Marshal(c.storage.MarshalOpts()...)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall token: %w", err)
	}

	// gzip encode and save token to storage
	compressedToken, err := compressToken(strToken)
	if err != nil {
		return nil, fmt.Errorf("unable to compress token: %w", err)
	}

	err = c.storage.Set(ctx, compressedToken)
	if err != nil {
		if errors.Is(err, keyring.ErrSetDataTooBig) {
			slog.Debug("Token too big, removing refresh token")

			strToken, err := token.Marshal(append(c.storage.MarshalOpts(), client.MarshalOptNoRefresh)...)
			if err != nil {
				return nil, fmt.Errorf("marshalling token: %w", err)
			}

			compressedToken, err = compressToken(strToken)
			if err != nil {
				return nil, fmt.Errorf("compressing token: %w", err)
			}

			err = c.storage.Set(ctx, compressedToken)
			if err != nil {
				return nil, fmt.Errorf("caching without the refresh token: %w", err)
			}
		} else {
			return nil, fmt.Errorf("caching the strToken: %w", err)
		}
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
	if cached == nil {
		return nil, nil
	}

	// decode gzip data
	decompressedStr, err := decompressToken(*cached)
	if err != nil {
		// if we fail to decompress the token we should treat it as a cache miss instead of returning an error
		slog.Debug(fmt.Errorf("failed to decompress token: %w", err).Error())
	}

	cachedToken, err := client.TokenFromString(decompressedStr)
	if err != nil {
		slog.Debug("fetching stored token", "error", err)
		err = c.storage.Delete(ctx) // can't read it, so attempt to purge it
		if err != nil {
			slog.Debug("clearing token from storage", "error", err)
		}
	}
	return cachedToken, nil
}

func compressToken(token string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(token)); err != nil {
		return "", fmt.Errorf("failed to write to gzip: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip: %w", err)
	}
	return buf.String(), nil
}

func decompressToken(token string) (*string, error) {
	reader := bytes.NewReader([]byte(token))
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	decompressed, err := io.ReadAll(gzreader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip data: %w", err)
	}
	if err := gzreader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}
	decompressedStr := string(decompressed)
	return &decompressedStr, nil
}
