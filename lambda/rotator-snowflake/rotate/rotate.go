package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/hashicorp/go-multierror"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func updateSnowflake(user string, db *sql.DB, pubKey string) error {
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, pubKey)
	_, err := snowflake.ExecNoRows(db, query)

	return err
}

func Rotate(ctx context.Context) error {
	databricksConnection, err := setup.Databricks()
	if err != nil {
		return errors.Wrap(err, "Unable to configure databricks")
	}

	oktaClient, err := oktaCfg.GetOktaClient(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to configure okta")
	}

	// Get users from databricks okta app ID
	databricksUsers, err := oktaCfg.GetOktaAppUsers(oktaClient.AppID, oktaClient.Client.Application.ListApplicationUsers)
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	// Use environment variables to get []SnowflakeAccount
	snowflakeApps, err := snowflakeCfg.LoadSnowflakeAccounts()
	if err != nil {
		return errors.Wrap(err, "Unable to get Snowflake Account information")
	}

	processUser := func(user, snowflakeAcctName string, snowflakeDB *sql.DB) error {
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

		snowflakeSecrets, err := buildSnowflakeSecrets(snowflakeDB, user, privKeyStr)
		if err != nil {
			return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
		}

		// Intentionally equating databricks scope and user here
		return updateDatabricks(user, snowflakeAcctName, snowflakeSecrets, databricksConnection.Secrets())
	}

	// // Collect errors for each user:
	userErrors := &multierror.Error{}
	processSnowflake := func(snowflakeAcct *snowflakeCfg.SnowflakeAccount, snowflakeDB *sql.DB) error {
		snowflakeUsers, err := oktaCfg.GetOktaAppUsers(snowflakeAcct.AppID, oktaClient.Client.Application.ListApplicationUsers)
		if err != nil {
			return errors.Wrap(err, "Unable to get list of users to rotate")
		}

		for _, user := range snowflakeUsers.List() {

			if databricksUsers.ContainsElement(user) {
				err = processUser(user, snowflakeAcct.Name, snowflakeDB)
				userErrors = multierror.Append(userErrors, err)
				continue
			}
			logrus.Debugf("%s not in databricks app", user)
		}

		return nil
	}

	for _, snowflakeApp := range snowflakeApps {
		snowflakeDB, err := setup.Snowflake(snowflakeApp.Name)
		if err != nil {
			return errors.Wrap(err, "Unable to configure snowflake")
		}

		err = processSnowflake(snowflakeApp, snowflakeDB)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors.ErrorOrNil()
}
