package setup

import (
	"context"
	"fmt"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/segmentio/chamber/store"
	"github.com/sirupsen/logrus"
)

// For getting the latest version of a secret, use -1
var latestSecretVersionIndex = -1

type SecretStore interface {
	Read(id store.SecretId, version int) (store.Secret, error)
}

func Okta(ctx context.Context, secrets SecretStore) (*oktaCfg.OktaClient, error) {
	env, err := oktaCfg.LoadOktaClientEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to load right Okta env variables")
	}

	service := env.PARAM_STORE_SERVICE
	tokenSecretID := store.SecretId{
		Service: service,
		Key:     "okta_private_key",
	}

	private_key, err := secrets.Read(tokenSecretID, latestSecretVersionIndex)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't find Okta Private Key in AWS Parameter Store in service (%s)", service)
	}

	client, err := okta.NewClient(
		ctx,
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(env.CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(*private_key.Value),
		okta.WithOrgUrl(env.ORG_URL),
		okta.WithCache(true),
	)

	return &oktaCfg.OktaClient{Client: client}, errors.Wrap(err, "Unable to configure Okta client")
}

func Databricks(ctx context.Context, secrets SecretStore) (*databricksCfg.Account, error) {
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
	token, err := secrets.Read(tokenSecretID, latestSecretVersionIndex)
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

func Snowflake(ctx context.Context, secrets SecretStore) ([]*snowflakeCfg.Account, error) {
	snowflakeAccts := &snowflakeCfg.Accounts{}

	err := envconfig.Process("SNOWFLAKE", snowflakeAccts)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}

	acctMapping := snowflakeAccts.OKTAMAP
	logrus.Debug("Snowflake: Okta Client ID Mapping: ", acctMapping)

	acctList := []*snowflakeCfg.Account{}

	for acctName, snowflakeAppID := range acctMapping {
		snowflakeEnv, err := snowflakeCfg.LoadSnowflakeEnv(acctName)
		if err != nil {
			currentError := errors.Wrapf(err, "Error configuring Snowflake %s account", acctName)
			return nil, currentError
		}

		cfg := snowflake.Config{
			Account: snowflakeEnv.NAME,
			User:    snowflakeEnv.USER,
			Role:    snowflakeEnv.ROLE,
			Region:  snowflakeEnv.REGION,
		}
		// Get the password through the account name
		service := snowflakeEnv.PARAM_STORE_SERVICE
		passwordName := fmt.Sprintf("snowflake_%s_password", snowflakeEnv.NAME)
		tokenSecretID := store.SecretId{
			Service: service,
			Key:     passwordName,
		}

		password, err := secrets.Read(tokenSecretID, latestSecretVersionIndex)
		if err != nil {
			return nil, errors.Wrapf(err, "Can't find %s's password in AWS Parameter Store in service (%s)", snowflakeEnv.NAME, service)
		}

		cfg.Password = *password.Value

		sqlDB, err := snowflakeCfg.ConfigureConnection(cfg)
		if err != nil {
			logrus.Debug(err)
			return nil, errors.Wrap(err, "Unable to configure SQL Connection")
		}

		acctList = append(acctList, &snowflakeCfg.Account{
			AppID: snowflakeAppID,
			Name:  snowflakeEnv.NAME,
			DB:    sqlDB,
		})
	}

	return acctList, nil
}

func ListSnowflakeUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, snowflakeAcct *snowflakeCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, snowflakeAcct.AppID, userGetter)
}

func ListDatabricksUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, databricksAccount *databricksCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, databricksAccount.AppID, userGetter)
}
