package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const example = `{"Records":[{"eventVersion":"1.05","userIdentity":{"type":"AssumedRole","principalId":"XXX","arn":"xxx","accountId":"XXX","accessKeyId":"xxx","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2019-01-01T02:50:42Z"},"sessionIssuer":{"type":"Role","principalId":"XXX","arn":"xxxx","accountId":"xxxx","userName":"xxx"}}},"eventTime":"2019-01-01T02:50:45Z","eventSource":"guardduty.amazonaws.com","eventName":"GetDetector","awsRegion":"xx","sourceIPAddress":"xxx","userAgent":"x","requestParameters":{"detectorId":"xxx"},"responseElements":null,"requestID":"","eventID":"xx","readOnly":true,"eventType":"AwsApiCall","recipientAccountId":"XX"},{"eventVersion":"1.05","userIdentity":{"type":"AssumedRole","principalId":"XXx","arn":"XXX","accountId":"123","accessKeyId":"XXX","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2019-01-01T02:50:18Z"},"sessionIssuer":{"type":"Role","principalId":"principal","arn":"role","accountId":"act","userName":"name"}}},"eventTime":"2019-01-01T02:50:19Z","eventSource":"ec2.amazonaws.com","eventName":"DescribeAccountAttributes","awsRegion":"","sourceIPAddress":"ip","userAgent":"user agent","requestParameters":{"accountAttributeNameSet":{"items":[{"attributeName":"supported-platforms"}]},"filterSet":{}},"responseElements":null,"requestID":"2","eventID":"2","eventType":"AwsApiCall","recipientAccountId":"2"}]}`

func TestUnfurl(t *testing.T) {
	a := assert.New(t)

	sess, serv := cziAws.NewMockSession()
	defer serv.Close()
	client, mockS3, mockS3Manager := cziAws.New(sess).WithMockS3()

	data := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(data)
	_, err := gzipWriter.Write([]byte(example))
	a.Nil(err)
	err = gzipWriter.Close()
	a.Nil(err)

	putObjectOutput := &s3.PutObjectOutput{}

	mockS3Manager.On("DownloadWithContext", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		writer := args.Get(0).(io.WriterAt)
		writer.WriteAt(data.Bytes(), int64(0))
	}).Return(int64(0), nil)

	mockS3.On("PutObjectWithContext", mock.Anything).Return(putObjectOutput, nil)

	err = processRecord(context.Background(), client, "foo", "bar", "baz", "kms key", "prefix")
	a.Nil(err)
}
