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

	oktaClient, err := setup.GetOktaClient(context.TODO())
	if err != nil {
		return err
	}

	// Get users from databricks okta app ID
	users, err := setup.GetOktaAppUsers(oktaClient.AppID, oktaClient.Client.Application.ListApplicationUsers)
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	// Get users from eachsnowflake okta app ID
	snowflake_app_ids := map[string]string{
		"acct1": "appID1",
		"acct2": "appID2",
	}

	processUser := func(user, snowflakeAcctName string) error {
		snowflakeDB, err := setup.Snowflake(snowflakeAcctName)
		if err != nil {
			return errors.Wrap(err, "Unable to configure snowflake")
		}

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
		// TODO: include snowflakeAcctName here rather than using snowflake_account environment variable
		snowflakeSecrets, err := buildSnowflakeSecrets(snowflakeDB, user, privKeyStr)
		if err != nil {
			return errors.Wrap(err, "Cannot generate Snowflake Secrets Map")
		}

		// Intentionally equating databricks scope and user here
		return updateDatabricks(user, snowflakeAcctName, snowflakeSecrets, databricksConnection.Secrets())
	}

	// // Collect errors for each user:
	userErrors := &multierror.Error{}
	processSnowflake := func(acctName, snowflakeAppID string) error {
		snowflakeUsers, err := setup.GetOktaAppUsers(snowflakeAppID, oktaClient.Client.Application.ListApplicationUsers)
		if err != nil {
			return errors.Wrap(err, "Unable to get list of users to rotate")
		}
		for _, user := range snowflakeUsers.List() {
			if users.ContainsElement(user) {
				userErrors = multierror.Append(userErrors, processUser(user, acctName))
				continue
			}
			logrus.Debugf("%s not in databricks app", user)
		}
		return nil
	}

	for acctName, snowflakeAppID := range snowflake_app_ids {
		err = processSnowflake(acctName, snowflakeAppID)
		if err != nil {
			userErrors = multierror.Append(userErrors, processSnowflake(acctName, snowflakeAppID))
		}
	}

	return userErrors.ErrorOrNil()

}
