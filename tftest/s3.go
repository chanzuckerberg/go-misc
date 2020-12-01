package tftest

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type AWSStrings []string

type Statement struct {
	Sid       string
	Effect    string
	Principal string
	Action    AWSStrings
	Resource  AWSStrings
	Condition map[string]map[string]string
}

type S3BucketPolicy struct {
	Statements []Statement
}

var UserArn = "arn:aws:iam::119435350371:user/ci/cztack-ci"

// General Unmarshal function for values that could be a string or []string, unmarshal as []string
func (a *AWSStrings) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		*a = []string{str}
		return nil
	}
	// If the error is not an unmarshal type error, then we return the error
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return errors.Wrap(err, "Unexpected error type from unmarshaling")
	}

	var strSlice []string
	err = json.Unmarshal(data, &strSlice)
	if err == nil {
		*a = strSlice
		return nil
	}
	return errors.Wrap(err, "Unable to unmarshal Action")
}

// UnmarshalS3BucketPolicy will parse an s3 bucket policy and return as a go struct. Only parts that
// have been used are supported so far
func UnmarshalS3BucketPolicy(in string) (*S3BucketPolicy, error) {
	p := &S3BucketPolicy{}
	err := json.Unmarshal([]byte(in), p)
	return p, err
}

// S3S3SimulateRequest uses the IAM policy simulator to run end-to-end tests on permissions
func S3SimulateRequest(t *testing.T, region, action, bucketArn, bucketPolicy string, secureTransport bool) *iam.EvaluationResult {
	r := require.New(t)

	iamClient := aws.NewIamClient(t, region)

	simRequest := &iam.SimulatePrincipalPolicyInput{
		ActionNames: []*string{&action},
		CallerArn:   &UserArn,
		ContextEntries: []*iam.ContextEntry{
			{
				ContextKeyName:   Strptr("aws:securetransport"),
				ContextKeyType:   Strptr("boolean"),
				ContextKeyValues: []*string{Strptr(strconv.FormatBool(secureTransport))},
			},
		},
		ResourceArns:    []*string{&bucketArn},
		PolicySourceArn: &UserArn,
		ResourcePolicy:  &bucketPolicy,
	}

	resp, err := iamClient.SimulatePrincipalPolicy(simRequest)
	r.NoError(err)
	r.NotNil(resp.EvaluationResults)
	r.Len(resp.EvaluationResults, 1)

	return resp.EvaluationResults[0]
}
