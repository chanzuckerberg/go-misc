package setup

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
	"github.com/pkg/errors"
)

type OktaClientEnvironment struct {
	PRIVATE_KEY string `required:"true"`
	ORG_URL     string `required:"true"`
	CLIENT_ID   string `required:"true"`
}

func loadOktaClientEnv() (*OktaClientEnvironment, error) {
	env := &OktaClientEnvironment{}
	err := envconfig.Process("OKTA", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

type OktaClient struct {
	client *okta.Client
}

func GetOktaClient(ctx context.Context, env *OktaClientEnvironment) (*OktaClient, error) {
	client, err := okta.NewClient(
		context.TODO(),
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(env.CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(env.PRIVATE_KEY),
		okta.WithOrgUrl(env.ORG_URL),
		okta.WithCache(true),
	)

	return &OktaClient{client}, err

	// if err != nil {
	// return []string{}, errors.Wrap(err, "Unable to create Okta Client Connection")
	// }
}

// TODO: Grab from Okta
func GetUsersForAPP(
	appID string,
	getter func(string, *query.Params) ([]*okta.AppUser, *okta.Response, error), // HACK: probably better to use an iface
) ([]string, error) {
	users, _, err := getter("TODO app id", nil)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to list users in okta app %s", "TODO app id")
	}

	assignedUserEmails := []string{} // This might be better as a sets.Strings instead (so you can do intersections)
	for _, user := range users {
		assignedUserEmails = append(assignedUserEmails, user.Credentials.UserName) // TODO: not sure if that is the right property, verify
	}

	return assignedUserEmails, nil
}
