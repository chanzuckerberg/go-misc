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
	// Generate tmp file
	// tmpFile, err := ioutil.TempFile("", "tmpKey")
	// r.Nil(err)

	// defer tmpFile.Close()
	defer os.Remove("private.pem")
	// This line assumes there is nothing wrong with GenerateKeypair()
	privKey, _, err := GenerateKeypair()
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
	// _, err = tmpFile.Write(privatePem)
	// r.NoError(err)
	// Plug in path
	filePrivKey, err := ParsePrivateKey("private.pem")
	r.NoError(err)
	r.NotNil(privKey)
	r.Equal(filePrivKey, privKey)
}

func TestGeneratePrivateKey(t *testing.T) {
	r := require.New(t)
	priv, pub, err := GenerateKeypair()
	r.NoError(err)
	r.NotNil(priv)
	r.NotNil(pub)
}
