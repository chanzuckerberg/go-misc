package snowflake

import (
	"database/sql"
	"strings"

	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func configureConnection(env *SnowflakeClientEnv) (*sql.DB, error) {
	cfg := snowflake.Config{
		Account:  env.NAME,
		User:     env.USER,
		Role:     env.ROLE,
		Region:   env.REGION,
		Password: env.PASSWORD,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake SQL connection")
	}
	if sqlDB == nil {
		return nil, errors.Errorf("Unable to create db connection with the %s snowflake account", env.NAME)
	}

	return sqlDB, nil
}

func LoadSnowflakeAccounts(accountMap map[string]string) ([]*Account, error) {
	snowflakeErrs := &multierror.Error{}
	acctList := []*Account{}

	for acctName, snowflakeAppID := range accountMap {
		// If acctName has "okta" or "databricks" in the name, print a warning for possible name collision
		oktaCollision := strings.Contains(acctName, "okta")
		if oktaCollision {
			logrus.Warnf("Snowflake Account %s will likely collide with okta Environment Variables", acctName)
		}

		databricksCollision := strings.Contains(acctName, "databricks")
		if databricksCollision {
			logrus.Warnf("Snowflake Account %s will likely collide with databricks Environment Variables", acctName)
		}

		snowflakeEnv := &SnowflakeClientEnv{}

		err := envconfig.Process(acctName, snowflakeEnv)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, errors.Wrap(err, "Error processing Snowflake environment variables"))
		}

		sqlDB, err := configureConnection(snowflakeEnv)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, err)
		}

		acctList = append(acctList, &Account{
			AppID: snowflakeAppID,
			Name:  snowflakeEnv.NAME,
			DB:    sqlDB,
		})
	}

	return acctList, snowflakeErrs.ErrorOrNil()
}
