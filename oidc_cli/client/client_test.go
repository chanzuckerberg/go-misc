package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateState(t *testing.T) {
	r := require.New(t)

	material, err := newOauthMaterial()
	r.NoError(err)

	c := &Client{
		oauthMaterial: material,
	}

	err = c.ValidateState("definitely doesn't match")
	r.Error(err)

	// matches, send the same value
	err = c.ValidateState(c.oauthMaterial.State)
	r.NoError(err)
}
