package keypair

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePrivateKey(t *testing.T) {
	r := require.New(t)

	defer os.Remove("private.pem")

	privKey, err := GenerateKeypair()
	r.NoError(err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privatePem, err := os.Create("private.pem")
	r.NoError(err)
	err = pem.Encode(privatePem, privateKeyBlock)
	r.NoError(err)
	// Plug in path
	filePrivKey, err := ParsePrivateKey("private.pem")
	r.NoError(err)
	r.NotNil(privKey)
	r.Equal(filePrivKey, privKey)
}

func TestGeneratePrivateKey(t *testing.T) {
	r := require.New(t)
	priv, err := GenerateKeypair()
	r.NoError(err)
	r.NotNil(priv)
}

func TestFileHandling(t *testing.T) {
	r := require.New(t)

	originalPriv, err := GenerateKeypair()
	r.NoError(err)

	originalPub := &originalPriv.PublicKey

	privKeyBuffer, pubKeyBuffer, err := SaveKeys(originalPriv)
	r.NoError(err)

	bufferPEMBlock, _ := pem.Decode(privKeyBuffer.Bytes())
	bufferPrivateKey, err := x509.ParsePKCS1PrivateKey(bufferPEMBlock.Bytes)
	r.NoError(err)

	r.Equal(bufferPrivateKey, originalPriv)

	bufferPEMBlock, _ = pem.Decode(pubKeyBuffer.Bytes())
	bufferPubKey, err := x509.ParsePKCS1PublicKey(bufferPEMBlock.Bytes)
	r.NoError(err)
	r.Equal(bufferPubKey, originalPub)
}
