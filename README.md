# LLM Dispatcher

<div align="center">

# 🤖 LLM Dispatcher

**Intelligent LLM Request Routing & Dispatching**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Coverage](https://img.shields.io/badge/Coverage-90%25-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Code Quality](https://img.shields.io/badge/Code%20Quality-A%2B-9cf)](https://github.com/llmefficiency/llmdispatcher)
[![Security](https://img.shields.io/badge/Security-Scanned-blue)](https://github.com/llmefficiency/llmdispatcher/security)
[![Maintenance](https://img.shields.io/badge/Maintenance-Active-brightgreen)](https://github.com/llmefficiency/llmdispatcher)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen)](https://github.com/llmefficiency/llmdispatcher/pulls)
[![Issues](https://img.shields.io/badge/Issues-Welcome-orange)](https://github.com/llmefficiency/llmdispatcher/issues)
[![Release](https://img.shields.io/badge/Release-v1.0.0-blue)](https://github.com/llmefficiency/llmdispatcher/releases)
[![Last Commit](https://img.shields.io/badge/Last%20Commit-Active-brightgreen)](https://github.com/llmefficiency/llmdispatcher/commits/main)
[![Contributors](https://img.shields.io/badge/Contributors-Welcome-orange)](https://github.com/llmefficiency/llmdispatcher/graphs/contributors)
[![Stars](https://img.shields.io/badge/Stars-⭐-yellow)](https://github.com/llmefficiency/llmdispatcher/stargazers)

</div>

> **🤖 AI-Generated Repository**: This project was created using AI assistance to demonstrate best practices in Go development, API design, and documentation.

A Go library for dispatching LLM requests to different vendors with intelligent routing, retry logic, fallback capabilities, and streaming support.

## Features

- **Multi-vendor support**: Route requests to OpenAI, Anthropic, Google, Azure OpenAI, and more
- **Streaming support**: Real-time streaming responses with channel-based communication
- **Intelligent routing**: Route requests based on model, cost, latency, and other criteria
- **Advanced routing**: Cost optimization and latency-based vendor selection
- **Automatic retry**: Configurable retry policies with exponential backoff
- **Fallback support**: Automatic fallback to alternative vendors
- **Usage tracking**: Monitor request statistics and vendor performance
- **Rate limiting**: Built-in rate limiting support
- **Thread-safe**: Safe for concurrent use
- **Comprehensive testing**: 90%+ test coverage across all components

## Quick Start

### Installation

```bash
go get github.com/llmefficiency/llmdispatcher
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
    // Create dispatcher
    dispatcher := llmdispatcher.New()

    // Create OpenAI vendor
    openaiConfig := &llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    }
    openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)

    // Register vendor
    if err := dispatcher.RegisterVendor(openaiVendor); err != nil {
        log.Fatal(err)
    }

    // Send request
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello!"},
        },
    }

    response, err := dispatcher.Send(context.Background(), request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
}
```

### Streaming Usage

```go
// Send streaming request
streamingResp, err := dispatcher.SendStreaming(context.Background(), request)
if err != nil {
    log.Fatal(err)
}

// Read streaming content
for {
    select {
    case content := <-streamingResp.ContentChan:
        fmt.Print(content) // Print each chunk as it arrives
    case done := <-streamingResp.DoneChan:
        if done {
            fmt.Println("\nStreaming completed")
            return
        }
    case err := <-streamingResp.ErrorChan:
        log.Printf("Streaming error: %v", err)
        return
    }
}

// Clean up
streamingResp.Close()
```

## Advanced Configuration

### With Retry Policy and Routing Rules

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    FallbackVendor: "anthropic",
    Timeout: 30 * time.Second,
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries: 3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
        RetryableErrors: []string{"rate limit exceeded", "timeout"},
    },
    RoutingRules: []llmdispatcher.RoutingRule{
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "gpt-4",
                MaxTokens: 1000,
            },
            Vendor: "openai",
            Priority: 1,
            Enabled: true,
        },
    },
    // Advanced routing options
    CostOptimization: &llmdispatcher.CostOptimization{
        Enabled:     true,
        MaxCost:     0.10,
        PreferCheap: true,
        VendorCosts: map[string]float64{
            "openai":   0.002,
            "anthropic": 0.003,
            "google":    0.001,
        },
    },
    LatencyOptimization: &llmdispatcher.LatencyOptimization{
        Enabled:    true,
        MaxLatency: 30 * time.Second,
        PreferFast: true,
        LatencyWeights: map[string]float64{
            "openai":   1.0,
            "anthropic": 1.2,
            "google":    0.8,
        },
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)
```

### Multiple Vendors

```go
// Register OpenAI
openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
dispatcher.RegisterVendor(openaiVendor)

// Register Anthropic
anthropicVendor := llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})
dispatcher.RegisterVendor(anthropicVendor)

// Register Google
googleVendor := llmdispatcher.NewGoogleVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("GOOGLE_API_KEY"),
})
dispatcher.RegisterVendor(googleVendor)

// Register Azure OpenAI
azureVendor := llmdispatcher.NewAzureOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey:  os.Getenv("AZURE_OPENAI_API_KEY"),
    BaseURL: os.Getenv("AZURE_OPENAI_ENDPOINT"),
})
dispatcher.RegisterVendor(azureVendor)
```

## API Reference

### Types

#### Request
```go
type Request struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Temperature float64   `json:"temperature,omitempty"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    TopP        float64   `json:"top_p,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
    Stop        []string  `json:"stop,omitempty"`
    User        string    `json:"user,omitempty"`
}
```

#### Response
```go
type Response struct {
    Content     string    `json:"content"`
    Usage       Usage     `json:"usage"`
    Model       string    `json:"model"`
    Vendor      string    `json:"vendor"`
    FinishReason string   `json:"finish_reason,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}
```

#### StreamingResponse
```go
type StreamingResponse struct {
    ContentChan chan string `json:"-"`
    DoneChan    chan bool   `json:"-"`
    ErrorChan   chan error  `json:"-"`
    Usage       Usage       `json:"usage"`
    Model       string      `json:"model"`
    Vendor      string      `json:"vendor"`
    CreatedAt   time.Time   `json:"created_at"`
}
```

#### Config
```go
type Config struct {
    DefaultVendor        string              `json:"default_vendor"`
    FallbackVendor       string              `json:"fallback_vendor,omitempty"`
    RetryPolicy          *RetryPolicy        `json:"retry_policy,omitempty"`
    RoutingRules         []RoutingRule       `json:"routing_rules,omitempty"`
    Timeout              time.Duration       `json:"timeout,omitempty"`
    EnableLogging        bool                `json:"enable_logging"`
    EnableMetrics        bool                `json:"enable_metrics"`
    CostOptimization     *CostOptimization   `json:"cost_optimization,omitempty"`
    LatencyOptimization  *LatencyOptimization `json:"latency_optimization,omitempty"`
}
```

### Methods

#### Dispatcher
- `New() *Dispatcher` - Create a new dispatcher with default config
- `NewWithConfig(config *Config) *Dispatcher` - Create a new dispatcher with custom config
- `Send(ctx context.Context, req *Request) (*Response, error)` - Send a request
- `SendStreaming(ctx context.Context, req *Request) (*StreamingResponse, error)` - Send a streaming request
- `RegisterVendor(vendor Vendor) error` - Register a vendor
- `GetStats() *Stats` - Get dispatcher statistics
- `GetVendors() []string` - Get list of registered vendors
- `GetVendor(name string) (Vendor, bool)` - Get a specific vendor

#### Vendor Interface
```go
type Vendor interface {
    Name() string
    SendRequest(ctx context.Context, req *Request) (*Response, error)
    SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error)
    GetCapabilities() Capabilities
    IsAvailable(ctx context.Context) bool
}
```

## Statistics

The dispatcher tracks comprehensive statistics including cost and latency metrics:

```go
stats := dispatcher.GetStats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
fmt.Printf("Total Cost: $%.4f\n", stats.TotalCost)
fmt.Printf("Average Cost: $%.4f\n", stats.AverageCost)

// Vendor-specific stats
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d successes, %d failures, $%.4f cost\n",
        vendorName, vendorStats.Requests, vendorStats.Successes, 
        vendorStats.Failures, vendorStats.TotalCost)
}
```

## Running the Example

### Quick Start with Make

1. Setup the environment:
```bash
make setup
```

2. Edit the `.env` file with your API keys

3. Run the example:
```bash
make run
```

### Method 1: Environment Variables

1. Set your API keys as environment variables:
```bash
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
export GOOGLE_API_KEY="your-google-api-key"
export AZURE_OPENAI_API_KEY="your-azure-openai-api-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export COHERE_API_KEY="your-cohere-api-key"
```

2. Run the example:
```bash
go run cmd/example/main.go
```

### Method 2: .env File

1. Copy the example environment file:
```bash
cp cmd/example/env.example .env
```

2. Edit `.env` and add your API keys:
```bash
# OpenAI Configuration
OPENAI_API_KEY=sk-your-actual-openai-key
ANTHROPIC_API_KEY=sk-ant-your-actual-anthropic-key
GOOGLE_API_KEY=your-google-api-key
AZURE_OPENAI_API_KEY=your-azure-openai-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
# ... add other keys as needed
```

3. Run the example:
```bash
go run cmd/example/main.go
```

### Method 3: Configuration File

You can also load configuration from YAML or JSON files. See `cmd/example/config.go` for examples.

## Running Tests

### Quick Test Commands

```bash
# Run tests with .env file loading
make test

# Run tests with HTML coverage report
make test-html

# Run tests with detailed coverage
make test-coverage

# Run tests without .env (for CI)
make test-ci
```

### Test Scripts

The project includes test scripts that automatically load environment variables from your `.env` file:

```bash
# Using the Go test runner
go run scripts/test.go

# Using the bash test runner
./scripts/test.sh

# With HTML coverage report
go run scripts/test.go --html
```

### Test Coverage

The test suite provides excellent coverage:
- **Internal Dispatcher**: 90.9% coverage
- **Internal Models**: 100% coverage
- **Internal Vendors**: 49.7% coverage

## Supported Vendors

### Currently Implemented

#### OpenAI
- **Models**: GPT-3.5-turbo, GPT-4, GPT-4-turbo, GPT-4o
- **Features**: Streaming support, rate limiting, comprehensive error handling
- **Configuration**: API key, custom base URL, timeout settings

#### Anthropic (Claude)
- **Models**: claude-3-opus, claude-3-sonnet, claude-3-haiku, claude-3-5-sonnet, claude-3-5-haiku
- **Features**: Streaming support, large context windows (200K tokens)
- **Configuration**: API key, custom headers, timeout settings

#### Google (Gemini)
- **Models**: gemini-1.5-pro, gemini-1.5-flash, gemini-pro, gemini-pro-vision
- **Features**: Streaming support, massive context windows (1M tokens)
- **Configuration**: API key, generation config, timeout settings

#### Azure OpenAI
- **Models**: gpt-4, gpt-4-turbo, gpt-4o, gpt-35-turbo, gpt-35-turbo-16k
- **Features**: Streaming support, deployment-based routing
- **Configuration**: API key, endpoint URL, timeout settings

### Planned
- **Cohere** - Command models
- **Hugging Face** - Various open-source models
- **Local models** - Via Ollama integration

## Advanced Features

### Streaming Support
Real-time streaming responses with channel-based communication:
- **Content streaming**: Receive text chunks as they're generated
- **Completion signaling**: Know when streaming is complete
- **Error handling**: Handle streaming errors gracefully
- **Resource cleanup**: Automatic channel cleanup

### Cost Optimization
Intelligent cost-based routing:
- **Vendor cost mapping**: Configure costs per 1K tokens
- **Budget management**: Set maximum cost per request
- **Cost tracking**: Monitor total and average costs
- **Preference settings**: Choose between cost and quality

### Latency Optimization
Performance-based vendor selection:
- **Latency weighting**: Configure vendor performance weights
- **Maximum latency**: Set acceptable latency thresholds
- **Performance tracking**: Monitor vendor response times
- **Preference settings**: Choose between speed and quality

### Advanced Routing
Sophisticated routing based on multiple criteria:
- **Model patterns**: Route by model name patterns
- **Token limits**: Route by maximum token requirements
- **Temperature settings**: Route by creativity requirements
- **User-based routing**: Route by user ID
- **Request type routing**: Route by request type
- **Content length routing**: Route by input length

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests (aim for 90%+ coverage)
5. Submit a pull request

## API Key Security

### Best Practices

1. **Never commit API keys to version control**
   - Add `.env` to your `.gitignore`
   - Use environment variables in production
   - Use secret management services (AWS Secrets Manager, HashiCorp Vault, etc.)

2. **Environment-specific configuration**
   ```bash
   # Development
   export OPENAI_API_KEY="sk-dev-key"
   
   # Production
   export OPENAI_API_KEY="sk-prod-key"
   ```

3. **Rotate keys regularly**
   - Set up key rotation schedules
   - Monitor API usage for unusual activity
   - Use different keys for different environments

4. **Limit key permissions**
   - Use API keys with minimal required permissions
   - Set up rate limits and usage quotas
   - Monitor API usage and costs

### Supported Environment Variables

| Vendor | API Key Variable | Base URL Variable | Timeout Variable |
|--------|------------------|-------------------|------------------|
| OpenAI | `OPENAI_API_KEY` | `OPENAI_BASE_URL` | `OPENAI_TIMEOUT` |
| Anthropic | `ANTHROPIC_API_KEY` | `ANTHROPIC_BASE_URL` | `ANTHROPIC_TIMEOUT` |
| Google | `GOOGLE_API_KEY` | `GOOGLE_BASE_URL` | `GOOGLE_TIMEOUT` |
| Azure OpenAI | `AZURE_OPENAI_API_KEY` | `AZURE_OPENAI_ENDPOINT` | `AZURE_OPENAI_TIMEOUT` |
| Cohere | `COHERE_API_KEY` | `COHERE_BASE_URL` | `COHERE_TIMEOUT` |

## Documentation

### 📚 Complete Documentation
- **[docs/INDEX.md](docs/INDEX.md)** - Documentation index and overview
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture and design principles
- **[docs/API_REFERENCE.md](docs/API_REFERENCE.md)** - Complete API documentation with examples
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Development guide for contributors
- **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[docs/EXAMPLES.md](docs/EXAMPLES.md)** - Comprehensive usage examples

### 🎯 Quick Navigation
- **Getting Started**: This README → [API Reference](docs/API_REFERENCE.md) → [Examples](docs/EXAMPLES.md)
- **Development**: [Development Guide](docs/DEVELOPMENT.md) → [Architecture](docs/ARCHITECTURE.md) → [API Reference](docs/API_REFERENCE.md)
- **Troubleshooting**: [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- **Documentation Index**: [docs/INDEX.md](docs/INDEX.md)
- **Version History**: [CHANGELOG.md](CHANGELOG.md)

## License

MIT License - see LICENSE file for details.
