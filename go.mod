module github.com/chanzuckerberg/go-misc

go 1.15

require (
	github.com/AlecAivazis/survey/v2 v2.2.9
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.38.19
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-errors/errors v1.1.1
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/mock v1.5.0
	github.com/google/go-github/v27 v27.0.6
	github.com/google/uuid v1.2.0
	github.com/gruntwork-io/terratest v0.32.19
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-tfe v0.13.0
	github.com/hashicorp/go-version v1.2.1
	github.com/honeycombio/libhoney-go v1.15.2
	github.com/jmoiron/sqlx v1.3.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pty v1.1.8 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v1.0.0
	github.com/nlopes/slack v0.6.0
	github.com/okta/okta-sdk-golang v1.1.0
	github.com/peterhellberg/link v1.1.0
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/snowflakedb/gosnowflake v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/xinsnake/databricks-sdk-golang v0.1.3
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210410081132-afb366fc7cd1 // indirect
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/square/go-jose.v2 v2.5.1
)

replace github.com/hashicorp/go-tfe => github.com/chanzuckerberg/go-tfe v0.9.1-0.20201023195027-6a99188f09d3
