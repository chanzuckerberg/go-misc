package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

const (
	timeSkew = 5 * time.Minute

	tokenVersion = 0
)

// Claims represent the oidc token claims
type Claims struct {
	Issuer                string   `json:"iss"`
	Audience              string   `json:"aud"`
	Subject               string   `json:"sub"`
	Name                  string   `json:"name"`
	AuthenticationMethods []string `json:"amr"`
	Email                 string   `json:"email"`
	PreferredUsername     string   `json:"preferred_username"`
}

// Token wraps the extracted claims, auth token, id token, refresh token
// so we can easily use it throughout our application
type Token struct {
	Version int

	Expiry time.Time `json:"expires,omitempty"`

	IDToken      string `json:"token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Claims       Claims `json:"claims,omitempty"`
}

func (vt *Token) IsFresh() bool {
	if vt == nil {
		return false
	}
	return vt.Expiry.After(time.Now().Add(timeSkew))
}

func TokenFromString(tokenString *string, opts ...MarshalOpts) (*Token, error) {
	if tokenString == nil {
		slog.Debug("nil token string")
		return nil, nil
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(*tokenString)
	if err != nil {
		return nil, fmt.Errorf("error b64 decoding token: %w", err)
	}
	token := &Token{
		Version: tokenVersion,
	}
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return nil, fmt.Errorf("could not json unmarshal token: %w", err)
	}

	for _, opt := range opts {
		opt(token)
	}
	return token, nil
}

func (vt *Token) Marshal(opts ...MarshalOpts) (string, error) {
	if vt == nil {
		return "", fmt.Errorf("error Marshalling nil token")
	}

	// apply any processing to the token
	for _, opt := range opts {
		opt(vt)
	}

	tokenBytes, err := json.Marshal(vt)
	if err != nil {
		return "", fmt.Errorf("could not marshal token: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(tokenBytes)
	return b64, nil
}

// MarshalOpts changes a token for marshaling
type MarshalOpts func(*Token)

// Disables the refresh oauth flow
func MarshalOptNoRefresh(t *Token) {
	if t == nil {
		return
	}
	t.RefreshToken = ""
}
