package rotate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chanzuckerberg/go-misc/snowflake"
)

func updateSnowflake(user string, db *sql.DB, pubKey string) error {
	query := fmt.Sprintf(`ALTER USER "%s" SET RSA_PUBLIC_KEY_2 = "%s"`, user, pubKey)
	_, err := snowflake.ExecNoRows(context.TODO(), db, query)

	return err
}
