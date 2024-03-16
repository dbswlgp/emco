# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

export GO111MODULE=on

all: clean
	@echo "Compiling Workflow Client"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o ../../../bin/workflowclient/workflowclient ./workflowclient/main.go
	go build -o ../../../bin/workflowclient/http_server ./http_server/main.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -race ./...

format:
	@go fmt ./...

clean:
	@echo "Cleaning up Workflow Client"
	@rm -rf ../../../bin/workflowclient coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html