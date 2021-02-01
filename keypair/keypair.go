package keypair

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

type Config struct {
	PrivateKey *rsa.PrivateKey
	KeyPrefix  string
	KeyPath    string
}

func (c *Config) GetPrivateKeyPath() string {
	return fmt.Sprintf("%s/%s_private.pem", c.KeyPath, c.KeyPrefix)
}
func (c *Config) GetPublicKeyPath() string {
	return fmt.Sprintf("%s/%s_public.pem", c.KeyPath, c.KeyPrefix)
}

func ParsePrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	expandedPrivateKeyPath, err := homedir.Expand(privateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid Path to private key")
	}

	privKeyBytes, err := ioutil.ReadFile(expandedPrivateKeyPath)
	if err != nil {
		return nil, errors.Errorf("Unable to read private key file")
	}

	privPemBlock, _ := pem.Decode(privKeyBytes)

	priv, err := x509.ParsePKCS1PrivateKey(privPemBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse private key file bytes")
	}

	if priv == nil {
		return nil, errors.Errorf("nil private key")
	}

	return priv, nil
}

// TODO(aku): Consider implementing this by reading public key file instead of reading the private key
func GetPublicKey(privateKeyPath string) (*rsa.PublicKey, error) {
	privateKey, err := ParsePrivateKey(privateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read private key path")
	}

	if privateKey.PublicKey.Size() == 0 {
		return nil, errors.Errorf("Private key does not contain corresponding public key. Private key path: %s", privateKeyPath)
	}

	return &privateKey.PublicKey, nil
}

func GenerateKeypair() (*rsa.PrivateKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot generate RSA key")
	}

	return privatekey, nil
}

func SaveKeys(config Config) error {
	if config.PrivateKey == nil {
		return errors.New("No private key set")
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(config.PrivateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privatePem, err := os.Create(config.GetPrivateKeyPath())
	if err != nil {
		return errors.Wrap(err, "Unable to create private key file")
	}

	err = pem.Encode(privatePem, privateKeyBlock)
	if err != nil {
		return errors.Wrap(err, "Unable to pem-encode private key")
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(&config.PrivateKey.PublicKey)

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicPem, err := os.Create(config.GetPublicKeyPath())
	if err != nil {
		return errors.Wrap(err, "Unable to create public key file")
	}

	err = pem.Encode(publicPem, publicKeyBlock)
	if err != nil {
		return errors.Wrap(err, "Unable to pem-encode public key")
	}

	return nil
}
