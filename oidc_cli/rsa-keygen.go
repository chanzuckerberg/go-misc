package oidc

import (
	"fmt"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"golang.org/x/crypto/ssh"
	"encoding/pem"
	"gopkg.in/square/go-jose.v2"
	"os"
)

func generateRSAKey() error {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil
	}

	out, err := os.Create("rsa")
	if err != nil {
		return err
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
		return err
	}

	sshPub, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return err
	}

	pub := jose.JSONWebKey{
		Key:       priv.Public(),
		KeyID:     ssh.FingerprintSHA256(sshPub),
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	b, err := pub.MarshalJSON()
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
