# Makefile for Water Pipe proxy

.PHONY: all build clean test lint

# Variables
BINARY_NAME=water_pipe
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_WINDOWS=$(BINARY_NAME)_windows.exe
GO=go
GOFLAGS=-ldflags="-s -w"

# Default target
all: build

# Build for current platform
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_LINUX) .

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_WINDOWS) .

# Build for all platforms
build-all: build-linux build-windows

# Run tests
test:
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_LINUX) $(BINARY_WINDOWS) coverage.out coverage.html
