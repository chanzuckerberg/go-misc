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
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func getUsers() ([]string, error) {
	// TODO(aku): get list from okta
	return []string{os.Getenv("CURRENT_USER")}, nil
}

func buildSnowflakeSecrets(connection *sql.DB, snowflake_user string, privateKey *bytes.Buffer) (map[string]string, error) {
	userQuery := fmt.Sprintf(`DESCRIBE USER "%s"`, snowflake_user)
	// TODO: write a snowflake.ExecWithRows()
	connectionProperties, err := snowflake.QueryRow(connection, userQuery)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to execute ")
	}

	default_role, ok := connectionProperties["DEFAULT_ROLE"]
	if !ok {
		return nil, errors.Errorf("Could not get user's DEFAULT_ROLE from Snowflake query: %s", userQuery)
	}

	userSecrets := map[string]string{
		"snowflake.user":            snowflake_user,
		"snowflake.role":            default_role,
		"snowflake.pem_private_key": base64.StdEncoding.EncodeToString(privateKey.Bytes()),
	}
	return userSecrets, nil
}

func updateDatabricks(user, scope string, databricks *aws.DBClient, privKeyBuffer *bytes.Buffer) error {
	secretsAPI := databricks.Secrets()
	secretsAPI.PutSecret(privKeyBuffer.Bytes(), scope, "RSA_PUBLIC_KEY_2")
	return nil
}

func updateSnowflake(user string, db *sql.DB, privKeyBuffer *bytes.Buffer) error {
	privPemBlock, _ := pem.Decode(privKeyBuffer.Bytes())

	privKeyIface, err := x509.ParsePKCS8PrivateKey(privPemBlock.Bytes)
	if err != nil {
		return errors.Wrap(err, "Unable to get pkcs8 private key")
	}

	privKey, ok := privKeyIface.(*rsa.PrivateKey)
	if !ok {
		return errors.New("Unable to get pkcs8 private key")
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal public key to bytes")
	}

	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, publicKeyStr)
	_, err = snowflake.ExecNoRows(db, query)

	return err
}

func Rotate(ctx context.Context) error {
	// snowflakeDB, err := setup.Snowflake()
	// if err != nil {
	// 	return errors.Wrap(err, "Unable to configure snowflake and databricks")
	// }
	databricksConnection, err := setup.Databricks()
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
		// err = updateSnowflake(user, snowflakeDB, privKeyBuffer)
		// if err != nil {
		// 	userErrors = multierror.Append(userErrors, err)
		// }

		err = updateDatabricks(user, user, databricksConnection, privKeyBuffer)
	}

	return userErrors
}
