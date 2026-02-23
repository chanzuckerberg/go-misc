package pidlock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLockUnlock(t *testing.T) {
	r := require.New(t)
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "test.lock")

	lock, err := NewLock(lockPath)
	r.NoError(err)

	err = lock.Lock()
	r.NoError(err)

	err = lock.Unlock()
	r.NoError(err)
}

func TestLockIsExclusive(t *testing.T) {
	r := require.New(t)
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "test.lock")

	lock1, err := NewLock(lockPath)
	r.NoError(err)

	lock2, err := NewLock(lockPath)
	r.NoError(err)

	err = lock1.Lock()
	r.NoError(err)

	locked, err := lock2.fl.TryLock()
	r.NoError(err)
	r.False(locked, "second lock should not be acquired while first is held")

	err = lock1.Unlock()
	r.NoError(err)

	locked, err = lock2.fl.TryLock()
	r.NoError(err)
	r.True(locked, "second lock should succeed after first is released")

	err = lock2.Unlock()
	r.NoError(err)
}

func TestNewLockRequiresAbsolutePath(t *testing.T) {
	r := require.New(t)
	_, err := NewLock("relative/path/test.lock")
	r.Error(err)
}

func TestNewLockCreatesDirectory(t *testing.T) {
	r := require.New(t)
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "subdir", "test.lock")

	lock, err := NewLock(lockPath)
	r.NoError(err)

	_, err = os.Stat(filepath.Dir(lockPath))
	r.NoError(err, "lock directory should be created")

	err = lock.Lock()
	r.NoError(err)
	err = lock.Unlock()
	r.NoError(err)
}
