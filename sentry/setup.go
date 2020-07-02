package sentry

import (
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

func Setup(env, dsn string) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: env,
	})
	if err != nil {
		return errors.Wrap(err, "sentry initialization failed")
	}
	return nil
}
