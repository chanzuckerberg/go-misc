# oidc

A Go library for obtaining OAuth2/OIDC access tokens using various authentication methods, with built-in support for AWS STS integration.

## Overview

`oidc` provides a comprehensive framework for authenticating with OAuth2 and OIDC providers. The library supports multiple authentication flows including interactive browser-based authentication with token caching and AWS KMS-signed JWT bearer assertions for service-to-service authentication.

### Key Features

- **Interactive Browser Authentication**: OIDC authentication flow with automatic browser launching
- **Token Caching & Storage**: Secure token storage using system keyring with automatic refresh
- **KMS-Signed JWT Provider**: Service-to-service authentication using AWS KMS for JWT signing
- **AWS STS Integration**: Built-in AWS credentials provider using OIDC tokens
- **Kubernetes Support**: Generates tokens in Kubernetes `ExecCredential` format
- **Production Ready**: Battle-tested with locking mechanisms to prevent concurrent authentication attempts

## Installation

```bash
go get github.com/chanzuckerberg/go-misc/oidc/v4
```

## Authentication Methods

### 1. Interactive Browser Flow

The interactive flow opens a browser for user authentication and securely caches tokens in the system keyring.

#### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chanzuckerberg/go-misc/oidc/v4/cli"
)

func main() {
    ctx := context.Background()

    // GetToken handles the full flow: checking cache, refreshing if needed,
    // or launching browser for new authentication
    token, err := cli.GetToken(
        ctx,
        "my-client-id",                        // OAuth2 client ID
        "https://auth.example.com",            // OIDC issuer URL
    )
    if err != nil {
        log.Fatalf("failed to get token: %v", err)
    }

    fmt.Printf("ID Token: %s\n", token.IDToken)
    fmt.Printf("Access Token: %s\n", token.AccessToken)
    fmt.Printf("Email: %s\n", token.Claims.Email)
}
```

#### With Custom Options

```go
import (
    "github.com/chanzuckerberg/go-misc/oidc/v4/cli"
    "github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
    "golang.org/x/oauth2"
)

token, err := cli.GetToken(
    ctx,
    clientID,
    issuerURL,
    // Customize OAuth2 scopes
    client.SetScopeOptions([]string{
        "openid",
        "profile",
        "email",
        "offline_access",
    }),
    // Customize success message shown in browser
    client.SetSuccessMessage("Authentication complete! Close this window."),
    // Set OAuth2 auth style
    client.SetOauth2AuthStyle(oauth2.AuthStyleInParams),
)
```

#### How It Works

1. **Check Cache**: First checks system keyring for cached valid token
2. **Refresh if Needed**: If token is expired but refresh token exists, automatically refreshes
3. **Browser Authentication**: If no valid token, launches browser to OIDC provider
4. **Local Callback Server**: Starts temporary local server (ports 49152-49215) to receive OAuth callback
5. **Token Storage**: Securely stores tokens in system keyring for future use
6. **File Locking**: Uses `/tmp/aws-oidc.lock` to prevent concurrent authentication attempts

### 2. AWS STS Integration

Get AWS credentials using OIDC tokens via STS AssumeRoleWithWebIdentity.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sts"
    "github.com/chanzuckerberg/go-misc/oidc/v4"
)

func main() {
    ctx := context.Background()

    // Create AWS session
    sess := session.Must(session.NewSession())
    stsClient := sts.New(sess)

    // Create OIDC credentials provider
    provider, err := oidc.NewAwsOIDCCredsProvider(
        ctx,
        stsClient,
        &oidc.AwsOIDCCredsProviderConfig{
            AWSRoleARN:    "arn:aws:iam::123456789012:role/MyOIDCRole",
            OIDCClientID:  "my-client-id",
            OIDCIssuerURL: "https://auth.example.com",
        },
    )
    if err != nil {
        log.Fatalf("failed to create provider: %v", err)
    }

    // Get AWS credentials
    creds, err := provider.Get()
    if err != nil {
        log.Fatalf("failed to get credentials: %v", err)
    }

    fmt.Printf("AWS Access Key: %s\n", creds.AccessKeyID)

    // You can also fetch the raw OIDC token
    token, err := provider.FetchOIDCToken(ctx)
    if err != nil {
        log.Fatalf("failed to fetch OIDC token: %v", err)
    }
    fmt.Printf("ID Token: %s\n", token.IDToken)
}
```

