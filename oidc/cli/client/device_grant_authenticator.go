package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// DeviceGrantAuthenticator implements the OAuth 2.0 Device Authorization Grant flow
type DeviceGrantAuthenticator struct {
	verifier *oidc.IDTokenVerifier
}

// NewDeviceGrantAuthenticator creates a new DeviceGrantClient
func NewDeviceGrantAuthenticator(verifier *oidc.IDTokenVerifier) *DeviceGrantAuthenticator {
	return &DeviceGrantAuthenticator{
		verifier: verifier,
	}
}

// Authenticate initiates the device authorization flow and waits for user authentication
func (c *DeviceGrantAuthenticator) Authenticate(ctx context.Context, config *oauth2.Config) (*Token, error) {
	response, err := config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}

	err = c.displayUserCode(response)
	if err != nil {
		return nil, err
	}

	token, err := config.DeviceAccessToken(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("requesting access token: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Successfully authenticated!\n")
	claims, _, verifiedIDToken, err := idTokenFromOauth2Token(ctx, token, c.verifier)

	if err != nil {
		return nil, fmt.Errorf("extracting id token: %w", err)
	}
	return &Token{
		IDToken: verifiedIDToken,
		Claims:  *claims,
		Token:   token,
	}, nil
}

// displayUserCode displays the user code and verification URL to the user
func (c *DeviceGrantAuthenticator) displayUserCode(deviceAuth *oauth2.DeviceAuthResponse) error {
	data := &deviceAuthTemplateData{
		VerificationURI:  deviceAuth.VerificationURI,
		UserCode:         deviceAuth.UserCode,
		ExpiresInMinutes: int(time.Until(deviceAuth.Expiry).Minutes()),
	}
	err := renderDeviceAuthTemplate(os.Stderr, data)
	if err != nil {
		return fmt.Errorf("rendering device auth template: %w", err)
	}

	return nil
}
