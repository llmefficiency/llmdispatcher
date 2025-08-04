# LLM Dispatcher

A unified interface for dispatching LLM requests to multiple vendors with intelligent routing, retry logic, and streaming support.

## 🚀 Features

### ✅ Implemented Features

- **Multi-Vendor Support**: OpenAI, Anthropic, Google, Azure, and Local (Ollama)
- **Intelligent Routing**: Basic routing rules based on model patterns, token limits, and temperature
- **Retry Logic**: Configurable retry policies with exponential backoff
- **Streaming Support**: Real-time streaming responses from all vendors
- **Web Service**: REST API with health checks, statistics, and vendor testing
- **CLI Interface**: Command-line tool for testing and development
- **Statistics**: Basic metrics and vendor performance tracking
- **Fallback Strategy**: Automatic fallback to secondary vendors

### ❌ Not Yet Implemented

- **Cost Optimization**: Cost-based routing and optimization
- **Latency Optimization**: Latency-based routing and optimization  
- **Advanced Routing**: User ID, request type, and content length routing
- **Rate Limiting**: Per-vendor rate limiting
- **Advanced Metrics**: Cost tracking and advanced performance metrics

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher

# Install dependencies
go mod tidy

# Build the application
make build
```

## 🏃‍♂️ Quick Start

### 1. Setup Environment

```bash
# Copy environment template
cp env.example .env

# Edit with your API keys
OPENAI_API_KEY=sk-your-openai-api-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key
GOOGLE_API_KEY=your-google-api-key
AZURE_OPENAI_API_KEY=your-azure-api-key
```

### 2. Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
    // Create dispatcher
    dispatcher := llmdispatcher.New()

    // Register OpenAI vendor
    openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: "your-openai-api-key",
    })
    dispatcher.RegisterVendor(openaiVendor)

    // Send request
    req := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
        Temperature: 0.7,
        MaxTokens:   1000,
    }

    resp, err := dispatcher.Send(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Vendor: %s\n", resp.Vendor)
}
```

### 3. Web Service

```bash
# Start the web service
make webservice

# Test the API
curl http://localhost:8080/api/v1/health
```

## 📚 Documentation

- **[API Reference](API_REFERENCE.md)**: Complete API documentation
- **[Examples](EXAMPLES.md)**: Comprehensive usage examples
- **[Architecture](ARCHITECTURE.md)**: System design and architecture
- **[Development](DEVELOPMENT.md)**: Development setup and guidelines
- **[Local Vendor](LOCAL_VENDOR.md)**: Local model integration guide
- **[Troubleshooting](TROUBLESHOOTING.md)**: Common issues and solutions

## 🔧 Configuration

### Basic Configuration

```go
config := &llmdispatcher.Config{
    DefaultVendor:  "openai",
    FallbackVendor: "anthropic",
    Timeout:        30 * time.Second,
    EnableLogging:  true,
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
    },
}
```

### Routing Rules

```go
config := &llmdispatcher.Config{
    RoutingRules: []llmdispatcher.RoutingRule{
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "gpt-4",     // Route GPT-4 to OpenAI
                MaxTokens:    2000,
            },
            Vendor:   "openai",
            Priority: 1,
            Enabled:  true,
        },
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "claude-3",  // Route Claude to Anthropic
                MaxTokens:    1000,
            },
            Vendor:   "anthropic",
            Priority: 2,
            Enabled:  true,
        },
    },
}
```

## 🌐 Web Service API

The dispatcher includes a REST API with the following endpoints:

- `GET /api/v1/health` - Health check
- `POST /api/v1/chat/completions` - Chat completions
- `POST /api/v1/chat/completions/stream` - Streaming chat
- `POST /api/v1/test/vendor` - Test specific vendor
- `GET /api/v1/stats` - Statistics
- `GET /api/v1/vendors` - Available vendors

## 🖥️ CLI Usage

```bash
# Run with all vendors
go run apps/cli/cli.go

# Vendor mode with specific vendor
go run apps/cli/cli.go --vendor --vendor-override anthropic

# Local mode with Ollama
go run apps/cli/cli.go --local --model llama2:7b
```

## 🏗️ Architecture

The dispatcher uses a modular architecture:

- **Core Dispatcher**: Main routing and coordination logic
- **Vendor Interfaces**: Unified interface for all vendors
- **Configuration System**: Flexible configuration management
- **Web Service**: REST API layer
- **CLI Interface**: Command-line interface

## 🔄 Vendor Support

### Cloud Vendors

- **OpenAI**: GPT-3.5, GPT-4, GPT-4o
- **Anthropic**: Claude-3 models
- **Google**: Gemini models
- **Azure OpenAI**: Azure-hosted OpenAI models

### Local Vendors

- **Ollama**: Local model inference
- **llama.cpp**: Direct model execution
- **Custom HTTP**: Custom model servers

## 📊 Statistics

The dispatcher provides comprehensive statistics:

```go
stats := dispatcher.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n", 
    float64(stats.SuccessfulRequests)/float64(stats.TotalRequests)*100)

// Vendor-specific stats
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d successes\n", 
        vendorName, vendorStats.Requests, vendorStats.Successes)
}
```

## 🚨 Error Handling

Comprehensive error handling with automatic retries:

```go
resp, err := dispatcher.Send(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, llmdispatcher.ErrNoVendorsRegistered):
        log.Fatal("No vendors registered")
    case errors.Is(err, llmdispatcher.ErrVendorUnavailable):
        log.Printf("All vendors unavailable")
    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

## 🔧 Development

### Prerequisites

- Go 1.24 or higher
- Make (for build scripts)
- golangci-lint (for linting)

### Setup

```bash
# Install dependencies
go mod tidy

# Run tests
make test

# Run linting
make lint

# Build
make build
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for CI
make test-ci
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting: `make check`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check the [docs](docs/) folder
- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join discussions for questions and ideas

## 🗺️ Roadmap

### Planned Features

- [ ] Cost optimization routing
- [ ] Latency optimization routing
- [ ] Advanced routing conditions
- [ ] Rate limiting
- [ ] Advanced cost tracking
- [ ] More vendor integrations
- [ ] Kubernetes deployment support
- [ ] Prometheus metrics integration

### Current Status

The dispatcher is currently in active development with core functionality implemented and stable. The focus is on improving reliability, adding advanced routing features, and expanding vendor support. 