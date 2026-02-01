package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
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
	log := slog.Default()
	log.Debug("NewCache: creating new cache instance",
		"storage_type", fmt.Sprintf("%T", storage),
	)
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
	log := slog.Default()
	startTime := time.Now()

	log.Debug("Cache.Read: attempting to read token from cache")

	cachedToken, err := c.readFromStorage(ctx)
	if err != nil {
		log.Error("Cache.Read: failed to read from storage",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	// if we have a valid token, use it
	if cachedToken.IsFresh() {
		log.Info("Cache.Read: found fresh token in cache",
			"token_expiry", cachedToken.Token.Expiry,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return cachedToken, nil
	}

	if cachedToken == nil {
		log.Debug("Cache.Read: no cached token found, will refresh")
	} else {
		log.Debug("Cache.Read: cached token is stale, will refresh",
			"token_expiry", cachedToken.Token.Expiry,
		)
	}

	// otherwise, try refreshing
	log.Debug("Cache.Read: initiating token refresh")
	token, err := c.refresh(ctx)
	if err != nil {
		log.Error("Cache.Read: refresh failed",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	log.Info("Cache.Read: successfully refreshed token",
		"token_expiry", token.Token.Expiry,
		"elapsed_ms", time.Since(startTime).Milliseconds(),
	)
	return token, nil
}

func (c *Cache) refresh(ctx context.Context) (*client.Token, error) {
	log := slog.Default()
	startTime := time.Now()

	log.Debug("Cache.refresh: acquiring lock")
	err := c.lock.Lock()
	if err != nil {
		log.Error("Cache.refresh: failed to acquire lock",
			"error", err,
		)
		return nil, err
	}
	log.Debug("Cache.refresh: lock acquired successfully")
	defer func() {
		log.Debug("Cache.refresh: releasing lock")
		c.lock.Unlock() //nolint:errcheck
	}()

	// acquire lock, try reading from cache again just in case
	// someone else got here first
	log.Debug("Cache.refresh: re-reading from storage after lock acquisition")
	cachedToken, err := c.readFromStorage(ctx)
	if err != nil {
		log.Error("Cache.refresh: failed to read from storage after lock",
			"error", err,
		)
		return nil, err
	}
	// if we have a valid token, use it
	if cachedToken.IsFresh() {
		log.Info("Cache.refresh: found fresh token after lock (another process refreshed)",
			"token_expiry", cachedToken.Token.Expiry,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return cachedToken, nil
	}

	// ok, at this point we have the lock and there are no good tokens around
	// fetch a new one and save it
	log.Debug("Cache.refresh: calling refreshToken function",
		"has_cached_token", cachedToken != nil,
	)
	token, err := c.refreshToken(ctx, cachedToken)
	if err != nil {
		log.Error("Cache.refresh: refreshToken function failed",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}
	log.Debug("Cache.refresh: refreshToken function succeeded",
		"token_expiry", token.Token.Expiry,
	)

	// check the new token is good to use
	if !token.IsFresh() {
		log.Error("Cache.refresh: fetched token is not fresh",
			"token_expiry", token.Token.Expiry,
		)
		return nil, fmt.Errorf("invalid token fetched")
	}

	// marshal token with options
	log.Debug("Cache.refresh: marshalling token")
	strToken, err := token.Marshal(c.storage.MarshalOpts()...)
	if err != nil {
		log.Error("Cache.refresh: failed to marshal token",
			"error", err,
		)
		return nil, fmt.Errorf("marshalling token: %w", err)
	}
	log.Debug("Cache.refresh: token marshalled",
		"token_length", len(strToken),
	)

	// gzip encode and save token to storage
	log.Debug("Cache.refresh: compressing token")
	compressedToken, err := compressToken(strToken)
	if err != nil {
		log.Error("Cache.refresh: failed to compress token",
			"error", err,
		)
		return nil, fmt.Errorf("compressing token: %w", err)
	}
	log.Debug("Cache.refresh: token compressed",
		"original_length", len(strToken),
		"compressed_length", len(compressedToken),
	)

	log.Debug("Cache.refresh: saving token to storage")
	err = c.storage.Set(ctx, compressedToken)
	if err != nil {
		if errors.Is(err, keyring.ErrSetDataTooBig) {
			log.Warn("Cache.refresh: token too big for storage, removing refresh token")

			strToken, err := token.Marshal(append(c.storage.MarshalOpts(), client.MarshalOptNoRefresh)...)
			if err != nil {
				log.Error("Cache.refresh: failed to marshal token without refresh",
					"error", err,
				)
				return nil, fmt.Errorf("marshalling token: %w", err)
			}

			compressedToken, err = compressToken(strToken)
			if err != nil {
				log.Error("Cache.refresh: failed to compress token without refresh",
					"error", err,
				)
				return nil, fmt.Errorf("compressing token: %w", err)
			}
			log.Debug("Cache.refresh: compressed token without refresh",
				"compressed_length", len(compressedToken),
			)

			err = c.storage.Set(ctx, compressedToken)
			if err != nil {
				log.Error("Cache.refresh: failed to save token without refresh",
					"error", err,
				)
				return nil, fmt.Errorf("caching without the refresh token: %w", err)
			}
			log.Debug("Cache.refresh: saved token without refresh token")
		} else {
			log.Error("Cache.refresh: failed to save token to storage",
				"error", err,
			)
			return nil, fmt.Errorf("caching the strToken: %w", err)
		}
	} else {
		log.Debug("Cache.refresh: token saved to storage successfully")
	}

	log.Info("Cache.refresh: completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Token.Expiry,
	)
	return token, nil
}

// reads token from storage, potentially returning a nil/expired token
// users must call IsFresh to check token validty
func (c *Cache) readFromStorage(ctx context.Context) (*client.Token, error) {
	log := slog.Default()

	log.Debug("Cache.readFromStorage: reading from storage backend")
	cached, err := c.storage.Read(ctx)
	if err != nil {
		log.Error("Cache.readFromStorage: failed to read from storage",
			"error", err,
		)
		return nil, err
	}
	if cached == nil {
		log.Debug("Cache.readFromStorage: no cached data found")
		return nil, nil
	}
	log.Debug("Cache.readFromStorage: cached data found",
		"cached_length", len(*cached),
	)

	// decode gzip data
	log.Debug("Cache.readFromStorage: decompressing token data")
	decompressedStr, err := decompressToken(*cached)
	if err != nil {
		// if we fail to decompress the token we should treat it as a cache miss instead of returning an error
		log.Warn("Cache.readFromStorage: failed to decompress token, treating as cache miss",
			"error", err,
		)
	} else {
		log.Debug("Cache.readFromStorage: token decompressed",
			"decompressed_length", len(*decompressedStr),
		)
	}

	log.Debug("Cache.readFromStorage: parsing token from string")
	cachedToken, err := client.TokenFromString(decompressedStr)
	if err != nil {
		log.Warn("Cache.readFromStorage: failed to parse token, purging from storage",
			"error", err,
		)
		err = c.storage.Delete(ctx) // can't read it, so attempt to purge it
		if err != nil {
			log.Warn("Cache.readFromStorage: failed to purge invalid token from storage",
				"error", err,
			)
		} else {
			log.Debug("Cache.readFromStorage: purged invalid token from storage")
		}
	} else if cachedToken != nil {
		log.Debug("Cache.readFromStorage: token parsed successfully",
			"token_expiry", cachedToken.Token.Expiry,
			"is_fresh", cachedToken.IsFresh(),
		)
	}
	return cachedToken, nil
}

func compressToken(token string) (string, error) {
	log := slog.Default()
	log.Debug("compressToken: starting compression",
		"input_length", len(token),
	)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(token)); err != nil {
		log.Error("compressToken: failed to write to gzip",
			"error", err,
		)
		return "", fmt.Errorf("failed to write to gzip: %w", err)
	}
	if err := gz.Close(); err != nil {
		log.Error("compressToken: failed to close gzip writer",
			"error", err,
		)
		return "", fmt.Errorf("failed to close gzip: %w", err)
	}

	result := buf.String()
	log.Debug("compressToken: compression complete",
		"input_length", len(token),
		"output_length", len(result),
		"compression_ratio", fmt.Sprintf("%.2f%%", float64(len(result))/float64(len(token))*100),
	)
	return result, nil
}

func decompressToken(token string) (*string, error) {
	log := slog.Default()
	log.Debug("decompressToken: starting decompression",
		"input_length", len(token),
	)

	reader := bytes.NewReader([]byte(token))
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		log.Error("decompressToken: failed to create gzip reader",
			"error", err,
		)
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	decompressed, err := io.ReadAll(gzreader)
	if err != nil {
		log.Error("decompressToken: failed to read gzip data",
			"error", err,
		)
		return nil, fmt.Errorf("failed to read gzip data: %w", err)
	}
	if err := gzreader.Close(); err != nil {
		log.Error("decompressToken: failed to close gzip reader",
			"error", err,
		)
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	decompressedStr := string(decompressed)
	log.Debug("decompressToken: decompression complete",
		"input_length", len(token),
		"output_length", len(decompressedStr),
	)
	return &decompressedStr, nil
}
