package snowflake

import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func LoadSnowflakeClientEnv() (*SnowflakeClientEnvironment, error) {
	env := &SnowflakeClientEnvironment{}
	err := envconfig.Process("SNOWFLAKE", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

func Snowflake() (*sql.DB, error) {
	snowflakeEnv, err := LoadSnowflakeClientEnv()
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
