package client

import "time"

const (
	defaultSuccessMessage = "Signed in successfully! You can now return to CLI."
)

type oidcStatus string

var oidcStatusSuccess oidcStatus = "success"

type Option func(*Config)

var SetSuccessMessage = func(successMessage string) Option {
	return func(c *Config) {
		c.customMessages[oidcStatusSuccess] = successMessage
	}
}

var SetServerTimeoutDuration = func(timeout time.Duration) Option {
	return func(c *Config) {
		c.ServerConfig.Timeout = timeout
	}
}
