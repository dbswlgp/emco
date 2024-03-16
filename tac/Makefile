# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

export GO111MODULE=on

all: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -tags netgo -o ../../bin/tac/tac ./cmd/main.go
build: clean test cover

deploy: build

.PHONY: test
test: clean
	@go test -race ./...

format:
	@go fmt ./...

clean:
	@rm -f ../../bin/tac/tac coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
