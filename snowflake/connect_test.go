package snowflake

import (
	"crypto/x509"
	"encoding/base64"
	"net/url"
	"os"
	"testing"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/stretchr/testify/require"
)

var testSnowflakeConfig = SnowflakeConfig{
	Account:          "test-acct",
	User:             "test-user",
	Password:         "test-password",
	BrowserAuth:      false,
	OauthAccessToken: "test-oauthaccesstoken",
	Region:           "test-region",
	Role:             "test-role",
}
var TestKeypairConfig = keypair.Config{
	KeyPrefix: "test",
	KeyPath:   ".",
}

func TestDSN(t *testing.T) {
	r := require.New(t)
	testPriv, err := keypair.GenerateKeypair()
	r.NoError(err)

	TestKeypairConfig.PrivateKey = testPriv
	err = keypair.SaveKeys(TestKeypairConfig)
	r.NoError(err)

	defer os.Remove(TestKeypairConfig.GetPrivateKeyPath())
	defer os.Remove(TestKeypairConfig.GetPublicKeyPath())

	testSnowflakeConfig.PrivateKeyPath = TestKeypairConfig.GetPrivateKeyPath()

	testDSN, err := DSN(&testSnowflakeConfig)
	r.NoError(err)
	r.NotNil(testDSN)

	r.Contains(testDSN, testSnowflakeConfig.User)
	r.Contains(testDSN, testSnowflakeConfig.Region)
	r.Contains(testDSN, testSnowflakeConfig.Account)
	r.Contains(testDSN, "privateKey=")

	// Looked into the gosnowflake code to identify how the private key marshaling worked.
	// Replicated here for testing
	goSnowflakePrivKeyBytes, err := x509.MarshalPKCS8PrivateKey(testPriv)
	r.NoError(err, "This custom key unmarshal process from gosnowflake doesn't work. Source: https://github.com/snowflakedb/gosnowflake/blob/52137ce8c32eaf93b0bd22fc5c7297beff339812/dsn.go#L131")

	// Added this block because the DSN() function would URL-encode equal signs as %3D
	keyBase64 := base64.URLEncoding.EncodeToString(goSnowflakePrivKeyBytes)
	decodedPrivKey := url.QueryEscape(keyBase64)
	r.Contains(testDSN, decodedPrivKey)
}

func TestConfigureProvider(t *testing.T) {
	r := require.New(t)
	testPriv, err := keypair.GenerateKeypair()
	r.NoError(err)

	TestKeypairConfig.PrivateKey = testPriv
	err = keypair.SaveKeys(TestKeypairConfig)
	r.NoError(err)

	defer os.Remove(TestKeypairConfig.GetPrivateKeyPath())
	defer os.Remove(TestKeypairConfig.GetPublicKeyPath())

	testSnowflakeConfig.PrivateKeyPath = TestKeypairConfig.GetPrivateKeyPath()

	testDB, err := ConfigureProvider(&testSnowflakeConfig)
	r.NoError(err)
	r.NotNil(testDB)
}
