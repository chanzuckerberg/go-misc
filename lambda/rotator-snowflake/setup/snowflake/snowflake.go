package snowflake

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/chamber/store"
	"github.com/sirupsen/logrus"
)

func configureConnection(env *SnowflakeClientEnv, secrets *store.SSMStore) (*sql.DB, error) {
	cfg := snowflake.SnowflakeConfig{
		Account: env.NAME,
		User:    env.USER,
		Role:    env.ROLE,
		Region:  env.REGION,
	}
	// Get the password through the account name
	service := env.PARAM_STORE_SERVICE
	tokenSecretID := store.SecretId{
		Service: service, // TODO(aku): Figure out how to feed this value in through environment variables
		Key:     fmt.Sprintf("%s_password", env.NAME),
	}

	password, err := secrets.Read(tokenSecretID, -1)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't find %s's password in AWS Parameter Store in service (%s)", env.NAME, service)
	}

	cfg.Password = *password.Value

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake SQL connection")
	}
	if sqlDB == nil {
		return nil, errors.Errorf("Unable to create db connection with the %s snowflake account", env.NAME)
	}

	return sqlDB, nil
}

func LoadSnowflakeAccounts(accountMap map[string]string, secrets *store.SSMStore) ([]*Account, error) {
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

		sqlDB, err := configureConnection(snowflakeEnv, secrets)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, err)

			continue
		}

		acctList = append(acctList, &Account{
			AppID: snowflakeAppID,
			Name:  snowflakeEnv.NAME,
			DB:    sqlDB,
		})
	}

	return acctList, snowflakeErrs.ErrorOrNil()
}
