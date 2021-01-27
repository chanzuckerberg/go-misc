package snowflake

// Originally from terraform-provider-snowflake/pkg/provider/provider.go
import (
	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"github.com/snowflakedb/gosnowflake"
)

func ConfigureProvider(s *schema.ResourceData) (interface{}, error) {
	account := s.Get("account").(string)
	user := s.Get("username").(string)
	password := s.Get("password").(string)
	browserAuth := s.Get("browser_auth").(bool)
	privateKeyPath := s.Get("private_key_path").(string)
	oauthAccessToken := s.Get("oauth_access_token").(string)
	region := s.Get("region").(string)
	role := s.Get("role").(string)

	dsn, err := DSN(account, user, password, browserAuth, privateKeyPath, oauthAccessToken, region, role)

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
