package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

// ASG is an autoscaling svc
type ASG struct {
	Svc autoscalingiface.AutoScalingAPI
}

// NewASG returns a new autoscaling service
func NewASG(c client.ConfigProvider, config *aws.Config) *ASG {
	return &ASG{Svc: autoscaling.New(c, config)}
}
