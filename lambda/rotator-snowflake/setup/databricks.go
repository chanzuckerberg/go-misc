package setup

import (
	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	DBAWS "github.com/xinsnake/databricks-sdk-golang/aws"
	DBModels "github.com/xinsnake/databricks-sdk-golang/aws/models"
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

type SecretsIface interface {
	ListSecretScopes() ([]DBModels.SecretScope, error)
	PutSecretACL(scope string, principal string, permission DBModels.AclPermission) error
	CreateSecretScope(scope string, initialManagePrincipal string) error
	PutSecret(bytesValue []byte, scope string, key string) error
}

func Databricks() (DatabricksConnection, error) {
	databricksEnv, err := loadDatabricksClientEnv()
	if err != nil {
		return nil, err
	}

	dbClient := databricks.NewAWSClient(databricksEnv.HOST, databricksEnv.TOKEN)

	return DatabricksConnection(dbClient), nil
}
