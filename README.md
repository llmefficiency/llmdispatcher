# LLM Dispatcher

A Go library for dispatching LLM requests to different vendors with intelligent routing, retry logic, and fallback capabilities.

## Features

- **Multi-vendor support**: Route requests to OpenAI, Anthropic, Google, and more
- **Intelligent routing**: Route requests based on model, cost, latency, and other criteria
- **Automatic retry**: Configurable retry policies with exponential backoff
- **Fallback support**: Automatic fallback to alternative vendors
- **Usage tracking**: Monitor request statistics and vendor performance
- **Rate limiting**: Built-in rate limiting support
- **Thread-safe**: Safe for concurrent use

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

// Register Anthropic (when implemented)
// anthropicVendor := llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
//     APIKey: os.Getenv("ANTHROPIC_API_KEY"),
// })
// dispatcher.RegisterVendor(anthropicVendor)
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

#### Config
```go
type Config struct {
    DefaultVendor   string       `json:"default_vendor"`
    FallbackVendor  string       `json:"fallback_vendor,omitempty"`
    RetryPolicy     *RetryPolicy `json:"retry_policy,omitempty"`
    RoutingRules    []RoutingRule `json:"routing_rules,omitempty"`
    Timeout         time.Duration `json:"timeout,omitempty"`
    EnableLogging   bool         `json:"enable_logging"`
    EnableMetrics   bool         `json:"enable_metrics"`
}
```

### Methods

#### Dispatcher
- `New() *Dispatcher` - Create a new dispatcher with default config
- `NewWithConfig(config *Config) *Dispatcher` - Create a new dispatcher with custom config
- `Send(ctx context.Context, req *Request) (*Response, error)` - Send a request
- `RegisterVendor(vendor Vendor) error` - Register a vendor
- `GetStats() *Stats` - Get dispatcher statistics
- `GetVendors() []string` - Get list of registered vendors
- `GetVendor(name string) (Vendor, bool)` - Get a specific vendor

#### Vendor Interface
```go
type Vendor interface {
    Name() string
    SendRequest(ctx context.Context, req *Request) (*Response, error)
    GetCapabilities() Capabilities
    IsAvailable(ctx context.Context) bool
}
```

## Statistics

The dispatcher tracks comprehensive statistics:

```go
stats := dispatcher.GetStats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

// Vendor-specific stats
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d successes, %d failures\n",
        vendorName, vendorStats.Requests, vendorStats.Successes, vendorStats.Failures)
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
- **Internal Dispatcher**: 92.6% coverage
- **Internal Vendors**: 90.5% coverage  
- **Public API**: 77.1% coverage

## Supported Vendors

### Currently Implemented
- **OpenAI** - GPT-3.5, GPT-4, and other OpenAI models

### Planned
- **Anthropic** - Claude models
- **Google** - Gemini models
- **Azure OpenAI** - Azure-hosted OpenAI models
- **Cohere** - Command models
- **Hugging Face** - Various open-source models
- **Local models** - Via Ollama integration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests
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

## License

MIT License - see LICENSE file for details.
