.PHONY: all build build-all dev test clean install deps help
.PHONY: build-linux build-darwin build-windows docker-build release

BINARY_NAME=vpsctl
VERSION?=0.1.0
BUILD_DIR=./bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Default target
all: clean deps build

## deps: Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## build: Build the binary for current platform
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

## dev: Run in development mode
dev:
	$(GOCMD) run main.go

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## install: Install binary to /usr/local/bin
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed $(BINARY_NAME) to /usr/local/bin/"

## uninstall: Remove binary from /usr/local/bin
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME) from /usr/local/bin/"

## fmt: Format code
fmt:
	$(GOFMT) ./...

## vet: Run go vet
vet:
	$(GOVET) ./...

## lint: Run all linters (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, please install it" && exit 1)
	golangci-lint run ./...

# Cross compilation targets

## build-linux: Build for Linux amd64
build-linux:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

## build-darwin: Build for macOS (amd64 and arm64)
build-darwin:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

## build-windows: Build for Windows amd64
build-windows:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All builds completed in $(BUILD_DIR)/"

# Docker targets

## docker-build: Build Docker image
docker-build:
	docker build -t vpsctl:$(VERSION) .

## docker-run: Run in Docker container
docker-run:
	docker run --rm -v /var/snap/lxd/common/lxd/unix.socket:/var/snap/lxd/common/lxd/unix.socket vpsctl:$(VERSION)

# Release targets

## release: Create release builds for all platforms
release: clean build-all
	@echo "Release $(VERSION) ready in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

## checksum: Generate checksums for release binaries
checksum:
	cd $(BUILD_DIR) && sha256sum * > checksums.txt
	@echo "Checksums generated in $(BUILD_DIR)/checksums.txt"

# Development helpers

## watch: Watch for changes and rebuild (requires reflex)
watch:
	@which reflex > /dev/null || (echo "reflex not found, install with: go install github.com/cespare/reflex@latest" && exit 1)
	reflex -s -r '\.go$$' make build

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

# Help target

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
