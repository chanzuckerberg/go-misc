package client

const (
	defaultSuccessMessage = "Signed in successfully! You can now return to CLI."
)

type oidcStatus string

var oidcStatusSuccess oidcStatus = "success"

type ClientOption func(*Client)

var SetSuccessMessage = func(successMessage string) ClientOption {
	return func(c *Client) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}
