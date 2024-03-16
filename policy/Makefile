# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Aarna Networks, Inc.

all: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -tags netgo -o ../../bin/policy/sacontroller ./cmd/sacontroller/main.go

build: clean test cover
deploy: build

.PHONY: test
test: clean
	@go test -race ./...

format:
	@go fmt ./...

clean:
	@rm -f ../../bin/policy/sacontroller coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
