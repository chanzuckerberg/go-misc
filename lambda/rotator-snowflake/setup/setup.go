package setup

import (
	"database/sql"
	"os"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func getSnowflakeAccount() (string, error) {
	acct, present := os.LookupEnv("SNOWFLAKE_ACCOUNT")
	if !present {
		return "", errors.New("Could not find SNOWFLAKE_ACCOUNT")
	}
	return acct, nil
}

func getSnowflakeRegion() (string, error) {
	region, present := os.LookupEnv("SNOWFLAKE_REGION")
	if !present {
		return "", errors.New("Could not find SNOWFLAKE_REGION")
	}
	return region, nil
}

func getSnowflakePassword() (string, error) {
	password, present := os.LookupEnv("SNOWFLAKE_PASSWORD")
	if !present {
		return "", errors.New("Could not find SNOWFLAKE_PASSWORD")
	}
	return password, nil
}

func getSnowflakeRole() (string, error) {
	role, present := os.LookupEnv("SNOWFLAKE_ROLE")
	if !present {
		return "", errors.New("Could not find SNOWFLAKE_ROLE")
	}
	return role, nil
}

func getSnowflakeUser() (string, error) {
	user, present := os.LookupEnv("SNOWFLAKE_USER")
	if !present {
		return "", errors.New("Could not find SNOWFLAKE_USER")
	}
	return user, nil
}

func Snowflake() (*sql.DB, error) {
	account, err := getSnowflakeAccount()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake account name")
	}

	region, err := getSnowflakeRegion()
	if err != nil {
		// Region should default to
		logrus.Debug("Snowflake Region not set")
	}

	password, err := getSnowflakePassword()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake password")
	}

	user, err := getSnowflakeUser()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake password")
	}

	role, err := getSnowflakeRole()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get snowflake password")
	}
	// Figure out what to fill in here:
	cfg := snowflake.SnowflakeConfig{
		Account:     account,
		User:        user,
		Role:        role,
		BrowserAuth: false,
		Region:      region,
		Password:    password,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake DB")
	}

	return sqlDB, nil
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
