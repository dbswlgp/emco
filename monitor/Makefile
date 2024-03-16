# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

export GO111MODULE=on

all: clean
	GOOS=linux GOARCH=amd64 \
    go build -tags static,system_libgit2,monitor -o ../../bin/monitor/monitor main.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -v ./...

format:
	 @go fmt ./...

clean:
	@rm -f ../../bin/monitor/monitor

.PHONY: cover
cover:
	@go test -p 2 ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

