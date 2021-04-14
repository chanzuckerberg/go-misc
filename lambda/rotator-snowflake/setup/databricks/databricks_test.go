package databricks

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestDatabricksSetup(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("DATABRICKS_HOST", "testhost")
	r.NoError(err)
	err = os.Setenv("DATABRICKS_APP_ID", "testappID")
	r.NoError(err)
	err = os.Setenv("DATABRICKS_PARAM_STORE_SERVICE", "testService")
	r.NoError(err)
	clientPtr, err := LoadDatabricksClientEnv()
	r.NoError(err)
	r.NotNil(clientPtr)

	os.Unsetenv("DATABRICKS_HOST")
	clientPtr, err = LoadDatabricksClientEnv()
	r.Error(err)
}
