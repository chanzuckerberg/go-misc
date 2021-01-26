package tfe

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/go-misc/ptr"
	"github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
)

func NewClientFromEnv(ctx context.Context) (*tfe.Client, error) {
	var tfeToken string

	if t := os.Getenv("TFE_TOKEN"); t != "" {
		tfeToken = t
	} else {
		tfeTokenARN := os.Getenv("TFE_TOKEN_SECRET_ARN")
		if tfeTokenARN == "" {
			return nil, errors.New("please set TFE_TOKEN_SECRET_ARN")
		}

		sess, err := session.NewSessionWithOptions(
			session.Options{
				SharedConfigState: session.SharedConfigEnable,
			},
		)

		if err != nil {
			return nil, err
		}

		awsClient := aws.New(sess).WithSecretsManager(sess.Config)

		token, err := awsClient.SecretsManager.ReadStringLatestVersion(ctx, tfeTokenARN)
		if err != nil {
			return nil, err
		}

		tfeToken = *token
	}
	config := tfe.DefaultConfig()
	config.Token = tfeToken

	return tfe.NewClient(config)
}

func ListWorkspacesInOrg(ctx context.Context, client *tfe.Client, organization string, include *string) ([]tfe.Workspace, error) {

	workspaces := []tfe.Workspace{}

	// https://www.terraform.io/docs/cloud/api/index.html#pagination
	page := 1

	// when there are no more pages, the api should return null, which gets the int zero value
	for page != 0 {
		ws, err := client.Workspaces.List(ctx, organization, tfe.WorkspaceListOptions{
			Include: ptr.String("current_run"),
			ListOptions: tfe.ListOptions{
				PageNumber: page,
			},
		})

		if err != nil {
			return nil, errors.Wrapf(err, "unable to list workspaces for %s", organization)
		}

		for _, workspace := range ws.Items {
			if workspace != nil {
				workspaces = append(workspaces, *workspace)
			}
		}
	}
	return workspaces, nil
}
