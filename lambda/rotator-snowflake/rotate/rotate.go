package rotate

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func getUsers() ([]string, error) {
	// TODO(aku): get list from okta
	return []string{os.Getenv("CURRENT_USER")}, nil
}

func updateSnowflake(user string, db *sql.DB, privKey *bytes.Buffer) error {
	privPemBlock, _ := pem.Decode(privKey.Bytes())

	privKeyIface, err := x509.ParsePKCS8PrivateKey(privPemBlock.Bytes)
	if err != nil {
		return errors.Wrap(err, "Unable to get pkcs8 private key")
	}

	// privKeyIface, err := x509.ParsePKCS8PrivateKey(der []byte)
	officialPrivateKey, ok := privKeyIface.(*rsa.PrivateKey)
	if !ok {
		return errors.New("Unable to get pkcs8 private key")
	}
	publicKeyBytes := x509.MarshalPKCS1PublicKey(&officialPrivateKey.PublicKey)
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)
	// so... how do we get the right public key format???
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, publicKeyStr)
	_, err = snowflake.ExecNoRows(db, query)

	return err
}

func Rotate(ctx context.Context) error {
	snowflakeDB, err := setup.Snowflake()
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake and databricks")
	}

	users, err := getUsers()
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	// Collect errors for each user:
	userErrors := &multierror.Error{}

	for _, user := range users {
		privKey, err := keypair.GenerateRSAKeypair()
		if err != nil {
			return errors.Wrap(err, "Unable to generate RSA keypair")
		}
		privKeyBuffer, err := keypair.SaveRSAKey(privKey)
		if err != nil {
			return errors.Wrap(err, "Unable to save RSA Keypair")
		}
		err = updateSnowflake(user, snowflakeDB, privKeyBuffer)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors
}
