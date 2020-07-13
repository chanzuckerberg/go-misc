package client

type oidcStatus string

var oidcStatusSuccess oidcStatus = "success"

type clientOpt func(*Client)

var ClientOptSetSuccessMessage = func(successMessage string) clientOpt {
	return func(c *Client) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}
