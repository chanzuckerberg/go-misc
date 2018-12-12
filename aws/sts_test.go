package aws_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite

	mockIAM *cziAws.MockIAMSvc
	mockSTS *cziAws.MockSTSSvc
	client  *cziAws.Client

	provider  *cziAws.UserTokenProvider
	cachePath string
	creds     *sts.Credentials

	pathsToRemove []string
	server        *httptest.Server
}

func (ts *ProviderTestSuite) TearDownTest() {
	for _, pathToRemove := range ts.pathsToRemove {
		os.RemoveAll(pathToRemove)
	}
	ts.server.Close()
}

func (ts *ProviderTestSuite) SetupTest() {
	t := ts.T()
	a := assert.New(t)

	f, err := ioutil.TempFile("", "cache")
	a.Nil(err)
	defer f.Close()
	ts.cachePath = f.Name()

	sess, serv := cziAws.NewMockSession()
	client, mockIAM := cziAws.New(sess).WithMockIAM()
	client, mockSTS := client.WithMockSTS()

	tokenProvider := func() (string, error) {
		return "mytoken", nil
	}

	ts.server = serv
	ts.client = client
	ts.mockIAM = mockIAM
	ts.mockSTS = mockSTS
	ts.provider = cziAws.NewUserTokenProvider(f.Name(), client, tokenProvider)

	user := &iam.User{}
	user.SetArn("my user arn").SetUserName("my username")
	getUserOutput := &iam.GetUserOutput{}
	getUserOutput.SetUser(user)

	ts.mockIAM.On("GetUserWithContext", mock.Anything).Return(getUserOutput, nil)
	mfaDevivies := []*iam.MFADevice{
		&iam.MFADevice{SerialNumber: aws.String("serial number")},
	}
	output := &iam.ListMFADevicesOutput{}
	output.SetMFADevices(mfaDevivies)
	ts.mockIAM.On("ListMFADevicesPagesWithContext", mock.Anything).Return(output, nil)

	creds := &sts.Credentials{}
	creds.SetAccessKeyId("access key id").SetExpiration(time.Now().Add(time.Hour)).SetSecretAccessKey("secret").SetSessionToken("Token")
	token := &sts.GetSessionTokenOutput{}
	token.SetCredentials(creds)
	ts.mockSTS.On("GetSessionTokenWithContext", mock.Anything).Return(token, nil)
	ts.creds = creds

	callerIdentity := &sts.GetCallerIdentityOutput{}
	callerIdentity.SetAccount("my account id")
	callerIdentity.SetArn("my user arn")
	callerIdentity.SetUserID("my user id")
	ts.mockSTS.On("GetCallerIdentity", mock.Anything).Return(callerIdentity, nil)

	ts.pathsToRemove = []string{f.Name()}
}

func (ts *ProviderTestSuite) TestNoCache() {
	t := ts.T()
	a := assert.New(t)
	err := os.Remove(ts.cachePath)
	a.Nil(err)
	c, err := ts.provider.Retrieve()
	a.Nil(err)
	a.Equal(*ts.creds.AccessKeyId, c.AccessKeyID)
}

func (ts *ProviderTestSuite) TestCacheCorrupted() {
	t := ts.T()
	a := assert.New(t)
	err := ioutil.WriteFile(ts.cachePath, []byte("corrupted"), 0644)
	a.Nil(err)
	c, err := ts.provider.Retrieve()
	a.Nil(err)
	a.Equal(*ts.creds.AccessKeyId, c.AccessKeyID)
}

func (ts *ProviderTestSuite) TestCached() {
	t := ts.T()
	a := assert.New(t)
	tokenCache := &cziAws.UserTokenProviderCache{
		Expiration:      aws.Time(time.Now().Add(time.Hour)),
		AccessKeyID:     aws.String("aki;"),
		SecretAccessKey: aws.String("sak"),
		SessionToken:    aws.String("st"),
	}
	b, err := json.Marshal(tokenCache)
	a.Nil(err)
	err = ioutil.WriteFile(ts.cachePath, b, 0644)
	a.Nil(err)
	c, err := ts.provider.Retrieve()
	a.Nil(err)
	a.Equal(*tokenCache.AccessKeyID, c.AccessKeyID)
	a.True(ts.mockIAM.Mock.AssertNotCalled(t, "GetUserWithContext", mock.Anything))
}

func (ts *ProviderTestSuite) TestCacheExpired() {
	t := ts.T()
	a := assert.New(t)
	tokenCache := &cziAws.UserTokenProviderCache{
		Expiration:      aws.Time(time.Now().Add(-1 * time.Hour)),
		AccessKeyID:     aws.String("aki;"),
		SecretAccessKey: aws.String("sak"),
		SessionToken:    aws.String("st"),
	}
	b, err := json.Marshal(tokenCache)
	a.Nil(err)
	err = ioutil.WriteFile(ts.cachePath, b, 0644)
	a.Nil(err)
	c, err := ts.provider.Retrieve()
	a.Nil(err)
	a.Equal(*ts.creds.AccessKeyId, c.AccessKeyID)
}

func (ts *ProviderTestSuite) TestGetCallerIdentity() {
	t := ts.T()
	a := assert.New(t)
	input := &sts.GetCallerIdentityInput{}
	res, err := ts.client.STS.GetCallerIdentity(input)
	a.Equal(res.UserId, "my user id")
}

func TestSTSProviderSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}
