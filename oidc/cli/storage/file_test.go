package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileReadWriteDefault(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	dir := t.TempDir()

	f, err := NewFile(ctx, dir, "client-id", "issuer-url")
	r.NoError(err)

	got, err := f.Read(ctx)
	r.NoError(err)
	r.Nil(got)

	err = f.Set(ctx, "hello")
	r.NoError(err)

	got, err = f.Read(ctx)
	r.NoError(err)
	r.NotNil(got)
	r.Equal("hello", *got)

	err = f.Delete(ctx)
	r.NoError(err)

	got, err = f.Read(ctx)
	r.NoError(err)
	r.Nil(got)
}

func TestFileCachePath(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	dir := t.TempDir()

	f, err := NewFile(ctx, dir, "client-id", "issuer-url")
	r.NoError(err)
	r.Contains(f.key, dir)
}

func TestFileCachePathLocal(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := t.TempDir()

	f, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	hostname, err := os.Hostname()
	r.NoError(err)

	r.Contains(f.key, localDir)
	r.Contains(f.key, hostname)
}

func TestFileBootstrapFromRoot(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := t.TempDir()

	root, err := NewFile(ctx, nfsDir, "client-id", "issuer-url")
	r.NoError(err)

	err = root.Set(ctx, "root-token-data")
	r.NoError(err)

	local, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	got, err := local.Read(ctx)
	r.NoError(err)
	r.NotNil(got)
	r.Equal("root-token-data", *got)

	r.Contains(local.key, localDir)
	_, err = os.Stat(local.key)
	r.NoError(err, "local cache file should exist after bootstrap")
}

func TestFileBootstrapNoRoot(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := t.TempDir()

	local, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	got, err := local.Read(ctx)
	r.NoError(err)
	r.Nil(got, "should return nil when root file doesn't exist")
}

func TestFileLocalWriteDoesNotTouchRoot(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := t.TempDir()

	root, err := NewFile(ctx, nfsDir, "client-id", "issuer-url")
	r.NoError(err)
	err = root.Set(ctx, "root-data")
	r.NoError(err)

	local, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	_, err = local.Read(ctx)
	r.NoError(err)

	err = local.Set(ctx, "local-updated-data")
	r.NoError(err)

	got, err := local.Read(ctx)
	r.NoError(err)
	r.NotNil(got)
	r.Equal("local-updated-data", *got)

	rootGot, err := root.Read(ctx)
	r.NoError(err)
	r.NotNil(rootGot)
	r.Equal("root-data", *rootGot, "root file should be unchanged")
}

func TestFileBootstrapSkipsWhenLocalExists(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := t.TempDir()

	root, err := NewFile(ctx, nfsDir, "client-id", "issuer-url")
	r.NoError(err)
	err = root.Set(ctx, "root-data")
	r.NoError(err)

	local, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	_, err = local.Read(ctx)
	r.NoError(err)

	err = local.Set(ctx, "local-data")
	r.NoError(err)

	err = root.Set(ctx, "root-data-v2")
	r.NoError(err)

	local2, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	got, err := local2.Read(ctx)
	r.NoError(err)
	r.NotNil(got)
	r.Equal("local-data", *got, "should use existing local file, not re-bootstrap from root")
}

func TestFileLocalDirCreatedOnBootstrap(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	nfsDir := t.TempDir()
	localDir := filepath.Join(t.TempDir(), "nested", "cache")

	root, err := NewFile(ctx, nfsDir, "client-id", "issuer-url")
	r.NoError(err)
	err = root.Set(ctx, "root-data")
	r.NoError(err)

	local, err := NewFile(ctx, nfsDir, "client-id", "issuer-url", WithLocalCacheDir(localDir))
	r.NoError(err)

	got, err := local.Read(ctx)
	r.NoError(err)
	r.NotNil(got)
	r.Equal("root-data", *got)

	_, err = os.Stat(localDir)
	r.NoError(err, "nested local dir should be created")
}
