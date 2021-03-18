package setup

import (
	"database/sql"
	"os"
	"strings"

	"github.com/chanzuckerberg/go-misc/databricks"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
)

// TODO: Grab from Okta
func GetUsers() ([]string, error) {
	usersList := os.Getenv("CURRENT_USERS") //Comma-delimited users list
	userSlice := strings.Split(usersList, ",")

	return userSlice, nil
}

func Databricks() (databricksCfg.DatabricksConnection, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return databricksCfg.DatabricksConnection(dbClient), nil
}

func Snowflake() (*sql.DB, error) {
	snowflakeEnv, err := snowflakeCfg.LoadSnowflakeClientEnv()
	if err != nil {
		return nil, err
	}

	cfg := snowflake.SnowflakeConfig{
		Account:  snowflakeEnv.ACCOUNT,
		User:     snowflakeEnv.USER,
		Role:     snowflakeEnv.ROLE,
		Region:   snowflakeEnv.REGION,
		Password: snowflakeEnv.PASSWORD,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)

	return sqlDB, errors.Wrap(err, "Unable to configure Snowflake DB")
}
