package rotate

import (
	"fmt"

	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws/models"
)

type snowflakeUserCredentials struct {
	user            string
	role            string
	pem_private_key string
	accountName     string
}

func (creds *snowflakeUserCredentials) writeSecrets(secretsClient databricksCfg.SecretsIface, currentScope string) error {
	if currentScope == "" {
		return errors.New("empty scope")
	}

	err := secretsClient.PutSecret([]byte(creds.user), currentScope, fmt.Sprintf("snowflake.%s.user", creds.accountName))
	if err != nil {
		return errors.Wrapf(err, "Unable to put secret username %s in scope %s", creds.user, currentScope)
	}

	err = secretsClient.PutSecret([]byte(creds.role), currentScope, fmt.Sprintf("snowflake.%s.role", creds.accountName))
	if err != nil {
		return errors.Wrapf(err, "Unable to put role %s in scope %s", creds.role, currentScope)
	}

	err = secretsClient.PutSecret([]byte(creds.pem_private_key), currentScope, fmt.Sprintf("snowflake.%s.pem_private_key", creds.accountName))
	return errors.Wrapf(err, "Unable to put private key in scope %s", currentScope)
}

func updateDatabricks(currentScope string, creds *snowflakeUserCredentials, secretsClient databricksCfg.SecretsIface) error {

	scopes, err := secretsClient.ListSecretScopes()
	if err != nil {
		return errors.Wrap(err, "Unable to list secret scopes")
	}

	// Check if scope exists under current name before
	scopeExists := false
	for _, scope := range scopes {
		scopeName := scope.Name
		if scopeName == currentScope {
			scopeExists = true
			break
		}
	}

	if !scopeExists {
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
	}

	return creds.writeSecrets(secretsClient, currentScope)
}
