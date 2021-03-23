package okta

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
		ctx,
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(env.CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(privKeyNoQuotes),
		okta.WithOrgUrl(env.ORG_URL),
		okta.WithCache(true),
	)

	return &OktaClient{
			Client: client,
			AppID:  env.DATABRICKS_APP_ID,
		},
		errors.Wrap(err, "Unable to configure Okta client")
}

func GetOktaAppUsers(
	appID string,
	getter func(string, *query.Params) ([]*okta.AppUser, *okta.Response, error), // HACK: probably better to use an iface
) (*sets.StringSet, error) {
	// TODO(aku): implement pagination steps
	users, _, err := getter(appID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list users in okta app %s", appID)
	}
	assignedUserEmails := sets.NewStringSet()
	for _, user := range users {
		assignedUserEmails.Add(user.Credentials.UserName)
	}

	return assignedUserEmails, nil
}
