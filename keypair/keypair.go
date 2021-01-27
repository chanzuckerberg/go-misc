package keypair

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

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

}
