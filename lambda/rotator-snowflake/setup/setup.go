package setup

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/chanzuckerberg/go-misc/databricks"
	"github.com/chanzuckerberg/go-misc/errors"
	databricksCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/databricks"
	oktaCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/okta"
	snowflakeCfg "github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup/snowflake"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
	"github.com/okta/okta-sdk-golang/okta"
	"github.com/segmentio/chamber/store"
	"github.com/sirupsen/logrus"
)

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
	private_key, err := secrets.Read(tokenSecretID, -1)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't find Okta Private Key in AWS Parameter Store in service (%s)", service)
	}

	privKeyNoQuotes := strings.ReplaceAll(*private_key.Value, `"`, ``)
	client, err := okta.NewClient(
		ctx,
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(env.CLIENT_ID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(privKeyNoQuotes),
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

func Snowflake(ctx context.Context, secrets SecretStore) ([]*snowflakeCfg.Account, error) {
	env := &snowflakeCfg.Accounts{}
	err := envconfig.Process("SNOWFLAKE", env)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse list of accounts from environment variables")
	}
	acctMapping := env.OKTAMAP

	// return snowflakeCfg.LoadSnowflakeAccounts(acctMapping, secrets)

	snowflakeErrs := &multierror.Error{}
	acctList := []*snowflakeCfg.Account{}

	for acctName, snowflakeAppID := range acctMapping {
		// If acctName has "okta" or "databricks" in the name, print a warning for possible name collision
		oktaCollision := strings.Contains(acctName, "okta")
		if oktaCollision {
			logrus.Warnf("Snowflake Account %s will likely collide with okta Environment Variables", acctName)
		}

		databricksCollision := strings.Contains(acctName, "databricks")
		if databricksCollision {
			logrus.Warnf("Snowflake Account %s will likely collide with databricks Environment Variables", acctName)
		}

		snowflakeEnv := &snowflakeCfg.SnowflakeClientEnv{}

		err := envconfig.Process(acctName, snowflakeEnv)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, errors.Wrap(err, "Error processing Snowflake environment variables"))
		}

		sqlDB, err := ConfigureConnection(snowflakeEnv, secrets)
		if err != nil {
			snowflakeErrs = multierror.Append(snowflakeErrs, err)

			continue
		}

		acctList = append(acctList, &snowflakeCfg.Account{
			AppID: snowflakeAppID,
			Name:  snowflakeEnv.NAME,
			DB:    sqlDB,
		})
	}

	return acctList, snowflakeErrs.ErrorOrNil()
}

func ConfigureConnection(env *snowflakeCfg.SnowflakeClientEnv, secrets SecretStore) (*sql.DB, error) {
	cfg := snowflake.SnowflakeConfig{
		Account: env.NAME,
		User:    env.USER,
		Role:    env.ROLE,
		Region:  env.REGION,
	}
	// Get the password through the account name
	service := env.PARAM_STORE_SERVICE
	tokenSecretID := store.SecretId{
		Service: service,
		Key:     fmt.Sprintf("%s_password", env.NAME),
	}

	password, err := secrets.Read(tokenSecretID, -1)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't find %s's password in AWS Parameter Store in service (%s)", env.NAME, service)
	}

	cfg.Password = *password.Value

	sqlDB, err := snowflake.ConfigureSnowflakeDB(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to configure Snowflake SQL connection")
	}
	if sqlDB == nil {
		return nil, errors.Errorf("Unable to create db connection with the %s snowflake account", env.NAME)
	}

	return sqlDB, nil
}

func ListSnowflakeUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, snowflakeAcct *snowflakeCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, snowflakeAcct.AppID, userGetter)
}

func ListDatabricksUsers(ctx context.Context, oktaClient *oktaCfg.OktaClient, databricksAccount *databricksCfg.Account) (*sets.StringSet, error) {
	userGetter := oktaClient.Client.Application.ListApplicationUsers
	return oktaCfg.GetOktaAppUsers(ctx, databricksAccount.AppID, userGetter)
}
