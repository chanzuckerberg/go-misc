# kms_jwt

A Go library for generating OAuth2 access tokens using AWS KMS-signed JWTs for client authentication.

## Overview

`kms_jwt` provides a secure way to authenticate with OAuth2 providers using the JWT Bearer assertion flow ([RFC 7523](https://tools.ietf.org/html/rfc7523)), where the JWT is signed using an AWS KMS asymmetric key instead of storing private keys locally. This approach offers enhanced security by keeping private key material within AWS KMS.

### Key Features

- **Secure Key Management**: Uses AWS KMS asymmetric keys for JWT signing, eliminating the need to store private keys locally
- **OAuth2 Client Credentials Flow**: Implements JWT bearer token authentication for service-to-service authentication
- **Kubernetes Integration**: Generates tokens in Kubernetes `ExecCredential` format for use with kubectl exec plugins
- **Customizable Claims**: Flexible claims interface allows custom JWT claim structures
- **AWS SDK v2**: Built on the latest AWS SDK for Go v2

## Installation

```bash
go get github.com/chanzuckerberg/go-misc/kms_jwt
```

## Requirements

- Go 1.24.5 or later
- An AWS KMS asymmetric key pair with `SIGN_VERIFY` usage (RSA_2048 or higher recommended)
- AWS credentials configured with permissions to use the KMS key (`kms:Sign`)
- An OAuth2 provider that supports JWT bearer token authentication

## Usage

### Basic Example

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/kms"
    "github.com/chanzuckerberg/go-misc/kms_jwt"
)

func main() {
    ctx := context.Background()
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        logger.Error("failed to load AWS config", "error", err)
        os.Exit(1)
    }

    // Create KMS client
    kmsClient := kms.NewFromConfig(cfg)

    // Configure claims
    claims := token.NewDefaultClaimsValues(
        "my-service",                          // issuer
        "https://oauth.example.com/token",     // audience (token endpoint)
        "read:data write:data",                // scopes
    )

    // Create token provider
    provider := token.NewKMSKeyTokenProvider(
        logger,
        kmsClient,
        "arn:aws:kms:us-east-1:123456789012:key/abcd1234-...", // KMS key ID
        &claims,
    )

    // Fetch access token
    accessToken, expiry, err := provider.fetchToken(ctx)
    if err != nil {
        logger.Error("failed to fetch token", "error", err)
        os.Exit(1)
    }

    logger.Info("token obtained", "expiry", expiry)
    // Use accessToken for API calls...
}
```

### Kubernetes ExecCredential Example

For use with Kubernetes `kubectl` exec credential plugins:

```go
package main

import (
    "context"
    "encoding/json"
    "log/slog"
    "os"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/kms"
    "github.com/chanzuckerberg/go-misc/kms_jwt"
)

func main() {
    ctx := context.Background()
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    cfg, _ := config.LoadDefaultConfig(ctx)
    kmsClient := kms.NewFromConfig(cfg)

    claims := token.NewDefaultClaimsValues(
        os.Getenv("CLIENT_ID"),
        os.Getenv("TOKEN_URL"),
        os.Getenv("SCOPES"),
    )

    provider := token.NewKMSKeyTokenProvider(
        logger,
        kmsClient,
        os.Getenv("KMS_KEY_ID"),
        &claims,
    )

    // Get token in ExecCredential format
    execCred, err := provider.GetExecToken(ctx, "client.authentication.k8s.io/v1")
    if err != nil {
        logger.Error("failed to get exec token", "error", err)
        os.Exit(1)
    }

    // Output to stdout for kubectl to consume
    json.NewEncoder(os.Stdout).Encode(execCred)
}
```

### Custom Claims

Implement the `ClaimsValues` interface for custom claim structures:

```go
type CustomClaims struct {
    issuer, audience, scope string
    customField             string
}

