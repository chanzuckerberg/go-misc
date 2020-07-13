export GO111MODULE=on
export CGO_ENABLED=1

all: test

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

lint: ## run the fast go linters
	@golangci-lint run --no-config \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=errcheck --enable=dupl --enable=unparam --enable=goimports \
		--enable=interfacer --enable=unconvert --enable=gosec --enable=megacheck \
		--skip-dirs aws/mocks
.PHONY: lint

deps:
	go get -u ./...
	go mod tidy
.PHONY: deps

test: ## run the tests
ifeq (, $(shell which gotest))
	go test -failfast -cover ./...
else
	gotest -failfast -cover ./...
endif

test-ci: ## run tests in ci (don't try to updated dependencies)
	CGO_ENABLED=1 go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test-ci

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

generate-mocks: deps ## will generate mocks
	@go get -d ./...
	@rm -rf aws/mocks/*
	@cd aws; go generate


.PHONY: build coverage test install lint lint-slow release help
