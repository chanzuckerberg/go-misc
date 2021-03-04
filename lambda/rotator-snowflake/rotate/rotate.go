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

type snowflakeAccount struct {
	appID string
	name  string
}

func getSnowflakeApps(oktaClient *setup.OktaClient, snowflakeAppIDs []string) ([]*snowflakeAccount, error) {
	accounts := []*snowflakeAccount{}
	for _, appID := range snowflakeAppIDs {
		accountName, err := setup.GetOktaAppAccount(oktaClient.AppID, oktaClient.Client.Application.GetApplicationKey) // TODO: Figure out what app name to "get"
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to get appID for appID %s", appID)
		}
		accounts = append(accounts, &snowflakeAccount{
			appID: appID,
			name:  accountName,
		})
	}
	return accounts, nil
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
	snowflake_apps, err := getSnowflakeApps(oktaClient, []string{})
	if err != nil {
		return errors.Wrap(err, "Unable to map snowflake appIDs with account names")
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
				err = processUser(user, acctName)
				userErrors = multierror.Append(userErrors, err)
				continue
			}
			logrus.Debugf("%s not in databricks app", user)
		}
		return nil
	}

	for _, snowflakeApp := range snowflake_apps {
		err = processSnowflake(snowflakeApp.name, snowflakeApp.appID)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors.ErrorOrNil()

}
