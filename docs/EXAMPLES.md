# Examples

This document provides comprehensive examples of how to use the LLM Dispatcher.

## CLI Usage

The example application (`cmd/example/cli.go`) provides a command-line interface for testing different configurations:

### Basic CLI Commands

```bash
# Run with all available vendors (default mode)
go run cmd/example/cli.go

# Vendor mode with default vendor (openai)
go run cmd/example/cli.go --vendor

# Vendor mode with specific vendor override
go run cmd/example/cli.go --vendor --vendor-override anthropic
go run cmd/example/cli.go --vendor --vendor-override openai

# Local mode with Ollama
go run cmd/example/cli.go --local

# Local mode with custom model
go run cmd/example/cli.go --local --model llama2:13b
go run cmd/example/cli.go --local --model mistral:7b

# Local mode with custom server
go run cmd/example/cli.go --local --server http://localhost:11434
```

### CLI Options

| Option | Description | Default |
|--------|-------------|---------|
| `--local` | Run in local mode with Ollama | false |
| `--vendor` | Run in vendor mode with cloud providers | false |
| `--vendor-override` | Specify vendor (anthropic, openai) | "" |
| `--model` | Model to use in local mode | "llama2:7b" |
| `--server` | Ollama server URL | "http://localhost:11434" |

### Environment Setup for CLI

```bash
# Copy environment template
cp cmd/example/env.example .env

# Edit .env with your API keys
export OPENAI_API_KEY="sk-your-openai-key"
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
export GOOGLE_API_KEY="your-google-api-key"
export AZURE_OPENAI_API_KEY="your-azure-key"
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
```

## Basic Usage

### Simple Configuration

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
    // Create dispatcher with basic configuration
    config := &llmdispatcher.Config{
        DefaultVendor: "openai",
        Timeout:       30 * time.Second,
        EnableLogging: true,
    }

    dispatcher := llmdispatcher.NewWithConfig(config)

    // Register OpenAI vendor
    openaiConfig := &llmdispatcher.VendorConfig{
        APIKey:  "your-openai-api-key",
        Timeout: 30 * time.Second,
    }

    openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
    if err := dispatcher.RegisterVendor(openaiVendor); err != nil {
        log.Fatalf("Failed to register OpenAI vendor: %v", err)
    }

    // Send a request
    ctx := context.Background()
    req := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
        Temperature: 0.7,
        MaxTokens:   100,
    }

    resp, err := dispatcher.Send(ctx, req)
    if err != nil {
        log.Fatalf("Failed to send request: %v", err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
}
```

## Advanced Configuration

### Cost Optimization

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    Timeout:       30 * time.Second,
    EnableLogging: true,
    CostOptimization: &llmdispatcher.CostOptimization{
        Enabled:     true,
        MaxCost:     0.10, // $0.10 per request
        PreferCheap: true,
        VendorCosts: map[string]float64{
            "openai":   0.0020, // $0.002 per 1K tokens
            "anthropic": 0.0015, // $0.0015 per 1K tokens
            "google":   0.0010, // $0.001 per 1K tokens
            "local":    0.0001, // $0.0001 per 1K tokens (cheapest)
        },
    },
}
```

### Routing Rules

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    RoutingRules: []llmdispatcher.RoutingRule{
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "gpt-4",
                MaxTokens:    1000,
            },
            Vendor:   "openai",
            Priority: 1,
            Enabled:  true,
        },
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "claude",
                CostThreshold: 0.05,
            },
            Vendor:   "anthropic",
            Priority: 2,
            Enabled:  true,
        },
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "llama",
                CostThreshold: 0.01,
            },
            Vendor:   "local",
            Priority: 3,
            Enabled:  true,
        },
    },
}
```

## Local Model Integration

The local vendor allows you to use local models like Ollama or llama.cpp as the cheapest option in your dispatcher.

### Ollama Setup (Recommended)

```go
// Configure local vendor with Ollama
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy", // Not used for local models
    Headers: map[string]string{
        "server_url": "http://localhost:11434", // Ollama default
        "model_path": "llama2:7b",             // Model name in Ollama
    },
    Timeout: 60 * time.Second,
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
if err := dispatcher.RegisterVendor(localVendor); err != nil {
    log.Fatalf("Failed to register local vendor: %v", err)
}
```

### llama.cpp Setup (Maximum Performance)

```go
// Configure local vendor with llama.cpp
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "executable": "/usr/local/bin/llama",     // llama.cpp executable
        "model_path": "/path/to/llama2-7b.gguf", // Model file path
    },
    Timeout: 120 * time.Second, // Longer timeout for direct execution
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
if err := dispatcher.RegisterVendor(localVendor); err != nil {
    log.Fatalf("Failed to register local vendor: %v", err)
}
```

### Custom HTTP Server Setup

```go
// Configure local vendor with custom HTTP server
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:8080", // Custom server
        "model_path": "mistral:7b",            // Model name
    },
    Timeout: 60 * time.Second,
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
if err := dispatcher.RegisterVendor(localVendor); err != nil {
    log.Fatalf("Failed to register local vendor: %v", err)
}
```

### GPU-Optimized Setup

```go
// Configure local vendor with GPU optimization
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:13b",
        "max_gpu_layers": "32",
        "max_memory_mb": "8192",
    },
    Timeout: 90 * time.Second,
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
if err := dispatcher.RegisterVendor(localVendor); err != nil {
    log.Fatalf("Failed to register local vendor: %v", err)
}
```

### Using Local Models with Cost Optimization

```go
// Create dispatcher with local model as cheapest option
config := &llmdispatcher.Config{
    DefaultVendor: "local",
    Timeout:       60 * time.Second,
    EnableLogging: true,
    CostOptimization: &llmdispatcher.CostOptimization{
        Enabled:     true,
        PreferCheap: true,
        VendorCosts: map[string]float64{
            "local":    0.0001, // $0.0001 per 1K tokens (cheapest)
            "openai":   0.0020, // $0.002 per 1K tokens
            "anthropic": 0.0015, // $0.0015 per 1K tokens
            "google":   0.0010, // $0.001 per 1K tokens
        },
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)

// Register local vendor
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:7b",
    },
    Timeout: 60 * time.Second,
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
if err := dispatcher.RegisterVendor(localVendor); err != nil {
    log.Fatalf("Failed to register local vendor: %v", err)
}

