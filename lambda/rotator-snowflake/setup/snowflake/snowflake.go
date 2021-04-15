package snowflake

import (
	"database/sql"
	"strings"

	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func LoadSnowflakeEnv(acctName string) (*SnowflakeClientEnv, error) {
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
		return nil, errors.Wrapf(err, "Unable to get right env variables for %s snowflake account", acctName)
	}

	return snowflakeEnv, nil
}

func ConfigureConnection(cfg snowflake.Config) (*sql.DB, error) {
	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake SQL connection")
	}
	if sqlDB == nil {
		return nil, errors.Errorf("Unable to create db connection with the %s snowflake account", cfg.Account)
	}

	return sqlDB, nil
}
