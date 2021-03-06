package okta

import (
	"context"
	"net/url"

	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
	"github.com/peterhellberg/link"
	"github.com/pkg/errors"
)

func GetOktaAppUsers(
	ctx context.Context,
	appID string,
	getter func(string, *query.Params) ([]*okta.AppUser, *okta.Response, error),
) (*sets.StringSet, error) {
	return paginateListUsers(ctx, appID, getter)
}

func paginateListUsers(
	ctx context.Context,
	appID string,
	getter func(string, *query.Params) ([]*okta.AppUser, *okta.Response, error),
) (*sets.StringSet, error) {
	qp := query.NewQueryParams()
	assignedUserEmails := sets.NewStringSet()

	for {
		users, resp, err := getter(appID, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to list users in okta app %s", appID)
		}

		for _, user := range users {
			assignedUserEmails.Add(user.Credentials.UserName)
		}

		if resp == nil {
			return nil, errors.New("Nil response from okta client")
		}

		links := link.ParseResponse(resp.Response)
		if links["next"] == nil {
			return assignedUserEmails, nil
		}

		nextLink := links["next"].String()

		nextLinkURL, err := url.Parse(nextLink)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing Link Header for the next page")
		}

		nextLinkMapping := nextLinkURL.Query()
		qp.After = nextLinkMapping.Get("after")
	}
}
