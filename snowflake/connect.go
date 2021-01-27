package snowflake

// Originally from terraform-provider-snowflake/pkg/provider/provider.go
import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/pkg/errors"
	"github.com/snowflakedb/gosnowflake"
)

type SnowflakeConfig struct {
	account          string `yaml:"account"`
	user             string `yaml:"username"`
	password         string `yaml:"password"`
	browserAuth      bool   `yaml:"browser_auth"`
	privateKeyPath   string `yaml:"private_key_path"`
	oauthAccessToken string `yaml:"oauth_access_token"`
	region           string `yaml:"region"`
	role             string `yaml:"role"`
}

func ConfigureProvider(s *SnowflakeConfig) (interface{}, error) {
	dsn, err := DSN(
		s.account,
		s.user,
		s.password,
		s.browserAuth,
		s.privateKeyPath,
		s.oauthAccessToken,
		s.region,
		s.role,
	)

	if err != nil {
		return nil, errors.Wrap(err, "could not build dsn for snowflake connection")
	}

	db, err := Open(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open snowflake database.")
	}

	return db, nil
}

func DSN(
	account,
	user,
	password string,
	browserAuth bool,
	privateKeyPath,
	oauthAccessToken,
	region,
	role string) (string, error) {
	// us-west-2 is their default region, but if you actually specify that it won't trigger their default code
	//  https://github.com/snowflakedb/gosnowflake/blob/52137ce8c32eaf93b0bd22fc5c7297beff339812/dsn.go#L61
	if region == "us-west-2" {
		region = ""
	}

	config := gosnowflake.Config{
		Account: account,
		User:    user,
		Region:  region,
		Role:    role,
	}

	if privateKeyPath != "" {
		rsaPrivateKey, err := keypair.ParsePrivateKey(privateKeyPath)
		if err != nil {
			return "", errors.Wrap(err, "Private Key could not be parsed")
		}

		config.PrivateKey = rsaPrivateKey
		config.Authenticator = gosnowflake.AuthTypeJwt
	} else if browserAuth {
		config.Authenticator = gosnowflake.AuthTypeExternalBrowser
	} else if oauthAccessToken != "" {
		config.Authenticator = gosnowflake.AuthTypeOAuth
		config.Token = oauthAccessToken
	} else if password != "" {
		config.Password = password
	} else {
		return "", errors.New("no authentication method provided")
	}

	return gosnowflake.DSN(&config)
}

func Open(dsn string) (*sql.DB, error) {
	return sql.Open("snowflake-instrumented", dsn)
}
