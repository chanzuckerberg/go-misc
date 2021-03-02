package setup

import (
	"database/sql"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func TestGetUsers(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("CURRENT_USERS", "testUser")
	r.NoError(err)
	strList, err := GetUsers()
	r.NoError(err)
	r.Len(strList, 1)
	r.Equal("testUser", strList[0])
}

func TestSnowflakeSetup(t *testing.T) {
	r := require.New(t)
	// keep the original values....
	defer util.ResetEnv(os.Environ()) // TODO: make ResetEnv() part of a go-misc package
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
	dbPtr, err := Snowflake()
	r.NoError(err)
	r.NotNil(dbPtr)
	r.IsType(&sql.DB{}, dbPtr)

	os.Unsetenv("SNOWFLAKE_ACCOUNT")
	dbPtr, err = Snowflake()
	r.Error(err)
	r.Nil(dbPtr)
}

func TestDatabricksSetup(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	err := os.Setenv("DATABRICKS_HOST", "testhost")
	r.NoError(err)
	os.Setenv("DATABRICKS_TOKEN", "testtoken")
	r.NoError(err)
	clientPtr, err := Databricks()
	r.NoError(err)
	r.NotNil(clientPtr)
	r.IsType(&aws.DBClient{}, clientPtr)

	os.Unsetenv("DATABRICKS_TOKEN")
	clientPtr, err = Databricks()
	r.Error(err)
	r.Nil(clientPtr)
}
