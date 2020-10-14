module github.com/chanzuckerberg/tfe-metrics

go 1.14

require (
	github.com/aws/aws-lambda-go v1.17.0
	github.com/aws/aws-sdk-go v1.33.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/chanzuckerberg/go-misc v0.0.0-20200702230056-488cdb62f023
	github.com/getsentry/sentry-go v0.6.1
	github.com/hashicorp/go-tfe v0.9.0
	github.com/honeycombio/libhoney-go v1.12.4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

// use this PR until it is merged https://github.com/hashicorp/go-tfe/pull/126
replace github.com/hashicorp/go-tfe => github.com/chanzuckerberg/go-tfe v0.6.1-0.20200701222238-d88b093cbe7e
