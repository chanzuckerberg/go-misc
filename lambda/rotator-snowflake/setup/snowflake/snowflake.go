package snowflake

import (
	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
)

// TODO(aku): Make everything snowflake account-specific
func LoadSnowflakeAccounts(accountList []string) ([]*SnowflakeAccount, error) {
	snowflakeErrs := &multierror.Error{}
	acctList := []*SnowflakeAccount{}
	for _, acctName := range accountList {
		env := &SnowflakeClientEnv{}
		snowflakeErrs = multierror.Append(snowflakeErrs, envconfig.Process(acctName, env))
		// Process acctList
		cfg := snowflake.SnowflakeConfig{
			Account:  env.ACCOUNT,
			User:     env.USER,
			Role:     env.ROLE,
			Region:   env.REGION,
			Password: env.PASSWORD, //TODO: see if we can use private key instead
		}
		sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, envconfig.Process(acctName, env))
			continue
		}
		if sqlDB != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, errors.Errorf("Unable to create db connection with the %s snowflake account", acctName))
			continue
		}
		// okay... now all the checks are done:
		acctList = append(acctList, &SnowflakeAccount{
			AppID: env.APP_ID,
			Name:  env.ACCOUNT,
			DB:    sqlDB,
		})
	}

	return acctList, nil
}
