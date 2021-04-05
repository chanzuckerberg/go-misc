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

func Databricks(ctx context.Context) (*databricksCfg.Account, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Databricks information from environment variables")
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)
	databricksAccount := &databricksCfg.Account{
		AppID:  databricksEnv.APP_ID,
		Client: dbClient,
	}
	return databricksAccount, nil
}

func Snowflake(ctx context.Context) ([]*snowflakeCfg.Account, error) {
	env := &snowflakeCfg.Accounts{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}
	acctMapping := env.OKTAMAP
	return snowflakeCfg.LoadSnowflakeAccounts(acctMapping)
}

func ListSnowflakeUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, snowflakeAcct *snowflakeCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, snowflakeAcct.AppID, userGetter)
}

func ListDatabricksUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, databricksAccount *databricksCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, databricksAccount.AppID, userGetter)
}