The provider automatically:
- Fetches OIDC token (with caching and refresh)
- Uses token to assume AWS role via STS
- Refreshes AWS credentials when they expire
- Thread-safe for concurrent use

### 3. KMS-Signed JWT Provider

For service-to-service authentication using AWS KMS to sign JWTs for OAuth2 client credentials flow.

#### Basic Usage

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "os"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/kms"
    "github.com/chanzuckerberg/go-misc/oidc/v4/kms"
)

func main() {
    ctx := context.Background()
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        log.Fatalf("failed to load AWS config: %v", err)
    }

    kmsClient := kms.NewFromConfig(cfg)

    // Configure JWT claims
    claims := kms.NewDefaultClaimsValues(
        "my-service",                          // issuer (client ID)
        "https://oauth.example.com/token",     // audience (token endpoint)
        "read:data write:data",                // scopes
    )

    // Create KMS token provider
    provider := kms.NewKMSKeyTokenProvider(
        logger,
        kmsClient,
        "arn:aws:kms:us-east-1:123456789012:key/abcd1234-...",
        &claims,
    )

    // Get token in Kubernetes ExecCredential format
    execCred, err := provider.GetExecToken(ctx, "client.authentication.k8s.io/v1")
    if err != nil {
        log.Fatalf("failed to get token: %v", err)
    }

    // Use the access token
    logger.Info("token obtained",
        "expiry", execCred.Status.ExpirationTimestamp,
        "token", execCred.Status.Token)
}
```

#### Custom Claims

Implement custom JWT claims by implementing the `ClaimsValues` interface:

```go
type CustomClaims struct {
    issuer, audience, scope string
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

### Interactive Browser Flow

#### `cli.GetToken`

```go
func GetToken(
    ctx context.Context,
    clientID string,
    issuerURL string,
    clientOptions ...client.Option,
) (*client.Token, error)
```

Main entry point for interactive OIDC authentication.

**Parameters:**
- `ctx`: Context for cancellation
- `clientID`: OAuth2 client ID registered with your OIDC provider
- `issuerURL`: OIDC provider's issuer URL (e.g., `https://accounts.google.com`)
- `clientOptions`: Optional configuration options

**Returns:**
- `*client.Token`: Token with ID token, access token, refresh token, and claims
- `error`: Any error during authentication

#### `client.Token`

```go
type Token struct {
    Version      int
    Expiry       time.Time
    IDToken      string
    AccessToken  string
    RefreshToken string
    Claims       Claims
}
```

**Methods:**
- `Marshal(...MarshalOpts) (string, error)` - Serialize token to base64-encoded JSON

#### `client.Claims`

```go
type Claims struct {
    Issuer                string   `json:"iss"`
    Audience              string   `json:"aud"`
    Subject               string   `json:"sub"`
    Name                  string   `json:"name"`
    AuthenticationMethods []string `json:"amr"`
    Email                 string   `json:"email"`
}
```

#### Client Options

```go
// Customize OAuth2 scopes (default: openid, offline_access, email, groups)
client.SetScopeOptions([]string{"openid", "email"})

// Customize success message shown in browser after authentication
client.SetSuccessMessage("You're all set! Return to your terminal.")

// Set OAuth2 authentication style
client.SetOauth2AuthStyle(oauth2.AuthStyleInParams)
```

### AWS STS Integration

#### `oidc.NewAwsOIDCCredsProvider`

```go
func NewAwsOIDCCredsProvider(
    ctx context.Context,
    svc stsiface.STSAPI,
    conf *AwsOIDCCredsProviderConfig,
) (*AWSOIDCCredsProvider, error)
```

Creates an AWS credentials provider that uses OIDC tokens.

**Configuration:**

```go
type AwsOIDCCredsProviderConfig struct {
    AWSRoleARN    string  // ARN of IAM role to assume
    OIDCClientID  string  // OIDC client ID
    OIDCIssuerURL string  // OIDC issuer URL
}
```

#### `AWSOIDCCredsProvider`

```go
type AWSOIDCCredsProvider struct {
    *credentials.Credentials
    // ... internal fields
}
```

**Methods:**
- `Get() (credentials.Value, error)` - Get AWS credentials (inherited from `credentials.Credentials`)
- `FetchOIDCToken(ctx context.Context) (*client.Token, error)` - Fetch the underlying OIDC token

The provider implements the AWS SDK's `credentials.Provider` interface and can be used anywhere AWS credentials are needed.

### KMS JWT Provider

#### `kms.NewKMSKeyTokenProvider`

```go
func NewKMSKeyTokenProvider(
    logger *slog.Logger,
    client *kms.Client,
    keyID string,
    claims ClaimsValues,
) *KMSKeyTokenProvider
```

Creates a token provider that uses AWS KMS for JWT signing.

**Parameters:**
- `logger`: Structured logger for debugging
- `client`: AWS KMS client (from AWS SDK v2)
- `keyID`: KMS key ID or ARN
- `claims`: Claims provider implementing `ClaimsValues` interface

#### `kms.ClaimsValues`

```go
type ClaimsValues interface {
    GetClaims() jwt.RegisteredClaims
    GetScope() string
    GetAudience() string
}
```

#### `kms.DefaultClaimsValues`

```go
func NewDefaultClaimsValues(issuer, audience, scopes string) DefaultClaimsValues
```

Default implementation with 1-hour token expiration.

#### `kms.ExecCredential`

```go
type ExecCredential struct {
    Kind       string      `json:"kind"`
    APIVersion string      `json:"apiVersion"`
    Spec       struct{}    `json:"spec"`
    Status     TokenStatus `json:"status"`
}
```

Kubernetes ExecCredential format for kubectl authentication plugins.

## Configuration

### Storage Locations

Tokens are stored securely in the system keyring:
- **macOS**: Keychain
- **Linux**: Secret Service API (GNOME Keyring, KWallet)
- **Windows**: Windows Credential Manager

Keyring service name: `oidc`
Keyring user format: `{clientID}@{issuerURL}`

### Lock File

Location: `/tmp/aws-oidc.lock`

Prevents multiple concurrent authentication flows from the same machine.

### Local Callback Server

**Port Range**: 49152-49215 (ephemeral port range)
**Timeout**: 30 seconds

The library attempts to bind to the first available port in this range.

## AWS KMS Setup (for KMS Provider)

### KMS Key Requirements

- **Key Type**: Asymmetric
- **Key Usage**: SIGN_VERIFY
- **Key Spec**: RSA_2048, RSA_3072, or RSA_4096
- **Signing Algorithm**: RSASSA_PKCS1_V1_5_SHA_256

### Create KMS Key

```bash
aws kms create-key \
  --key-usage SIGN_VERIFY \
  --key-spec RSA_2048 \
  --description "OIDC JWT Signing Key"
```

### IAM Permissions

Your application needs permission to sign with the KMS key:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:Sign",
        "kms:GetPublicKey"
      ],
      "Resource": "arn:aws:kms:us-east-1:123456789012:key/abcd1234-..."
    }
  ]
}
```

### Export Public Key

Register the public key with your OAuth2/OIDC provider:

```bash
# Export public key
aws kms get-public-key \
  --key-id abcd1234-... \
  --output text \
  --query PublicKey | base64 -d > public_key.der

