package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
)

type cloudtrail struct {
	Records []map[string]interface{} `json:"records,omitempty"`
}

func processRecord(
	ctx context.Context,
	awsClient *cziAWS.Client,
	sourceBucket string,
	key string,
	destinationBucket string) error {

	getObject := &s3.GetObjectInput{}
	output, err := awsClient.S3.Svc.GetObjectWithContext(ctx, getObject)
	if err != nil {
		return errors.Wrapf(err, "Could not get %s/%s", sourceBucket, key)
	}

	gzipReader, err := gzip.NewReader(output.Body)
	if err != nil {
		return errors.Wrap(err, "Could not create gzip reader")
	}
	defer gzipReader.Close()

	data, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return errors.Wrap(err, "Could not read all data")
	}
	parsed := &cloudtrail{}
	err = json.Unmarshal(data, parsed)
	if err != nil {
		return errors.Wrap(err, "Error unmarshalling data")
	}

	outputData := bytes.NewBuffer(nil)
	outputGzipWriter := gzip.NewWriter(outputData)
	for _, record := range parsed.Records {
		line, err := json.Marshal(record)
		if err != nil {
			return errors.Wrap(err, "Error marshalling sub-record")
		}
		_, err = outputGzipWriter.Write(line)
		if err != nil {
			errors.Wrap(err, "Error writing line")
		}
		_, err = outputGzipWriter.Write([]byte{byte('\n')})
		if err != nil {
			return errors.Wrap(err, "Error writing newline")
		}
	}

	err = outputGzipWriter.Close()
	if err != nil {
		return errors.Wrap(err, "Could not finalize gzip archive")
	}

	outputBytes := outputData.Bytes()
	putObjectInput := &s3.PutObjectInput{
		Bucket:               aws.String(destinationBucket),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		ContentLength:        aws.Int64(int64(len(outputBytes))),
		ContentType:          aws.String(http.DetectContentType(outputBytes)),
		Body:                 bytes.NewReader(outputBytes),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	}

	_, err = awsClient.S3.Svc.PutObjectWithContext(ctx, putObjectInput)
	return errors.Wrapf(err, "Error uploading to %s/%s", destinationBucket, key)
}

func handler(ctx context.Context, s3Event events.S3Event) (err error) {
	destinationBucket := os.Getenv("DESTINATION_BUCKET")

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return errors.Wrap(err, "Could not create aws session")
	}

	client := cziAWS.New(sess).WithS3(nil)
	for _, event := range s3Event.Records {
		err = processRecord(ctx, client, event.S3.Bucket.Name, event.S3.Object.Key, destinationBucket)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	logrus.Info("Processing started")
	lambda.Start(handler)
}
