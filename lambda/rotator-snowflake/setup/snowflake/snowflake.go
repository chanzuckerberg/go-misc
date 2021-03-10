package snowflake

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func LoadSnowflakeClientEnv() (*SnowflakeClientEnvironment, error) {
	env := &SnowflakeClientEnvironment{}
	err := envconfig.Process("SNOWFLAKE", env)

	return env, errors.Wrap(err, "Unable to load all the environment variables")
}

func LoadSnowflakeAccounts() ([]*SnowflakeAccount, error) {
	env := &SnowflakeAccountEnvironment{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Snowflake Account information not defined in environment")
	}
	acctList := []*SnowflakeAccount{}
	for acctName, appID := range env.OktaMap {
		acctList = append(acctList, &SnowflakeAccount{AppID: string(appID), Name: string(acctName)})
	}
	return acctList, nil
}
