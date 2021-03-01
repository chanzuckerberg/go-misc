package setup

import (
	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

type DatabricksClientEnvironment struct {
	HOST  string
	TOKEN string
}

func loadDatabricksClientEnv() (*DatabricksClientEnvironment, error) {
	env := &DatabricksClientEnvironment{}
	err := envconfig.Process("DATABRICKS", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

func Databricks() (*aws.DBClient, error) {
	databricksEnv, err := loadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return dbClient, nil
}
