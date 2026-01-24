# Provider development settings
HOSTNAME=registry.terraform.io
NAMESPACE=reserve-protocol
NAME=dokploy
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

.PHONY: build
build:
	go build -v ./...

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	go build -o ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY}

.PHONY: test
test:
	go test -v ./internal/client/...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -timeout 120m ./internal/...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w .

.PHONY: docs
docs:
	go generate ./...

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development."
	go test -v -sweep=all ./internal/acctest

.PHONY: clean
clean:
	rm -rf dist/
	rm -f ${BINARY}
