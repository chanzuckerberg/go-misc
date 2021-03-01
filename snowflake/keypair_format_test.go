package snowflake

import (
	"testing"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/stretchr/testify/require"
)

func TestNoBigErrors(t *testing.T) {
	r := require.New(t)
	privKey, err := keypair.GenerateRSAKeypair()
	r.NoError(err)

	privKeyStr, pubKeyStr, err := RSAKeypairToString(privKey)
	r.NoError(err)
	r.NotEmpty(privKeyStr)
	r.NotEmpty(pubKeyStr)
}
