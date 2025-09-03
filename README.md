# 🤖 LLM Dispatcher

<div align="center">

**Intelligent LLM Request Routing & Dispatching**

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Code Quality](https://img.shields.io/badge/Code%20Quality-A%2B-9cf)](https://github.com/llmefficiency/llmdispatcher)
[![Security](https://img.shields.io/badge/Security-Scanned-blue)](https://github.com/llmefficiency/llmdispatcher/security)
[![Maintenance](https://img.shields.io/badge/Maintenance-Active-brightgreen)](https://github.com/llmefficiency/llmdispatcher)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen)](https://github.com/llmefficiency/llmdispatcher/pulls)
[![Issues](https://img.shields.io/badge/Issues-Welcome-orange)](https://github.com/llmefficiency/llmdispatcher/issues)
[![Release](https://img.shields.io/badge/Release-v0.3.0-blue)](https://github.com/llmefficiency/llmdispatcher/releases)
[![Last Commit](https://img.shields.io/badge/Last%20Commit-Active-brightgreen)](https://github.com/llmefficiency/llmdispatcher/commits/main)
[![Contributors](https://img.shields.io/badge/Contributors-Welcome-orange)](https://github.com/llmefficiency/llmdispatcher/graphs/contributors)
[![Stars](https://img.shields.io/badge/Stars-⭐-yellow)](https://github.com/llmefficiency/llmdispatcher/stargazers)

</div>

## 🔹 What it does

**A Go library that intelligently routes LLM requests across multiple vendors (OpenAI, Anthropic, Google, Azure) with automatic fallback, retry logic, and cost optimization.**

## 🔹 Why it exists

**The Problem**: Managing multiple LLM vendors is painful:
- ❌ **Vendor lock-in** - Stuck with one provider
- ❌ **API failures** - No fallback when one vendor is down
- ❌ **Cost inefficiency** - Can't optimize for cost vs quality
- ❌ **Complex routing** - Manual vendor selection logic
- ❌ **Rate limits** - No automatic retry and fallback
- ❌ **Monitoring gaps** - No unified metrics across vendors

**The Solution**: LLM Dispatcher provides:
- ✅ **Multi-vendor support** - Route to any combination of vendors
- ✅ **Intelligent routing** - Automatic vendor selection based on model, cost, latency
- ✅ **Automatic fallback** - Seamless failover when vendors are unavailable
- ✅ **Cost optimization** - Route to cheapest vendor for your use case
- ✅ **Streaming support** - Real-time responses with vendor-agnostic interface
- ✅ **Unified monitoring** - Single dashboard for all vendor metrics

## 🔹 Quickstart Installation

### Method 1: From Source (Recommended)
```bash
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
go mod download
make build
```

### Method 2: Docker
```bash
# Build the image
docker build -t llmdispatcher .

# Run with environment variables
docker run -e OPENAI_API_KEY=your-key -e ANTHROPIC_API_KEY=your-key llmdispatcher
```

### Method 3: Using Makefile
```bash
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
make setup  # Sets up dependencies and .env file
make run    # Builds and runs the CLI application
```

## 🔹 Usage Example

### Basic Usage (CLI)

The simplest way to get started is using the CLI application:

```bash
# Set up your API keys
cp env.example .env
# Edit .env with your API keys

# Run with different modes
go run apps/cli/cli.go --vendor                          # Use cloud vendors
go run apps/cli/cli.go --vendor --vendor-override openai # Force OpenAI
go run apps/cli/cli.go --local                           # Use local models (Ollama)
go run apps/cli/cli.go --compare                         # Compare all modes
```

### Basic Usage (Go API)

```go
package main

import (
    "context"
    "log"
    "github.com/llmefficiency/llmdispatcher/internal/dispatcher"
    "github.com/llmefficiency/llmdispatcher/internal/models"
    "github.com/llmefficiency/llmdispatcher/internal/vendors"
)

func main() {
    // 1. Create dispatcher
    disp := dispatcher.New()
    
    // 2. Register a vendor
    config := &models.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    }
    vendor := vendors.NewOpenAI(config)
    disp.RegisterVendor(vendor)
    
    // 3. Send request
    req := &models.Request{
        Model: "gpt-3.5-turbo",
        Messages: []models.Message{{Role: "user", Content: "Hello!"}},
    }
    
    response, err := disp.Send(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s\n", response.Content)
}
```

### Advanced Usage with Strategy-Based Routing

```go
// Configure dispatcher with strategy-based routing
config := &models.Config{
    Strategy:      models.BalancedStrategy, // or SpeedStrategy, QualityStrategy, BudgetStrategy
    Timeout:       30 * time.Second,
    EnableLogging: true,
    EnableMetrics: true,
    RetryPolicy: &models.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: models.ExponentialBackoff,
        RetryableErrors: []string{"rate limit exceeded", "timeout"},
    },
}

disp := dispatcher.NewWithConfig(config)

// Register multiple vendors
openaiVendor := vendors.NewOpenAI(&models.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
disp.RegisterVendor(openaiVendor)

anthropicVendor := vendors.NewAnthropic(&models.VendorConfig{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})
disp.RegisterVendor(anthropicVendor)

// Send request - automatically routes based on strategy
req := &models.Request{
    Model: "gpt-3.5-turbo",
    Messages: []models.Message{{Role: "user", Content: "Hello!"}},
    Strategy: "fast", // Override strategy for this request
}
response, err := disp.Send(context.Background(), req)
```

### Design Choices

**1. Vendor-Agnostic Interface**
- Single API regardless of underlying vendor
- Consistent request/response format
- Automatic vendor-specific translation

**2. Strategy-Based Routing Engine**
- Strategy-based routing (SpeedStrategy, QualityStrategy, BudgetStrategy, BalancedStrategy)
- Model pattern matching for vendor selection
- Basic cost awareness (built into strategy implementations)
- ⚠️ **Note**: Advanced cost and latency optimization features are planned but not yet fully implemented

**3. Resilience & Reliability**
- Automatic retry with exponential backoff
- Seamless fallback to alternative vendors
- Rate limit handling and backoff
- Circuit breaker pattern for failing vendors

**4. Performance & Scalability**
- Thread-safe concurrent operations
- Connection pooling for HTTP clients
- Streaming support for real-time responses
- Minimal memory footprint

**5. Observability & Monitoring**
- Basic statistics tracking
- Vendor performance metrics
- Request success/failure rates
- ⚠️ **Note**: Detailed cost tracking and latency monitoring are planned features

## 🔹 Live Demo

**🚀 Try it now**: [Interactive Demo Web App](https://github.com/llmefficiency/llmdispatcher/tree/main/apps/server)

```bash
# Clone and run the demo web app
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
cp env.example .env
# Edit .env with your API keys

# Start the demo web app
make webservice

# Open http://localhost:8080 in your browser
```

**Demo Web App Features:**
- ✅ **Interactive Chat Interface** - Test different optimization strategies
- ✅ **Real-time Strategy Comparison** - See all strategies stats together in a clear table
- ✅ **Per-strategy Statistics** - Detailed breakdown for each strategy
- ✅ **Streaming Support** - Real-time streaming responses
- ✅ **Multi-vendor Routing** - Automatic vendor selection based on strategy
- ✅ **Cost & Performance Tracking** - Monitor costs and latency across strategies
- ✅ **No Dropdown Selection** - All strategies displayed together for easy comparison

**CLI Demo Features:**
- ✅ Multi-vendor request routing
- ✅ Cost optimization examples
- ✅ Streaming response demo
- ✅ Fallback scenarios
- ✅ Statistics and metrics
- ✅ Local model integration with Ollama

```bash
# CLI demo with different strategies:
# Vendor strategy with automatic vendor selection
go run apps/cli/cli.go --vendor

# Vendor strategy with specific vendor override
go run apps/cli/cli.go --vendor --vendor-override anthropic
go run apps/cli/cli.go --vendor --vendor-override openai

# Local strategy with Ollama (default model: llama2:7b)
go run apps/cli/cli.go --local

# Local strategy with custom model and server
go run apps/cli/cli.go --local --model llama2:13b --server http://localhost:11434

# Strategy comparison test across all optimization strategies
go run apps/cli/cli.go --compare
```

## Features

### 🚀 Core Features
- **Multi-vendor support**: OpenAI, Anthropic, Google, Azure OpenAI, Local (Ollama)
- **Strategy-based routing**: Automatic vendor selection based on optimization strategies (Fast, Sophisticated, Cost-Saving, Auto)
- **Automatic fallback**: Seamless failover when vendors are unavailable
- **Streaming support**: Real-time responses with vendor-agnostic interface
- **Basic cost awareness**: Built into strategy selection (advanced cost optimization planned)
- **Advanced retry**: Configurable retry policies with exponential backoff

### 📊 Monitoring & Analytics
- **Basic metrics**: Request counts, success/failure rates
- **Vendor performance**: Track which vendors are used
- **Basic statistics**: Total requests, response times
- ⚠️ **Planned**: Advanced cost tracking, detailed latency monitoring, usage analytics

### 🔧 Configuration
- **Mode-based routing**: Built-in optimization modes (Fast, Sophisticated, Cost-Saving, Auto)
- **Basic routing rules**: Route by model patterns, token limits, temperature
- **Retry configuration**: Configurable retry policies with exponential backoff
- **Vendor management**: Easy vendor registration and configuration
- **Security**: API key management and secure configuration
- ⚠️ **Planned**: Advanced routing rules, rate limiting, budget controls

## Supported Vendors

| Vendor | Models | Features | Status |
|--------|--------|----------|--------|
| **OpenAI** | GPT-4, GPT-3.5-turbo, GPT-4o, GPT-4o-mini | ✅ Streaming, Rate limiting | ✅ Implemented |
| **Anthropic** | Claude-3-opus, Claude-3-sonnet, Claude-3-haiku | ✅ Large context, Streaming | ✅ Implemented |
| **Google** | Gemini-1.5-pro, Gemini-1.5-flash, Gemini-pro | ✅ Massive context, Streaming | ✅ Implemented |
| **Azure OpenAI** | GPT-4, GPT-3.5-turbo, GPT-4-turbo | ✅ Enterprise deployment | ✅ Implemented |
| **Local (Ollama)** | llama2, mistral, custom models | ✅ Local inference, Free | ✅ Implemented |

## Quick Examples

### Streaming Response
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
    }
}
```

### Get Statistics
```go
stats := dispatcher.GetStats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Successful: %d, Failed: %d\n", stats.SuccessfulRequests, stats.FailedRequests)
fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
fmt.Printf("Total Cost: $%.4f\n", stats.TotalCost)
```

## CLI Usage

The example application supports multiple modes for testing different configurations:

### Vendor Mode
Test with cloud vendors (OpenAI, Anthropic, etc.):

```bash
# Use automatic vendor selection
go run apps/cli/cli.go --vendor

# Use specific vendor override
go run apps/cli/cli.go --vendor --vendor-override anthropic
go run apps/cli/cli.go --vendor --vendor-override openai
```

### Local Mode
Test with local models using Ollama:

```bash
# Use default local model (llama2:7b)
go run apps/cli/cli.go --local

# Use custom model
go run apps/cli/cli.go --local --model llama2:13b
go run apps/cli/cli.go --local --model mistral:7b

# Use custom Ollama server
go run apps/cli/cli.go --local --server http://localhost:11434
```

### Available Options
```bash
go run apps/cli/cli.go --help
```

**Options:**
- `--local` - Run in local mode with Ollama
- `--vendor` - Run in vendor mode with cloud providers
- `--vendor-override` - Specify vendor (anthropic, openai)
- `--model` - Model to use in local mode (default: llama2:7b)
- `--server` - Ollama server URL (default: http://localhost:11434)

## Environment Setup

### 1. Set API Keys
```bash
export OPENAI_API_KEY="sk-your-openai-key"
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
export GOOGLE_API_KEY="your-google-api-key"
export AZURE_OPENAI_API_KEY="your-azure-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
```

### 2. Or use .env file
```bash
cp env.example .env
# Edit .env with your API keys
```

### 3. Run the example
```bash
# Default mode (all vendors, automatic mode)
go run apps/cli/cli.go

# Or use specific modes as shown above
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
./scripts/test.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Documentation

- **[API Reference](docs/API_REFERENCE.md)** - Complete API documentation
- **[Architecture](docs/ARCHITECTURE.md)** - System design and principles
- **[Examples](docs/EXAMPLES.md)** - Comprehensive usage examples
- **[Development](docs/DEVELOPMENT.md)** - Contributing guide
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ by the LLM Efficiency Team**

[![GitHub stars](https://img.shields.io/github/stars/llmefficiency/llmdispatcher?style=social)](https://github.com/llmefficiency/llmdispatcher/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/llmefficiency/llmdispatcher?style=social)](https://github.com/llmefficiency/llmdispatcher/network/members)
[![GitHub issues](https://img.shields.io/github/issues/llmefficiency/llmdispatcher)](https://github.com/llmefficiency/llmdispatcher/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/llmefficiency/llmdispatcher)](https://github.com/llmefficiency/llmdispatcher/pulls)

</div>
