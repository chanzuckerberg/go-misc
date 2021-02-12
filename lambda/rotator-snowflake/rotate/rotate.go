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

func buildSnowflakeSecrets(connection *sql.DB, username string, privateKey *bytes.Buffer) (map[string]string, error) {
	userQuery := fmt.Sprintf(`SHOW USERS LIKE '%s'`, username)

	connectionRow := snowflake.QueryRow(connection, userQuery)

	snowflakeUser, err := snowflake.ScanUser(connectionRow)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create snowflake user from userQuery %s", userQuery)
	}

	userSecrets := map[string]string{
		"snowflake.user":            username,
		"snowflake.role":            snowflakeUser.DefaultRole.String,
		"snowflake.pem_private_key": base64.StdEncoding.EncodeToString(privateKey.Bytes()),
	}

	return userSecrets, nil
}

func updateDatabricks(scope string, snowflake *sql.DB, databricks *aws.DBClient, privKeyBuffer *bytes.Buffer) error {
	secretsAPI := databricks.Secrets()

	secrets, err := buildSnowflakeSecrets(snowflake, scope, privKeyBuffer)
	if err != nil {
		return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
	}

	// TODO: write logic for when we don't have the scope yet
	for key, secret := range secrets {
		err = secretsAPI.PutSecret([]byte(secret), scope, key)
		if err != nil {
			// TODO: create a scope if we get RESOURCE_DOES_NOT_EXIST error
			return errors.Wrapf(err, "Unable to put secret %s in scope %s", secret, scope)
		}
	}

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
	snowflakeDB, err := setup.Snowflake()
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake and databricks")
	}

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

		err = updateSnowflake(user, snowflakeDB, privKeyBuffer)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}

		scope := user

		err = updateDatabricks(scope, snowflakeDB, databricksConnection, privKeyBuffer)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}
	fmt.Println(userErrors)
	return userErrors
}
