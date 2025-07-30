#!/bin/bash

# Test script for LLM Dispatcher
# This script loads environment variables and runs tests with coverage

set -e  # Exit on any error

echo "🧪 LLM Dispatcher Test Suite"
echo "============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to load .env file
load_env_file() {
    if [ -f ".env" ]; then
        print_status "Loading environment variables from .env file..."
        
        # Read .env file and export variables
        while IFS= read -r line; do
            # Skip empty lines and comments
            if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*# ]]; then
                # Export the variable
                export "$line"
            fi
        done < .env
        
        print_success "Loaded environment variables from .env file"
        
        # Verify that key environment variables are loaded
        if [ -n "$OPENAI_API_KEY" ]; then
            print_success "OpenAI API key loaded"
        else
            print_warning "OPENAI_API_KEY not found in .env file"
        fi
        
        if [ -n "$ANTHROPIC_API_KEY" ]; then
            print_success "Anthropic API key loaded"
        else
            print_warning "ANTHROPIC_API_KEY not found in .env file"
        fi
        
        if [ -n "$GOOGLE_API_KEY" ]; then
            print_success "Google API key loaded"
        else
            print_warning "GOOGLE_API_KEY not found in .env file"
        fi
        
    else
        print_warning ".env file not found. Using system environment variables."
        print_status "To use API keys for testing, create a .env file with your API keys:"
        echo "  cp cmd/example/env.example .env"
        echo "  # Edit .env and add your API keys"
    fi
}

# Load environment variables
load_env_file

echo ""
print_status "Running tests with coverage..."

# Run tests with coverage
if go test -coverprofile=coverage.out -covermode=atomic ./...; then
    print_success "All tests passed!"
    
    # Generate coverage report
    if command -v go tool cover >/dev/null 2>&1; then
        echo ""
        print_status "Generating coverage report..."
        go tool cover -func=coverage.out
        
        # Show coverage summary
        echo ""
        print_status "Coverage Summary:"
        go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $3}'
        
        # Generate HTML coverage report
        if [ "$1" = "--html" ]; then
            print_status "Generating HTML coverage report..."
            go tool cover -html=coverage.out -o coverage.html
            print_success "HTML coverage report saved to coverage.html"
        fi
    fi
    
    # Clean up coverage file
    rm -f coverage.out
    
else
    print_error "Tests failed!"
    exit 1
fi

echo ""
print_success "Test suite completed successfully!" 