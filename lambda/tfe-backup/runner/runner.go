package runner

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

type Config struct {
	SentryDSN string `required:"true" envconfig:"sentry_dsn"`
	SentryEnv string `required:"true" envconfig:"sentry_env"`

	TFETokenSecretARN string `required:"true" envconfig:"tfe_token_secret_arn"`
	TFEHostname       string `required:"true" envconfig:"tfe_hostname"`

	KMSKeyARN string `required:"true" envconfig:"kms_key_arn"`
	S3Prefix  string `required:"true" envconfig:"s3_prefix"`
	S3Bucket  string `required:"true" envconfig:"s3_bucket"`
}

func SetupSentry(sentryDSN, env string) (func(), error) {
	if sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         sentryDSN,
			Environment: env,
		})
		if err != nil {
			f := func() {}
			return f, errors.Wrap(err, "Sentry initialization failed")
		}

		f := func() {
			sentry.Flush(time.Second * 5)
			sentry.Recover()
		}
		return f, nil
	}
	return func() {}, nil
}
