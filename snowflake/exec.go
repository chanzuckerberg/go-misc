package snowflake

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ExecNoRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	logrus.Debugf("[DEBUG] exec stmt (%s)", query)

	// if no args, then don't prepare:
	if len(args) == 0 {
		return db.ExecContext(ctx, query)
	}

	// Prepare the statement
	preparedStmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to prepare sql statement (%s)", query)
	}
	defer preparedStmt.Close()

	sqlResult, err := preparedStmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to execute prepared statement")
	}

	// Run statement with arguments
	return sqlResult, nil
}

func ExecMulti(ctx context.Context, db *sql.DB, queries []string, args ...[]interface{}) error {
	logrus.Debug("[DEBUG] exec stmts ", queries)

	if len(queries) != len(args) {
		return errors.Errorf("Number of queries (%d) does not match with number of argument groups (%d)", len(queries), len(args))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Unable to begin transaction with database")
	}

	queryErrors := &multierror.Error{}

	for i, query := range queries {
		logrus.Debugf("[DEBUG] exec (%s) ", query)
		// if no args, then don't prepare:
		if len(args[i]) == 0 {
			_, err = db.ExecContext(ctx, query)
			queryErrors = multierror.Append(queryErrors, errors.Wrap(err, "Unable to execute statement"))

			continue
		}
		// Prepare the statement
		preparedStmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			queryErrors = multierror.Append(queryErrors, errors.Wrapf(err, "Unable to prepare query (%s)", query))
		}
		defer preparedStmt.Close()

		// Run statement with arguments
		_, err = preparedStmt.ExecContext(ctx, args[i]...)
		if err != nil {
			queryErrors = multierror.Append(queryErrors, errors.Wrapf(err, "Unable to execute query (%s)", query))

			// If executing this statement returns an error, rollback, collect error from rollback if any, and return all errors
			err = tx.Rollback()
			if err != nil {
				queryErrors = multierror.Append(queryErrors, errors.Wrapf(err, "Unable to roll back transaction"))
			}
			return queryErrors.ErrorOrNil()
		}
	}

	err = tx.Commit()
	queryErrors = multierror.Append(queryErrors, err)

	return queryErrors.ErrorOrNil()
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
func Query(ctx context.Context, db *sql.DB, stmt string, args ...interface{}) (*sqlx.Rows, error) {
	logrus.Debug("[DEBUG] query stmt ", stmt)

	sdb := sqlx.NewDb(db, "snowflake").Unsafe()

	if len(args) == 0 {
		return sdb.QueryxContext(ctx, stmt)
	}

	preparedSQLxStmt, err := sdb.PreparexContext(ctx, stmt)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to prepare query (%s)", stmt)
	}
	defer preparedSQLxStmt.Close()

	return preparedSQLxStmt.QueryxContext(ctx, args...)
}
