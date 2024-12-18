package client

import "golang.org/x/oauth2"

const (
	defaultSuccessMessage = "Signed in successfully! You can now return to CLI."
)

type oidcStatus string

var oidcStatusSuccess oidcStatus = "success"

type Option func(*Client)

var SetSuccessMessage = func(successMessage string) Option {
	return func(c *Client) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}

var SetOauth2AuthStyle = func(authStyle oauth2.AuthStyle) Option {
	return func(c *Client) {
		c.OauthConfig.Endpoint.AuthStyle = authStyle
	}
}

var SetScopeOptions = func(scopes []string) Option {
	return func(c *Client) {
		c.OauthConfig.Scopes = scopes
	}
}
