package snowflake

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

func Exec(db *sql.DB, query string) (sql.Result, error) {
	logrus.Debug("[DEBUG] exec stmt ", query)

	return db.Exec(query)
}