// Send request - will automatically use local model due to cost optimization
ctx := context.Background()
req := &llmdispatcher.Request{
    Model: "llama2:7b",
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "What is the capital of France?"},
    },
    Temperature: 0.7,
    MaxTokens:   100,
}

resp, err := dispatcher.Send(ctx, req)
if err != nil {
    log.Fatalf("Failed to send request: %v", err)
}

fmt.Printf("Response from %s: %s\n", resp.Vendor, resp.Content)
```

## Streaming

### Basic Streaming

```go
req := &llmdispatcher.Request{
    Model: "gpt-3.5-turbo",
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "Write a story about a robot."},
    },
    Temperature: 0.8,
    MaxTokens:   500,
}

streamResp, err := dispatcher.SendStreaming(ctx, req)
if err != nil {
    log.Fatalf("Failed to start streaming: %v", err)
}

for {
    select {
    case content := <-streamResp.ContentChan:
        fmt.Print(content)
    case <-streamResp.DoneChan:
        fmt.Println("\n[Stream completed]")
        return
    case err := <-streamResp.ErrorChan:
        log.Printf("Streaming error: %v", err)
        return
    }
}
```

### Streaming with Local Models

```go
req := &llmdispatcher.Request{
    Model: "llama2:7b",
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "Write a poem about AI."},
    },
    Temperature: 0.8,
    MaxTokens:   200,
}

streamResp, err := dispatcher.SendStreaming(ctx, req)
if err != nil {
    log.Fatalf("Failed to start streaming: %v", err)
}

fmt.Print("Streaming response: ")
for {
    select {
    case content := <-streamResp.ContentChan:
        fmt.Print(content)
    case <-streamResp.DoneChan:
        fmt.Println("\n[Stream completed]")
        return
    case err := <-streamResp.ErrorChan:
        log.Printf("Streaming error: %v", err)
        return
    }
}
```

## Error Handling

### Retry Policy

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
        RetryableErrors: []string{
            "rate limit exceeded",
            "timeout",
            "server error",
        },
    },
}
```

### Fallback Strategy

```go
config := &llmdispatcher.Config{
    DefaultVendor:  "openai",
    FallbackVendor: "anthropic",
    RoutingRules: []llmdispatcher.RoutingRule{
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "gpt-4",
            },
            Vendor:   "openai",
            Priority: 1,
            Enabled:  true,
        },
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "claude",
            },
            Vendor:   "anthropic",
            Priority: 2,
            Enabled:  true,
        },
    },
}
```

## Metrics and Monitoring

### Basic Metrics

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    EnableMetrics: true,
    EnableLogging: true,
}

dispatcher := llmdispatcher.NewWithConfig(config)

