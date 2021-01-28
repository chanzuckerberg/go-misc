package snowflake

import (
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
	testPriv, testPub, err := keypair.GenerateKeypair()
	r.NoError(err)

	TestKeypairConfig.PrivateKey = testPriv
	TestKeypairConfig.PublicKey = testPub
	err = keypair.SaveKeys(TestKeypairConfig)
	r.NoError(err)

	defer os.Remove(TestKeypairConfig.GetPrivateKeyPath())
	defer os.Remove(TestKeypairConfig.GetPublicKeyPath())

	testSnowflakeConfig.PrivateKeyPath = TestKeypairConfig.GetPrivateKeyPath()

	testDSN, err := DSN(&testSnowflakeConfig)
	r.NoError(err)
	r.NotNil(testDSN)
}

func TestConfigureProvider(t *testing.T) {
	r := require.New(t)
	testPriv, testPub, err := keypair.GenerateKeypair()
	r.NoError(err)

	TestKeypairConfig.PrivateKey = testPriv
	TestKeypairConfig.PublicKey = testPub
	err = keypair.SaveKeys(TestKeypairConfig)
	r.NoError(err)

	defer os.Remove(TestKeypairConfig.GetPrivateKeyPath())
	defer os.Remove(TestKeypairConfig.GetPublicKeyPath())

	testSnowflakeConfig.PrivateKeyPath = TestKeypairConfig.GetPrivateKeyPath()

	testDB, err := ConfigureProvider(&testSnowflakeConfig)
	r.NoError(err)
	r.NotNil(testDB)
}