func (c *CustomClaims) GetClaims() jwt.RegisteredClaims {
    return jwt.RegisteredClaims{
        Issuer:    c.issuer,
        Subject:   c.issuer,
        Audience:  []string{c.audience},
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
    }
}

func (c *CustomClaims) GetScope() string {
    return c.scope
}

func (c *CustomClaims) GetAudience() string {
    return c.audience
}
```

## API Reference

### Types

#### `KMSKeyTokenProvider`

The main token provider that uses AWS KMS for JWT signing.

**Constructor:**
```go
func NewKMSKeyTokenProvider(
    logger *slog.Logger,
    client *kms.Client,
    keyID string,
    claims ClaimsValues,
) *KMSKeyTokenProvider
```

**Methods:**
- `GetExecToken(ctx context.Context, apiVersion string) (*ExecCredential, error)` - Returns a token in Kubernetes ExecCredential format
- `fetchToken(ctx context.Context) (string, time.Time, error)` - Returns the access token and its expiration time (note: currently not exported)

#### `ClaimsValues` (Interface)

Defines the contract for providing JWT claims.

```go
type ClaimsValues interface {
    GetClaims() jwt.RegisteredClaims
    GetScope() string
    GetAudience() string
}
```

#### `DefaultClaimsValues`

Default implementation of `ClaimsValues`.

**Constructor:**
```go
func NewDefaultClaimsValues(issuer, audience, scopes string) DefaultClaimsValues
```

- `issuer`: Client ID or service identifier
- `audience`: OAuth2 token endpoint URL
- `scopes`: Space-separated list of OAuth2 scopes

#### `AccessTokenResponse`

Response from the OAuth2 token endpoint.

```go
type AccessTokenResponse struct {
    TokenType        string `json:"token_type"`
    ExpiresInSeconds int    `json:"expires_in"`
    AccessToken      string `json:"access_token"`
    Scope            string `json:"scope"`
}
```

#### `ExecCredential`

Kubernetes ExecCredential format for kubectl authentication.

```go
type ExecCredential struct {
    Kind       string      `json:"kind"`
    APIVersion string      `json:"apiVersion"`
    Spec       struct{}    `json:"spec"`
    Status     TokenStatus `json:"status"`
}
```

## How It Works

1. **JWT Creation**: Creates a JWT with the specified claims using the RS256 signing algorithm
2. **KMS Signing**: Uses AWS KMS to sign the JWT's signing string with your asymmetric key
3. **Token Exchange**: Sends the signed JWT to the OAuth2 provider using the `client_credentials` grant type with JWT bearer assertion
4. **Access Token**: Returns the access token received from the OAuth2 provider

The flow follows [RFC 7523](https://tools.ietf.org/html/rfc7523) for JWT Bearer Token authentication.

## AWS KMS Key Setup

Your KMS key must be configured as follows:

1. **Key Type**: Asymmetric
2. **Key Usage**: SIGN_VERIFY
3. **Key Spec**: RSA_2048 (or RSA_3072, RSA_4096)
4. **Signing Algorithm**: RSASSA_PKCS1_V1_5_SHA_256

Example KMS key policy to allow signing:

```json
{
  "Effect": "Allow",
  "Principal": {
    "AWS": "arn:aws:iam::123456789012:role/MyServiceRole"
  },
  "Action": "kms:Sign",
  "Resource": "*"
}
```

## Security Considerations

- **Key Rotation**: Regularly rotate your KMS keys and update the OAuth2 provider with the new public key
- **IAM Permissions**: Restrict `kms:Sign` permissions to only the services that need to generate tokens
- **Token Expiration**: The default implementation sets JWT expiration to 1 hour; adjust based on your security requirements
- **Audit Logging**: Enable AWS CloudTrail to audit KMS key usage

## License

See the [LICENSE](../LICENSE) file in the root of this repository.

## Contributing

Contributions are welcome! Please see the repository's main README for contribution guidelines.

