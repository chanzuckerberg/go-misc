package rotate

import (
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
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
