package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const example = `{"Records":[{"eventVersion":"1.05","userIdentity":{"type":"AssumedRole","principalId":"XXX","arn":"xxx","accountId":"XXX","accessKeyId":"xxx","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2019-01-01T02:50:42Z"},"sessionIssuer":{"type":"Role","principalId":"XXX","arn":"xxxx","accountId":"xxxx","userName":"xxx"}}},"eventTime":"2019-01-01T02:50:45Z","eventSource":"guardduty.amazonaws.com","eventName":"GetDetector","awsRegion":"xx","sourceIPAddress":"xxx","userAgent":"x","requestParameters":{"detectorId":"xxx"},"responseElements":null,"requestID":"","eventID":"xx","readOnly":true,"eventType":"AwsApiCall","recipientAccountId":"XX"},{"eventVersion":"1.05","userIdentity":{"type":"AssumedRole","principalId":"XXx","arn":"XXX","accountId":"123","accessKeyId":"XXX","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2019-01-01T02:50:18Z"},"sessionIssuer":{"type":"Role","principalId":"principal","arn":"role","accountId":"act","userName":"name"}}},"eventTime":"2019-01-01T02:50:19Z","eventSource":"ec2.amazonaws.com","eventName":"DescribeAccountAttributes","awsRegion":"","sourceIPAddress":"ip","userAgent":"user agent","requestParameters":{"accountAttributeNameSet":{"items":[{"attributeName":"supported-platforms"}]},"filterSet":{}},"responseElements":null,"requestID":"2","eventID":"2","eventType":"AwsApiCall","recipientAccountId":"2"}]}`

func TestUnfurl(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sess, serv := cziAws.NewMockSession()
	defer serv.Close()
	client, mockS3, mockS3Manager := cziAws.New(sess).WithMockS3(ctrl)

	data := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(data)
	_, err := gzipWriter.Write([]byte(example))
	a.Nil(err)
	err = gzipWriter.Close()
	a.Nil(err)

	mockS3Manager.EXPECT().
		DownloadWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(arg0 aws.Context, writer io.WriterAt, arg2 *s3.GetObjectInput, arg3 ...func(*s3manager.Downloader)) (int64, error) {
			_, err := writer.WriteAt(data.Bytes(), int64(0))
			return int64(0), err
		})

	putObjectOutput := &s3.PutObjectOutput{}
	mockS3.EXPECT().PutObjectWithContext(gomock.Any(), gomock.Any()).Return(putObjectOutput, nil)

	err = processRecord(context.Background(), client, "foo", "bar", "baz", "kms key", "prefix")
	a.Nil(err)
}
