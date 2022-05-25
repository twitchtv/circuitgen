export SHELL := /bin/bash
export PATH := $(PWD)/bin:$(PWD):$(PATH)
export GOBIN := $(PWD)/bin

default: test

clean:
	@find internal -name "*.gen.go" -exec rm {} \;
.PHONY: clean

generate: clean
	go build
	go generate ./...
.PHONY: generate

test: lint generate
	go test -race ./...
.PHONY: test

install-tools:
	@mkdir -p bin
	go install github.com/kisielk/errcheck
	go install golang.org/x/lint/golint
	go install golang.org/x/tools/cmd/goimports
	go install github.com/securego/gosec/cmd/gosec
.PHONY: install-tools

lint: install-tools
	go vet ./...
	errcheck -asserts -blank ./...
	golint -set_exit_status ./...
	gosec -quiet ./...
.PHONY: lint

fix: install-tools
	go fmt ./...
	find . -iname "*.go" -print0 | xargs -0 goimports -w
.PHONY: fix

