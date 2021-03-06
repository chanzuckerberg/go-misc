package okta

import (
	"context"
	"net/http"
	"testing"

	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
	"github.com/stretchr/testify/require"
)

var testIndexedApplications = map[int][]*okta.AppUser{
	1: {
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user1",
			}},
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user2",
			}},
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user3",
			},
		},
	},
	2: {
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user4",
			},
		},
	},
	3: {
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user5",
			},
		},
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user6",
			},
		},
		&okta.AppUser{
			Credentials: &okta.AppUserCredentials{
				UserName: "user7",
			},
		},
	},
}

var testAppIndex = 1

func testGetterFunc(appID string, qp *query.Params) ([]*okta.AppUser, *okta.Response, error) {
	AppUsers, ok := testIndexedApplications[testAppIndex]
	if !ok {
		return nil, &okta.Response{}, nil
	}

	respWithNext := &okta.Response{
		&http.Response{
			Header: http.Header{
				"Link": []string{"<localhost?after=afterToken&limit=2>; rel=\"next\""},
			},
		},
	}
	testAppIndex = testAppIndex + 1

	return AppUsers, respWithNext, nil
}

func TestListUsersPagination(t *testing.T) {
	r := require.New(t)
	users, err := paginateListUsers(context.TODO(), "testAppID", testGetterFunc)
	r.NoError(err)
	r.Len(users.List(), 7)
}
