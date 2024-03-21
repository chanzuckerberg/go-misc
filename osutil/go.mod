module github.com/chanzuckerberg/go-misc/osutil

go 1.21

toolchain go1.21.1

require (
	github.com/chanzuckerberg/go-misc/errors v0.0.0-20240321155940-238914650ee4
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/chanzuckerberg/go-misc/errors => ../errors
