package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/golang-jwt/jwt/v4"
)

type ClaimsValues interface {
	GetClaims() jwt.RegisteredClaims
	GetScope() string
	GetIssuerURL() string
}

type DefaultClaimsValues struct {
	clientID, issuerURL, scope string
}

var _ ClaimsValues = DefaultClaimsValues{}
var _ ClaimsValues = &DefaultClaimsValues{}

func NewDefaultClaimsValues(clientID, issuerURL, scopes string) DefaultClaimsValues {
	return DefaultClaimsValues{
		clientID:  clientID,
		issuerURL: issuerURL,
		scope:     scopes,
	}
}

func (d DefaultClaimsValues) GetClaims() jwt.RegisteredClaims {
	// client id is the issuer and subject
	// issuer url is the audience
	// https://developer.okta.com/docs/guides/implement-oauth-for-okta-serviceapp/main/#create-and-sign-the-jwt
	return jwt.RegisteredClaims{
		Issuer:    d.clientID,
		Subject:   d.clientID,
		Audience:  []string{d.issuerURL},
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
}

func (d DefaultClaimsValues) GetScope() string {
	return d.scope
}

func (d DefaultClaimsValues) GetIssuerURL() string {
	return d.issuerURL
}

type KMSKeyTokenProvider struct {
	logger *slog.Logger
	client *kms.Client
	keyID  string
	claims ClaimsValues
}

type AccessTokenResponse struct {
	TokenType        string `json:"token_type"`
	ExpiresInSeconds int    `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
}

type TokenStatus struct {
	ExpirationTimestamp string `json:"expirationTimestamp"`
	Token               string `json:"token"`
}

type ExecCredential struct {
	Kind       string      `json:"kind"`
	APIVersion string      `json:"apiVersion"`
	Spec       struct{}    `json:"spec"`
	Status     TokenStatus `json:"status"`
}

func makeExecCredential(token string, expiry time.Time, apiVersion string) *ExecCredential {
	return &ExecCredential{
		Kind:       "ExecCredential",
		APIVersion: apiVersion,
		Spec:       struct{}{},
		Status: TokenStatus{
			ExpirationTimestamp: expiry.UTC().Format(time.RFC3339),
			Token:               token,
		},
	}
}

func NewKMSKeyTokenProvider(logger *slog.Logger, client *kms.Client, keyID string, claims ClaimsValues) *KMSKeyTokenProvider {
	return &KMSKeyTokenProvider{
		logger: logger,
		client: client,
		keyID:  keyID,
		claims: claims,
	}
}

func (k *KMSKeyTokenProvider) GetExecToken(ctx context.Context, apiVersion string) (*ExecCredential, error) {
	token, expiry, err := k.fetchToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch token: %w", err)
	}
	return makeExecCredential(token, expiry, apiVersion), nil
}

func (k *KMSKeyTokenProvider) fetchToken(ctx context.Context) (string, time.Time, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, k.claims.GetClaims())
	signingStr, err := token.SigningString()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to make signing string: %w", err)
	}

	signResponse, err := k.client.Sign(ctx, &kms.SignInput{
		Message:          []byte(signingStr),
		KeyId:            aws.String(k.keyID),
		SigningAlgorithm: types.SigningAlgorithmSpecRsassaPkcs1V15Sha256,
		MessageType:      types.MessageTypeRaw,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to sign JWT with KMS key %s: %w", k.keyID, err)
	}

	accessTokenResp, err := requestAccessToken(k.claims.GetScope(), k.claims.GetIssuerURL(), fmt.Sprintf("%s.%s", signingStr, base64.RawStdEncoding.EncodeToString(signResponse.Signature)))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("unable to request access token: %w", err)
	}

	expiration := time.Now().Add(time.Duration(accessTokenResp.ExpiresInSeconds) * time.Second)

	k.logger.Debug("Access token response:", "access_token", accessTokenResp.AccessToken, "expiration", expiration)

	return accessTokenResp.AccessToken, expiration, nil
}

func requestAccessToken(scope, issuerURL, signedToken string) (*AccessTokenResponse, error) {
	values := url.Values{}
	values.Add("grant_type", "client_credentials")
	values.Add("scope", scope)
	values.Add("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	values.Add("client_assertion", signedToken)

	params := values.Encode()

	req, err := http.NewRequest(http.MethodPost, issuerURL, strings.NewReader(params))
	if err != nil {
		return nil, fmt.Errorf("error talking to %s: %w", issuerURL, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to %s: %w", issuerURL, err)
	}
	if resp.StatusCode >= 300 {
		respOut, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("got status code %s %+v: %w", resp.Status, resp.StatusCode, err)
		}
		return nil, fmt.Errorf("got status code %s %+v %s", resp.Status, resp.StatusCode, string(respOut))
	}

	accessTokenResp := AccessTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&accessTokenResp)
	if err != nil {
		return nil, fmt.Errorf("unable to decode access token response: %w", err)
	}

	return &accessTokenResp, nil
}
