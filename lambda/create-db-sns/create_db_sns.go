package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chanzuckerberg/go-misc/sentry"
	_ "github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

// Design choices:
// * The return value should be a URI, containing all the info needed to connect to the database.
//   This allows the flexibility to be able to have the code decide how it wants to implement the
//   database copying without requiring the caller to make any assumptions about what this code does.
//   This code could in theory create a new user, clone the underlying DB to a new DB, even change
//   DB engines in the most extreme scenario. The caller shouldn't need to know how to construct a
//   connection string.
// * The source string identifies a DB to clone. We do not require the caller to know the username or
//   password to the source DB. This DB identifier can be decoupled from the underlying
//   database implementation, but for now we map that name directly to the Postgres database name. In
//   the future, this can refer to potentially other database instances, via some naming convention
//   or some central DB of valid sources.
// * v1 does not necessarily need to take full advantage of this flexibility, as long as it preserves
//   the interface to allow it. In v1, we can make assumptions about the source and target DB sitting
//   on the same postgres instance. We also assume a fixed source Postgres instance, for which we get
//   the connection string from an environment variable DATABASE_URI.

// Open questions:
// * How do we report back the URI for the DB? For now, writing to a key in AWS Secrets Manager, replacing
//   whatever key is there and merging with existing values. However, this makes the assumptions that the
//   caller is using secrets manager, that the secret is JSON structured, and that it is easy to give our
//   code the AWS permissions to write to arbitrary secrets. (The last one is tricky.) If we have no way
//   of directly reporting the Database URI back, then we will be forced to violate our above design choice
//   and have the client aware of the DB so that it can construct its own connection string.

// CreateDB contains the information about the source DB to copy and the target to copy to, and where to report the new location
type CreateDB struct {
	Version   int    `json:"version"` // Only valid value is 1 for now
	Source    string `json:"source"`
	Target    string `json:"target"`
	SecretKey string `json:"secret_key"` // Database secret key to write to
	SecretARN string `json:"secret_arn"` // Database secret to write connection string to; URI value will be merged into most recent version
}

// aws sns publish --message '{"version": 1, "source": "template_db_name", "target": "target_db_name", "secret_key": "database_uri", "secret_arn": ""}'

func run(ctx context.Context, snsEvent events.SNSEvent) {
	sentry.Run(ctx, func(innerCtx context.Context) error { return handler(innerCtx, snsEvent) })
}

func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	for _, record := range snsEvent.Records {
		// SNSEvent structure for reference: https://github.com/aws/aws-lambda-go/blob/master/events/sns.go
		snsRecord := record.SNS
		fmt.Printf("[%s %s] Message = %s \n", record.EventSource, snsRecord.Timestamp, snsRecord.Message)

		// Parse the message
		var createDB CreateDB
		json.Unmarshal([]byte(snsRecord.Message), &createDB)
		if createDB.Version != 1 {
			logrus.Fatalf("Unexpected payload version. Expected 1, got %d", createDb.Version)
		}

		db, err := sql.Open("pgx", os.Getenv("DATABASE_URI"))
		if err != nil {
			logrus.WithError(err).Fatal("Unable to connect to database")
		}
		defer db.Close()

		conn, err := db.Conn(ctx)
		if err != nil {
			logrus.WithError(err).Fatal("Unable to create connection")
		}
		defer conn.Close()

		// TODO(mbarrien): Ensure source DB template exists
		// TODO(mbarrien): Ensure caller has permissions to clone this particular source DB?
		// TODO(mbarrien): Ensure target DB does not exist?
		// TODO(mbarrien): Construct insert db clone command to copy from source to target

		result, err := conn.ExecContext(ctx, "INSERT DB CLONE COMMAND HERE")
		if err != nil {
			logrus.WithError(err).Fatal("Unable to clone database")
		}

		rows, err := result.RowsAffected()
		if err != nil {
			logrus.WithError(err).Fatal("Error getting rows affected")
		}
		if rows != 1 {
			log.Fatalf("expected single row affected, got %d rows affected", rows)
		}

		// TODO(mbarrien): Construct new postgres connection string here
		// TODO(mbarrien): Read existing secret, and merge connection string into secret
		// TODO(mbarrien): Write new connection string to secret
	}
	return nil
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	flag.Parse()
	logrus.Debugf("arg: %s", flag.Arg(0))

	// cheap and simple local-mode for lambda
	if flag.Arg(0) == "-local" {
		message := events.SNSEvent{
			Records: []events.SNSEventRecord{{
				EventVersion: "1",
				SNS: events.SNSEntity{
					Timestamp: time.Now(),
					Message:   "Insert stuff from cmd here",
				},
			}},
		}
		run(context.Background(), message)
	} else {
		lambda.Start(run)
	}
}
