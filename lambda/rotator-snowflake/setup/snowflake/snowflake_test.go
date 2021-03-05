package snowflake

import (
	"database/sql"
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestSnowflakeSetup(t *testing.T) {
	r := require.New(t)
	// keep the original values....
	defer util.ResetEnv(os.Environ()) // TODO: make ResetEnv() part of a go-misc package
	err := os.Setenv("SNOWFLAKE_PASSWORD", "testpassword")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_USER", "testuser")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_ROLE", "testrole")
	r.NoError(err)
	err = os.Setenv("SNOWFLAKE_REGION", "testregion")
	r.NoError(err)

	dbPtr, err := Snowflake("testAccount")
	r.NoError(err)
	r.NotNil(dbPtr)
	r.IsType(&sql.DB{}, dbPtr)
}
