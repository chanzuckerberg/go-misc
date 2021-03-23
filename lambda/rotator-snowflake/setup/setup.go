package setup

import (
	"context"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/kelseyhightower/envconfig"
)

func Okta(ctx context.Context) (*oktaCfg.OktaClient, error) {
	return oktaCfg.GetOktaClient(ctx)
}

func Databricks(ctx context.Context) (databricksCfg.DatabricksConnection, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Databricks information from environment variables")
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return databricksCfg.DatabricksConnection(dbClient), nil
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
