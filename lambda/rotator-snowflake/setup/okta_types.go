package setup

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/pkg/errors"
)

type OktaClientEnvironment struct {
	PRIVATE_KEY       string `required:"true"`
	ORG_URL           string `required:"true"`
	CLIENT_ID         string `required:"true"`
	DATABRICKS_APP_ID string `required:"true"`
	SNOWFLAKE_APP_IDS []string
}

func loadOktaClientEnv() (*OktaClientEnvironment, error) {
	env := &OktaClientEnvironment{}
	err := envconfig.Process("OKTA", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

type OktaClient struct {
	Client          *okta.Client
	AppID           string
	SnowflakeAppIDs []string
}
