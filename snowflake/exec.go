package snowflake

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func ExecNoRows(db *sql.DB, query string) (sql.Result, error) {
	logrus.Debug("[DEBUG] exec stmt ", query)

	return db.Exec(query)
}

func ExecMulti(db *sql.DB, queries []string) error {
	logrus.Debug("[DEBUG] exec stmts ", queries)

	tx, err := db.Begin()
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
func QueryRow(db *sql.DB, stmt string) *sqlx.Row {
	logrus.Debug("[DEBUG] query stmt ", stmt)

	sdb := sqlx.NewDb(db, "snowflake").Unsafe()

	return sdb.QueryRowx(stmt)
}

// Query will run stmt against the db and return the rows. We use
// [DB.Unsafe](https://godoc.org/github.com/jmoiron/sqlx#DB.Unsafe) so that we can scan to structs
// without worrying about newly introduced columns
func Query(db *sql.DB, stmt string) (*sqlx.Rows, error) {
	sdb := sqlx.NewDb(db, "snowflake").Unsafe()

	return sdb.Queryx(stmt)
}
