GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GOBIN?=$(shell go env GOBIN)
GOFLAGS_MOD ?=
VERSION=$(shell git describe --tags --always)
DOCKER := $(shell type -P podman || type -P docker)
GOENV=GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 GOFLAGS="${GOFLAGS_MOD}"
GOBUILDFLAGS=-gcflags="all=-trimpath=${GOPATH}" -asmflags="all=-trimpath=${GOPATH}"

.PHONY: all
all: lint test

.PHONY: help
help: ## Display this help screen
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build example binaries
	@echo "Building gRPC example"
	@go build -o bin/grpc-example examples/grpc/main.go

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint"
	@$(DOCKER) run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.1 golangci-lint run -v ./examples/... ./kessel/config/... ./kessel/errors/... ./kessel/inventory/...

.PHONY: test
test: ## Run all tests
	@echo "Running tests"
	@go test -v ./kessel/config/... ./kessel/errors/... ./kessel/inventory/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage"
	@go test -coverprofile=coverage.out ./kessel/config/... ./kessel/errors/... ./kessel/inventory/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts"
	@rm -rf bin/
	@rm -f coverage.out coverage.html

.PHONY: mod-tidy
mod-tidy: ## Run go mod tidy
	@echo "Running go mod tidy"
	@go mod tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code"
	@go fmt ./...

.PHONY: generate
generate: ## Generate protobuf files
	@echo "Generating protobuf files"
	@buf generate 