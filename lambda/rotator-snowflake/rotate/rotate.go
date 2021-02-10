package rotate

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func getUsers() ([]string, error) {
	// TODO(aku): get list from okta
	return []string{os.Getenv("CURRENT_USER")}, nil
}

func updateSnowflake(user string, db *sql.DB, pubKey *rsa.PublicKey) error {
	// TODO(aku): figure out how to
	query := fmt.Sprintf("ALTER USER %s SET RSA_PUBLIC_KEY_2 = '%s'", user, pubKey)
	snowflake.ExecNoRows(db, query)
	return nil
}

func updateDatabricks(user string, databricks *aws.DBClient) error {

	return nil
}

func Rotate(ctx context.Context) error {
	snowflakeDB, err := setup.SetupSnowflake()
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake and databricks")
	}

	databricksConnection, err := setup.SetupDatabricks()
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake and databricks")
	}

	users, err := getUsers()
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	for _, user := range users {
		// Generate new keypair
		privKey, err := keypair.GenerateRSAKeypair()
		if err != nil {
			return errors.Wrap(err, "Unable to generate RSA keypair")
		}
		err = updateSnowflake(user, snowflakeDB, &privKey.PublicKey)
		logrus.Warn(err)
		err = updateDatabricks(user, databricksConnection)
		logrus.Warn(err)
	}

	return nil
}
