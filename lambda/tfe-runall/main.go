package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/go-misc/ptr"
	"github.com/chanzuckerberg/go-misc/sentry"
	"github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func run(ctx context.Context) {
	sentry.Run(ctx, run0)
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
	logrus.Debugf("org: %v", org)

	// https://www.terraform.io/docs/cloud/api/index.html#pagination
	page := 1

	// when there are no more pages, the api should return null, which gets the int zero value
	for page != 0 {
		workspaces, err := tfeClient.Workspaces.List(ctx, org.Name, tfe.WorkspaceListOptions{
			Include: ptr.String("current_run"),
		})

		if err != nil {
			return errors.Wrapf(err, "unable to list workspaces for %s", org.Name)
		}

		for _, workspace := range workspaces.Items {
			if time.Since(workspace.CurrentRun.CreatedAt) <= (24 * time.Hour) {
				logrus.Debugf("skipping %s", workspace.Name)
				continue
			}
			logrus.Debugf("running workspace %#v", workspace.Name)
			logrus.Debugf("current run %#v", workspace.CurrentRun.CreatedAt)
			_, err := tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
				Message:   tfe.String("scheduled auto-run"),
				Workspace: workspace,
			})

			if err != nil {
				return errors.Wrapf(err, "Unable to create run for %s", workspace.Name)
			}

		}

		page = workspaces.NextPage
	}

	return nil
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
