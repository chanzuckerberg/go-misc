package setup

import (
	"context"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/chamber/store"
)

func Okta(ctx context.Context, secrets *store.SSMStore) (*oktaCfg.OktaClient, error) {
	return oktaCfg.GetOktaClient(ctx, secrets)
}

func Databricks(ctx context.Context, secrets *store.SSMStore) (*databricksCfg.Account, error) {
	databricksEnv, err := databricksCfg.LoadDatabricksClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Databricks information from environment variables")
	}
	host := databricksEnv.HOST

	service := databricksEnv.PARAM_STORE_SERVICE
	tokenSecretID := store.SecretId{
		Service: service,
		Key:     "databricks_token",
	}
	token, err := secrets.Read(tokenSecretID, -1)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't find Databricks Token in AWS Parameter Store in service (%s)", service)
	}

	dbClient := databricks.NewAWSClient(host, *token.Value)
	databricksAccount := &databricksCfg.Account{
		AppID:  databricksEnv.APP_ID,
		Client: dbClient,
	}
	return databricksAccount, nil
}

func Snowflake(ctx context.Context, secrets *store.SSMStore) ([]*snowflakeCfg.Account, error) {
	env := &snowflakeCfg.Accounts{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}
	acctMapping := env.OKTAMAP
	return snowflakeCfg.LoadSnowflakeAccounts(acctMapping, secrets)
}

func ListSnowflakeUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, snowflakeAcct *snowflakeCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, snowflakeAcct.AppID, userGetter)
}

func ListDatabricksUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, databricksAccount *databricksCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, databricksAccount.AppID, userGetter)
}
