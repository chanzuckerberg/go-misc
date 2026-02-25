// Package logging provides session-aware logging utilities for the OIDC CLI.
// It enables log correlation across concurrent invocations by attaching a
// unique session ID to all log entries.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
)

// sessionIDKey is the context key for the session ID
type sessionIDKey struct{}

// generateSessionID creates a short random session ID for log correlation
func generateSessionID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// Fall back to the zero-initialized buffer if randomness is unavailable.
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b)
}

// WithSessionID returns a new context with a session ID and a logger with the session ID attached.
// Use this at the entry points of the package (GetToken, RefreshToken) to start a new session.
func WithSessionID(ctx context.Context) (context.Context, *slog.Logger) {
	sessionID := generateSessionID()
	ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)
	hostname, _ := os.Hostname()
	log := slog.Default().With("session_id", sessionID, "hostname", hostname)
	return ctx, log
}

// FromContext returns a logger with the session ID from context, or the default logger.
// Use this in nested functions to get a logger that includes the session ID.
func FromContext(ctx context.Context) *slog.Logger {
	if sessionID, ok := ctx.Value(sessionIDKey{}).(string); ok {
		return slog.Default().With("session_id", sessionID)
	}
	return slog.Default()
}
