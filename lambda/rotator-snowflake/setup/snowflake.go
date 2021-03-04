package setup

import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type SnowflakeClientEnvironment struct {
	ACCOUNT  string `required:"true"`
	REGION   string
	PASSWORD string `required:"true"`
	USER     string `required:"true"`
	ROLE     string `required:"true"`
}

func loadSnowflakeClientEnv() (*SnowflakeClientEnvironment, error) {
	env := &SnowflakeClientEnvironment{}
	err := envconfig.Process("SNOWFLAKE", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

func Snowflake(snowflakeAcct string) (*sql.DB, error) {
	snowflakeEnv, err := loadSnowflakeClientEnv()
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
