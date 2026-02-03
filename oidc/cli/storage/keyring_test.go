package storage

import (
	"context"
	"testing"

	guuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func TestKeyringNilIfMissing(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	id := guuid.New()
	k := NewKeyring(ctx, id.String(), "testo")

	res, err := k.Read(ctx)
	r.Nil(err)
	r.Nil(res)

	err = k.Delete(ctx)
	r.Nil(err)
}

func TestKeyringSetReadDelete(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	val := "testoasfdasdf"

	id := guuid.New()
	k := NewKeyring(ctx, id.String(), "testo")

	err := k.Set(ctx, val)
	r.Nil(err)

	got, err := k.Read(ctx)
	r.Nil(err)
	r.NotNil(got)
	r.Equal(val, *got)

	err = k.Delete(ctx)
	r.Nil(err)
}
