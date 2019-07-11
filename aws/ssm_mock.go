package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/mock"
)

// This is a mock for the SSM service - mock more functions here as needed

// MockSSMSvc is a mock STS service
type MockSSMSvc struct {
	ssmiface.SSMAPI
	mock.Mock
}

// NewMockSTS returns a new mock sts svc
func NewMockSSM() *MockSSMSvc {
	return &MockSSMSvc{}
}

// // GetSessionTokenWithContext mocks GetSessionToken
// func (s *MockSTSSvc) GetSessionTokenWithContext(ctx aws.Context, in *sts.GetSessionTokenInput, ro ...request.Option) (*sts.GetSessionTokenOutput, error) {
// 	args := s.Called(in)
// 	out := args.Get(0).(*sts.GetSessionTokenOutput)
// 	return out, args.Error(1)
// }

// // GetCallerIdentityWithContext mocks GetCallerIdentity
// func (s *MockSTSSvc) GetCallerIdentityWithContext(ctx aws.Context, in *sts.GetCallerIdentityInput, ro ...request.Option) (*sts.GetCallerIdentityOutput, error) {
// 	args := s.Called(in)
// 	out := args.Get(0).(*sts.GetCallerIdentityOutput)
// 	return out, args.Error(1)
// }
