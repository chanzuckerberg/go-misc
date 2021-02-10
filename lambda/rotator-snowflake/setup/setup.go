package setup

import (
	"database/sql"
	"os"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func getSnowflakeAccount() (string, error) {
	return os.Getenv("SNOWFLAKE_ACCOUNT"), nil
}

func getSnowflakeRegion() (string, error) {
	return os.Getenv("SNOWFLAKE_REGION"), nil
}

func getSnowflakePassword() (string, error) {
	return os.Getenv("SNOWFLAKE_PASSWORD"), nil
}

func SetupSnowflake() (*sql.DB, error) {
	account, err := getSnowflakeAccount()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake account name")
	}

	region, err := getSnowflakeRegion()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake region")
	}

	password, err := getSnowflakePassword()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake password")
	}

	// Figure out what to fill in here:
	cfg := snowflake.SnowflakeConfig{
		Account:  account,
		Region:   region,
		Password: password,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake DB")
	}

	return sqlDB, nil
}

func SetupDatabricks() (*aws.DBClient, error) {

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
