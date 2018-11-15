LDFLAGS=-ldflags ""

all: test

setup:
	go get github.com/rakyll/gotest
	go install github.com/rakyll/gotest
	go get -u golang.org/x/lint/golint

lint: ## run the fast go linters
	gometalinter --vendor --fast ./...

lint-slow: ## run all linters, even the slow ones
	gometalinter --vendor --deadline 120s ./...

build: ## build the binary
	go build ${LDFLAGS} .

test: ## run the tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

install: ## install the fogg binary in $GOPATH/bin
	go install ${LDFLAGS} .

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint: ## lint the go code
	golint -set_exit_status $(go list ./... | grep -v /vendor/)

.PHONY: build coverage test install lint lint-slow release help
