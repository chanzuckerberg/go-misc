package setup

import (
	"context"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/pkg/errors"
)

type OktaClientEnvironment struct {
	PRIVATE_KEY       string `required:"true"`
	ISSUER_URL        string // required?
	SERVICE_CLIENT_ID string `required:"true"`
}

func loadOktaClientEnv() (*OktaClientEnvironment, error) {
	env := &OktaClientEnvironment{}
	err := envconfig.Process("OKTA", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

// TODO: Grab from Okta
func GetUsers() ([]string, error) {

	oktaEnv, err := loadOktaClientEnv()
	if err != nil {
		return []string{}, err
	}
	spew.Dump(oktaEnv)

	privateKeyNoQuotes := strings.ReplaceAll(oktaEnv.PRIVATE_KEY, `"`, ``)

	// TODO: see if this can be part of a go-misc package
	client, err := okta.NewClient(
		context.TODO(),
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(oktaEnv.SERVICE_CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(privateKeyNoQuotes),
		okta.WithOrgUrl(oktaEnv.ISSUER_URL),
		okta.WithCache(true),
	)
	if err != nil {
		return []string{}, errors.Wrap(err, "Unable to create Okta Client Connection")
	}

	_, resp, err := client.User.ListUsers(nil)
	if err != nil {
		fmt.Println(resp.Status)
		return []string{}, errors.Wrap(err, "Unable to list users in okta app")
	}

	return []string{}, nil
}
