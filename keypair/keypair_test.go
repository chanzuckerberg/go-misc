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

var TestKeypairConfig = Config{
	KeyPrefix: "test",
	KeyPath:   ".",
}

func TestFileHandling(t *testing.T) {
	r := require.New(t)

	originalPriv, originalPub, err := GenerateKeypair()
	r.NoError(err)

	TestKeypairConfig.PrivateKey = originalPriv
	TestKeypairConfig.PublicKey = originalPub

	err = SaveKeys(TestKeypairConfig)
	r.NoError(err)

	defer os.Remove(TestKeypairConfig.GetPrivateKeyPath())
	defer os.Remove(TestKeypairConfig.GetPublicKeyPath())

	err = SaveKeys(TestKeypairConfig)
	r.NoError(err)

	priv, pub, err := FromFiles(TestKeypairConfig)
	r.NoError(err)
	r.NotNil(priv)
	r.NotNil(pub)
	r.Equal(originalPriv, priv)
	r.Equal(originalPub, pub)
}
