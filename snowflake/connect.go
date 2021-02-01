package snowflake

import (
	"database/sql"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/pkg/errors"
	"github.com/snowflakedb/gosnowflake"
)

type SnowflakeConfig struct {
	Account          string `yaml:"account"`
	User             string `yaml:"username"`
	Password         string `yaml:"password"`
	BrowserAuth      bool   `yaml:"browser_auth"`
	PrivateKeyPath   string `yaml:"private_key_path"`
	OauthAccessToken string `yaml:"oauth_access_token"`
	Region           string `yaml:"region"`
	Role             string `yaml:"role"`
}

func ConfigureSnowflakeDB(s *SnowflakeConfig) (*sql.DB, error) {
	dsn, err := DSN(s)

	if err != nil {
		return nil, errors.Wrap(err, "could not build dsn for snowflake connection")
	}

	return Open(dsn)
}

func DSN(conf *SnowflakeConfig) (string, error) {
	// us-west-2 is their default region, but if you actually specify that it won't trigger their default code
	//  https://github.com/snowflakedb/gosnowflake/blob/52137ce8c32eaf93b0bd22fc5c7297beff339812/dsn.go#L61
	if conf.Region == "us-west-2" {
		conf.Region = ""
	}

	config := gosnowflake.Config{
		Account: conf.Account,
		User:    conf.User,
		Region:  conf.Region,
		Role:    conf.Role,
	}

	if conf.PrivateKeyPath != "" {
		rsaPrivateKey, err := keypair.ParseRSAPrivateKey(conf.PrivateKeyPath)
		if err != nil {
			return "", errors.Wrap(err, "Private Key could not be parsed")
		}

		config.PrivateKey = rsaPrivateKey
		config.Authenticator = gosnowflake.AuthTypeJwt
	} else if conf.BrowserAuth {
		config.Authenticator = gosnowflake.AuthTypeExternalBrowser
	} else if conf.OauthAccessToken != "" {
		config.Authenticator = gosnowflake.AuthTypeOAuth
		config.Token = conf.OauthAccessToken
	} else if conf.Password != "" {
		config.Password = conf.Password
	} else {
		return "", errors.New("no authentication method provided")
	}

	return gosnowflake.DSN(&config)
}
