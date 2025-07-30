.PHONY: test test-html test-coverage build clean setup help

# Default target
all: test

# Run tests with environment variable loading
test:
	@echo "🧪 Running tests with .env file loading..."
	@go run scripts/test.go

# Run tests with HTML coverage report
test-html:
	@echo "🧪 Running tests with HTML coverage report..."
	@go run scripts/test.go --html

# Run tests with detailed coverage
test-coverage:
	@echo "🧪 Running tests with detailed coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out
	@rm -f coverage.out

# Run tests without .env loading (for CI)
test-ci:
	@echo "🧪 Running tests in CI mode..."
	@go test -v ./...

# Build the example
build:
	@echo "🔨 Building example..."
	@go build -o bin/example cmd/example/*.go

# Run the example
run:
	@echo "🚀 Running example..."
	@go run cmd/example/*.go

# Setup development environment
setup:
	@echo "🔧 Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from template..."; \
		cp cmd/example/env.example .env; \
		echo "✅ Created .env file. Please edit it with your API keys."; \
	else \
		echo "✅ .env file already exists."; \
	fi
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@echo "✅ Setup complete!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Format code
fmt:
	@echo "🔧 Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Check code quality
check: fmt lint test

# Show help
help:
	@echo "LLM Dispatcher - Available Commands"
	@echo "=================================="
	@echo ""
	@echo "  make test          - Run tests with .env file loading"
	@echo "  make test-html     - Run tests with HTML coverage report"
	@echo "  make test-coverage - Run tests with detailed coverage"
	@echo "  make test-ci       - Run tests without .env (for CI)"
	@echo "  make build         - Build the example application"
	@echo "  make run           - Run the example application"
	@echo "  make setup         - Setup development environment"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make fmt           - Format code with go fmt"
	@echo "  make lint          - Lint code with go vet and golangci-lint"
	@echo "  make check         - Run fmt, lint, and test"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Environment Setup:"
	@echo "  cp cmd/example/env.example .env"
	@echo "  # Edit .env and add your API keys"
	@echo ""
	@echo "Example Usage:"
	@echo "  make setup         # First time setup"
	@echo "  make check         # Run all quality checks"
	@echo "  make test          # Run tests"
	@echo "  make run           # Run example" 