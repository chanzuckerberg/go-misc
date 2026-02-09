package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

// Cache to cache credentials
type Cache struct {
	storage storage.Storage
	lock    *pidlock.Lock
	log     *slog.Logger

	refreshToken func(context.Context, *client.Token) (*client.Token, error)
}

// NewCache returns a new cache
func NewCache(
	ctx context.Context,
	storage storage.Storage,
	refreshToken func(context.Context, *client.Token) (*client.Token, error),
	lock *pidlock.Lock,
) *Cache {
	return &Cache{
		storage:      storage,
		refreshToken: refreshToken,
		lock:         lock,
		log:          logging.FromContext(ctx),
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
	if cachedToken.Valid() {
		c.log.Debug("Cache.Read: using cached token",
			"token_expiry", cachedToken.Expiry,
			"has_refresh_token", cachedToken.RefreshToken != "",
			"email", cachedToken.Claims.Email,
		)
		return cachedToken, nil
	}

	c.log.Debug("Cache.Read: cached token is stale or empty, will refresh",
		"token_expiry", cachedToken.Expiry,
		"has_refresh_token", cachedToken.RefreshToken != "",
	)

	return c.refresh(ctx)
}

func (c *Cache) refresh(ctx context.Context) (*client.Token, error) {
	c.log.Debug("Cache.refresh: acquiring lock")
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

	// Check if another process refreshed while we waited for lock
	if cachedToken.Valid() {
		c.log.Debug("Cache.refresh: token was refreshed by another process",
			"token_expiry", cachedToken.Expiry,
		)
		return cachedToken, nil
	}

	c.log.Debug("Cache.refresh: calling refresh function",
		"has_access_token", cachedToken.AccessToken != "",
		"has_refresh_token", cachedToken.RefreshToken != "",
	)

	token, err := c.refreshToken(ctx, cachedToken)
	if err != nil {
		return nil, err
	}

	// marshal and save token
	err = c.saveToken(ctx, token)
	if err != nil {
		return nil, err
	}

	c.log.Debug("Cache.refresh: completed",
		"token_expiry", token.Token.Expiry,
		"has_refresh_token", token.Token.RefreshToken != "",
		"email", token.Claims.Email,
	)
	return token, nil
}

// saveToken marshals, compresses, and saves the token to storage
func (c *Cache) saveToken(ctx context.Context, token *client.Token) error {
	strToken, err := token.Marshal(c.storage.MarshalOpts()...)
	if err != nil {
		return fmt.Errorf("marshalling token: %w", err)
	}

	compressedToken, err := compressToken(strToken)
	if err != nil {
		return fmt.Errorf("compressing token: %w", err)
	}

	err = c.storage.Set(ctx, compressedToken)
	if err != nil {
		if errors.Is(err, keyring.ErrSetDataTooBig) {
			c.log.Warn("Cache.saveToken: token too big for storage, retrying without refresh token")
			return c.saveTokenWithoutRefresh(ctx, token)
		}
		return fmt.Errorf("caching the token: %w", err)
	}

	return nil
}

// saveTokenWithoutRefresh saves the token without the refresh token
func (c *Cache) saveTokenWithoutRefresh(ctx context.Context, token *client.Token) error {
	strToken, err := token.Marshal(append(c.storage.MarshalOpts(), client.MarshalOptNoRefresh)...)
	if err != nil {
		return fmt.Errorf("marshalling token: %w", err)
	}

	compressedToken, err := compressToken(strToken)
	if err != nil {
		return fmt.Errorf("compressing token: %w", err)
	}

	err = c.storage.Set(ctx, compressedToken)
	if err != nil {
		return fmt.Errorf("caching without the refresh token: %w", err)
	}

	return nil
}

// reads token from storage, potentially returning an empty/expired token
// users must call Valid to check token validity
func (c *Cache) readFromStorage(ctx context.Context) (*client.Token, error) {
	cached, err := c.storage.Read(ctx)
	if err != nil {
		return nil, err
	}
	if cached == nil {
		c.log.Debug("Cache.readFromStorage: no cached data found")
		return &client.Token{Token: &oauth2.Token{}}, nil
	}

	// decode gzip data
	decompressedStr, err := decompressToken(*cached)
	if err != nil {
		// if we fail to decompress the token we should treat it as a cache miss
		c.log.Warn("Cache.readFromStorage: failed to decompress cached token, treating as cache miss", "error", err)
		return &client.Token{Token: &oauth2.Token{}}, nil
	}

	cachedToken, err := client.TokenFromString(decompressedStr)
	if err != nil {
		c.log.Warn("Cache.readFromStorage: failed to parse cached token, purging", "error", err)
		deleteErr := c.storage.Delete(ctx)
		if deleteErr != nil {
			c.log.Warn("Cache.readFromStorage: failed to purge invalid token", "error", deleteErr)
		}
		return &client.Token{Token: &oauth2.Token{}}, nil
	}

	// Restore the id_token to the oauth2.Token extras so it can be extracted
	// via Token.Extra("id_token"). The IDToken field is persisted separately
	// since oauth2.Token extras don't survive JSON serialization.
	if cachedToken.IDToken != "" {
		c.log.Debug("Cache.readFromStorage: restoring id_token to oauth2.Token extras",
			"id_token_length", len(cachedToken.IDToken),
		)
		cachedToken.Token = cachedToken.WithExtra(map[string]interface{}{
			"id_token": cachedToken.IDToken,
		})
	}

	c.log.Debug("Cache.readFromStorage: loaded token from cache",
		"token_expiry", cachedToken.Expiry,
		"is_valid", cachedToken.Valid(),
		"has_refresh_token", cachedToken.RefreshToken != "",
	)
	return cachedToken, nil
}

func compressToken(token string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(token)); err != nil {
		return "", fmt.Errorf("failed to write to gzip: %w", err)
	}
	err := gz.Close()
	if err != nil {
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
	err = gzreader.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}
	decompressedStr := string(decompressed)
	return &decompressedStr, nil
}