# Convert to PEM (if needed)
openssl rsa -pubin -inform DER -in public_key.der \
  -outform PEM -out public_key.pem
```

## AWS IAM Setup (for STS Integration)

### OIDC Identity Provider

First, register your OIDC provider with AWS IAM:

```bash
aws iam create-open-id-connect-provider \
  --url https://auth.example.com \
  --client-id-list my-client-id \
  --thumbprint-list <certificate-thumbprint>
```

### IAM Role Trust Policy

Create an IAM role with a trust policy that allows your OIDC provider:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/auth.example.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "auth.example.com:aud": "my-client-id",
          "auth.example.com:sub": "user@example.com"
        }
      }
    }
  ]
}
```

## Security Considerations

### General

- **Token Storage**: Tokens are stored in system keyring with OS-level encryption
- **Token Expiration**: Tokens include 5-minute time skew tolerance for clock differences
- **Secure Transmission**: All OAuth2/OIDC communication uses HTTPS
- **State Validation**: OAuth2 state parameter prevents CSRF attacks
- **PKCE**: Uses Proof Key for Code Exchange (PKCE) for additional security
- **Nonce Validation**: OIDC nonce prevents replay attacks

### KMS Provider

- **Key Rotation**: Regularly rotate KMS keys and update OAuth2 provider configuration
- **IAM Least Privilege**: Grant `kms:Sign` only to services that need it
- **CloudTrail**: Enable AWS CloudTrail to audit all KMS key usage
- **Key Policies**: Use KMS key policies to restrict access by IAM principal

