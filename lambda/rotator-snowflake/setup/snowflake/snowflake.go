package snowflake

import (
	"database/sql"
	"fmt"

	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	"github.com/chanzuckerberg/go-misc/snowflake"
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

func Snowflake(snowflakeAcct string) (*sql.DB, error) {
	snowflakeEnv, err := loadSnowflakeClientEnv()
	if err != nil {
		return nil, err
	}

	cfg := snowflake.SnowflakeConfig{
		Account:  snowflakeAcct,
		User:     snowflakeEnv.USER,
		Role:     snowflakeEnv.ROLE,
		Region:   snowflakeEnv.REGION,
		Password: snowflakeEnv.PASSWORD,
	}

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)

	return sqlDB, errors.Wrap(err, "Unable to configure Snowflake DB")
}
