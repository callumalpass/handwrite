# Handwrite - Go Build System

.PHONY: build test clean install help

# Build the binary
build:
	@echo "Building handwrite..."
	@go build -o handwrite .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/handwrite-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 go build -o dist/handwrite-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build -o dist/handwrite-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build -o dist/handwrite-windows-amd64.exe .
	@echo "Built binaries in dist/"

# Run tests
test:
	@echo "Running tests..."
	@go test ./tests/... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./tests/... -v -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Install the binary globally
install: build
	@echo "Installing handwrite to GOBIN..."
	@go install .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f handwrite
	@rm -rf dist/
	@rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Quick development build and test
dev: fmt lint test build
	@echo "Development build complete!"

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  deps         - Install dependencies"
	@echo "  install      - Install binary globally"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  dev-setup    - Setup development environment"
	@echo "  dev          - Quick development build (fmt+lint+test+build)"
	@echo "  help         - Show this help message"