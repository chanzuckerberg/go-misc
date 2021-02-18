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
	"strings"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
	"github.com/xinsnake/databricks-sdk-golang/aws/models"
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
	defaultRole := snowflakeUser.DefaultRole.String
	if defaultRole == "" {
		defaultRole = "PUBLIC"
	}
	privateKeyStr := privateKey.String()
	stripHeaders := strings.ReplaceAll(privateKeyStr, "-----BEGIN RSA PRIVATE KEY-----\n", "")
	stripFooters := strings.ReplaceAll(stripHeaders, "\n-----END RSA PRIVATE KEY-----", "")
	keyNoWhitespace := strings.TrimSpace(stripFooters)
	userSecrets := map[string]string{
		"snowflake.user":            username,
		"snowflake.role":            snowflakeUser.DefaultRole.String,
		"snowflake.pem_private_key": keyNoWhitespace,
	}

	return userSecrets, nil
}

func updateDatabricks(currentScope string, secrets map[string]string, databricks *aws.DBClient) error {
	secretsAPI := databricks.Secrets()

	scopes, err := databricks.Secrets().ListSecretScopes()
	if err != nil {
		return errors.Wrap(err, "Unable to list secret scopes")
	}

	// Check if scope exists under current name before
	for _, scope := range scopes {
		scopeName := scope.Name
		if scopeName == currentScope {
			for key, secret := range secrets {
				if secret == "" {
					return errors.Wrapf(err, "key %s has an empty secret", key)
				}
				err = secretsAPI.PutSecret([]byte(secret), currentScope, key)
				if err != nil {
					return errors.Wrapf(err, "Unable to put secret %s in scope %s", secret, scopeName)
				}
			}

			return nil
		}
	}

	// create scope called currentScope
	err = databricks.Secrets().CreateSecretScope(currentScope, "")
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s scope with %s perms", currentScope, "")
	}
	// Allow admins to manage this secret
	err = databricks.Secrets().PutSecretACL(currentScope, "admins", models.AclPermissionManage)
	if err != nil {
		return errors.Wrapf(err, "Unable to make admins control this scope: %s", currentScope)
	}

	return updateDatabricks(currentScope, secrets, databricks)
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
		return errors.Wrap(err, "Unable to configure snowflake")
	}

	databricksConnection, err := setup.Databricks()
	if err != nil {
		return errors.Wrap(err, "Unable to configure databricks")
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

		snowflakeSecrets, err := buildSnowflakeSecrets(snowflakeDB, user, privKeyBuffer)
		if err != nil {
			return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
		}

		// Intentionally equating databricks scope and user here
		err = updateDatabricks(user, snowflakeSecrets, databricksConnection)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors
}
