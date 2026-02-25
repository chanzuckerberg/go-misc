package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/logging"
)

type File struct {
	// key is the active cache file path (local when localCacheDir is set, otherwise the default/NFS path).
	key string
	// dir is the parent directory of key.
	dir string
	// rootKey is the default/NFS cache file path used for bootstrapping.
	// Empty when no localCacheDir is configured.
	rootKey string

	log *slog.Logger
}

type fileConfig struct {
	localCacheDir string
}

// FileOption configures the File storage backend.
type FileOption func(*fileConfig)

// WithLocalCacheDir stores a per-hostname cache file on node-local disk,
// bootstrapped from the default (e.g. NFS) cache on first access.
func WithLocalCacheDir(dir string) FileOption {
	return func(c *fileConfig) {
		c.localCacheDir = dir
	}
}

func NewFile(ctx context.Context, dir string, clientID string, issuerURL string, opts ...FileOption) (*File, error) {
	var cfg fileConfig
	for _, o := range opts {
		o(&cfg)
	}

	rootKey := GenerateKey(dir, clientID, issuerURL)
	log := logging.FromContext(ctx)

	activeKey := rootKey
	activeDir := dir
	var rootKeyForBootstrap string

	if cfg.localCacheDir != "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("getting hostname: %w", err)
		}
		hashBase := filepath.Base(rootKey)
		activeKey = filepath.Join(cfg.localCacheDir, fmt.Sprintf("%s-%s", hashBase, hostname))
		activeDir = cfg.localCacheDir
		rootKeyForBootstrap = rootKey
	}

	log.Debug("File storage initialized",
		"cache_dir", activeDir,
		"cache_file", activeKey,
		"root_file", rootKeyForBootstrap,
		"client_id", clientID,
	)
	return &File{
		dir:     activeDir,
		key:     activeKey,
		rootKey: rootKeyForBootstrap,
		log:     log,
	}, nil
}

func GenerateKey(dir string, clientID string, issuerURL string) string {
	k := fmt.Sprintf("%s %s %s", storageVersion, clientID, issuerURL)
	h := sha256.Sum256([]byte(k))
	return filepath.Join(dir, hex.EncodeToString(h[:]))
}

// atomicFileWrite writes data to dest via a temp file in dir, using
// fsync + rename to ensure readers never observe a partial write.
func atomicFileWrite(dir string, dest string, data []byte) error {
	tmp, err := os.CreateTemp(dir, ".oidc-cache-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	cleanup := func() {
		tmp.Close()
		os.Remove(tmpName)
	}

	_, err = tmp.Write(data)
	if err != nil {
		cleanup()
		return fmt.Errorf("writing temp file: %w", err)
	}

	err = tmp.Sync()
	if err != nil {
		cleanup()
		return fmt.Errorf("syncing temp file: %w", err)
	}

	err = tmp.Close()
	if err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	err = os.Chmod(tmpName, 0600)
	if err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	err = os.Rename(tmpName, dest)
	if err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

const (
	maxAttempts = 5
	retryDelay  = 500 * time.Millisecond
)

func (f *File) Read(_ context.Context) (*string, error) {
	err := f.bootstrap()
	if err != nil {
		return nil, err
	}

	contents, err := f.readFileWithRetry(maxAttempts, retryDelay)
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

// bootstrap copies the root (NFS) cache file into the local cache path
// on first access. No-op when localCacheDir is not configured or the
// local file already exists.
func (f *File) bootstrap() error {
	if f.rootKey == "" {
		return nil
	}

	_, err := os.Stat(f.key)
	if err == nil {
		return nil
	}

	contents, err := os.ReadFile(f.rootKey)
	if err != nil {
		if os.IsNotExist(err) {
			f.log.Debug("File.bootstrap: root file not found, skipping", "root", f.rootKey)
			return nil
		}
		return fmt.Errorf("reading root cache file: %w", err)
	}

	err = os.MkdirAll(f.dir, 0700)
	if err != nil {
		return fmt.Errorf("creating local cache dir: %w", err)
	}

	err = atomicFileWrite(f.dir, f.key, contents)
	if err != nil {
		return fmt.Errorf("bootstrap write: %w", err)
	}

	f.log.Debug("File.bootstrap: copied root to local cache",
		"root", f.rootKey,
		"local", f.key,
		"size_bytes", len(contents),
	)
	return nil
}

// readFileWithRetry reads the cache file, retrying up to 10 times with a delay between
// attempts. The NFS backend may be eventually consistent across server
// frontends, so transient ENOENT or stale-handle errors are retried rather
// than treated as terminal.
func (f *File) readFileWithRetry(maxAttempts int, retryDelay time.Duration) ([]byte, error) {
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
	if os.IsNotExist(lastErr) {
		return nil, nil
	}
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

func (f *File) Set(_ context.Context, value string) error {
	err := os.MkdirAll(f.dir, 0700)
	if err != nil {
		return fmt.Errorf("could not create cache dir %s: %w", f.dir, err)
	}

	err = atomicFileWrite(f.dir, f.key, []byte(value))
	if err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	f.log.Debug("File.Set: saved to cache file", "path", f.key, "size_bytes", len(value))
	return nil
}

func (f *File) Delete(_ context.Context) error {
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
