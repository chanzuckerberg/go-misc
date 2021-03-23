package okta

import (
	"context"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
	"github.com/stretchr/testify/require"
)

func TestGetOktaClient(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())

	os.Setenv("OKTA_PRIVATE_KEY", "testPrivKey")
	os.Setenv("OKTA_ORG_URL", "https://www.testOrgURL.com")
	os.Setenv("OKTA_CLIENT_ID", "testClientID")
	os.Setenv("OKTA_DATABRICKS_APP_ID", "databricksAppID")
	os.Setenv("OKTA_SNOWFLAKE_APP_IDS", "snowflakeAppID1,snowflakeAppID2,snowflakeAppID3")
	oktaClient, err := GetOktaClient(context.Background())
	r.NoError(err)
	r.NotNil(oktaClient)
}

// Ugh... how do I use indices to control the "next" process?
func testGetterFunc(appID string, qp *query.Params) ([]*okta.AppUser, *okta.Response, error) {
	return nil, nil, nil
}

// TODO(aku): define testGetterFunc with some nil but paginated outputs?
func TestListUsersPagination(t *testing.T) {
	r := require.New(t)
	_, err := paginateListUsers(context.TODO(), "testAppID", testGetterFunc)
	r.Error(err) // r.NoError(err)
}
