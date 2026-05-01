# Makefile for Go project

default: build

BIN := bin/aviatrix-network-policy-controller
GO_LD_FLAGS := "-s -w"

build:
	mkdir -p bin
	go build -ldflags=$(GO_LD_FLAGS) -o $(BIN) .

test:
	go test -v ./...

fmt:
	go fmt ./...

generate:
	go generate ./...

lint: lint-go

GOLANGCI_LINT_VERSION ?= v2.11.4
setup-env:
	if ! command -v golangci-lint &> /dev/null; then \
		echo "Could not find golangci-lint, installing version $(GOLANGCI_LINT_VERSION)."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

lint-go: setup-env
	golangci-lint run

.PHONY: build test lint lint-go fmt generate setup-env
