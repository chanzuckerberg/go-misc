module github.com/chanzuckerberg/go-misc

go 1.14

require (
	cloud.google.com/go v0.61.0 // indirect
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/aws/aws-lambda-go v1.19.1
	github.com/aws/aws-sdk-go v1.36.30
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/getsentry/sentry-go v0.7.0
	github.com/go-errors/errors v1.1.1
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/go-github/v27 v27.0.6
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gruntwork-io/terratest v0.29.0
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-tfe v0.10.2
	github.com/hashicorp/go-version v1.2.1
	github.com/honeycombio/libhoney-go v1.13.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jmoiron/sqlx v1.3.1
	github.com/klauspost/compress v1.10.11 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/nightlyone/lockfile v1.0.0
	github.com/nlopes/slack v0.6.0
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20200819021114-67c6ae64274f // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/snowflakedb/gosnowflake v1.3.13
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/xinsnake/databricks-sdk-golang v0.1.3
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/square/go-jose.v2 v2.5.1
)

replace github.com/hashicorp/go-tfe => github.com/chanzuckerberg/go-tfe v0.9.1-0.20201023195027-6a99188f09d3
