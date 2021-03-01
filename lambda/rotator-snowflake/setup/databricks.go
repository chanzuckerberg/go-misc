package setup

import (
	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	DBAWS "github.com/xinsnake/databricks-sdk-golang/aws"
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

type DatabricksConnection interface {
	Secrets() DBAWS.SecretsAPI
}

func Databricks() (DatabricksConnection, error) {
	databricksEnv, err := loadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return DatabricksConnection(dbClient), nil
}
