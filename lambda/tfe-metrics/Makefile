SHA=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)
DIRTY=false
# TODO add release flag
GO_PACKAGE=$(shell go list)
LDFLAGS=-ldflags "-w -s -X $(GO_PACKAGE)/util.GitSha=${SHA} -X $(GO_PACKAGE)/util.Version=${VERSION} -X $(GO_PACKAGE)/util.Dirty=${DIRTY}"
export AWS_PROFILE=czi-tfe

all: test install

setup: ## setup development dependencies
	curl -sfL https://raw.githubusercontent.com/chanzuckerberg/bff/main/download.sh | sh
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh
.PHONY: setup

fmt:
	goimports -w -d $$(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./dist/*")
.PHONY: fmt

lint: ## run the fast go linters
	./bin/reviewdog -conf .reviewdog.yml  -diff "git diff origin/master" -tee
.PHONY: lint

lint-ci: ## run the fast go linters
	./bin/reviewdog -conf .reviewdog.yml  -reporter=github-pr-review -tee -level=info
.PHONY: lint-ci

lint-all: ## run the fast go linters
	# doesn't seem to be a way to get reviewdog to not filter by diff
	./bin/golangci-lint run
.PHONY: lint

docker: ## check to be sure docker is running
	@docker ps
.PHONY: docker

release: setup docker ## run a release
	./bin/bff bump
	git push
	goreleaser release
.PHONY: release

release-prerelease: setup ## release to github as a 'pre-release'
	go build ${LDFLAGS} .
	version=`./tfe-metrics version`; \
	git tag v"$$version"; \
	git push
	git push --tags
	goreleaser release -f .goreleaser.prerelease.yml --debug
.PHONY: release-prerelease

release-snapshot: setup ## run a release
	goreleaser release --snapshot
.PHONY: release-snapshot

build: ## build the binary
	go build ${LDFLAGS} .
.PHONY: build

deps:
	go mod tidy
.PHONY: deps

coverage: ## run the go coverage tool, reading file coverage.out
	go tool cover -html=coverage.out
.PHONY: coverage

test: fmt deps ## run tests
 ifeq (, $(shell which gotest))
	go test -failfast -cover ./...
 else
	gotest -failfast -cover ./...
 endif
.PHONY: test

test-ci: ## run tests
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
.PHONY: test-ci

test-coverage:  ## run the test with proper coverage reporting
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out
.PHONY: test-coverage

install: ## install the fogg binary in $GOPATH/bin
	go install ${LDFLAGS} .
.PHONY: install

help: ## display help for this makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

clean: ## clean the repo
	rm tfe-metrics 2>/dev/null || true
	go clean
	go clean -testcache
	rm -rf dist 2>/dev/null || true
	rm coverage.out 2>/dev/null || true
.PHONY: clean

package:
	GOARCH=amd64 GOOS=linux go build ${LDFLAGS} .
	zip tfe-metrics.zip tfe-metrics
.PHONY: package

deploy-dev: package
	aws s3 cp tfe-metrics.zip s3://tfe-metrics-packages/dev.zip
	aws --region us-west-2 --profile czi-tfe  lambda update-function-code --function-name tfe-dev-metrics --s3-bucket tfe-metrics-packages --s3-key dev.zip
.PHONY: deploy-dev

deploy-prod: package
	aws s3 cp tfe-metrics.zip s3://tfe-metrics-packages/prod.zip
	aws --region us-west-2 --profile czi-tfe  lambda update-function-code --function-name tfe-prod-metrics --s3-bucket tfe-metrics-packages --s3-key prod.zip
.PHONY: deploy-prod
