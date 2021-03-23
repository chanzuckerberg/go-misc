package setup

import (
	"context"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/kelseyhightower/envconfig"
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

func TestSnowflakeConfigure(t *testing.T) {
	r := require.New(t)
	// keep the original values....
	defer util.ResetEnv(os.Environ()) // TODO: make ResetEnv() part of a go-misc package
	// First define the snowflake Account Names
	err := os.Setenv("SNOWFLAKE_ACCOUNTS", "test1,test2")
	r.NoError(err)

	err = os.Setenv("TEST1_NAME", "test1name")
	r.NoError(err)
	err = os.Setenv("TEST1_PASSWORD", "test1password")
	r.NoError(err)
	err = os.Setenv("TEST1_USER", "test1user")
	r.NoError(err)
	err = os.Setenv("TEST1_ROLE", "test1role")
	r.NoError(err)
	err = os.Setenv("TEST1_REGION", "test1region")
	r.NoError(err)

	err = os.Setenv("TEST2_NAME", "test2name")
	r.NoError(err)
	err = os.Setenv("TEST2_PASSWORD", "test2password")
	r.NoError(err)
	err = os.Setenv("TEST2_USER", "test2user")
	r.NoError(err)
	err = os.Setenv("TEST2_ROLE", "test2role")
	r.NoError(err)
	err = os.Setenv("TEST2_REGION", "test2region")
	r.NoError(err)

	// Contents of the Snowflake() function
	env := &snowflakeCfg.Accounts{}
	err = envconfig.Process("SNOWFLAKE", env)
	r.NoError(err)
	r.Len(env.ACCOUNTS, 2)
	for _, accountName := range env.ACCOUNTS {
		env := &snowflakeCfg.SnowflakeClientEnv{}
		err = envconfig.Process(accountName, env)
		r.NoError(err)
	}

	// Despite having all the environment variables defined, LoadSnowflakeAccounts() won't work
	// 	Because these credentials are bogus
	accountInfo, err := snowflakeCfg.LoadSnowflakeAccounts(env.ACCOUNTS)
	r.Error(err) // These credentials are dummy anyway. They shouldn't work in this case
	r.Len(accountInfo, 0)
}
