package main

import (
	"context"
	"flag"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chanzuckerberg/go-misc/errors"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/rotate"
	"github.com/chanzuckerberg/go-misc/lambda/rotator-snowflake/setup"
	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/chamber/store"
	"github.com/sirupsen/logrus"
)

var localFlag = flag.Bool("local", false, "Whether this lambda should be run locally")

func Rotate(ctx context.Context) error {
	numRetries := 2
	secretStore := store.NewSSMStore(numRetries)

	databricksAccount, err := setup.Databricks(ctx, secretStore)
	if err != nil {
		return errors.Wrap(err, "Unable to configure databricks")
	}

	oktaClient, err := setup.Okta(ctx, secretStore)
	if err != nil {
		return errors.Wrap(err, "Unable to configure okta")
	}

	snowflakeAccounts, err := setup.Snowflake(ctx, secretStore)
	if err != nil {
		return errors.Wrap(err, "Unable to configure snowflake accounts")
	}

	databricksUsers, err := setup.ListDatabricksUsers(ctx, oktaClient, databricksAccount)
	if err != nil {
		return errors.Wrap(err, "Unable to get list of users to rotate")
	}

	userErrors := &multierror.Error{}
	for _, snowflakeAccount := range snowflakeAccounts {
		snowflakeUsers, err := setup.ListSnowflakeUsers(ctx, oktaClient, snowflakeAccount)
		if err != nil {
			return errors.Wrap(err, "Unable to list Snowflake Users")
		}
		if snowflakeUsers == nil {
			return errors.Errorf("Got no snowflakeUsers listed for %s account", snowflakeAccount.Name)
		}
		for _, user := range snowflakeUsers.List() {
			if databricksUsers.ContainsElement(user) {
				err = rotate.ProcessUser(ctx, user, snowflakeAccount, databricksAccount)
				userErrors = multierror.Append(userErrors, errors.Wrapf(err, "Unable to rotate %s's credentials", user))
			}
		}
	}

	return userErrors.ErrorOrNil()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	flag.Parse()
	logrus.Debugf("localFlag: %t", *localFlag)

	// local-mode for lambda
	if *localFlag {
		err := Rotate(context.Background())
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		lambda.Start(Rotate)
	}
}
