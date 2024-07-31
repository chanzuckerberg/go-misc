
include ../scripts/common.mk

generate-mocks: ## will generate mocks
	go install github.com/golang/mock/mockgen@v1.6.0
	go get -u github.com/aws/aws-sdk-go/...
	go get -u github.com/aws/aws-sdk-go-v2/...
	rm -rf mocks/*
	go generate
	go mod tidy
.PHONY: generate-mocks
