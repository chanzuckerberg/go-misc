package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// STS is an STS client
type STS struct {
	Svc stsiface.STSAPI
}

// NewSTS returns an sts client
func NewSTS(c client.ConfigProvider, config *aws.Config) *STS {
	return &STS{Svc: sts.New(c, config)}
}
