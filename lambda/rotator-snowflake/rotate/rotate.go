package rotate

import (
	"context"
	"crypto/rsa"
	"database/sql"
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

func updateSnowflake(user string, db *sql.DB, pubKey *rsa.PublicKey) error {
	// Convert publicKey to []bytes

	// See if the bytes.Buffer approach solves everything after we convert everything to PKCS#8 format
	// publicKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	// // add a pem encoding step
	// publicKeyBlock := &pem.Block{
	// 	Type:  "PUBLIC KEY",
	// 	Bytes: publicKeyBytes,
	// }
	// pemBytes := pem.EncodeToMemory(publicKeyBlock)
	// publicKeyStr := base64.StdEncoding.EncodeToString(pemBytes)
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, publicKeyStr)
	_, err := snowflake.ExecNoRows(db, query)

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

		err = updateSnowflake(user, snowflakeDB, &privKey.PublicKey)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors
}
