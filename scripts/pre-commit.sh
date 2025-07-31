#!/bin/bash

# Pre-commit script to ensure code quality
# This script runs linting and tests before allowing commits

set -e  # Exit on any error

echo "🔍 Running pre-commit checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Not in a Go module directory. Please run this from the project root."
    exit 1
fi

# Step 1: Format code
echo "📝 Formatting code..."
if go fmt ./...; then
    print_status "Code formatting completed"
else
    print_error "Code formatting failed"
    exit 1
fi

# Step 2: Run go vet
echo "🔍 Running go vet..."
if go vet ./...; then
    print_status "Go vet passed"
else
    print_error "Go vet failed"
    exit 1
fi

# Step 3: Run golangci-lint
echo "🔍 Running golangci-lint..."
GOLANGCI_LINT="${HOME}/go/bin/golangci-lint"
if [ -f "$GOLANGCI_LINT" ]; then
    if "$GOLANGCI_LINT" run; then
        print_status "Golangci-lint passed"
    else
        print_error "Golangci-lint failed"
        exit 1
    fi
else
    print_warning "golangci-lint not found at $GOLANGCI_LINT"
    print_warning "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    print_warning "Skipping linting..."
fi

# Step 4: Run tests
echo "🧪 Running unit tests..."
if go test -v -race -cover ./...; then
    print_status "All tests passed"
else
    print_error "Tests failed"
    exit 1
fi

# Step 5: Check test coverage
echo "📊 Checking test coverage..."
COVERAGE=$(go test -coverprofile=coverage.out ./... | grep "coverage:" | awk '{print $2}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    print_status "Test coverage: ${COVERAGE}% (above 80% threshold)"
else
    print_warning "Test coverage: ${COVERAGE}% (below 80% threshold)"
fi

# Step 6: Run integration tests if they exist
if [ -f "scripts/test.sh" ]; then
    echo "🔧 Running integration tests..."
    if ./scripts/test.sh; then
        print_status "Integration tests passed"
    else
        print_error "Integration tests failed"
        exit 1
    fi
fi

# Step 7: Check for any uncommitted changes after formatting
if [ -n "$(git status --porcelain)" ]; then
    print_warning "Code was reformatted. Please review and commit changes."
    git status --short
    echo ""
    print_warning "You may need to run: git add . && git commit -m 'style: format code'"
fi

print_status "All pre-commit checks passed! 🎉"
print_status "Ready to commit and push to remote." 