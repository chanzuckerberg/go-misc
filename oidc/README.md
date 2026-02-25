# oidc

A Go library for obtaining OAuth2/OIDC access tokens using various authentication methods, with built-in support for AWS STS integration.

## Overview

`oidc` provides a comprehensive framework for authenticating with OAuth2 and OIDC providers. The library supports multiple authentication flows including interactive browser-based authentication with token caching and AWS KMS-signed JWT bearer assertions for service-to-service authentication.

### Key Features

- **Interactive Browser Authentication**: OIDC authentication flow with automatic browser launching
- **Token Caching & Storage**: Secure token storage using system keyring or file-based cache with automatic refresh
- **Node-Local Caching**: Optional local cache for distributed/NFS environments to avoid cross-host contention
- **KMS-Signed JWT Provider**: Service-to-service authentication using AWS KMS for JWT signing
- **AWS STS Integration**: Built-in AWS credentials provider using OIDC tokens
- **Kubernetes Support**: Generates tokens in Kubernetes `ExecCredential` format
- **Structured Logging**: Session-correlated logging with session ID, hostname, and PID

## Installation

```bash
go get github.com/chanzuckerberg/go-misc/oidc/v5
```

## Authentication Methods

### 1. Interactive Browser Flow

The interactive flow opens a browser for user authentication and caches tokens in the system keyring (desktop) or a file (headless/WSL).

#### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chanzuckerberg/go-misc/oidc/v5/cli"
)

func main() {
    ctx := context.Background()

    token, err := cli.GetToken(
        ctx,
        "my-client-id",
        "https://auth.example.com",
    )
    if err != nil {
        log.Fatalf("failed to get token: %v", err)
    }

    fmt.Printf("ID Token: %s\n", token.IDToken)
    fmt.Printf("Access Token: %s\n", token.AccessToken)
    fmt.Printf("Email: %s\n", token.Claims.Email)
}
```

#### With Options

```go
import (
    "github.com/chanzuckerberg/go-misc/oidc/v5/cli"
    "github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
)

token, err := cli.GetToken(
    ctx,
    clientID,
    issuerURL,
    // Use a node-local cache directory (see "Distributed / NFS Environments" below)
    cli.WithLocalCacheDir("/tmp/oidc-cache"),
    // Customize OAuth2 scopes
    cli.WithClientOptions(client.WithScopes([]string{
        "openid",
        "profile",
        "email",
        "offline_access",
    })),
)
```

#### How It Works

1. **Check Cache**: Checks storage backend (keyring or file) for a cached valid token
2. **Refresh if Needed**: If the token is expired but a refresh token exists, acquires a file lock and refreshes
3. **Browser Authentication**: If no valid token exists, launches the browser to the OIDC provider
4. **Local Callback Server**: Starts a temporary local server (ports 49152-49215) to receive the OAuth callback
5. **Token Storage**: Stores tokens in the storage backend for future use

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
    "github.com/chanzuckerberg/go-misc/oidc/v5"
    "github.com/chanzuckerberg/go-misc/oidc/v5/cli"
)

func main() {
    ctx := context.Background()

    sess := session.Must(session.NewSession())
    stsClient := sts.New(sess)

    provider, err := oidc.NewAwsOIDCCredsProvider(
        ctx,
        stsClient,
        &oidc.AwsOIDCCredsProviderConfig{
            AWSRoleARN:    "arn:aws:iam::123456789012:role/MyOIDCRole",
            OIDCClientID:  "my-client-id",
            OIDCIssuerURL: "https://auth.example.com",
            // Pass GetToken options through to the underlying token flow
            GetTokenOptions: []cli.GetTokenOption{
                cli.WithLocalCacheDir("/tmp/oidc-cache"),
            },
        },
    )
    if err != nil {
        log.Fatalf("failed to create provider: %v", err)
    }

    creds, err := provider.Get()
    if err != nil {
        log.Fatalf("failed to get credentials: %v", err)
    }

    fmt.Printf("AWS Access Key: %s\n", creds.AccessKeyID)

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
    awskms "github.com/aws/aws-sdk-go-v2/service/kms"
    "github.com/chanzuckerberg/go-misc/oidc/v5/kms"
)

func main() {
    ctx := context.Background()
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        log.Fatalf("failed to load AWS config: %v", err)
    }

    kmsClient := awskms.NewFromConfig(cfg)

    claims := kms.NewDefaultClaimsValues(
        "my-service",
        "https://oauth.example.com/token",
        "read:data write:data",
    )

    provider := kms.NewKMSKeyTokenProvider(
        logger,
        kmsClient,
        "arn:aws:kms:us-east-1:123456789012:key/abcd1234-...",
        &claims,
    )

    execCred, err := provider.GetExecToken(ctx, "client.authentication.k8s.io/v1")
    if err != nil {
        log.Fatalf("failed to get token: %v", err)
    }

    logger.Info("token obtained",
        "expiry", execCred.Status.ExpirationTimestamp,
        "token", execCred.Status.Token)
}
```

