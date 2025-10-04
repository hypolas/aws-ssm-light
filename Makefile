# Makefile for AWS SSM

# Version information
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)")
GIT_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MAINTAINER ?= Nicolas HYPOLITE

# Build flags
LDFLAGS = -w -s -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.Maintainer="$(MAINTAINER)"

.PHONY: help test test-race test-coverage build clean lint fmt vet install-tools benchmark

# Default target
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: install-tools ## Run golangci-lint
	golangci-lint run --timeout=5m

# Testing
test: ## Run all tests
	go test -v ./...

test-short: ## Run short tests only
	go test -short -v ./...

test-race: ## Run tests with race detection
	go test -race -short ./...

test-coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-integration: ## Run integration tests (requires AWS credentials)
	AWS_SKIP_INTEGRATION_TESTS=false go test -v ./...

benchmark: ## Run benchmarks
	go test -bench=. -benchmem -run=^$$ ./...

# Building
build: ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o aws-ssm main.go

build-all: ## Build for all platforms
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/aws-ssm-windows-arm64.exe main.go

install: ## Install the binary to $GOPATH/bin
	go install -ldflags="$(LDFLAGS)"

# Maintenance
clean: ## Clean build artifacts
	go clean
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f aws-ssm aws-ssm.exe

deps: ## Download dependencies
	go mod download
	go mod tidy

# Tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing gosec..."
	@curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest || \
	go install github.com/securecodewarrior/gosec/cmd/gosec@latest

security: install-tools ## Run security analysis
	gosec ./...

# CI/CD helpers
ci-test: ## Run all CI tests
	$(MAKE) fmt vet lint test test-race test-coverage

# Docker
docker-build: ## Build Docker image
	docker build -t aws-ssm:latest .

# Release helpers
tag: ## Create a new git tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required. Usage: make tag VERSION=v1.0.0"; exit 1; fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Development workflow
dev: ## Run development checks (fmt, vet, test)
	$(MAKE) fmt vet test-short

all: ## Run all checks and build
	$(MAKE) ci-test build