package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

// This is a mock for the IAM Svc - mock more functions here as needed

// MockIAMSvc is a mock of IAM service
type MockIAMSvc struct {
	iamiface.IAMAPI
	Mock
}

// NewMockIAM returns a mock IAM SVC
func NewMockIAM() *MockIAMSvc {
	return &MockIAMSvc{}
}

// GetUser mocks getuser
func (i *MockIAMSvc) GetUser(in *iam.GetUserInput) (*iam.GetUserOutput, error) {
	resp, read := <-i.Resp
	if !read {
		panic("No response read - queue one up")
	}
	err, read := <-i.Errs
	if !read {
		panic("No err read - queue one up")
	}
	typed, ok := resp.(*iam.GetUserOutput)
	if !ok {
		panic("resp of wrong type")
	}
	return typed, err
}

// ListMFADevicesPages lists
func (i *MockIAMSvc) ListMFADevicesPages(in *iam.ListMFADevicesInput, fn func(*iam.ListMFADevicesOutput, bool) bool) error {
	resp, read := <-i.Resp
	if !read {
		panic("No response read - queue one up")
	}
	err, read := <-i.Errs
	if !read {
		panic("No err read - queue one up")
	}
	typed, ok := resp.(*iam.ListMFADevicesOutput)
	if !ok {
		panic("resp of wrong type")
	}

	if err != nil {
		return err
	}
	fn(typed, true)
	return nil
}
