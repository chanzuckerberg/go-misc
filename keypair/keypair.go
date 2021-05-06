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
	"golang.org/x/crypto/ssh"
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

	return UnmarshalRSAPrivateKey(privKeyBytes)
}

func UnmarshalRSAPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	key, err := ssh.ParseRawPrivateKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse RSA private key")
	}

	k, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("unable to convert private key interface as rsa.PrivateKey. Got %T", key)
	}

	return k, nil
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

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY", // "RSA PRIVATE KEY" is associated with PKCS1, whereas "PRIVATE KEY" is associated with PKCS8
		Bytes: privateKeyBytes,
	}
	privKeyBuffer := bytes.NewBuffer(pem.EncodeToMemory(privateKeyBlock))

	return privKeyBuffer, nil
}
