package setup

import (
	"context"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/kelseyhightower/envconfig"
)

func Okta(ctx context.Context) (*oktaCfg.OktaClient, error) {
	return oktaCfg.GetOktaClient(ctx)
}

func Databricks(ctx context.Context) (*databricksCfg.DatabricksAccount, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Databricks information from environment variables")
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)
	databricksAccount := &databricksCfg.DatabricksAccount{
		AppID:  databricksEnv.APP_ID,
		Client: dbClient,
	}
	return databricksAccount, nil
}

func Snowflake(ctx context.Context) ([]*snowflakeCfg.SnowflakeAccount, error) {
	env := &snowflakeCfg.Accounts{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}
	acctList := env.ACCOUNTS
	return snowflakeCfg.LoadSnowflakeAccounts(acctList)
}

func ListSnowflakeUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, snowflakeAcct *snowflakeCfg.SnowflakeAccount) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(snowflakeAcct.AppID, userGetter)
}

func ListDatabricksUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, databricksAccount *databricksCfg.DatabricksAccount) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(databricksAccount.AppID, userGetter)
}
