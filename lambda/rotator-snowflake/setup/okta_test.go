package setup

import (
	"context"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestGetOktaClient(t *testing.T) {
	r := require.New(t)
	r.Nil(nil)
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
