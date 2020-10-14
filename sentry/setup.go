package sentry

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

// Run sets up sentry, runs your func and then tears down sentry
func Run(ctx context.Context, f func(context.Context) error) {
	sentryDSN := os.Getenv("SENTRY_DSN")
	sentryEnv := os.Getenv("SENTRY_ENV")

	logrus.Info("setting up sentry")

	teardown, e := Setup(sentryDSN, sentryEnv)
	if e != nil {
		log.Fatal(e)
	}

	defer teardown()

	err := f(ctx)

	if err != nil {
		logrus.Errorf("%+v\n", err)
		sentry.CaptureException(err)
	}
}
