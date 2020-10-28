package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/go-misc/lambda/tfe-metrics/runner"
	"github.com/chanzuckerberg/go-misc/lambda/tfe-metrics/state"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

func run(ctx context.Context) {
	sentryDSN := os.Getenv("SENTRY_DSN")
	sentryEnv := os.Getenv("SENTRY_ENV")

	logrus.Info("setting up sentry")
	f, e := runner.SetupSentry(sentryDSN, sentryEnv)
	if e != nil {
		log.Fatal(e)
	}

	defer f()

	err := run0(ctx)

	if err != nil {
		logrus.Errorf("%+v\n", err)
		sentry.CaptureException(err)
	}
}

func run0(ctx context.Context) error {
	// TODO
	// ingress attributes support https://github.com/hashicorp/go-tfe/issues/79

	sess, err := session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	)

	if err != nil {
		return err
	}

	tfeTokenARN := os.Getenv("TFE_TOKEN_SECRET_ARN")
	if tfeTokenARN == "" {
		return errors.New("please set TFE_TOKEN_SECRET_ARN")
	}

	client := aws.New(sess).WithSecretsManager(sess.Config)

	token, err := client.SecretsManager.ReadStringLatestVersion(ctx, tfeTokenARN)
	if err != nil {
		return err
	}

	tfeToken := *token

	honeycombWriteKeyArn := os.Getenv("HONEYCOMB_WRITE_KEY_SECRET_ARN")
	if honeycombWriteKeyArn == "" {
		return errors.New("please set HONEYCOMB_WRITE_KEY_SECRET_ARN")
	}

	key, err := client.SecretsManager.ReadStringLatestVersion(ctx, honeycombWriteKeyArn)
	if err != nil {
		return err
	}

	honeycombWriteKey := *key

	honeycombDataset := os.Getenv("HONEYCOMB_DATASET")

	if honeycombDataset == "" {
		return errors.New("please set HONEYCOMB_DATASET")
	}

	backend := os.Getenv("BACKEND")

	if backend == "" {
		backend = "local"
	}

	if backend != "local" && backend != "dynamo" {
		return errors.New("backend must be one of 'local' or dynamo")
	}

	var dynamoTable string
	if backend == "dynamo" {
		dynamoTable = os.Getenv("DYNAMO_TABLE")
		if dynamoTable == "" {
			return errors.New("if backend=='dynamo', DYNAMO_TABLE env var must be set")
		}
	}

	var stater state.Stater

	if backend == "local" {
		stater = state.NewFileStater(".", "tfe-metrics")
	} else if backend == "dynamo" {
		sess, err := session.NewSession()
		if err != nil {
			return err
		}
		stater, err = state.NewDynamoDBStater(sess, dynamoTable)
		if err != nil {
			return err
		}
	}

	runner := runner.NewRunner(tfeToken, honeycombDataset, honeycombWriteKey, stater)

	return runner.RunOnce()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	flag.Parse()
	logrus.Debugf("arg: %s", flag.Arg(0))

	// cheap and simple local-mode for lambda
	if flag.Arg(0) == "-local" {
		run(context.Background())
	} else {
		lambda.Start(run)
	}
}
