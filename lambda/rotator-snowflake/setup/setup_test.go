package setup

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/segmentio/chamber/store"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	params map[string]string
}

func (s *mockStore) Write(id store.SecretId, value string) error {
	s.params[id.Key] = value

	return nil
}

func (s *mockStore) Read(id store.SecretId, version int) (store.Secret, error) {
	return store.Secret{
		// note: we're disregarding the "service" in id for now.
		Value: aws.String("testValue"),
		Meta:  store.SecretMetadata{},
	}, nil
}

func TestGetOktaClient(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	os.Setenv("OKTA_ORG_URL", "https://www.testOrgURL.com")
	os.Setenv("OKTA_CLIENT_ID", "testClientID")
	os.Setenv("OKTA_PARAM_STORE_SERVICE", "testService")
	store := &mockStore{}
	oktaClient, err := Okta(context.Background(), store)
	r.NoError(err)
	r.NotNil(oktaClient)
}

func TestGetDatabricksAccount(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	os.Setenv("DATABRICKS_HOST", "testhost")
	os.Setenv("DATABRICKS_APP_ID", "testappID")
	os.Setenv("DATABRICKS_PARAM_STORE_SERVICE", "testService")
	store := &mockStore{}
	databricksAccount, err := Databricks(context.Background(), store)
	r.NoError(err)
	r.NotNil(databricksAccount)
}

func TestGetSnowflakeAccounts(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("SNOWFLAKE_OKTAMAP", "TEST1:clientID1,TEST2:clientID2")
	r.NoError(err)

	err = os.Setenv("SNOWFLAKE_TEST1_NAME", "test1name")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST1_USER", "test1user")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST1_ROLE", "test1role")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST1_REGION", "test1region")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST1_PARAM_STORE_SERVICE", "test1Service")
	r.NoError(err)

	err = os.Setenv("SNOWFLAKE_TEST2_NAME", "test2name")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST2_USER", "test2user")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST2_ROLE", "test2role")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST2_REGION", "test2region")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_TEST2_PARAM_STORE_SERVICE", "test2Service")
	r.NoError(err)

	store := &mockStore{}
	snowflakeAccounts, err := Snowflake(context.Background(), store)
	r.NoError(err)
	r.NotNil(snowflakeAccounts)
	r.Len(snowflakeAccounts, 2)
}
