package rotate

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/stretchr/testify/require"
)

func TestBuildSnowflakeSecrets(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	r.Nil(nil)

	// There should be an error with dummy credentials...
	err := os.Setenv("SNOWFLAKE_ACCOUNT", "testaccount")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_PASSWORD", "testpassword")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_USER", "testuser")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_ROLE", "testrole")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_REGION", "testregion")
	r.NoError(err)
	err = os.Setenv("DATABRICKS_HOST", "testhost")
	r.NoError(err)

	dbPtr, err := setup.Snowflake()
	r.NoError(err)
	secretsMap, err := buildSnowflakeSecrets(dbPtr, "testuser", "test private key")
	r.Error(err)
	r.Contains(secretsMap, "snowflake.user")
}
