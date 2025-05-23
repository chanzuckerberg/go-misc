module github.com/chanzuckerberg/go-misc/oidc_cli/v2

go 1.23.0

toolchain go1.23.8

require (
	github.com/aws/aws-sdk-go v1.51.4
	github.com/chanzuckerberg/go-misc/osutil v0.0.0-20240404182313-43e397411f6e
	github.com/chanzuckerberg/go-misc/pidlock v0.0.0-20240320212149-709d6d5c338b
	github.com/coreos/go-oidc v2.3.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	github.com/zalando/go-keyring v0.2.4
	golang.org/x/crypto v0.35.0
	golang.org/x/oauth2 v0.25.0
	gopkg.in/go-jose/go-jose.v2 v2.6.3
)

require (
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// breaking change for mac keychains
exclude github.com/zalando/go-keyring v0.2.0

exclude github.com/zalando/go-keyring v0.2.1

replace github.com/chanzuckerberg/go-misc/errors => ../errors

replace github.com/chanzuckerberg/go-misc/osutil => ../osutil

replace github.com/chanzuckerberg/go-misc/pidlock => ../pidlock
