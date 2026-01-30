package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// LogFilePath is the default log file path in /tmp
	LogFilePath = "/tmp/oidc-cli.log"
)

var (
	logger   *slog.Logger
	logFile  *os.File
	initOnce sync.Once
	initErr  error
)

// Init initializes the file logger. It creates or appends to the log file in /tmp.
// This function is safe to call multiple times; initialization only happens once.
func Init() (*slog.Logger, error) {
	initOnce.Do(func() {
		// Ensure the directory exists
		dir := filepath.Dir(LogFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			initErr = fmt.Errorf("creating log directory: %w", err)
			return
		}

		// Open log file for append (create if not exists)
		var err error
		logFile, err = os.OpenFile(LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("opening log file: %w", err)
			return
		}

		// Create a JSON handler that writes to the file
		handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Add more precision to timestamps
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format(time.RFC3339Nano))
					}
				}
				return a
			},
		})

		logger = slog.New(handler)

		// Log initialization
		logger.Info("OIDC CLI logger initialized",
			"log_file", LogFilePath,
			"pid", os.Getpid(),
		)
	})

	if initErr != nil {
		return nil, initErr
	}
	return logger, nil
}

// Get returns the initialized logger. Panics if Init() was not called or failed.
func Get() *slog.Logger {
	if logger == nil {
		// Try to initialize if not already done
		l, err := Init()
		if err != nil {
			// Return a no-op logger to avoid panics
			return slog.New(slog.NewTextHandler(os.Stderr, nil))
		}
		return l
	}
	return logger
}

// Close closes the log file. Should be called at program exit.
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}
