package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Client is an aws client
type Client struct {
	session *session.Session

	IAM *IAM
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
func (c *Client) WithMockIAM() *Client {
	mock := &IAM{Svc: NewMockIAM()}
	c.IAM = mock

	return c
}
