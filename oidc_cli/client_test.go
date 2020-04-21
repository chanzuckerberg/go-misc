package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateState(t *testing.T) {
	r := require.New(t)

	c := &Client{
		oauthMaterial: &oauthMaterial{
			State: "qwerlkajsdflkasjfdoiquwer",
		},
	}

	err := c.ValidateState("definitely doesn't match")
	r.Error(err)

	err = c.ValidateState(c.oauthMaterial.State)
	r.NoError(err)
}
