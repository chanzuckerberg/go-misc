LDFLAGS=-ldflags ""

all: test

setup:
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest

lint: ## run the fast go linters
	gometalinter --vendor --fast ./...

lint-slow: ## run all linters, even the slow ones
	gometalinter --vendor --deadline 120s ./...

release:
	goreleaser release --rm-dist

build: ## build the binary
	go build ${LDFLAGS} .

coverage: ## run the go coverage tool, reading file coverage.out
	go tool cover -html=coverage.out

test: ## run the tests
	gotest -cover ./...

install: ## install the fogg binary in $GOPATH/bin
	go install ${LDFLAGS} .

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build coverage test install lint lint-slow release help
