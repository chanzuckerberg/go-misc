package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func processRecord(
	ctx context.Context,
	awsClient *cziAWS.Client,
	sourceBucket string,
	key string,
	destinationBucket string,
	destinationKMSKeyID string,
	destinationPrefix string) error {

	getObject := &s3.GetObjectInput{
		Bucket: aws.String(sourceBucket),
		Key:    aws.String(key),
	}
	downloadData := aws.NewWriteAtBuffer([]byte{})
	_, err := awsClient.S3.Downloader.DownloadWithContext(ctx, downloadData, getObject)
	if err != nil {
		return errors.Wrapf(err, "Could not get %s/%s", sourceBucket, key)
	}
	gzipReader, err := gzip.NewReader(bytes.NewBuffer(downloadData.Bytes()))
	if err != nil {
		return errors.Wrap(err, "Could not create gzip reader")
	}
	defer gzipReader.Close()

	data, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return errors.Wrap(err, "Could not read all data")
	}
	parsed := map[string]interface{}{}
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return errors.Wrap(err, "Error unmarshalling data")
	}

	outputData := bytes.NewBuffer(nil)
	outputGzipWriter := gzip.NewWriter(outputData)
	defer outputGzipWriter.Close()

	// TODO(el): This will ignore the digest files, do we care about that?
	records, ok := parsed["Records"]
	if !ok {
		logrus.Infof("Malformed event, skipping. 'Records' key not present")
		return nil
	}

	recordList, ok := records.([]interface{})
	if !ok {
		logrus.Infof("Malformed event, skipping. Records not a list")
		return nil
	}

	line := []byte{} //nolint
	for _, record := range recordList {
		line, err = json.Marshal(record)
		if err != nil {
			return errors.Wrap(err, "Error marshalling sub-record")
		}
		_, err = outputGzipWriter.Write(line)
		if err != nil {
			return errors.Wrap(err, "Error writing line")
		}
		_, err = outputGzipWriter.Write([]byte{byte('\n')})
		if err != nil {
			return errors.Wrap(err, "Error writing newline")
		}
	}

	// Explicitly close here since we're done with the data
	// Ok to call close twice (with defer)
	err = outputGzipWriter.Close()
	if err != nil {
		return errors.Wrap(err, "Could not finalize gzip archive")
	}

	destinationKey := path.Join(destinationPrefix, key)
	outputBytes := outputData.Bytes()
	putObjectInput := &s3.PutObjectInput{
		Bucket:               aws.String(destinationBucket),
		Key:                  aws.String(destinationKey),
		ACL:                  aws.String("private"),
		ContentLength:        aws.Int64(int64(len(outputBytes))),
		ContentType:          aws.String(http.DetectContentType(outputBytes)),
		Body:                 bytes.NewReader(outputBytes),
		ServerSideEncryption: aws.String("aws:kms"),
		SSEKMSKeyId:          aws.String(destinationKMSKeyID),
	}

	_, err = awsClient.S3.Svc.PutObjectWithContext(ctx, putObjectInput)
	return errors.Wrapf(err, "Error uploading to %s", destinationKey)
}

func handler(ctx context.Context, s3Event events.S3Event) (err error) {
	destinationBucket := os.Getenv("DESTINATION_BUCKET")
	destinationKMSKeyID := os.Getenv("DESTINATION_KMS_KEY_ID")
	destinationPrefix := os.Getenv("DESTINATION_PREFIX")

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return errors.Wrap(err, "Could not create aws session")
	}
	client := cziAWS.New(sess).WithS3(nil)
	for _, event := range s3Event.Records {
		err = processRecord(
			ctx,
			client,
			event.S3.Bucket.Name,
			event.S3.Object.Key,
			destinationBucket,
			destinationKMSKeyID,
			destinationPrefix,
		)
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
