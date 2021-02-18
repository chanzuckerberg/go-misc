package keypair

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func ParseRSAPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	expandedPrivateKeyPath, err := homedir.Expand(privateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid Path to private key")
	}

	privKeyBytes, err := ioutil.ReadFile(expandedPrivateKeyPath)
	if err != nil {
		return nil, errors.Errorf("Unable to read private key file")
	}

	privPemBlock, _ := pem.Decode(privKeyBytes)

	priv, err := x509.ParsePKCS8PrivateKey(privPemBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse private key file bytes")
	}

	if priv == nil {
		return nil, errors.Errorf("nil private key")
	}

	pkcs8Key, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("Unable to convert private key interface as rsa.PrivateKey. Got %T", priv)
	}

	return pkcs8Key, nil
}

func GetRSAPublicKey(privateKeyPath string) (*rsa.PublicKey, error) {
	privateKey, err := ParseRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read private key path")
	}

	if privateKey.PublicKey.Size() == 0 {
		return nil, errors.Errorf("Private key does not contain corresponding public key. Private key path: %s", privateKeyPath)
	}

	return &privateKey.PublicKey, nil
}

func GenerateRSAKeypair() (*rsa.PrivateKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot generate RSA key")
	}

	return privatekey, nil
}

func SaveRSAKey(privateKey *rsa.PrivateKey) (privateKeyBuffer *bytes.Buffer, err error) {
	if privateKey == nil {
		return &bytes.Buffer{}, errors.New("No private key set")
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return &bytes.Buffer{}, errors.New("Unable to marshal private key to pkcs8 format")
	}
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privKeyBuffer := bytes.NewBuffer(pem.EncodeToMemory(privateKeyBlock))

	return privKeyBuffer, nil
}
