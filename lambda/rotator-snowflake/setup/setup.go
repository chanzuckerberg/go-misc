package setup

import (
	"database/sql"
	"os"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func Snowflake() (*sql.DB, error) {
	account, present := os.LookupEnv("SNOWFLAKE_ACCOUNT")
	if !present {
		return nil, errors.New("Could not find SNOWFLAKE_ACCOUNT")
	}

	region, present := os.LookupEnv("SNOWFLAKE_REGION")
	if !present {
		return nil, errors.New("Could not find SNOWFLAKE_REGION")
	}

	password, present := os.LookupEnv("SNOWFLAKE_PASSWORD")
	if !present {
		return nil, errors.New("Could not find SNOWFLAKE_PASSWORD")
	}

	user, present := os.LookupEnv("SNOWFLAKE_USER")
	if !present {
		return nil, errors.New("Could not find SNOWFLAKE_USER")
	}

	role, present := os.LookupEnv("SNOWFLAKE_ROLE")
	if !present {
		return nil, errors.New("Could not find SNOWFLAKE_ROLE")
	}

	// Figure out what to fill in here:
	cfg := snowflake.SnowflakeConfig{
		Account:     account,
		User:        user,
		Role:        role,
		Region:      region,
		Password:    password,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	return sqlDB, errors.Wrap(err, "Unable to configure Snowflake DB")
}

func Databricks() (*aws.DBClient, error) {
	host, present := os.LookupEnv("DATABRICKS_HOST")
	if !present {
		return nil, errors.New("We can't find the DATABRICKS_HOST")
	}
	token, present := os.LookupEnv("DATABRICKS_TOKEN")
	if !present {
		return nil, errors.New("We can't find the DATABRICKS_HOST")
	}

	dbClient := databricks.NewAWSClient(host, token)

	return dbClient, nil
}
