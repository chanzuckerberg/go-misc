package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/go-misc/aws/mocks"
	"github.com/golang/mock/gomock"
)

// Client is an aws client
type Client struct {
	session *session.Session

	// services
	ASG            *ASG
	EC2            *EC2
	IAM            *IAM
	KMS            *KMS
	Lambda         *Lambda
	S3             *S3
	Firehose       *Firehose
	SecretsManager *SecretsManager
	SSM            *SSM
	STS            *STS
	Support        *Support
}

// New returns a new aws client
func New(sess *session.Session) *Client {
	return &Client{session: sess}
}

// WithAllServices Convenience method that configures all services with the same aws.Config
func (c *Client) WithAllServices(conf *aws.Config) *Client {
	return c.
		WithASG(conf).
		WithEC2(conf).
		WithIAM(conf).
		WithKMS(conf).
		WithLambda(conf).
		WithS3(conf).
		WithFirehose(conf).
		WithSecretsManager(conf).
		WithSSM(conf).
		WithSTS(conf).
		WithSupport(conf)
}

// ------- Autoscaling -----------

// WithASG configures an autoscaling Svc
func (c *Client) WithASG(conf *aws.Config) *Client {
	c.ASG = NewASG(c.session, conf)
	return c
}

// WithMockASG mocks the ASG svc
func (c *Client) WithMockASG(ctrl *gomock.Controller) (*Client, *mocks.MockAutoScalingAPI) {
	mock := mocks.NewMockAutoScalingAPI(ctrl)
	c.ASG = &ASG{Svc: mock}
	return c, mock
}

// ------- SecretsManager -----------

// WithSecretsManager configures a Secrets Manager svc
func (c *Client) WithSecretsManager(conf *aws.Config) *Client {
	c.SecretsManager = NewSecretsManager(c.session, conf)
	return c
}

// WithMockSecretsManager mocks the Secrets Manager svc
func (c *Client) WithMockSecretsManager(ctrl *gomock.Controller) (*Client, *mocks.MockSecretsManagerAPI) {
	mock := mocks.NewMockSecretsManagerAPI(ctrl)
	c.SecretsManager = &SecretsManager{Svc: mock}
	return c, mock
}

// ------- Firehose -----------

// WithFirehose configures the firehose service
func (c *Client) WithFirehose(conf *aws.Config) *Client {
	c.Firehose = NewFirehose(c.session, conf)
	return c
}

func (c *Client) WithMockFirehose(ctrl *gomock.Controller) (*Client, *mocks.MockFirehoseAPI) {
	mock := mocks.NewMockFirehoseAPI(ctrl)
	c.Firehose = &Firehose{Svc: mock}
	return c, mock
}

// ------- S3 -----------

// WithS3 configures the s3 client
func (c *Client) WithS3(conf *aws.Config) *Client {
	c.S3 = NewS3(c.session, conf)
	return c
}

// WithMockS3 mocks s3 svc
func (c *Client) WithMockS3(ctrl *gomock.Controller) (*Client, *mocks.MockS3API, *mocks.MockDownloaderAPI) {
	mock := mocks.NewMockS3API(ctrl)
	mockDownloader := mocks.NewMockDownloaderAPI(ctrl)

	c.S3 = &S3{Svc: mock, Downloader: mockDownloader}
	return c, mock, mockDownloader
}

// ------- IAM -----------

// WithIAM configures the IAM SVC
func (c *Client) WithIAM(conf *aws.Config) *Client {
	c.IAM = NewIAM(c.session, conf)
	return c
}

// WithMockIAM mocks iam svc
func (c *Client) WithMockIAM(ctrl *gomock.Controller) (*Client, *mocks.MockIAMAPI) {
	mock := mocks.NewMockIAMAPI(ctrl)
	c.IAM = &IAM{Svc: mock}
	return c, mock
}

// ------- SSM -----------

// WithSSM configures the SSM service
func (c *Client) WithSSM(conf *aws.Config) *Client {
	c.SSM = NewSSM(c.session, conf)
	return c
}

// WithMockSSM mocks the SSM service
func (c *Client) WithMockSSM(ctrl *gomock.Controller) (*Client, *mocks.MockSSMAPI) {
	mock := mocks.NewMockSSMAPI(ctrl)
	c.SSM = &SSM{Svc: mock}
	return c, mock
}

// ------- STS -----------

// WithSTS configures the STS service
func (c *Client) WithSTS(conf *aws.Config) *Client {
	c.STS = NewSTS(c.session, conf)
	return c
}

// WithMockSTS mocks the STS service
func (c *Client) WithMockSTS(ctrl *gomock.Controller) (*Client, *mocks.MockSTSAPI) {
	mock := mocks.NewMockSTSAPI(ctrl)
	c.STS = &STS{Svc: mock}
	return c, mock
}

// ------- Lambda -----------

// WithLambda configures the lambda service
func (c *Client) WithLambda(conf *aws.Config) *Client {
	c.Lambda = NewLambda(c.session, conf)
	return c
}

// WithMockLambda mocks the lambda service
func (c *Client) WithMockLambda(ctrl *gomock.Controller) (*Client, *mocks.MockLambdaAPI) {
	mock := mocks.NewMockLambdaAPI(ctrl)
	c.Lambda = &Lambda{Svc: mock}
	return c, mock
}

// ------- KMS -----------

// WithKMS configures the kms service
func (c *Client) WithKMS(conf *aws.Config) *Client {
	c.KMS = NewKMS(c.session, conf)
	return c
}

// WithMockKMS mocks the kms service
func (c *Client) WithMockKMS(ctrl *gomock.Controller) (*Client, *mocks.MockKMSAPI) {
	mock := mocks.NewMockKMSAPI(ctrl)
	c.KMS = &KMS{Svc: mock}
	return c, mock
}

// ------- EC2 -----------

// WithEC2 configures an EC2 svc
func (c *Client) WithEC2(conf *aws.Config) *Client {
	c.EC2 = NewEC2(c.session, conf)
	return c
}

func (c *Client) WithMockEC2(ctrl *gomock.Controller) (*Client, *mocks.MockEC2API) {
	mock := mocks.NewMockEC2API(ctrl)
	c.EC2 = &EC2{Svc: mock}
	return c, mock
}

// ------- Support -----------

// WithSupport configures an Support svc
func (c *Client) WithSupport(conf *aws.Config) *Client {
	c.Support = NewSupport(c.session, conf)
	return c
}

func (c *Client) WithMockSupport(ctrl *gomock.Controller) (*Client, *mocks.MockSupportAPI) {
	mock := mocks.NewMockSupportAPI(ctrl)
	c.Support = &Support{Svc: mock}
	return c, mock
}
