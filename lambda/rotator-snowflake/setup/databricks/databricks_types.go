package databricks

import (
	DBAWS "github.com/xinsnake/databricks-sdk-golang/aws"
	DBModels "github.com/xinsnake/databricks-sdk-golang/aws/models"
)

type DatabricksClientEnvironment struct {
	HOST   string `required:"true"`
	TOKEN  string `required:"true"`
	APP_ID string `required:"true"`
}

type DatabricksAccount struct {
	AppID  string
	Client *DBAWS.DBClient
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
