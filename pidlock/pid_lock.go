package pidlock

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gofrs/flock"
)

func defaultBackoff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 2 * time.Minute
	b.InitialInterval = 10 * time.Millisecond
	b.MaxInterval = time.Second
	return b
}

// Lock represents a file lock backed by flock(2) on Unix and LockFileEx on Windows.
// Unlike PID-based locks, flock locks are kernel-mediated and work correctly
// across multiple hosts on NFS. The lock is automatically released if the
// process dies.
type Lock struct {
	fl      *flock.Flock
	backoff backoff.BackOff
}

// NewLock returns a new lock for the given file path.
// The path must be absolute. The parent directory is created if needed.
func NewLock(lockFilePath string) (*Lock, error) {
	if !path.IsAbs(lockFilePath) {
		return nil, fmt.Errorf("%s must be an absolute path", lockFilePath)
	}

	if err := os.MkdirAll(path.Dir(lockFilePath), 0755); err != nil { // #nosec
		return nil, fmt.Errorf("creating lock directory %s: %w", path.Dir(lockFilePath), err)
	}

	slog.Debug("Creating flock", "lock_path", lockFilePath)

	return &Lock{
		fl:      flock.New(lockFilePath),
		backoff: defaultBackoff(),
	}, nil
}

// Lock acquires an exclusive lock with retries using exponential backoff.
// An optional backoff strategy can be provided to override the default.
func (l *Lock) Lock(optBackoff ...backoff.BackOff) error {
	b := l.backoff
	if len(optBackoff) == 1 {
		b = optBackoff[0]
	}

	err := backoff.Retry(func() error {
		locked, err := l.fl.TryLock()
		if err != nil {
			return fmt.Errorf("acquiring lock: %w", err)
		}
		if !locked {
			return fmt.Errorf("lock is held by another process")
		}
		return nil
	}, b)
	if err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	return nil
}

// Unlock releases the file lock.
func (l *Lock) Unlock() error {
	if err := l.fl.Unlock(); err != nil {
		return fmt.Errorf("releasing lock: %w", err)
	}
	return nil
}
