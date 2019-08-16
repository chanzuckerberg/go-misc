GO_PACKAGE=$(shell go list)
export GOFLAGS=-mod=vendor
export GO111MODULE=on

LDFLAGS=-ldflags ""

all: test

setup:
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

lint: ## run the fast go linters
	@golangci-lint run --no-config \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck
.PHONY: lint

deps:
	go mod tidy
	go mod vendor
.PHONY: deps

build: ## build the binary
	go build ${LDFLAGS} .

test: deps ## run the tests
	gotest -race -coverprofile=coverage.txt -covermode=atomic ./...

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build coverage test install lint lint-slow release help
