# 🤖 LLM Dispatcher

<div align="center">

**Intelligent LLM Request Routing & Dispatching**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Coverage](https://img.shields.io/badge/Coverage-45.9%25-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/llmefficiency/llmdispatcher/actions)
[![Code Quality](https://img.shields.io/badge/Code%20Quality-A%2B-9cf)](https://github.com/llmefficiency/llmdispatcher)
[![Security](https://img.shields.io/badge/Security-Scanned-blue)](https://github.com/llmefficiency/llmdispatcher/security)
[![Maintenance](https://img.shields.io/badge/Maintenance-Active-brightgreen)](https://github.com/llmefficiency/llmdispatcher)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen)](https://github.com/llmefficiency/llmdispatcher/pulls)
[![Issues](https://img.shields.io/badge/Issues-Welcome-orange)](https://github.com/llmefficiency/llmdispatcher/issues)
[![Release](https://img.shields.io/badge/Release-v0.1.0-blue)](https://github.com/llmefficiency/llmdispatcher/releases)
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

### Method 1: Go Install (Recommended)
```bash
go install github.com/llmefficiency/llmdispatcher/cmd/example@latest
```

### Method 2: Docker
```bash
# Build the image
docker build -t llmdispatcher .

# Run with environment variables
docker run -e OPENAI_API_KEY=your-key -e ANTHROPIC_API_KEY=your-key llmdispatcher
```

### Method 3: From Source
```bash
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
go mod download
go run cmd/example/cli.go
```

## 🔹 Usage Example

### Basic Usage (5 lines of code)

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
    // 1. Create dispatcher
    dispatcher := llmdispatcher.New()
    
    // 2. Register vendors
    openai := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    dispatcher.RegisterVendor(openai)
    
    // 3. Send request (automatic routing & fallback)
    response, err := dispatcher.Send(context.Background(), &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{{Role: "user", Content: "Hello!"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Response: %s\n", response.Content)
}
```

### Advanced Usage with Cost Optimization

```go
// Configure intelligent routing
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    FallbackVendor: "anthropic",
    CostOptimization: &llmdispatcher.CostOptimization{
        Enabled: true,
        MaxCost: 0.10,
        VendorCosts: map[string]float64{
            "openai":   0.002, // $0.002 per 1K tokens
            "anthropic": 0.003, // $0.003 per 1K tokens
            "google":    0.001, // $0.001 per 1K tokens
        },
    },
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries: 3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)

// Register multiple vendors
dispatcher.RegisterVendor(llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
}))
dispatcher.RegisterVendor(llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
}))

// Send request - automatically routes to cheapest available vendor
response, err := dispatcher.Send(context.Background(), request)
```

### Design Choices

**1. Vendor-Agnostic Interface**
- Single API regardless of underlying vendor
- Consistent request/response format
- Automatic vendor-specific translation

**2. Intelligent Routing Engine**
- Model-based routing (GPT-4 → OpenAI, Claude → Anthropic)
- Cost optimization (route to cheapest vendor)
- Latency optimization (route to fastest vendor)
- Custom routing rules (user-defined logic)

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
- Comprehensive statistics tracking
- Vendor performance metrics
- Cost and latency monitoring
- Request success/failure rates

## 🔹 Live Demo

**🚀 Try it now**: [Interactive Demo](https://github.com/llmefficiency/llmdispatcher/tree/main/cmd/example)

```bash
# Clone and run the demo
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
cp cmd/example/env.example .env
# Edit .env with your API keys

# Run with different modes:
# Vendor mode with default vendor (openai)
go run cmd/example/cli.go --vendor

# Vendor mode with specific vendor override
go run cmd/example/cli.go --vendor --vendor-override anthropic

# Local mode with Ollama
go run cmd/example/cli.go --local

# Local mode with custom model
go run cmd/example/cli.go --local --model llama2:13b
```

**Demo Features:**
- ✅ Multi-vendor request routing
- ✅ Cost optimization examples
- ✅ Streaming response demo
- ✅ Fallback scenarios
- ✅ Statistics and metrics
- ✅ Local model integration with Ollama

## Features

### 🚀 Core Features
- **Multi-vendor support**: OpenAI, Anthropic, Google, Azure OpenAI
- **Intelligent routing**: Automatic vendor selection based on model, cost, latency
- **Automatic fallback**: Seamless failover when vendors are unavailable
- **Streaming support**: Real-time responses with vendor-agnostic interface
- **Cost optimization**: Route to cheapest vendor for your use case
- **Advanced retry**: Configurable retry policies with exponential backoff

### 📊 Monitoring & Analytics
- **Unified metrics**: Single dashboard for all vendor performance
- **Cost tracking**: Monitor total and per-request costs
- **Latency monitoring**: Track response times across vendors
- **Success rates**: Monitor vendor reliability and uptime
- **Usage statistics**: Detailed request and token usage

### 🔧 Advanced Configuration
- **Custom routing rules**: Route by model, tokens, temperature, user
- **Cost optimization**: Set budgets and vendor cost preferences
- **Latency optimization**: Configure performance-based routing
- **Rate limiting**: Built-in rate limit handling and backoff
- **Security**: API key management and secure configuration

## Supported Vendors

| Vendor | Models | Features | Cost (per 1K tokens) |
|--------|--------|----------|----------------------|
| **OpenAI** | GPT-4, GPT-3.5-turbo, GPT-4o | Streaming, Rate limiting | $0.002-0.03 |
| **Anthropic** | Claude-3-opus, Claude-3-sonnet | Large context (200K) | $0.003-0.015 |
| **Google** | Gemini-1.5-pro, Gemini-pro | Massive context (1M) | $0.001-0.007 |
| **Azure OpenAI** | GPT-4, GPT-3.5-turbo | Enterprise features | $0.002-0.03 |

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
# Use default vendor (openai)
go run cmd/example/cli.go --vendor

# Use specific vendor override
go run cmd/example/cli.go --vendor --vendor-override anthropic
go run cmd/example/cli.go --vendor --vendor-override openai
```

### Local Mode
Test with local models using Ollama:

```bash
# Use default local model (llama2:7b)
go run cmd/example/cli.go --local

# Use custom model
go run cmd/example/cli.go --local --model llama2:13b
go run cmd/example/cli.go --local --model mistral:7b

# Use custom Ollama server
go run cmd/example/cli.go --local --server http://localhost:11434
```

### Available Options
```bash
go run cmd/example/cli.go --help
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
cp cmd/example/env.example .env
# Edit .env with your API keys
```

### 3. Run the example
```bash
# Default mode (all vendors)
go run cmd/example/cli.go

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
