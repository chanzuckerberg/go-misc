package oidc

import (
	"context"
	"sync"

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

// AWSOIDCCredsProvider providers OIDC tokens and aws:STS credentials
type AWSOIDCCredsProvider struct {
	credentials.ProviderWithContext

	fetcher *tokenFetcher
}

// FetchOIDCToken will fetch an oidc token
func (a *AWSOIDCCredsProvider) FetchOIDCToken(ctx context.Context) ([]byte, error) {
	return a.fetcher.FetchToken(ctx)
}

// NewAWSOIDCCredsProvider returns an AWS credential provider
// using OIDC.
func NewAwsOIDCCredsProvider(
	svc stsiface.STSAPI,
	conf *AwsOIDCCredsProviderConfig,
) *AWSOIDCCredsProvider {

	tokenFetcher := &tokenFetcher{
		conf: conf,
	}

	provider := stscreds.NewWebIdentityRoleProviderWithToken(
		svc,
		conf.AWS.roleARN,
		conf.AWS.sessionName,
		tokenFetcher,
	)

	return &AWSOIDCCredsProvider{
		ProviderWithContext: provider,
		fetcher:             tokenFetcher,
	}
}

type tokenFetcher struct {
	conf *AwsOIDCCredsProviderConfig
	mu   sync.Mutex
}

// safe for concurrent use
func (tf *tokenFetcher) FetchToken(ctx credentials.Context) ([]byte, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	token, err := GetToken(ctx, tf.conf.OIDC.clientID, tf.conf.OIDC.issuerURL)
	if err != nil {
		return nil, err
	}
	return []byte(token.IDToken), nil
}
