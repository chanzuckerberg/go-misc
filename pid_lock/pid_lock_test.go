package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func fileLockWithUnlock(lockPath string, r *require.Assertions) {
	fileLock, err := NewLock(lockPath)
	defer fileLock.Unlock() //nolint: errcheck
	r.NoError(err)
	err = fileLock.Lock()
	r.NoError(err)
}

// defer fileLock included
func TestIncludeDeferFileUnlock(t *testing.T) {
	r := require.New(t)
	dir, err := ioutil.TempDir("", "pid_lock_test")
	r.NoError(err)
	defer os.RemoveAll(dir) //nolint: errcheck
	testLockPath := filepath.Join(dir, "testlock.lock")

	fileLockWithUnlock(testLockPath, r)
	_, err = NewLock(testLockPath)
	r.NoError(err)
}
