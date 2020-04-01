export GOFLAGS=-mod=vendor
export GO111MODULE=on

all: test

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

lint: ## run the fast go linters
	@golangci-lint run --no-config \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck
.PHONY: lint

deps:
	go get -u ./...
	go mod tidy
	go mod vendor
.PHONY: deps

test: deps ## run the tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

test-ci: ## run tests in ci (don't try to updated dependencies)
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test-ci

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

generate-mocks: deps## will generate mocks
	rm -rf aws/mocks/*
	_bin/generate_mocks.sh

.PHONY: build coverage test install lint lint-slow release help
