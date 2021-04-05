package rotate

import (
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/keypair"
	databricksConfig "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	snowflakeConfig "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
)

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

func ProcessUser(user string, snowflakeAccount *snowflakeConfig.Account, databricksAccount *databricksConfig.Account) error {
	snowflakeDB := snowflakeAccount.DB
	snowflakeAcctName := snowflakeAccount.Name
	databricksSecretsAPI := databricksAccount.Client.Secrets()

	privKey, err := keypair.GenerateRSAKeypair()
	if err != nil {
		return errors.Wrap(err, "Unable to generate RSA keypair")
	}

	privKeyStr, pubKeyStr, err := snowflake.RSAKeypairToString(privKey)
	if err != nil {
		return errors.Wrap(err, "Unable to format new keypair for snowflake and databricks")
	}

	err = updateSnowflake(user, snowflakeDB, pubKeyStr)
	if err != nil {
		return err
	}

	formattedSecrets, err := buildSnowflakeSecrets(snowflakeDB, user, privKeyStr)
	if err != nil {
		return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
	}

	// Intentionally equating databricks scope and user here
	databricksScope := user
	return updateDatabricks(databricksScope, snowflakeAcctName, formattedSecrets, databricksSecretsAPI)
}
