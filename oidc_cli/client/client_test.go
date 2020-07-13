package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateState(t *testing.T) {
	r := require.New(t)

	material, err := newOauthMaterial()
	r.NoError(err)

	c := &Client{}

	err = c.ValidateState(material.StateBytes, []byte("definitely doesn't match"))
	r.Error(err)

	// matches, send the same value
	err = c.ValidateState(material.StateBytes, material.StateBytes)
	r.NoError(err)
}

func TestClientConfig(t *testing.T) {
	r := require.New(t)

	addSuccessMsgFunc := func(c *Client) *Client {
		c.customMsgs["success"] = "foo"
	}

	NewClient(context.Background())
}
