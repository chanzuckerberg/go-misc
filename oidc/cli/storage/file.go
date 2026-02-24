package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
)

type File struct {
	key string
	dir string
	log *slog.Logger

	mu sync.Mutex
}

func NewFile(ctx context.Context, dir string, clientID string, issuerURL string) *File {
	key := generateKey(dir, clientID, issuerURL)
	log := logging.FromContext(ctx)

	log.Debug("File storage initialized",
		"cache_dir", dir,
		"cache_file", key,
		"client_id", clientID,
	)
	return &File{
		dir: dir,
		key: key,
		log: log,
	}
}

func generateKey(dir string, clientID string, issuerURL string) string {
	k := fmt.Sprintf("%s %s %s", storageVersion, clientID, issuerURL)
	h := sha256.Sum256([]byte(k))
	return path.Join(dir, hex.EncodeToString(h[:]))
}

func (f *File) Read(_ context.Context) (*string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	contents, err := f.readFile()
	if err != nil {
		return nil, err
	}
	if contents == nil {
		return nil, nil
	}

	f.log.Debug("File.Read: loaded from cache file", "path", f.key, "size_bytes", len(contents))
	stringContents := string(contents)
	return &stringContents, nil
}

// readFile reads the cache file, retrying up to 10 times with a delay between
// attempts. The NFS backend may be eventually consistent across server
// frontends, so transient ENOENT or stale-handle errors are retried rather
// than treated as terminal.
func (f *File) readFile() ([]byte, error) {
	const maxAttempts = 10
	const retryDelay = 500 * time.Millisecond

	var lastErr error
	for attempt := range maxAttempts {
		contents, err := os.ReadFile(f.key)
		if err == nil {
			return contents, nil
		}

		lastErr = err
		f.log.Debug("File.readFile: read failed",
			"path", f.key,
			"attempt", attempt+1,
			"max_attempts", maxAttempts,
			"error", err,
		)

		if attempt < maxAttempts-1 {
			time.Sleep(retryDelay)
		}
	}

	f.logCacheMiss()
	return nil, fmt.Errorf("could not read file after %d attempts: %w", maxAttempts, lastErr)
}

// logCacheMiss logs diagnostic information when the cache file is missing.
// Inspects the parent directory to help distinguish between "file was never
// written" (empty dir) and "file was deleted" (dir exists but file is gone).
func (f *File) logCacheMiss() {
	dirInfo, statErr := os.Stat(f.dir)
	if statErr != nil {
		f.log.Debug("File.readFile: cache miss, cache dir inaccessible",
			"path", f.key,
			"dir", f.dir,
			"dir_error", statErr,
		)
		return
	}

	entries, readErr := os.ReadDir(f.dir)
	if readErr != nil {
		f.log.Debug("File.readFile: cache miss, could not list cache dir",
			"path", f.key,
			"dir", f.dir,
			"dir_mod_time", dirInfo.ModTime(),
			"dir_error", readErr,
		)
		return
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}

	f.log.Debug("File.readFile: cache miss",
		"path", f.key,
		"dir", f.dir,
		"dir_mod_time", dirInfo.ModTime(),
		"dir_file_count", len(entries),
		"dir_files", names,
	)
}

func (f *File) Set(_ context.Context, value string) (err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	err = os.MkdirAll(f.dir, 0700)
	if err != nil {
		return fmt.Errorf("could not create cache dir %s: %w", f.dir, err)
	}

	// Write to a temp file then atomically rename into place so that
	// concurrent readers never observe a truncated/partial file.
	tmp, err := os.CreateTemp(f.dir, ".oidc-cache-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	closed := false
	renamed := false
	defer func() {
		if !closed {
			closeErr := tmp.Close()
			if closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}
		if !renamed {
			removeErr := os.Remove(tmpName)
			if removeErr != nil {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	_, err = tmp.Write([]byte(value))
	if err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	err = tmp.Sync()
	if err != nil {
		return fmt.Errorf("syncing temp file: %w", err)
	}

	err = tmp.Close()
	if err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	closed = true

	err = os.Chmod(tmpName, 0600)
	if err != nil {
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	err = os.Rename(tmpName, f.key)
	if err != nil {
		return fmt.Errorf("renaming temp file to cache file: %w", err)
	}
	renamed = true

	f.log.Debug("File.Set: saved to cache file", "path", f.key, "size_bytes", len(value))
	return nil
}

func (f *File) Delete(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := os.Remove(f.key)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not delete from file: %w", err)
	}

	f.log.Debug("File.Delete: removed cache file", "path", f.key)
	return nil
}

func (f *File) MarshalOpts() []client.MarshalOpts {
	return []client.MarshalOpts{}
}
