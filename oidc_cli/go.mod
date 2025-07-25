module github.com/chanzuckerberg/go-misc/oidc_cli/v3

go 1.24.0

require (
	github.com/aws/aws-sdk-go v1.55.7
	github.com/chanzuckerberg/go-misc/osutil v0.0.0-20250725153648-47b0707ca4b9
	github.com/chanzuckerberg/go-misc/pidlock v0.0.0-20250725155314-6a5b915d3532
	github.com/coreos/go-oidc/v3 v3.14.1
	github.com/go-jose/go-jose/v4 v4.1.1
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/zalando/go-keyring v0.2.6
	golang.org/x/crypto v0.40.0
	golang.org/x/oauth2 v0.30.0
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// breaking change for mac keychains
exclude github.com/zalando/go-keyring v0.2.0

exclude github.com/zalando/go-keyring v0.2.1

replace github.com/chanzuckerberg/go-misc/errors => ../errors

replace github.com/chanzuckerberg/go-misc/osutil => ../osutil

replace github.com/chanzuckerberg/go-misc/pidlock => ../pidlock
