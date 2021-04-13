package rotate

import (
	"fmt"

	"github.com/chanzuckerberg/go-misc/keypair"
	databricksConfig "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	snowflakeConfig "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
)

func buildSnowflakeSecrets(snowflakeAccount *snowflakeConfig.Account, username string, privKey string) (*snowflakeUserCredentials, error) {
	if username == "" {
		return nil, errors.New("Empty username. Snowflake secrets cannot be built")
	}
	snowflakeDB := snowflakeAccount.DB

	userQuery := fmt.Sprintf(`SHOW USERS LIKE '%s'`, username)
	connectionRow := snowflake.QueryRow(snowflakeDB, userQuery)
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
		accountName:     snowflakeAccount.Name,
	}

	return &userSecrets, nil
}

func ProcessUser(user string, snowflakeAccount *snowflakeConfig.Account, databricksAccount *databricksConfig.Account) error {

	privKey, err := keypair.GenerateRSAKeypair()
	if err != nil {
		return errors.Wrap(err, "Unable to generate RSA keypair")
	}

	privKeyStr, pubKeyStr, err := snowflake.RSAKeypairToString(privKey)
	if err != nil {
		return errors.Wrap(err, "Unable to format new keypair for snowflake and databricks")
	}

	snowflakeDB := snowflakeAccount.DB
	err = updateSnowflake(user, snowflakeDB, pubKeyStr)
	if err != nil {
		return err
	}

	formattedSecrets, err := buildSnowflakeSecrets(snowflakeAccount, user, privKeyStr)
	if err != nil {
		return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
	}

	databricksSecretsAPI := databricksAccount.Client.Secrets()
	databricksScope := user // Intentionally equating databricks scope and user here
	return updateDatabricks(databricksScope, formattedSecrets, databricksSecretsAPI)
}