// After sending requests, get statistics
stats := dispatcher.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Successful requests: %d\n", stats.SuccessfulRequests)
fmt.Printf("Failed requests: %d\n", stats.FailedRequests)
fmt.Printf("Average latency: %v\n", stats.AverageLatency)
fmt.Printf("Total cost: $%.4f\n", stats.TotalCost)
```

### Vendor-Specific Metrics

```go
stats := dispatcher.GetStats()
for vendor, vendorStats := range stats.VendorStats {
    fmt.Printf("Vendor %s:\n", vendor)
    fmt.Printf("  Requests: %d\n", vendorStats.Requests)
    fmt.Printf("  Successes: %d\n", vendorStats.Successes)
    fmt.Printf("  Failures: %d\n", vendorStats.Failures)
    fmt.Printf("  Average latency: %v\n", vendorStats.AverageLatency)
    fmt.Printf("  Total cost: $%.4f\n", vendorStats.TotalCost)
}
```

## Complete Example

Here's a complete example that demonstrates all the features:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
    // Get API keys from environment
    openaiAPIKey := os.Getenv("OPENAI_API_KEY")
    anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")

    // Create dispatcher with advanced configuration
    config := &llmdispatcher.Config{
        DefaultVendor:  "local",
        FallbackVendor: "openai",
        Timeout:        60 * time.Second,
        EnableLogging:  true,
        EnableMetrics:  true,
        CostOptimization: &llmdispatcher.CostOptimization{
            Enabled:     true,
            PreferCheap: true,
            VendorCosts: map[string]float64{
                "local":    0.0001,
                "openai":   0.0020,
                "anthropic": 0.0015,
            },
        },
        RetryPolicy: &llmdispatcher.RetryPolicy{
            MaxRetries:      3,
            BackoffStrategy: llmdispatcher.ExponentialBackoff,
            RetryableErrors: []string{"rate limit exceeded", "timeout"},
        },
        RoutingRules: []llmdispatcher.RoutingRule{
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "gpt-4",
                },
                Vendor:   "openai",
                Priority: 1,
                Enabled:  true,
            },
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "claude",
                },
                Vendor:   "anthropic",
                Priority: 2,
                Enabled:  true,
            },
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "llama",
                },
                Vendor:   "local",
                Priority: 3,
                Enabled:  true,
            },
        },
    }

    dispatcher := llmdispatcher.NewWithConfig(config)

    // Register local vendor (cheapest option)
    localConfig := &llmdispatcher.VendorConfig{
        APIKey: "dummy",
        Headers: map[string]string{
            "server_url": "http://localhost:11434",
            "model_path": "llama2:7b",
        },
        Timeout: 60 * time.Second,
    }

    localVendor := llmdispatcher.NewLocalVendor(localConfig)
    if err := dispatcher.RegisterVendor(localVendor); err != nil {
        log.Printf("Failed to register local vendor: %v", err)
    }

    // Register OpenAI vendor
    if openaiAPIKey != "" {
        openaiConfig := &llmdispatcher.VendorConfig{
            APIKey:  openaiAPIKey,
            Timeout: 30 * time.Second,
        }

        openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
        if err := dispatcher.RegisterVendor(openaiVendor); err != nil {
            log.Printf("Failed to register OpenAI vendor: %v", err)
        }
    }

    // Register Anthropic vendor
    if anthropicAPIKey != "" {
        anthropicConfig := &llmdispatcher.VendorConfig{
            APIKey:  anthropicAPIKey,
            Timeout: 30 * time.Second,
        }

        anthropicVendor := llmdispatcher.NewAnthropicVendor(anthropicConfig)
        if err := dispatcher.RegisterVendor(anthropicVendor); err != nil {
            log.Printf("Failed to register Anthropic vendor: %v", err)
        }
    }

    // Send requests
    ctx := context.Background()
    requests := []llmdispatcher.Request{
        {
            Model: "llama2:7b",
            Messages: []llmdispatcher.Message{
                {Role: "user", Content: "What is the capital of France?"},
            },
            Temperature: 0.7,
            MaxTokens:   100,
        },
        {
            Model: "gpt-4",
            Messages: []llmdispatcher.Message{
                {Role: "user", Content: "Explain quantum computing."},
            },
            Temperature: 0.8,
            MaxTokens:   200,
        },
    }

    for i, req := range requests {
        fmt.Printf("\nRequest %d:\n", i+1)
        resp, err := dispatcher.Send(ctx, &req)
        if err != nil {
            log.Printf("Request %d failed: %v", i+1, err)
            continue
        }

        fmt.Printf("Vendor: %s\n", resp.Vendor)
        fmt.Printf("Model: %s\n", resp.Model)
        fmt.Printf("Response: %s\n", resp.Content)
        fmt.Printf("Usage: %+v\n", resp.Usage)
    }

    // Print final statistics
    stats := dispatcher.GetStats()
    fmt.Printf("\nFinal Statistics:\n")
    fmt.Printf("Total requests: %d\n", stats.TotalRequests)
    fmt.Printf("Successful requests: %d\n", stats.SuccessfulRequests)
    fmt.Printf("Failed requests: %d\n", stats.FailedRequests)
    fmt.Printf("Average latency: %v\n", stats.AverageLatency)
    fmt.Printf("Total cost: $%.4f\n", stats.TotalCost)
}
```

This example demonstrates:
- Local model integration as the cheapest option
- Cost optimization with multiple vendors
- Routing rules for different models
- Retry policies and error handling
- Metrics and monitoring
- Fallback strategies 