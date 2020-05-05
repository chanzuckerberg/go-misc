package cache

import (
	"context"
	"testing"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func genStorage() storage.Storage {
	u := uuid.New()
	return storage.NewKeyring(u.String(), "testo")
}

func TestNewCache(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()

	u := uuid.New()

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		return &client.Token{IDToken: u.String()}, nil
	}

	c := NewCache(s, refresh)

	token, err := c.Read(ctx)
	r.NoError(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())
}

func TestCorruptedCache(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()
	err := s.Set(ctx, "garbage token")
	r.NoError(err)

	u := uuid.New()

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		return &client.Token{IDToken: u.String()}, nil
	}

	c := NewCache(s, refresh)

	token, err := c.Read(ctx)
	r.NoError(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())

	cachedToken, err := s.Read(ctx)
	r.NoError(err)
	r.NotNil(cachedToken)

	tok, err := client.TokenFromString(cachedToken)
	r.NoError(err)
	r.NotNil(t)

	r.Equal(u.String(), tok.IDToken)
}

func TestCachedToken(t *testing.T) {
	r := require.New(t)
	s := genStorage()
	ctx := context.Background()

	u := uuid.New()

	freshToken := &client.Token{
		IDToken: u.String(),
		Expiry:  time.Now().Add(time.Hour), // should always be fresh in this context... unless the tests are so slow
	}

	marshalled, err := freshToken.Marshal()
	r.NoError(err)

	err = s.Set(ctx, marshalled)
	r.NoError(err)

	refresh := func(ctx context.Context, c *client.Token) (*client.Token, error) {
		return nil, errors.New("always error")
	}

	c := NewCache(s, refresh)

	token, err := c.Read(ctx)
	r.Nil(err)
	r.NotNil(token)
	r.Equal(token.IDToken, u.String())
}
