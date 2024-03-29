export GO111MODULE=on
export CGO_ENABLED=0

all: test

clean: ## clean the repo
	rm go-misc 2>/dev/null || true
	go clean
	go clean -testcache
	rm -rf dist 2>/dev/null || true
	rm coverage.out 2>/dev/null || true
.PHONY: clean

setup:
	brew install golangci-lint
	brew upgrade golangci-lint
	brew install reviewdog/tap/reviewdog
	brew upgrade reviewdog/tap/reviewdog
.PHONY: setup

lint: ## run lint on changes from main
	reviewdog -conf ../.reviewdog.yml  -diff "git diff main"
.PHONY: lint

lint-all: ## run linters across all code
	reviewdog -conf ../.reviewdog.yml  -filter-mode nofilter
.PHONY: lint-all

deps:
	go get ./...
	go mod tidy
.PHONY: deps

test: ## run the tests
ifeq (, $(shell which gotest))
	go test -failfast -cover ./...
else
	gotest -failfast -cover ./...
endif
.PHONY: test

test-ci: ## run tests in ci (don't try to updated dependencies)
	CGO_ENABLED=1 go test -race -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test-ci

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod

fmt:
	gofmt -w -s .
.PHONY: fmt
