package oidc

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
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
	*credentials.Credentials

	fetcher *tokenFetcher
}

// FetchOIDCToken will fetch an oidc token
func (a *AWSOIDCCredsProvider) FetchOIDCToken(ctx context.Context) (*client.Token, error) {
	return a.fetcher.fetchFullToken(ctx)
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
		Credentials: credentials.NewCredentials(provider),
		fetcher:     tokenFetcher,
	}
}

type tokenFetcher struct {
	conf *AwsOIDCCredsProviderConfig
	mu   sync.Mutex
}

// safe for concurrent use
// fetches a full token
func (tf *tokenFetcher) fetchFullToken(ctx context.Context) (*client.Token, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	return GetToken(ctx, tf.conf.OIDC.clientID, tf.conf.OIDC.issuerURL)
}

func (tf *tokenFetcher) FetchToken(ctx context.Context) ([]byte, error) {
	token, err := tf.fetchFullToken(ctx)
	if err != nil {
		return nil, err
	}

	return []byte(token.IDToken), nil
}
