.PHONY: build clean install test lint

# Build variables
BINARY_NAME=lockbox
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/yourusername/lockbox/internal/version.Version=${VERSION} -s -w"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install

# Determine the operating system
ifeq ($(OS),Windows_NT)
	BINARY_SUFFIX=.exe
else
	BINARY_SUFFIX=
endif

# Default target
all: clean build

# Build the application
build:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)$(BINARY_SUFFIX) ./cmd/lockbox

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/

# Install the application
install: build
	cp bin/$(BINARY_NAME)$(BINARY_SUFFIX) /usr/local/bin/$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run linter
lint:
	golangci-lint run

# Development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Cross compilation
.PHONY: build-all
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/lockbox
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/lockbox

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/lockbox
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/lockbox

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/lockbox 