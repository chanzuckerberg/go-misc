package oidc

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/chanzuckerberg/go-misc/oidc_cli/v3/oidc_impl"
	"github.com/chanzuckerberg/go-misc/oidc_cli/v3/oidc_impl/client"
)

type AwsOIDCCredsProviderConfig struct {
	AWSRoleARN    string
	OIDCClientID  string
	OIDCIssuerURL string
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
	ctx context.Context,
	svc stsiface.STSAPI,
	conf *AwsOIDCCredsProviderConfig,
) (*AWSOIDCCredsProvider, error) {

	tokenFetcher := &tokenFetcher{
		conf: conf,
	}

	// fetch a token to get a relevant role session name
	token, err := tokenFetcher.fetchFullToken(ctx)
	if err != nil {
		return nil, err
	}

	provider := stscreds.NewWebIdentityRoleProviderWithToken(
		svc,
		conf.AWSRoleARN,
		token.Claims.Email,
		tokenFetcher,
	)

	return &AWSOIDCCredsProvider{
		Credentials: credentials.NewCredentials(provider),
		fetcher:     tokenFetcher,
	}, nil
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

	return oidc_impl.GetToken(ctx, tf.conf.OIDCClientID, tf.conf.OIDCIssuerURL)
}

func (tf *tokenFetcher) FetchToken(ctx context.Context) ([]byte, error) {
	token, err := tf.fetchFullToken(ctx)
	if err != nil {
		return nil, err
	}

	return []byte(token.IDToken), nil
}