#### Custom Claims

Implement custom JWT claims by implementing the `ClaimsValues` interface:

```go
type CustomClaims struct {
    clientID, issuerURL, scope string
}

func (c *CustomClaims) GetClaims() jwt.RegisteredClaims {
    return jwt.RegisteredClaims{
        Issuer:    c.clientID,
        Subject:   c.clientID,
        Audience:  []string{c.issuerURL},
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
    }
}

func (c *CustomClaims) GetScope() string     { return c.scope }
func (c *CustomClaims) GetIssuerURL() string { return c.issuerURL }
```

## Caching and Concurrency

### Storage Backends

The library automatically selects a storage backend based on the environment:

| Environment | Backend | Location |
|---|---|---|
| Desktop (macOS, Linux with GUI) | System keyring | macOS Keychain, GNOME Keyring, KWallet |
| WSL / Headless | File | `~/.cache/oidc-cli/<hash>` |

Cache filenames are a SHA-256 hash of the storage version, client ID, and issuer URL, so different OIDC clients do not collide.

### Locking

The library uses `flock`-based file locking (via the `pidlock` package) to coordinate processes on the same host. The lock is scoped to a specific client ID and issuer URL combination.

**What the lock protects:**
- **Token refresh**: When a cached token expires, the lock prevents multiple processes from calling the IdP's refresh endpoint simultaneously. The first process to acquire the lock performs the refresh and writes the new token; other processes waiting on the lock re-read the cache and find the fresh token.
- **Cache writes**: The token is written to storage while the lock is held, so concurrent refreshes do not race on the write.

**What the lock does not protect:**
- **Reads**: Cache reads are lock-free. Reads do not need lock protection because writes use atomic file operations (write to temp file, fsync, rename), so readers always see either the old complete file or the new complete file, never a partial write.

**Lock file location:**
- Default: `~/.cache/oidc-cli/<hash>.lock`
- With `WithLocalCacheDir`: `<localCacheDir>/<hash>.lock`

The lock file uses the same hash as the cache file, with a `.lock` suffix appended.

### Distributed / NFS Environments

On NFS filesystems, `flock` is unreliable across hosts and `rename` may not be atomic across NFS server frontends. If you are running on a distributed system where multiple hosts share a home directory over NFS, use `WithLocalCacheDir` to store the cache and lock files on node-local disk:

```go
token, err := cli.GetToken(
    ctx,
    clientID,
    issuerURL,
    cli.WithLocalCacheDir("/tmp/oidc-cache"),
)
```

When `WithLocalCacheDir` is set:

1. The first time a process on a given host reads the cache, it **bootstraps** by copying the root cache file (on NFS at `~/.cache/oidc-cli/<hash>`) into the local directory as `<localCacheDir>/<hash>-<hostname>`.
2. All subsequent reads and writes use the local copy.
3. The lock file is also placed in the local directory (`<localCacheDir>/<hash>.lock`), so locking uses the local filesystem instead of NFS.

This means each host independently refreshes its own token. Since most IdP refresh tokens (e.g. Okta) are not single-use, concurrent refreshes from different hosts are safe -- they simply produce valid but independent access tokens.

The initial root cache file is created by the first human login (interactive browser flow) and serves as the bootstrap source for all hosts.

## API Reference

### Interactive Browser Flow

#### `cli.GetToken`

```go
func GetToken(
    ctx context.Context,
    clientID string,
    issuerURL string,
    opts ...GetTokenOption,
) (*client.Token, error)
```

Main entry point for interactive OIDC authentication.

