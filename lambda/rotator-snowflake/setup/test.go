package setup

import (
	"context"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func TestDatabricksSetup(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("DATABRICKS_HOST", "testhost")
	r.NoError(err)
	os.Setenv("DATABRICKS_TOKEN", "testtoken")
	r.NoError(err)
	clientPtr, err := Databricks(context.Background())
	r.NoError(err)
	r.NotNil(clientPtr)
	r.IsType(&aws.DBClient{}, clientPtr)

	os.Unsetenv("DATABRICKS_TOKEN")
	clientPtr, err = Databricks(context.Background())
	r.Error(err)
	r.Nil(clientPtr)
}
