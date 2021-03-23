package setup

import (
	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/kelseyhightower/envconfig"
)

func Databricks() (databricksCfg.DatabricksConnection, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Databricks information from environment variables")
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return databricksCfg.DatabricksConnection(dbClient), nil
}

func Snowflake() ([]*snowflakeCfg.SnowflakeAccount, error) {
	env := &snowflakeCfg.Accounts{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}
	acctList := env.ACCOUNTS
	return snowflakeCfg.LoadSnowflakeAccounts(acctList)
}