**Parameters:**
- `ctx`: Context for cancellation
- `clientID`: OAuth2 client ID registered with your OIDC provider
- `issuerURL`: OIDC provider's issuer URL (e.g., `https://accounts.google.com`)
- `opts`: Optional `GetTokenOption` values to configure behavior

**Returns:**
- `*client.Token`: Token with ID token, access token, refresh token, and claims
- `error`: Any error during authentication

#### `cli.GetTokenOption`

```go
// Store cache and lock files on node-local disk instead of the default
// directory. Bootstraps from the default cache on first access.
cli.WithLocalCacheDir(dir string) GetTokenOption

// Pass an OIDCClientOption through to the underlying OIDC client.
cli.WithClientOptions(opts ...client.OIDCClientOption) GetTokenOption
```

#### `client.OIDCClientOption`

```go
// Customize OAuth2 scopes (default: openid, offline_access, email, groups)
client.WithScopes([]string{"openid", "email"})

// Use device grant flow instead of browser-based authorization grant
client.WithDeviceGrantAuthenticator(a *DeviceGrantAuthenticator)

// Use authorization grant flow with custom config
client.WithAuthzGrantAuthenticator(a *AuthorizationGrantConfig, opts ...AuthorizationGrantAuthenticatorOption)
```

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
- `Valid() bool` - Reports whether the token is currently valid (not expired) according to the underlying `oauth2.Token` semantics

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
    AWSRoleARN      string              // ARN of IAM role to assume
    OIDCClientID    string              // OIDC client ID
    OIDCIssuerURL   string              // OIDC issuer URL
    GetTokenOptions []cli.GetTokenOption // Options forwarded to cli.GetToken
}
```

#### `AWSOIDCCredsProvider`

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
    GetIssuerURL() string
}
```

#### `kms.DefaultClaimsValues`

```go
func NewDefaultClaimsValues(clientID, issuerURL, scopes string) DefaultClaimsValues
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

## Logging

The library uses Go's `log/slog` for structured logging. All log entries within a `GetToken` call share the same attributes for correlation:

| Attribute | Description |
|---|---|
| `session_id` | Random 8-character hex string unique to each `GetToken` invocation |
| `hostname` | Machine hostname (or a random pet name if hostname is unavailable) |
| `pid` | Process ID |

To see debug logs, configure `slog` before calling `GetToken`:

```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})))
```

## Configuration

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
aws kms get-public-key \
  --key-id abcd1234-... \
  --output text \
  --query PublicKey | base64 -d > public_key.der

openssl rsa -pubin -inform DER -in public_key.der \
  -outform PEM -out public_key.pem
```

## AWS IAM Setup (for STS Integration)

### OIDC Identity Provider

Register your OIDC provider with AWS IAM:

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

- **Token Storage**: Tokens are stored in system keyring with OS-level encryption (desktop) or in permission-restricted files (headless)
- **Atomic Writes**: File-based cache writes use temp file + fsync + rename to prevent partial reads
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

- **File Lock**: Prevents race conditions during concurrent token refresh
- **Browser Security**: Uses system default browser with HTTPS OAuth endpoints
- **Local Server**: Callback server binds to localhost only (not externally accessible)
- **Token Refresh**: Automatic refresh minimizes exposure window for access tokens

## Troubleshooting

### Interactive Flow

**"acquiring lock" timeout or error**
- Check that the lock directory (`~/.cache/oidc-cli/` or your `WithLocalCacheDir` path) is writable
- Verify no stale `.lock` files exist in that directory
- If on NFS, use `WithLocalCacheDir` to place locks on local disk

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

    "github.com/chanzuckerberg/go-misc/oidc/v5/cli"
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
    "github.com/chanzuckerberg/go-misc/oidc/v5"
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

### Distributed Environment with Node-Local Cache

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chanzuckerberg/go-misc/oidc/v5/cli"
)

func main() {
    ctx := context.Background()

    token, err := cli.GetToken(
        ctx,
        "my-client-id",
        "https://auth.example.com",
        // Place cache and lock on local disk to avoid NFS issues
        cli.WithLocalCacheDir("/tmp/oidc-cache"),
    )
    if err != nil {
        log.Fatalf("failed to get token: %v", err)
    }

    fmt.Printf("Email: %s\n", token.Claims.Email)
}
```

## License

See the [LICENSE](../LICENSE) file in the root of this repository.

## Contributing

Contributions are welcome! Please see the repository's main README for contribution guidelines.
