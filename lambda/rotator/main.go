package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/go-misc/ptr"
	cziTfe "github.com/chanzuckerberg/go-misc/tfe"
	"github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// config options
// iam user name
// TODO access key minimum rotation age - don't rotate if younger than
// TODO access key maximum age - rotate if older than
// TODO access key minimum idle time

// target type(s)

// tfe
// tfe organization
// tfe access key secret arn

func run(ctx context.Context) error {
	iamUser := os.Getenv("IAM_USER")
	if iamUser == "" {
		return fmt.Errorf("IAM_USER env variable must be set")
	}

	sess, err := session.NewSession(aws.Config())
	if err != nil {
		logrus.WithError(err).WithContext(ctx).Error("could not create aws session")
		return errors.Wrap(err, "")
	}

	client := aws.New(sess).WithIAM(nil)

	// read iam user
	user, err := client.IAM.GetUser(ctx, ptr.String(iamUser))
	if err != nil {
		return errors.Wrap(err, "unable to read IAM user")
	}

	// read access keys
	keys, err := client.IAM.GetAccessKeysForUser(ctx, iamUser)
	if err != nil {
		return nil
	}

	target := os.Getenv("TARGET_TYPE")

	switch target {
	case "":
		return errors.New("Must set TARGET_TYPE")
	case "tfe":
		err = runTFE(ctx, client, user, keys)
	}

	return err
}

func runTFE(ctx context.Context, client *aws.Client, user *iam.User, keys []*iam.AccessKeyMetadata) error {

	tfeClient, err := cziTfe.NewClientFromEnv(ctx)
	if err != nil {
		return err
	}

	tfeOrg := os.Getenv("TFE_ORG")
	if tfeOrg == "" {
		return errors.New("please set TFE_ORG to the name of the organization")
	}

	ws, err := cziTfe.ListWorkspacesInOrg(ctx, tfeClient, tfeOrg, nil)
	if err != nil {
		return err
	}

	rotation := false

	for _, workspace := range ws {
		// NOTE - we assume there is <= 1 page worth of variables
		variables, err := tfeClient.Variables.List(ctx, workspace.ID, tfe.VariableListOptions{})
		if err != nil {
			return err
		}

		key := false
		secret := false

		for _, v := range variables.Items {
			if v != nil && v.Category == tfe.CategoryEnv {
				if v.Key == "AWS_ACCESS_KEY_ID" {
					rotation = true
				} else if v.Key == "AWS_SECRET_ACCESS_KEY" {
					rotation = true
				}
			}
		}

		// if at least one of these is missing, we need to rotate
		if !(key && secret) {
			rotation = true

			break
		}
	}

	if rotation {
		// delete oldest key
		oldestKey := keys[0]

		for _, k := range keys {
			if k.CreateDate.Before(*oldestKey.CreateDate) {
				oldestKey = k
			}
		}

		_, err = client.IAM.Svc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: oldestKey.AccessKeyId},
		)
		if err != nil {
			return errors.Wrapf(err, "unable to delete access key %s", *oldestKey.AccessKeyId)
		}

		// create new aws key
		newKey, err := client.IAM.Svc.CreateAccessKey(&iam.CreateAccessKeyInput{
			UserName: user.UserId,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to create new access key for %s", *user.UserId)
		}

		// push new credentials to TFE
		for _, workspace := range ws {

			// delete access key
			err := tfeClient.Variables.Delete(ctx, workspace.ID, "AWS_ACCESS_KEY_ID")
			if err != nil {
				return errors.Wrapf(err, "unable to delete AWS_ACCESS_KEY_ID in %s", workspace.Name)
			}

			// delete secret
			err = tfeClient.Variables.Delete(ctx, workspace.ID, "AWS_SECRET_ACCESS_KEY")
			if err != nil {
				return errors.Wrapf(err, "unable to delete AWS_SECRET_ACCESS_KEY in %s", workspace.Name)
			}

			// add new access key
			_, err = tfeClient.Variables.Create(ctx, workspace.ID, tfe.VariableCreateOptions{
				Key:       ptr.String("AWS_ACCESS_KEY_ID"),
				Value:     newKey.AccessKey.AccessKeyId,
				HCL:       ptr.Bool(false),
				Sensitive: ptr.Bool(false),
				Category:  tfe.Category(tfe.CategoryEnv),
			})

			if err != nil {
				return errors.Wrapf(err, "unable to create new AWS_ACCESS_KEY_ID in %s", workspace.Name)
			}

			// add new secret
			_, err = tfeClient.Variables.Create(ctx, workspace.ID, tfe.VariableCreateOptions{
				Key:       ptr.String("AWS_SECRET_ACCESS_KEY"),
				Value:     newKey.AccessKey.AccessKeyId,
				HCL:       ptr.Bool(false),
				Sensitive: ptr.Bool(true),
				Category:  tfe.Category(tfe.CategoryEnv),
			})

			if err != nil {
				return errors.Wrapf(err, "unable to create new AWS_SECRET_ACCESS_KEY in %s", workspace.Name)
			}
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
		err := run(context.Background())
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		lambda.Start(run)
	}
}
