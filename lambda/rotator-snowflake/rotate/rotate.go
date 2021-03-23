package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
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
	databricksAccount, err := setup.Databricks(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to configure databricks")
	}

	oktaClient, err := setup.Okta(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to configure okta")
	}

	databricksUsers, err := setup.ListDatabricksUsers(ctx, oktaClient, databricksAccount)
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	snowflakeApps, err := setup.Snowflake(ctx)
	if err != nil {
		return err
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
		return updateDatabricks(user, snowflakeAcctName, snowflakeSecrets, databricksAccount.Client.Secrets())
	}

	// // Collect errors for each user:
	userErrors := &multierror.Error{}
	processSnowflake := func(snowflakeAcct *snowflakeCfg.SnowflakeAccount) error {
		snowflakeUsers, err := setup.ListSnowflakeUsers(ctx, oktaClient, snowflakeAcct)
		if err != nil {
			return errors.Wrap(err, "Unable to get list of users to rotate")
		}

		for _, user := range snowflakeUsers.List() {
			if databricksUsers.ContainsElement(user) {
				err = processUser(user, snowflakeAcct.Name, snowflakeAcct.DB)
				userErrors = multierror.Append(userErrors, err)

				continue
			}
			logrus.Debugf("%s not in databricks app", user)
		}

		return nil
	}

	for _, snowflakeApp := range snowflakeApps {
		err = processSnowflake(snowflakeApp)
		if err != nil {
			userErrors = multierror.Append(userErrors, err)
		}
	}

	return userErrors.ErrorOrNil()
}
