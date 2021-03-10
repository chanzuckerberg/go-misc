package setup

import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/databricks"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
)

func Databricks() (databricksCfg.DatabricksConnection, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return databricksCfg.DatabricksConnection(dbClient), nil
}

func Snowflake(snowflakeAcct string) (*sql.DB, error) {
	if snowflakeAcct == "" {
		return nil, errors.New("snowflake account input cannot be empty")
	}
	snowflakeEnv, err := snowflakeCfg.LoadSnowflakeClientEnv()
	if err != nil {
		return nil, err
	}

	cfg := snowflake.SnowflakeConfig{
		Account:  snowflakeAcct,
		User:     snowflakeEnv.USER,
		Role:     snowflakeEnv.ROLE,
		Region:   snowflakeEnv.REGION,
		Password: snowflakeEnv.PASSWORD,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)

	return sqlDB, errors.Wrap(err, "Unable to configure Snowflake DB")
}
