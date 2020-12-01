package tftest

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
)

var multipleActionsResourcesPolicy = S3BucketPolicy{
	Statements: []Statement{
		{
			Sid:       "",
			Effect:    "Allow",
			Principal: "*",
			Action:    AWSStrings{"sts:AssumeRoleWithWebIdentity", "sts:AnotherAction"},
			Resource:  AWSStrings{"Resource1", "Resource2"},
			Condition: map[string]map[string]string{
				"StringEquals": {"localhost:aud": "clientIDValue3"},
			},
		},
	},
}

func policyDocumentToString(policyDoc *S3BucketPolicy) *string {
	jsonPolicyData, err := json.Marshal(policyDoc)
	if err != nil {
		panic(err)
	}
	return aws.String(url.PathEscape(string(jsonPolicyData)))
}

// TODO(aku): Finish this test!
func TestParseMultipleActionsPrincipals(t *testing.T) {
	r := require.New(t)

	policyData, err := json.Marshal(multipleActionsResourcesPolicy)
	r.NoError(err)

	policyDoc := S3BucketPolicy{}
	err = json.Unmarshal(policyData, &policyDoc)
	r.NoError(err)

	r.NotEmpty(policyDoc)
	r.Len(policyDoc.Statements, 1)
	r.Len(policyDoc.Statements[0].Action, 2)
	r.Equal(policyDoc.Statements[0].Action, AWSStrings{"sts:AssumeRoleWithWebIdentity", "sts:AnotherAction"})
	r.Len(policyDoc.Statements[0].Resource, 2)
	r.Equal(policyDoc.Statements[0].Resource, AWSStrings{"Resource1", "Resource2"})
}

type AlternateStatementEntry struct {
	Effect    string                       `json:"Effect"`
	Action    AWSStrings                   `json:"Action"`
	Sid       string                       `json:"Sid"`
	Principal string                       `json:"Principal"`
	Resource  AWSStrings                   `json:"Resource"`
	Condition map[string]map[string]string `json:"Condition"`
}
type AlternatePolicyDocument struct {
	Version    string                    `json:"Version"`
	Statements []AlternateStatementEntry `json:"Statement"`
}

var singleActionResourcePolicy = &AlternatePolicyDocument{
	Statements: []AlternateStatementEntry{
		{
			Effect:    "Allow",
			Action:    AWSStrings{"sts:AssumeRoleWithWebIdentity"},
			Sid:       "",
			Principal: "*",
			Resource:  AWSStrings{"Resource1"},
			Condition: map[string]map[string]string{
				"StringEquals": {"localhost:aud": "clientIDValue4"},
			},
		},
	},
}

func TestSingleStringAction(t *testing.T) {
	r := require.New(t)

	policyData, err := json.Marshal(singleActionResourcePolicy)
	r.NoError(err)

	policyDoc := AlternatePolicyDocument{}
	err = json.Unmarshal(policyData, &policyDoc)
	r.NoError(err)

	r.NotEmpty(policyDoc)
	r.Len(policyDoc.Statements, 1)
	r.Len(policyDoc.Statements[0].Action, 1)
	r.Equal(policyDoc.Statements[0].Action, AWSStrings{"sts:AssumeRoleWithWebIdentity"})
	r.Len(policyDoc.Statements[0].Resource, 1)
	r.Equal(policyDoc.Statements[0].Resource, AWSStrings{"Resource1"})
}
