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
	"golang.org/x/crypto/ssh"
)

type Config struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
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
	// TODO(aku): figure out a decryption process if privateKey is encrypted with a passphrase
	privateKeyBytes, err := ioutil.ReadFile(expandedPrivateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "Could not read private key")
	}

	if len(privateKeyBytes) == 0 {
		return nil, errors.New("Private key is empty")
	}

	privateKey, err := ssh.ParseRawPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse private key")
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("privateKey not of type RSA")
	}

	return rsaPrivateKey, nil
}

func GenerateKeypair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Cannot generate RSA key")
	}

	publickey := &privatekey.PublicKey

	return privatekey, publickey, nil
}

func SaveKeys(config Config) error {
	if config.PrivateKey == nil {
		return errors.New("No private key set")
	}

	if config.PublicKey == nil {
		return errors.New("No public key set")
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

	publicKeyBytes := x509.MarshalPKCS1PublicKey(config.PublicKey)

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

func FromFiles(config Config) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Load private key
	privKeyBytes, err := ioutil.ReadFile(config.GetPrivateKeyPath())
	if err != nil {
		return nil, nil, errors.Errorf("Unable to read private key file")
	}

	privPemBlock, _ := pem.Decode(privKeyBytes)

	priv, err := x509.ParsePKCS1PrivateKey(privPemBlock.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to parse private key file bytes")
	}

	if priv == nil {
		return nil, nil, errors.Errorf("nil private key")
	}

	// Load public key
	pubKeyBytes, err := ioutil.ReadFile(config.GetPublicKeyPath())
	if err != nil {
		return nil, nil, errors.Errorf("Unable to read public key file")
	}

	pubPemBlock, _ := pem.Decode(pubKeyBytes)

	pub, err := x509.ParsePKCS1PublicKey(pubPemBlock.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to parse public key file bytes")
	}

	if pub == nil {
		return nil, nil, errors.Errorf("nil public key")
	}

	return priv, pub, nil
}
