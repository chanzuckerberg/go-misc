package snowflake

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ExecNoRows(ctx context.Context, db *sql.DB, query string) (sql.Result, error) {
	logrus.Debug("[DEBUG] exec stmt ", query)

	return db.ExecContext(ctx, query)
}

func ExecMulti(ctx context.Context, db *sql.DB, queries []string) error {
	logrus.Debug("[DEBUG] exec stmts ", queries)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, query := range queries {
		_, err = tx.Exec(query)
		if err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

// QueryRow will run stmt against the db and return the row. We use
// [DB.Unsafe](https://godoc.org/github.com/jmoiron/sqlx#DB.Unsafe) so that we can scan to structs
// without worrying about newly introduced columns
func QueryRow(ctx context.Context, db *sql.DB, stmt string, args ...interface{}) *sqlx.Row {
	logrus.Debug("[DEBUG] query stmt ", stmt)

	sdb := sqlx.NewDb(db, "snowflake").Unsafe()

	if len(args) == 0 {
		// Don't prepare the statement
		return sdb.QueryRowxContext(ctx, stmt)
	}

	// TODO: use this Prepare SQL pattern for the other functions in this file
	preparedStmt, err := sdb.PreparexContext(ctx, stmt)
	if err != nil {
		logrus.Warn(errors.Wrapf(err, "Unable to prepare query (%s)", stmt))

		return nil
	}
	defer preparedStmt.Close()

	return preparedStmt.QueryRowxContext(ctx, args)
}

// Query will run stmt against the db and return the rows. We use
// [DB.Unsafe](https://godoc.org/github.com/jmoiron/sqlx#DB.Unsafe) so that we can scan to structs
// without worrying about newly introduced columns
func Query(ctx context.Context, db *sql.DB, stmt string) (*sqlx.Rows, error) {
	sdb := sqlx.NewDb(db, "snowflake").Unsafe()

	return sdb.QueryxContext(ctx, stmt)
}
