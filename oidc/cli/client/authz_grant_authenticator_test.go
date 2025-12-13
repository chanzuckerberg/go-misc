package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateState(t *testing.T) {
	r := require.New(t)

	material, err := newOauthMaterial()
	r.NoError(err)

	areEqual := bytesAreEqual(material.StateBytes, []byte("definitely doesn't match"))
	r.False(areEqual)

	// matches, send the same value
	areEqual = bytesAreEqual(material.StateBytes, material.StateBytes)
	r.True(areEqual)
}
