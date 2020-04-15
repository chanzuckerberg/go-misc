package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
)

// Organizations is a organizations interface
type Organizations struct {
	Svc organizationsiface.OrganizationsAPI
}

// NewOrganizations will return Organizations
func NewOrganizations(c client.ConfigProvider, config *aws.Config) *Organizations {
	return &Organizations{Svc: organizations.New(c, config)}
}
