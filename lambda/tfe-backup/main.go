package main

import (
	"context"
	"flag"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/go-misc/lambda/tfe-backup/runner"
	"github.com/getsentry/sentry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func run(ctx context.Context) error {
	config := &runner.Config{}
	envconfig.MustProcess("", config)

	flushSentry, err := runner.SetupSentry(config.SentryDSN, config.SentryEnv)
	if err != nil {
		return err
	}
	defer flushSentry()

	err = run0(ctx, config)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	return nil
}

func run0(ctx context.Context, config *runner.Config) error {
	sess, err := session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	)
	if err != nil {
		return errors.Wrap(err, "could not instantiate AWS session")
	}

	client := cziAWS.New(sess).WithSecretsManager(sess.Config).WithKMS(sess.Config)
	uploader := s3manager.NewUploader(sess)

	tfeToken, err := client.SecretsManager.ReadStringLatestVersion(ctx, config.TFETokenSecretARN)
	if err != nil {
		return errors.Wrapf(err, "failed to read secret %s", config.TFETokenSecretARN)
	}

	dataKey, err := runner.GenerateDataKey(ctx, client.KMS.Svc, config.KMSKeyARN)
	if err != nil {
		return err
	}

	tfeClient := runner.NewTFE(*tfeToken, config.TFEHostname)

	return tfeClient.Backup(ctx, uploader, dataKey, config)
}

func main() {
	local := flag.Bool("local", false, "Set to true if you intent to run this local or docker (not lambda).")
	flag.Parse()

	if local == nil || !*local {
		logrus.Info("executing in lambda environment")
		lambda.Start(run)
		return
	}

	logrus.Info("executing in local environment")
	err := run(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
}
