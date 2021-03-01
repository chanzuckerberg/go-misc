package snowflake

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"strings"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/pkg/errors"
)

// Formatting based on requirements here: https://docs.snowflake.com/en/user-guide/key-pair-auth.html
// Important parts:
//  - pkcs8 format
//  - 2048 bits
//  - pem encoding
//  - no whitespace and headers in private key string
//  - base64 encoding
func KeypairToString(rsaPrivKey *rsa.PrivateKey) (snowflakePrivateKey string, snowflakePublicKey string, err error) {
	// get public key string
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(rsaPrivKey.PublicKey)
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to marshal public key to bytes")
	}

	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)

	// get private key string
	privKeyBuffer, err := keypair.SaveRSAKey(rsaPrivKey)
	if err != nil {
		return "", "", err
	}

	privateKeyStr := privKeyBuffer.String()
	stripHeaders := strings.ReplaceAll(privateKeyStr, "-----BEGIN RSA PRIVATE KEY-----", "")
	stripFooters := strings.ReplaceAll(stripHeaders, "-----END RSA PRIVATE KEY-----", "")
	privKeyNoWhitespace := strings.TrimSpace(stripFooters)

	return privKeyNoWhitespace, publicKeyStr, nil
}
