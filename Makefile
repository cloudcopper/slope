BINARY_NAME = slope
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     = -ldflags "-X main.Version=$(VERSION)"

.PHONY: build test cover clean

## build: Build the slope binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/slope

## test: Run all tests
test:
	go test -v -race ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)

## cover: Run tests with coverage and show per-package results
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
