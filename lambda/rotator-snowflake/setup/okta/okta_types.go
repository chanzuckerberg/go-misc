package okta

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/pkg/errors"
)

type OktaClientEnvironment struct {
	ORG_URL             string `required:"true"`
	CLIENT_ID           string `required:"true"`
	PARAM_STORE_SERVICE string `required:"true"`
}

func LoadOktaClientEnv() (*OktaClientEnvironment, error) {
	env := &OktaClientEnvironment{}
	err := envconfig.Process("OKTA", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

type OktaClient struct {
	Client *okta.Client
	AppID  string
}
