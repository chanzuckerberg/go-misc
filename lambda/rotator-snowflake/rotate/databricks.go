package rotate

import (
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws/models"
)

type snowflakeUserCredentials struct {
	user            string
	role            string
	pem_private_key string
}

func (creds *snowflakeUserCredentials) writeSecrets(secretsClient setup.SecretsIface, currentScope string) error {
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
	if connectionRow == nil {
		return nil, errors.New("Couldn't get a row output from snowflake")
	}
	snowflakeUser, err := snowflake.ScanUser(connectionRow)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read snowflake user from (%s)", userQuery)
	}
	if snowflakeUser == nil {
		return nil, errors.New("Could not create snowflake User profile")
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

func updateDatabricks(currentScope string, creds *snowflakeUserCredentials, secretsClient setup.SecretsIface) error {

	scopes, err := secretsClient.ListSecretScopes()
	if err != nil {
		return errors.Wrap(err, "Unable to list secret scopes")
	}

	// Check if scope exists under current name before
	for _, scope := range scopes {
		scopeName := scope.Name
		if scopeName == currentScope {
			return creds.writeSecrets(secretsClient, scopeName)
		}
	}

	// create scope called currentScope
	err = secretsClient.CreateSecretScope(currentScope, "")
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s scope with %s perms", currentScope, "")
	}
	// Allow admins to manage this secret
	err = secretsClient.PutSecretACL(currentScope, "admins", models.AclPermissionManage)
	if err != nil {
		return errors.Wrapf(err, "Unable to make admins control this scope: %s", currentScope)
	}
	// Allow user to read this secret
	err = secretsClient.PutSecretACL(currentScope, creds.user, models.AclPermissionRead)
	if err != nil {
		return errors.Wrapf(err, "Unable to make %s control this scope: %s", creds.user, currentScope)
	}
	return creds.writeSecrets(secretsClient, currentScope)
}
