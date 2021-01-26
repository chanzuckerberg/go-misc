package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chanzuckerberg/go-misc/ptr"
	"github.com/chanzuckerberg/go-misc/sentry"
	cziTfe "github.com/chanzuckerberg/go-misc/tfe"
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

	tfeClient, err := cziTfe.NewClientFromEnv(ctx)
	if err != nil {
		return err
	}

	org, err := tfeClient.Organizations.Read(ctx, tfeOrg)
	if err != nil {
		return errors.Wrap(err, "could not list TFE orgs")
	}
	logrus.Debugf("org: %v", org)

	var force bool

	if f := os.Getenv("FORCE"); f == "true" || f == "false" {
		force, _ = strconv.ParseBool(f)
	}

	workspaces, err := cziTfe.ListWorkspacesInOrg(ctx, tfeClient, org.Name, ptr.String("current_run"))
	if err != nil {
		return err
	}

	for _, workspace := range workspaces {
		if !force && time.Since(workspace.CurrentRun.CreatedAt) <= (24*time.Hour) {
			logrus.Debugf("skipping %s", workspace.Name)
			continue
		}
		logrus.Debugf("running workspace %#v", workspace.Name)
		logrus.Debugf("current run %#v", workspace.CurrentRun.CreatedAt)
		_, err := tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
			Message:   tfe.String("scheduled auto-run"),
			Workspace: &workspace,
		})

		if err != nil {
			return errors.Wrapf(err, "Unable to create run for %s", workspace.Name)
		}
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
