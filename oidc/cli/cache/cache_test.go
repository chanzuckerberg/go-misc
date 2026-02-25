package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"golang.org/x/oauth2"

	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func genStorage() storage.Storage {
	u := uuid.New()
	return storage.NewKeyring(context.Background(), u.String(), "testo")
}

func TestNewCache(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()

	u := uuid.New()

	fileLockPath := filepath.Join(os.TempDir(), u.String())
	defer os.Remove(fileLockPath)

	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		// returns a "valid" token
		return &client.Token{IDToken: u.String(), Token: &oauth2.Token{Expiry: time.Now().Add(time.Hour)}}, nil
	}

	c := NewCache(ctx, s, refresh, fileLock)

	token, err := c.Read(ctx)
	r.NoError(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())
}

func TestCorruptedCache(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()
	compressed, err := compressToken("garbage token")
	r.NoError(err)
	err = s.Set(ctx, compressed)
	r.NoError(err)

	u := uuid.New()

	fileLockPath := filepath.Join(os.TempDir(), u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		// returns a "fresh" token
		return &client.Token{IDToken: u.String(), Token: &oauth2.Token{Expiry: time.Now().Add(time.Hour)}}, nil
	}

	c := NewCache(ctx, s, refresh, fileLock)

	token, err := c.Read(ctx)
	r.NoError(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())

	cachedToken, err := s.Read(ctx)
	r.NoError(err)
	r.NotNil(cachedToken)

	decompressedToken, err := decompressToken(*cachedToken)
	r.NoError(err)

	tok, err := client.TokenFromString(decompressedToken)
	r.NoError(err)
	r.NotNil(t)

	r.Equal(u.String(), tok.IDToken)
}

func TestCachedToken(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()

	u := uuid.New()

	fileLockPath := filepath.Join(os.TempDir(), u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	freshToken := &client.Token{
		IDToken: u.String(),
		Token: &oauth2.Token{
			AccessToken: "test-access-token",
			Expiry:      time.Now().Add(time.Hour), // should always be fresh in this context... unless the tests are so slow
		},
	}

	marshalled, err := freshToken.Marshal()
	r.NoError(err)

	compressed, err := compressToken(marshalled)
	r.NoError(err)

	err = s.Set(ctx, compressed)
	r.NoError(err)

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		return nil, fmt.Errorf("always error")
	}

	c := NewCache(ctx, s, refresh, fileLock)

	token, err := c.Read(ctx)
	r.Nil(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())
}

func TestFileCache(t *testing.T) {
	r := require.New(t)
	u := uuid.New()
	ctx := context.Background()

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		// returns a "fresh" token
		return &client.Token{
			IDToken: u.String(),
			Token: &oauth2.Token{
				Expiry:       time.Now().Add(time.Hour),
				RefreshToken: "some refresh token",
			},
		}, nil
	}

	dir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.Remove(dir)

	fileLockPath := filepath.Join(os.TempDir(), u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	s, err := storage.NewFile(ctx, dir, "client-id", "issuer-url")
	r.NoError(err)

	c := NewCache(ctx, s, refresh, fileLock)

	token, err := c.Read(ctx)
	r.NoError(err)

	r.NotNil(token)
	r.NotEmpty(token.RefreshToken)

	token, err = c.Read(ctx)
	r.NoError(err)

	r.NotNil(token)
	r.NotEmpty(token.RefreshToken)
}

// TestFileCacheIDTokenRestored verifies that when a valid token is read from the file cache,
// the id_token is properly restored to the oauth2.Token extras so that Token.Extra("id_token")
// returns the cached id_token value. This is critical because oauth2.Token extras don't survive
// JSON serialization, so we store IDToken separately and restore it on read.
func TestFileCacheIDTokenRestored(t *testing.T) {
	r := require.New(t)
	u := uuid.New()
	ctx := context.Background()

	expectedIDToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test-id-token." + u.String()

	// Create a valid token with an id_token
	originalToken := &client.Token{
		IDToken: expectedIDToken,
		Token: &oauth2.Token{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			Expiry:       time.Now().Add(time.Hour), // valid for 1 hour
		},
		Claims: client.Claims{
			Email:   "test@example.com",
			Subject: "test-subject",
		},
	}

	// Set up file storage
	dir, err := os.MkdirTemp("", "oidc-cache-test-*")
	r.NoError(err)
	defer os.RemoveAll(dir)

	fileLockPath := filepath.Join(os.TempDir(), u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	s, err := storage.NewFile(ctx, dir, "client-id", "issuer-url")
	r.NoError(err)

	// Manually write the token to storage (simulating a previous save)
	marshalled, err := originalToken.Marshal()
	r.NoError(err)
	compressed, err := compressToken(marshalled)
	r.NoError(err)
	err = s.Set(ctx, compressed)
	r.NoError(err)

	// Create cache with a refresh function that should NOT be called
	refreshCalled := false
	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		refreshCalled = true
		return nil, fmt.Errorf("refresh should not be called for valid cached token")
	}

	c := NewCache(ctx, s, refresh, fileLock)

	// Read from cache - should return the cached token without calling refresh
	token, err := c.Read(ctx)
	r.NoError(err)
	r.NotNil(token)

	// Verify refresh was NOT called (token was valid)
	r.False(refreshCalled, "refresh function should not be called for valid cached token")

	// Verify the token is valid
	r.True(token.Valid(), "cached token should be valid")

	// Verify the IDToken field is preserved
	r.Equal(expectedIDToken, token.IDToken, "IDToken field should be preserved")

	// Verify the id_token is accessible via Extra("id_token") - this is the critical test
	// This ensures the id_token was restored to oauth2.Token extras after deserialization
	extractedIDToken := token.Token.Extra("id_token")
	r.NotNil(extractedIDToken, "id_token should be accessible via Extra()")
	extractedIDTokenStr, ok := extractedIDToken.(string)
	r.True(ok, "id_token should be a string")
	r.Equal(expectedIDToken, extractedIDTokenStr, "id_token from Extra() should match original")

	// Verify other token fields are preserved
	r.Equal("test-access-token", token.AccessToken)
	r.Equal("test-refresh-token", token.RefreshToken)
	r.Equal("test@example.com", token.Claims.Email)
}
