// Package logging provides session-aware logging utilities for the OIDC CLI.
// It enables log correlation across concurrent invocations by attaching a
// unique session ID, hostname, and PID to all log entries.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"

	petname "github.com/dustinkirkland/golang-petname"
)

type loggerKey struct{}

func generateSessionID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b)
}

// NewLogger returns a logger enriched with session_id, hostname, and pid,
// and a context that carries it. Downstream code retrieves the logger
// via FromContext.
func NewLogger(ctx context.Context) (context.Context, *slog.Logger) {
	hostname, err := os.Hostname()
	hostnameGenerated := false
	if err != nil || hostname == "" {
		hostname = petname.Generate(2, "-")
		hostnameGenerated = true
	}

	logger := slog.Default().With(
		"session_id", generateSessionID(),
		"hostname", hostname,
		"pid", os.Getpid(),
		"uid", os.Getuid(),
	)

	if hostnameGenerated {
		logger.Warn("hostname unavailable, assigned pet name",
			"generated_hostname", hostname,
			"hostname_error", err,
		)
	}

	return context.WithValue(ctx, loggerKey{}, logger), logger
}

// FromContext returns the logger stored in ctx by NewLogger,
// falling back to slog.Default() if none is present.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
