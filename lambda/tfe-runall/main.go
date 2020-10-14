package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/go-misc/aws"
	czi_sentry "github.com/chanzuckerberg/go-misc/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func run(ctx context.Context) {
	sentryDSN := os.Getenv("SENTRY_DSN")
	sentryEnv := os.Getenv("SENTRY_ENV")

	logrus.Info("setting up sentry")

	f, e := czi_sentry.Setup(sentryDSN, sentryEnv)
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
	tfeOrg := os.Getenv("TFE_ORG")
	if tfeOrg == "" {
		return errors.New("please set TFE_ORG to the name of the organization")
	}

	tfeTokenARN := os.Getenv("TFE_TOKEN_SECRET_ARN")
	if tfeTokenARN == "" {
		return errors.New("please set TFE_TOKEN_SECRET_ARN")
	}

	sess, err := session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	)

	if err != nil {
		return err
	}

	awsClient := aws.New(sess).WithSecretsManager(sess.Config)

	token, err := awsClient.SecretsManager.ReadStringLatestVersion(ctx, tfeTokenARN)
	if err != nil {
		return err
	}

	tfeToken := *token

	config := tfe.DefaultConfig()
	config.Token = tfeToken

	tfeClient, err := tfe.NewClient(config)

	if err != nil {
		return errors.Wrap(err, "unable to create tfe client")
	}

	org, err := tfeClient.Organizations.Read(ctx, tfeOrg)
	if err != nil {
		return errors.Wrap(err, "could not list TFE orgs")
	}
	logrus.Debugf("org: %s", org)

	// https://www.terraform.io/docs/cloud/api/index.html#pagination
	page := 1

	for page != 0 {
		workspaces, err := tfeClient.Workspaces.List(ctx, org.Name, tfe.WorkspaceListOptions{})

		if err != nil {
			return errors.Wrapf(err, "unable to list workspaces for %s", org.Name)
		}

		for _, workspace := range workspaces.Items {
			logrus.Infof("workspace %#v", workspace.Name)
			tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
				Message:   tfe.String("scheduled auto-run"),
				Workspace: workspace,
			})
		}

		// when there are no more pages, the api should return null, which gets the int zero value
		page = workspaces.NextPage

	}

	return nil

}

func main() {
	flag.Parse()

	logrus.Debugf("arg: %s", flag.Arg(0))
	if flag.Arg(0) == "-local" {
		run(context.Background())
	} else {
		lambda.Start(run)
	}

}
