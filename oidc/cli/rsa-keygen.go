package cli

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/crypto/ssh"
)

// Generate new RSA keys.
// One of the intended use-cases is to generate keys for the
// Okta oauth [api authentication](https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/create-publicprivate-keypair/).
func GenerateRSAKey() (*jose.JSONWebKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	out, err := os.Create("rsa")
	if err != nil {
		return nil, err
	}
	defer out.Close()

	err = pem.Encode(
		out,
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)
	if err != nil {
		return nil, err
	}

	sshPub, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return nil, err
	}

	return &jose.JSONWebKey{
		Key:       priv.Public(),
		KeyID:     ssh.FingerprintSHA256(sshPub),
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}, nil
}
