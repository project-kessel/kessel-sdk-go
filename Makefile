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
all: lint

.PHONY: lint
# run go linter with the repositories lint config
lint:
	@echo "Running golangci-lint"
	@$(DOCKER) run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.1 golangci-lint run -v

.PHONE: gen
gen:
	@echo "Exporting protos from Buf.build"
	./gen.sh

