package databricks

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func LoadDatabricksClientEnv() (*DatabricksClientEnvironment, error) {
	env := &DatabricksClientEnvironment{}
	err := envconfig.Process("DATABRICKS", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}
