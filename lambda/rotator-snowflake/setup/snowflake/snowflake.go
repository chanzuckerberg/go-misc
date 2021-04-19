package snowflake

import (
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/kelseyhightower/envconfig"
)

func LoadSnowflakeEnv(acctName string) (*SnowflakeClientEnv, error) {
	envVarPrefix := fmt.Sprintf("snowflake_%s", acctName)

	snowflakeEnv := &SnowflakeClientEnv{}

	err := envconfig.Process(envVarPrefix, snowflakeEnv)
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
