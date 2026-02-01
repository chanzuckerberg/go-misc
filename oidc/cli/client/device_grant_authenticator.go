package client

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// DeviceGrantAuthenticator implements the OAuth 2.0 Device Authorization Grant flow
type DeviceGrantAuthenticator struct {
}

// NewDeviceGrantAuthenticator creates a new DeviceGrantClient
func NewDeviceGrantAuthenticator() *DeviceGrantAuthenticator {
	log := slog.Default()
	log.Debug("NewDeviceGrantAuthenticator: creating device grant authenticator")
	return &DeviceGrantAuthenticator{}
}

// Authenticate initiates the device authorization flow and waits for user authentication
func (c *DeviceGrantAuthenticator) Authenticate(ctx context.Context, client *OIDCClient) (*Token, error) {
	log := slog.Default()
	startTime := time.Now()

	log.Debug("DeviceGrantAuthenticator.Authenticate: starting device authorization flow")

	log.Debug("DeviceGrantAuthenticator.Authenticate: requesting device code")
	response, err := client.DeviceAuth(ctx)
	if err != nil {
		log.Error("DeviceGrantAuthenticator.Authenticate: requesting device code",
			"error", err,
		)
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	log.Debug("DeviceGrantAuthenticator.Authenticate: device code received",
		"verification_uri", response.VerificationURI,
		"expires_in", time.Until(response.Expiry),
	)

	log.Debug("DeviceGrantAuthenticator.Authenticate: displaying user code")
	err = c.displayUserCode(response)
	if err != nil {
		log.Error("DeviceGrantAuthenticator.Authenticate: displaying user code",
			"error", err,
		)
		return nil, err
	}

	log.Debug("DeviceGrantAuthenticator.Authenticate: polling for access token")
	token, err := client.DeviceAccessToken(ctx, response)
	if err != nil {
		log.Error("DeviceGrantAuthenticator.Authenticate: getting access token",
			"error", err,
			"elapsed_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("requesting access token: %w", err)
	}
	log.Debug("DeviceGrantAuthenticator.Authenticate: access token received",
		"token_expiry", token.Expiry,
	)

	fmt.Fprintf(os.Stderr, "\n✓ Successfully authenticated!\n")

	log.Debug("DeviceGrantAuthenticator.Authenticate: parsing token as ID token")
	claims, _, verifiedIDToken, err := client.ParseAsIDToken(ctx, token)
	if err != nil {
		log.Error("DeviceGrantAuthenticator.Authenticate: parsing ID token",
			"error", err,
		)
		return nil, fmt.Errorf("extracting id token: %w", err)
	}

	log.Debug("DeviceGrantAuthenticator.Authenticate: device flow completed successfully",
		"elapsed_ms", time.Since(startTime).Milliseconds(),
		"token_expiry", token.Expiry,
	)
	return &Token{
		IDToken: verifiedIDToken,
		Claims:  *claims,
		Token:   token,
	}, nil
}

// displayUserCode displays the user code and verification URL to the user
func (c *DeviceGrantAuthenticator) displayUserCode(deviceAuth *oauth2.DeviceAuthResponse) error {
	log := slog.Default()
	log.Debug("displayUserCode: rendering device auth template",
		"verification_uri", deviceAuth.VerificationURI,
		"user_code", deviceAuth.UserCode,
		"expires_in_minutes", int(time.Until(deviceAuth.Expiry).Minutes()),
	)

	data := &deviceAuthTemplateData{
		VerificationURI:  deviceAuth.VerificationURI,
		UserCode:         deviceAuth.UserCode,
		ExpiresInMinutes: int(time.Until(deviceAuth.Expiry).Minutes()),
	}
	err := renderDeviceAuthTemplate(os.Stderr, data)
	if err != nil {
		log.Error("displayUserCode: rendering template",
			"error", err,
		)
		return fmt.Errorf("rendering device auth template: %w", err)
	}

	log.Debug("displayUserCode: template rendered successfully")
	return nil
}
