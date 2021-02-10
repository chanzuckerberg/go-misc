package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/keypair"
	"github.com/chanzuckerberg/go-misc/snowflake"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xinsnake/databricks-sdk-golang/aws"
)

func getUsers() ([]string, error) {
	// TODO(aku): get list from okta
	return []string{os.Getenv("CURRENT_USER")}, nil
}

func updateUserKeys(user string, sqlDB *sql.DB, databricks *aws.DBClient) error {
	// make an multierror collection & return all the users that didn't get to be updated
	privKey, err := keypair.GenerateRSAKeypair()
	if err != nil {
		return errors.Wrao(err, "Unable to create a keypair")
	}
	// convert public key into a string
	snowflakeQuery := fmt.Sprintf("ALTER USER {%s} SET {PUBLIC_KEY_NAME} = '{%s}'", user, privKey.PublicKey)
	_, err = snowflake.Exec(sqlDB, snowflakeQuery)
	if err != nil {
		return errors.Wrap(err, "Error executing snowflake query")
	}

}

func rotate(ctx context.Context) error {
	snowflake, databricks, err := setupSnowflakeDatabricks()
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake and databricks")
	}

	users, err := getUsers()
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}
	for _, user := range users {
		err = updateUserKeys(user, snowflake, databricks) // In this function, we generate the keypair & run ALTER USER {snowflake_user} SET {PUBLIC_KEY_NAME} = '{public_key}'
		logrus.Warn(err)
	}

	return nil
}
