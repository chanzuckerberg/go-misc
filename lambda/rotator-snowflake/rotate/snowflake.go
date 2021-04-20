package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/chanzuckerberg/go-misc/rand"
	"github.com/pkg/errors"
	"github.com/segmentio/chamber/store"

	"github.com/chanzuckerberg/go-misc/snowflake"
)

func updateSnowflake(ctx context.Context, user string, db *sql.DB, pubKey string) error {
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, pubKey)
	// Snowflake does not support preparation for ALTER statements
	// https://docs.snowflake.com/en/user-guide/sql-prepare.html
	_, err := snowflake.ExecNoRows(ctx, db, query)

	return err
}

func RotatePassword(ctx context.Context, user string, db *sql.DB) (err error) {
	// Generate a random string
	passwordLength := 50
	newPassword, err := rand.GenerateRandomString(passwordLength)
	if err != nil {
		return errors.Wrapf(err, "Unable to generate a random string for the %s's snowflake account", user)
	}

	query := fmt.Sprintf(`ALTER USER "%s" SET PASSWORD = '%s'`, user, newPassword)
	_, pwdChangeErr := snowflake.ExecNoRows(ctx, db, query)
	if err != nil {
		return errors.Wrapf(pwdChangeErr, "Unable to set new Snowflake password for %s user", user)
	}
	// If successful, write to chamber? Should I return a function? Instead of returning a password

	return nil
}

// TODO(aku): Think of a better name
// Try (your best) to find a good way to plug in placeholder values
func WritePasswordSecretStore(ctx context.Context, user string, secrets setup.SecretStore, newPassword string) error {
	passwordName := fmt.Sprintf("snowflake_%s_password", "placeholder accountname")
	tokenSecretID := store.SecretId{
		Service: "placeholder service",
		Key:     passwordName,
	}

	return errors.Wrap(
		secrets.Write(tokenSecretID, newPassword),
		"Unable to write secret to chamber",
	)
}
