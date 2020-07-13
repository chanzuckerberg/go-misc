package client

import (
	"context"
	"testing"
	"time"

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

	testClientConfig := &Config{
		ClientID:  "dummyClientID",
		IssuerURL: "localhost",
		ServerConfig: &ServerConfig{
			FromPort: 0,
			ToPort:   0,
			Timeout:  time.Second,
		},
	}

	client, err := NewClient(context.Background(), testClientConfig)
	r.NoError(err)
	r.Equal(client.customMessages[oidcStatusSuccess], "success")

	// Set the success status
}
