package keypair

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePrivateKey(t *testing.T) {
	r := require.New(t)

	// TempFile replaces * with a random number
	privKeyFile, err := ioutil.TempFile("", "*PrivKey.pem")
	fmt.Println("priv key location:", privKeyFile.Name())
	r.Nil(err)

	defer privKeyFile.Close()
	defer os.Remove(privKeyFile.Name())

	privKey, err := GenerateRSAKeypair()
	r.NoError(err)

	privKeyBuffer, _, err := SaveRSAKeys(privKey)
	r.NoError(err)

	_, err = privKeyFile.Write(privKeyBuffer.Bytes())
	r.NoError(err)

	// Plug in path
	filePrivKey, err := ParseRSAPrivateKey(privKeyFile.Name())
	r.NoError(err)
	r.NotNil(privKey)
	r.Equal(filePrivKey, privKey)
}

func TestGeneratePrivateKey(t *testing.T) {
	r := require.New(t)
	priv, err := GenerateRSAKeypair()
	r.NoError(err)
	r.NotNil(priv)
}

func TestBufferHandling(t *testing.T) {
	r := require.New(t)

	originalPriv, err := GenerateRSAKeypair()
	r.NoError(err)

	originalPub := &originalPriv.PublicKey

	privKeyBuffer, pubKeyBuffer, err := SaveRSAKeys(originalPriv)
	r.NoError(err)

	// Decode Private block buffer and ensure its the same as original private key
	bufferPEMBlock, _ := pem.Decode(privKeyBuffer.Bytes())
	bufferPrivateKey, err := x509.ParsePKCS1PrivateKey(bufferPEMBlock.Bytes)
	r.NoError(err)
	r.Equal(bufferPrivateKey, originalPriv)

	// Decode Public block buffer and ensure its the same as original private key
	bufferPEMBlock, _ = pem.Decode(pubKeyBuffer.Bytes())
	bufferPubKey, err := x509.ParsePKCS1PublicKey(bufferPEMBlock.Bytes)
	r.NoError(err)
	r.Equal(bufferPubKey, originalPub)
}
