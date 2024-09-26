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
		c.oauthConfig.Endpoint.AuthStyle = authStyle
	}
}

// This Helper helps you customize the scopes you're sending. It will format the list of strings
//
//	example: https://www.oauth.com/oauth2-servers/server-side-apps/authorization-code/#:~:text=The%20authorization%20URL%20is%20usually%20in%20a%20format%20such%20as%3A
var AddScope = func(scope string) Option {
	return func(c *Client) {
		c.oauthConfig.Scopes = append(c.oauthConfig.Scopes, scope)
	}
}

var AddClientSecret = func(secret string) Option {
	return func(c *Client) {
		c.oauthConfig.ClientSecret = secret
	}
}
