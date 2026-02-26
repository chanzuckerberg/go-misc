package cli

import (
	"context"
	"testing"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/cache"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/compress"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func storeToken(t *testing.T, s storage.Storage, tok *client.Token) {
	t.Helper()
	r := require.New(t)

	b64, err := tok.Marshal()
	r.NoError(err)

	compressed, err := compress.GzipStr(b64)
	r.NoError(err)

	r.NoError(s.Set(context.Background(), compressed))
}

func TestSyncFromRootIfNewerUsesRoot(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	rootDir := t.TempDir()
	localDir := t.TempDir()

	olderExpiry := time.Now().Add(24 * time.Hour)
	newerExpiry := time.Now().Add(7 * 24 * time.Hour)

	rootStorage, err := storage.NewFile(ctx, rootDir, "cid", "issuer")
	r.NoError(err)
	storeToken(t, rootStorage, &client.Token{
		Token: &oauth2.Token{
			AccessToken:  "root-access",
			RefreshToken: "root-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
		RefreshTokenExpiry: &newerExpiry,
	})

	localStorage, err := storage.NewFile(ctx, localDir, "cid", "issuer")
	r.NoError(err)
	storeToken(t, localStorage, &client.Token{
		Token: &oauth2.Token{
			AccessToken:  "local-access",
			RefreshToken: "local-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
		RefreshTokenExpiry: &olderExpiry,
	})

	lockPath, err := lockFilePath("cid", "issuer", localDir)
	r.NoError(err)
	fileLock, err := pidlock.NewLock(lockPath)
	r.NoError(err)

	trySyncFromRootIfNewer(ctx, fileLock, rootStorage, localStorage)

	localCache := cache.NewCache(ctx, localStorage, nil, nil)
	tok, err := localCache.DecodeFromStorage(ctx)
	r.NoError(err)
	r.NotNil(tok.RefreshTokenExpiry)
	r.WithinDuration(newerExpiry, *tok.RefreshTokenExpiry, time.Second,
		"local should now have root's newer refresh token expiry")
	r.Equal("root-access", tok.AccessToken)
}

func TestSyncFromRootIfNewerKeepsLocal(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	rootDir := t.TempDir()
	localDir := t.TempDir()

	olderExpiry := time.Now().Add(24 * time.Hour)
	newerExpiry := time.Now().Add(7 * 24 * time.Hour)

	rootStorage, err := storage.NewFile(ctx, rootDir, "cid", "issuer")
	r.NoError(err)
	storeToken(t, rootStorage, &client.Token{
		Token: &oauth2.Token{
			AccessToken:  "root-access",
			RefreshToken: "root-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
		RefreshTokenExpiry: &olderExpiry,
	})

	localStorage, err := storage.NewFile(ctx, localDir, "cid", "issuer")
	r.NoError(err)
	storeToken(t, localStorage, &client.Token{
		Token: &oauth2.Token{
			AccessToken:  "local-access",
			RefreshToken: "local-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
		RefreshTokenExpiry: &newerExpiry,
	})

	lockPath, err := lockFilePath("cid", "issuer", localDir)
	r.NoError(err)
	fileLock, err := pidlock.NewLock(lockPath)
	r.NoError(err)

	trySyncFromRootIfNewer(ctx, fileLock, rootStorage, localStorage)

	localCache := cache.NewCache(ctx, localStorage, nil, nil)
	tok, err := localCache.DecodeFromStorage(ctx)
	r.NoError(err)
	r.NotNil(tok.RefreshTokenExpiry)
	r.WithinDuration(newerExpiry, *tok.RefreshTokenExpiry, time.Second,
		"local should keep its own newer refresh token expiry")
	r.Equal("local-access", tok.AccessToken)
}

func TestSyncFromRootIfNewerNoopWhenRootEmpty(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	rootDir := t.TempDir()
	localDir := t.TempDir()

	rootStorage, err := storage.NewFile(ctx, rootDir, "cid", "issuer")
	r.NoError(err)

	localStorage, err := storage.NewFile(ctx, localDir, "cid", "issuer")
	r.NoError(err)
	localExpiry := time.Now().Add(24 * time.Hour)
	storeToken(t, localStorage, &client.Token{
		Token: &oauth2.Token{
			AccessToken:  "local-access",
			RefreshToken: "local-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
		RefreshTokenExpiry: &localExpiry,
	})

	lockPath, err := lockFilePath("cid", "issuer", localDir)
	r.NoError(err)
	fileLock, err := pidlock.NewLock(lockPath)
	r.NoError(err)

	trySyncFromRootIfNewer(ctx, fileLock, rootStorage, localStorage)

	localCache := cache.NewCache(ctx, localStorage, nil, nil)
	tok, err := localCache.DecodeFromStorage(ctx)
	r.NoError(err)
	r.Equal("local-access", tok.AccessToken, "local should be unchanged when root is empty")
}
