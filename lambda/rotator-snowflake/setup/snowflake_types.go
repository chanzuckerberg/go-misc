package setup

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type SnowflakeClientEnvironment struct {
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
