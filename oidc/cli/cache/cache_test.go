package cache

import (
	"context"
	"fmt"
	"os"
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

	fileLockPath := fmt.Sprintf("/tmp/%s", u.String())
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

	fileLockPath := fmt.Sprintf("/tmp/%s", u.String())
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

	fileLockPath := fmt.Sprintf("/tmp/%s", u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	freshToken := &client.Token{
		IDToken: u.String(),
		Token: &oauth2.Token{
			Expiry: time.Now().Add(time.Hour), // should always be fresh in this context... unless the tests are so slow
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

	fileLockPath := fmt.Sprintf("/tmp/%s", u.String())
	defer os.Remove(fileLockPath)
	fileLock, err := pidlock.NewLock(fileLockPath)
	r.NoError(err)

	s := storage.NewFile(ctx, dir, "client-id", "issuer-url")

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
