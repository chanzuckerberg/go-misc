package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
	"github.com/xinsnake/databricks-sdk-golang/aws/models"
)

type snowflakeUserCredentials struct {
	user            string
	role            string
	pem_private_key string
}

func (creds *snowflakeUserCredentials) writeSecrets(secretsClient *aws.SecretsAPI, currentScope string) error {
	if currentScope == "" {
		return errors.New("empty scope")
	}

	err := secretsClient.PutSecret([]byte(creds.user), currentScope, "snowflake.user")
	if err != nil {
		return errors.Wrapf(err, "Unable to put secret username %s in scope %s", creds.user, currentScope)
	}

	err = secretsClient.PutSecret([]byte(creds.role), currentScope, "snowflake.role")
	if err != nil {
		return errors.Wrapf(err, "Unable to put role %s in scope %s", creds.role, currentScope)
	}

	err = secretsClient.PutSecret([]byte(creds.pem_private_key), currentScope, "snowflake.pem_private_key")
	return errors.Wrapf(err, "Unable to put private key %s in scope %s", creds.pem_private_key, currentScope)
}

func buildSnowflakeSecrets(connection *sql.DB, username string, privKey string) (*snowflakeUserCredentials, error) {
	if username == "" {
		return nil, errors.New("Empty username. Snowflake secrets cannot be built")
	}

	userQuery := fmt.Sprintf(`SHOW USERS LIKE '%s'`, username)
	connectionRow := snowflake.QueryRow(connection, userQuery)
	snowflakeUser, err := snowflake.ScanUser(connectionRow)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read snowflake user from (%s)", userQuery)
	}
	defaultRole := snowflakeUser.DefaultRole.String
	if defaultRole == "" {
		defaultRole = "PUBLIC"
	}

	userSecrets := snowflakeUserCredentials{
		user:            username,
		role:            defaultRole,
		pem_private_key: privKey,
	}

	return &userSecrets, nil
}

func updateDatabricks(currentScope string, creds *snowflakeUserCredentials, databricks *aws.DBClient) error {
	secretsAPI := databricks.Secrets()

	scopes, err := secretsAPI.ListSecretScopes()
	if err != nil {
		return errors.Wrap(err, "Unable to list secret scopes")
	}

	// Check if scope exists under current name before
	for _, scope := range scopes {
		scopeName := scope.Name
		if scopeName == currentScope {
			return creds.writeSecrets(&secretsAPI, scopeName)
		}
	}

	// create scope called currentScope
	err = secretsAPI.CreateSecretScope(currentScope, "")
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s scope with %s perms", currentScope, "")
	}
	// Allow admins to manage this secret
	err = secretsAPI.PutSecretACL(currentScope, "admins", models.AclPermissionManage)
	if err != nil {
		return errors.Wrapf(err, "Unable to make admins control this scope: %s", currentScope)
	}
	// Allow user to read this secret
	err = secretsAPI.PutSecretACL(currentScope, creds.user, models.AclPermissionRead)
	if err != nil {
		return errors.Wrapf(err, "Unable to make %s control this scope: %s", creds.user, currentScope)
	}
	return creds.writeSecrets(&secretsAPI, currentScope)
}

func updateSnowflake(user string, db *sql.DB, pubKey string) error {
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, pubKey)
	_, err := snowflake.ExecNoRows(db, query)

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

	users, err := setup.GetUsers()
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

		privKeyStr, pubKeyStr, err := snowflake.RSAKeypairToString(privKey)

		err = updateSnowflake(user, snowflakeDB, pubKeyStr)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}

		snowflakeSecrets, err := buildSnowflakeSecrets(snowflakeDB, user, privKeyStr)
		if err != nil {
			return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
		}

		// Intentionally equating databricks scope and user here
		err = updateDatabricks(user, snowflakeSecrets, databricksConnection)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors.ErrorOrNil()
}
