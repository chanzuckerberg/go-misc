package client

const (
	defaultSuccessMessage = "Signed in successfully! You can now return to CLI."
)

type oidcStatus string

var oidcStatusSuccess oidcStatus = "success"

type clientOption func(*Client)

var ClientOptionSetSuccessMessage = func(successMessage string) clientOption {
	return func(c *Client) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}
