package oidc

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
)

type AwsOIDCCredsProviderConfig struct {
	AWSRoleARN      string
	OIDCClientID    string
	OIDCIssuerURL   string
	GetTokenOptions []cli.GetTokenOption
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

	return cli.GetToken(ctx, tf.conf.OIDCClientID, tf.conf.OIDCIssuerURL, tf.conf.GetTokenOptions...)
}

func (tf *tokenFetcher) FetchToken(ctx context.Context) ([]byte, error) {
	token, err := tf.fetchFullToken(ctx)
	if err != nil {
		return nil, err
	}

	return []byte(token.IDToken), nil
}
