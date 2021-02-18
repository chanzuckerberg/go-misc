package keypair

import (
	"crypto/rsa"
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

	privKeyBuffer, err := SaveRSAKey(privKey)
	r.NoError(err)

	_, err = privKeyFile.Write(privKeyBuffer.Bytes())
	r.NoError(err)

	// Plug in path
	filePrivKey, err := ParseRSAPrivateKeyFile(privKeyFile.Name())
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

	privKeyBuffer, err := SaveRSAKey(originalPriv)
	r.NoError(err)

	// Decode Private block buffer and ensure its the same as original private key
	bufferPEMBlock, _ := pem.Decode(privKeyBuffer.Bytes())
	bufferPrivateKey, err := x509.ParsePKCS8PrivateKey(bufferPEMBlock.Bytes)
	pkcs8Key, ok := bufferPrivateKey.(*rsa.PrivateKey)
	r.True(ok)
	r.NoError(err)
	r.Equal(pkcs8Key, originalPriv)

	r.Equal(pkcs8Key.PublicKey, *originalPub)
}