### Interactive Flow

- **File Lock**: Prevents race conditions during concurrent authentication attempts
- **Browser Security**: Uses system default browser with HTTPS OAuth endpoints
- **Local Server**: Callback server binds to localhost only (not externally accessible)
- **Token Refresh**: Automatic refresh minimizes exposure window for access tokens

## Troubleshooting

### Interactive Flow

**"unable to create lock"**
- Check that `/tmp` is writable
- Verify no stale lock files exist: `rm /tmp/aws-oidc.lock`

**"could not open browser"**
- Ensure a browser is installed and in PATH
- Set `BROWSER` environment variable to specify browser
- Check that DISPLAY is set (Linux) or windowing system is available

**"could not bind to port"**
- Verify ports 49152-49215 are not all in use
- Check firewall allows localhost connections
- Try closing other applications using ephemeral ports

### KMS Provider

**"unable to sign JWT with KMS key"**
- Verify AWS credentials are configured
- Check IAM permissions include `kms:Sign`
- Confirm KMS key ID/ARN is correct
- Ensure KMS key is in the same region (or region is explicitly set)

**"got status code 401" from token endpoint**
- Verify public key is registered with OAuth2 provider
- Check issuer (client ID) matches provider configuration
- Ensure audience is the correct token endpoint URL
- Confirm key ID in JWT header matches registered key

### AWS STS Integration

**"AssumeRoleWithWebIdentity failed"**
- Verify IAM role ARN is correct
- Check OIDC provider is registered in IAM
- Confirm trust policy allows your OIDC provider and client ID
- Ensure OIDC token contains required claims (sub, aud, etc.)

**"Token expired"**
- Check system clock is synchronized
- Verify OIDC token hasn't expired (tokens typically valid 1 hour)
- Token auto-refreshes, but may need manual re-authentication if refresh token expired

## Examples

### CLI Tool with OIDC Authentication

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/chanzuckerberg/go-misc/oidc/v4/cli"
)

func main() {
    ctx := context.Background()

    token, err := cli.GetToken(
        ctx,
        os.Getenv("OIDC_CLIENT_ID"),
        os.Getenv("OIDC_ISSUER_URL"),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
        os.Exit(1)
    }

    // Output token in Kubernetes ExecCredential format
    execCred := map[string]interface{}{
        "apiVersion": "client.authentication.k8s.io/v1",
        "kind":       "ExecCredential",
        "status": map[string]interface{}{
            "token":               token.IDToken,
            "expirationTimestamp": token.Expiry.Format("2006-01-02T15:04:05Z"),
        },
    }

    json.NewEncoder(os.Stdout).Encode(execCred)
}
```

### AWS SDK with OIDC Credentials

```go
package main

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/sts"
    "github.com/chanzuckerberg/go-misc/oidc/v4"
)

func main() {
    ctx := context.Background()

    sess := session.Must(session.NewSession())
    stsClient := sts.New(sess)

    provider, err := oidc.NewAwsOIDCCredsProvider(
        ctx,
        stsClient,
        &oidc.AwsOIDCCredsProviderConfig{
            AWSRoleARN:    "arn:aws:iam::123456789012:role/MyRole",
            OIDCClientID:  "my-client-id",
            OIDCIssuerURL: "https://auth.example.com",
        },
    )
    if err != nil {
        panic(err)
    }

    // Use the credentials with AWS SDK
    s3Client := s3.New(sess, &aws.Config{
        Credentials: provider.Credentials,
    })

    result, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
    if err != nil {
        panic(err)
    }

    fmt.Println("S3 Buckets:")
    for _, bucket := range result.Buckets {
        fmt.Printf("  - %s\n", *bucket.Name)
    }
}
```

## License

See the [LICENSE](../LICENSE) file in the root of this repository.

## Contributing

Contributions are welcome! Please see the repository's main README for contribution guidelines.
