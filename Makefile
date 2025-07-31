# Makefile for LLM Dispatcher

.PHONY: help test build run setup clean fmt lint check pre-commit

# Default target
help:
	@echo "Available commands:"
	@echo "  test        - Run tests with .env file loading"
	@echo "  test-html   - Run tests with HTML coverage report"
	@echo "  test-coverage - Run tests with detailed coverage"
	@echo "  test-ci     - Run tests without .env (for CI)"
	@echo "  build       - Build the application"
	@echo "  run         - Run the example application (default mode)"
	@echo "  run-vendor  - Run in vendor mode with default vendor"
	@echo "  run-local   - Run in local mode with Ollama"
	@echo "  run-anthropic - Run with Anthropic vendor"
	@echo "  setup       - Setup environment and dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  check       - Run fmt, lint, and test"
	@echo "  pre-commit  - Run pre-commit checks (lint + test)"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "CLI Usage Examples:"
	@echo "  go run cmd/example/cli.go --vendor"
	@echo "  go run cmd/example/cli.go --vendor --vendor-override anthropic"
	@echo "  go run cmd/example/cli.go --local"
	@echo "  go run cmd/example/cli.go --local --model llama2:13b"

# Run tests with .env file loading
test:
	@echo "🧪 Running tests with .env file loading..."
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && go test -v -race -cover ./...; \
	else \
		echo "⚠️  No .env file found. Running tests without environment variables..."; \
		go test -v -race -cover ./...; \
	fi

# Run tests with HTML coverage report
test-html: test
	@echo "📊 Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📄 Coverage report saved to coverage.html"

# Run tests with detailed coverage
test-coverage:
	@echo "📊 Running tests with detailed coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "📊 Coverage summary:"
	@go tool cover -func=coverage.out

# Run tests without .env (for CI)
test-ci:
	@echo "🧪 Running tests for CI (without .env)..."
	@go test -v -race -cover ./...

# Build the application
build:
	@echo "🔨 Building application..."
	@go build -o bin/llmdispatcher cmd/example/cli.go

# Run the example application
run: build
	@echo "🚀 Running example application..."
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && ./bin/llmdispatcher; \
	else \
		echo "⚠️  No .env file found. Please create one or set environment variables."; \
		echo "💡 Run 'make setup' to create a template .env file."; \
	fi

# Run with vendor mode
run-vendor: build
	@echo "🚀 Running in vendor mode..."
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && ./bin/llmdispatcher --vendor; \
	else \
		echo "⚠️  No .env file found. Please create one or set environment variables."; \
	fi

# Run with local mode
run-local: build
	@echo "🚀 Running in local mode..."
	@./bin/llmdispatcher --local

# Run with specific vendor
run-anthropic: build
	@echo "🚀 Running with Anthropic vendor..."
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && ./bin/llmdispatcher --vendor --vendor-override anthropic; \
	else \
		echo "⚠️  No .env file found. Please create one or set environment variables."; \
	fi

# Setup environment and dependencies
setup:
	@echo "🔧 Setting up environment..."
	@go mod tidy
	@if [ ! -f .env ]; then \
		cp cmd/example/env.example .env; \
		echo "📝 Created .env file from template. Please edit it with your API keys."; \
	else \
		echo "📝 .env file already exists."; \
	fi
	@echo "✅ Setup complete!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "✅ Clean complete!"

# Format code
fmt:
	@echo "🔧 Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@go vet ./...
	@GOLANGCI_LINT="$(HOME)/go/bin/golangci-lint" && \
	if [ -f "$$GOLANGCI_LINT" ]; then \
		"$$GOLANGCI_LINT" run; \
	else \
		echo "⚠️  golangci-lint not found at $$GOLANGCI_LINT"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Check code quality
check: fmt lint test

# Pre-commit checks
pre-commit:
	@echo "🔍 Running pre-commit checks..."
	@./scripts/pre-commit.sh 