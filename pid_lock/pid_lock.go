package util

import (
	"os"
	"path"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// lockPath returns the lock path given a path to the configPath
func lockPath(lockPath string) (string, error) {
	if !path.IsAbs(lockPath) {
		return "", errors.Errorf("%s must be an absolute path", lockPath)
	}
	return lockPath, nil
}

func defaultBackoff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 20 * time.Second
	b.InitialInterval = 10 * time.Millisecond
	b.MaxInterval = 100 * time.Millisecond
	return b
}

// Lock represents a pid lock
type Lock struct {
	lock    lockfile.Lockfile
	backoff backoff.BackOff
}

// NewLock returns a new lock
func NewLock(lockFilePath string) (*Lock, error) {
	lockPath, err := lockPath(lockFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not calculate lockfile path from %s", lockFilePath)
	}

	// Create the lock directory if needed
	err = os.MkdirAll(path.Dir(lockPath), 0755) // #nosec
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create %s", path.Dir(lockPath))
	}

	logrus.WithField("lock_path", lockPath).Debug("Creating pid lock")
	lock, err := lockfile.New(lockPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get lock from path %s", lockPath)
	}

	return &Lock{
		lock:    lock,
		backoff: defaultBackoff(),
	}, nil
}

// Lock will lock with retries.
func (l *Lock) Lock(optBackoff ...backoff.BackOff) error {
	b := l.backoff
	if len(optBackoff) == 1 {
		b = optBackoff[0]
	}

	return errors.Wrap(backoff.Retry(l.lock.TryLock, b), "Error acquiring lock")
}

// Unlock will unlock the pid lockfile
func (l *Lock) Unlock() error {
	return errors.Wrap(l.lock.Unlock(), "Error releasing lock")
}
