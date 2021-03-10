package snowflake

import (
	"fmt"

	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// TODO: figure out okta process. It'll have to work with oktaClient
func GetSnowflakeApps(oktaClient *oktaCfg.OktaClient, snowflakeAppIDs []string) ([]*SnowflakeAccount, error) {
	accounts := []*SnowflakeAccount{}
	for _, appID := range snowflakeAppIDs {
		accountName := fmt.Sprintf("%s_account", appID) //TODO: temporary fix until we have something solid
		accounts = append(accounts, &SnowflakeAccount{
			AppID: appID,
			Name:  accountName,
		})
	}
	return accounts, nil
}

func LoadSnowflakeClientEnv() (*SnowflakeClientEnvironment, error) {
	env := &SnowflakeClientEnvironment{}
	err := envconfig.Process("SNOWFLAKE", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}
