package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/snowflake"
)

func updateSnowflake(ctx context.Context, user string, db *sql.DB, pubKey string) error {
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, pubKey)
	// Snowflake does not support preparation for ALTER statements
	// https://docs.snowflake.com/en/user-guide/sql-prepare.html
	_, err := snowflake.ExecNoRows(ctx, db, query)

	return err
}
