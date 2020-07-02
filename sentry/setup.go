package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// Setup will initialize the sentry client and return a cleanup func to be used in a defer.Setup
//  Example:
//
//  f, e := sentry.Setup("...", "...")
//  if e != nil { ... }
//  defer f()
func Setup(env, dsn string) (func(), error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: env,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sentry initialization failed")
	}

	f := func() {
		sentry.Flush(time.Second * 5)
		sentry.Recover()
	}
	return f, nil
}
