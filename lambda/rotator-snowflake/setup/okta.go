package setup

import (
	"context"
	"strings"

	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
	"github.com/pkg/errors"
)

func GetOktaClient(ctx context.Context) (*OktaClient, error) {
	env, err := loadOktaClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to load right Okta env variables")
	}

	privKeyNoQuotes := strings.ReplaceAll(env.PRIVATE_KEY, `"`, ``)
	client, err := okta.NewClient(
		context.TODO(),
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(env.CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(privKeyNoQuotes),
		okta.WithOrgUrl(env.ORG_URL),
		okta.WithCache(true),
	)

	return &OktaClient{
			Client:          client,
			AppID:           env.DATABRICKS_APP_ID,
			SnowflakeAppIDs: env.SNOWFLAKE_APP_IDS},
		errors.Wrap(err, "Unable to configure Okta client")
}

func GetOktaAppUsers(
	appID string,
	getter func(string, *query.Params) ([]*okta.AppUser, *okta.Response, error), // HACK: probably better to use an iface
) (*sets.StringSet, error) {
	users, _, err := getter(appID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list users in okta app %s", appID)
	}
	assignedUserEmails := sets.NewStringSet()
	for _, user := range users {
		assignedUserEmails.Add(user.Credentials.UserName) // TODO: not sure if that is the right property, verify
	}

	return assignedUserEmails, nil
}
