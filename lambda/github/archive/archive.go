package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	envGithubSecret           = "GITHUB_SECRET"
	envFirehoseDeliveryStream = "FIREHOSE_DELIVERY_STREAM"
)

const (
	githubHeaderSignature = "X-Hub-Signature"
	githubHeaderEvent     = "X-GitHub-Event"
	githubHeaderDelivery  = "X-GitHub-Delivery"
)

type githubWebhook struct {
	signature string
	event     string
	id        string
	body      []byte
}

func newWebhook(event *events.APIGatewayProxyRequest) *githubWebhook {
	signature, ok := event.Headers[githubHeaderSignature]
	if !ok {
		logrus.Warn("no signature found")
		return nil
	}
	eventType, ok := event.Headers[githubHeaderEvent]
	if !ok {
		logrus.Warn("no event type found")
		return nil
	}
	id, ok := event.Headers[githubHeaderDelivery]
	if !ok {
		logrus.Warn("no id found")
		return nil
	}

	return &githubWebhook{
		signature: signature,
		event:     eventType,
		id:        id,
		body:      []byte(event.Body),
	}
}

func (w *githubWebhook) validate(secret []byte) error {
	err := github.ValidateSignature(w.signature, w.body, secret)
	return errors.Wrap(err, "could not validate webhook signature")
}

func processWebhook(ctx context.Context, awsClient *cziAWS.Client, event *events.APIGatewayProxyRequest) error {
	secret := os.Getenv(envGithubSecret)
	deliveryStream := os.Getenv(envFirehoseDeliveryStream)

	webhook := newWebhook(event)
	if webhook == nil {
		logrus.Info("nil webhook, nothing to do")
		return nil
	}

	// Do we trust this webhook?
	err := webhook.validate([]byte(secret))
	if err != nil {
		return errors.Wrap(err, "error validating webhook")
	}

	putInput := &firehose.PutRecordInput{
		DeliveryStreamName: aws.String(deliveryStream),
		Record: &firehose.Record{
			Data: webhook.body,
		},
	}
	_, err = awsClient.Firehose.Svc.PutRecordWithContext(ctx, putInput)
	return errors.Wrap(err, "error sending firehose record")
}

func handle(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		logrus.WithError(err).WithContext(ctx).Error("could not create aws session")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusTeapot}, nil
	}

	client := cziAWS.New(sess).WithFirehose(nil)
	err = processWebhook(ctx, client, &event)
	if err != nil {
		logrus.WithError(err).WithContext(ctx).Error("error handling webhook") // we do not return the actual error to not leak info
		return events.APIGatewayProxyResponse{StatusCode: http.StatusTeapot}, fmt.Errorf("Error")
	}
	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	logrus.Info("Processing started")
	lambda.Start(handle)
}
