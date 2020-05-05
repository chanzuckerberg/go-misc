package oidc

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type AwsOIDCCredsProviderConfig struct {
	AWS struct {
		roleARN     string
		sessionName string
	}
	OIDC struct {
		clientID  string
		issuerURL string
	}
}

// NewAWSOIDCCredsProvider returns an AWS credential provider
// using OIDC.
func NewAwsOIDCCredsProvider(
	svc stsiface.STSAPI,
	conf *AwsOIDCCredsProviderConfig,
) credentials.Provider {

	tokenFetcher := &tokenFetcher{
		conf: conf,
	}

	return stscreds.NewWebIdentityRoleProviderWithToken(
		svc,
		conf.AWS.roleARN,
		conf.AWS.sessionName,
		tokenFetcher,
	)
}

type tokenFetcher struct {
	conf *AwsOIDCCredsProviderConfig
}

func (tf *tokenFetcher) FetchToken(ctx credentials.Context) ([]byte, error) {
	token, err := GetToken(ctx, tf.conf.OIDC.clientID, tf.conf.OIDC.issuerURL)
	if err != nil {
		return nil, err
	}
	return []byte(token.IDToken), nil
}
