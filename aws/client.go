package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Client is an aws client
type Client struct {
	session *session.Session

	IAM    *IAM
	STS    *STS
	Lambda *Lambda
}

// New returns a new aws client
func New(sess *session.Session) *Client {
	return &Client{session: sess}
}

// WithIAM configures the IAM SVC
func (c *Client) WithIAM(conf *aws.Config) *Client {
	c.IAM = NewIAM(c.session, conf)
	return c
}

// WithMockIAM mocks iam svc
func (c *Client) WithMockIAM() (*Client, *MockIAMSvc) {
	mock := NewMockIAM()
	c.IAM = &IAM{Svc: mock}
	return c, mock
}

// WithSTS configures the STS service
func (c *Client) WithSTS(conf *aws.Config) *Client {
	c.STS = NewSTS(c.session, conf)
	return c
}

// WithMockSTS mocks the STS service
func (c *Client) WithMockSTS() (*Client, *MockSTSSvc) {
	mock := NewMockSTS()
	c.STS = &STS{Svc: mock}
	return c, mock
}

// WithLambda configures the lambda service
func (c *Client) WithLambda(conf *aws.Config) *Client {
	c.Lambda = NewLambda(c.session, conf)
	return c
}

// WithMockLambda mocks the lambda service
func (c *Client) WithMockLambda() (*Client, *MockLambdaSvc) {
	mock := NewMockLambda()
	c.Lambda = &Lambda{Svc: mock}
	return c, mock
}
