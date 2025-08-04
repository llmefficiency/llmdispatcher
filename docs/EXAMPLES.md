# Examples

This document provides comprehensive examples of how to use the LLM Dispatcher.

## Implementation Status

**✅ Implemented Features:**
- Basic vendor selection (default/fallback)
- Simple routing rules (model pattern, max tokens, temperature)
- Retry logic with configurable backoff strategies
- All vendor integrations (OpenAI, Anthropic, Google, Azure, Local)
- Streaming support
- Basic statistics and metrics
- Web service with REST API

**❌ Not Yet Implemented:**
- Cost optimization routing
- Latency optimization routing
- Advanced routing conditions (user ID, request type, content length)
- Rate limiting
- Advanced cost tracking

## CLI Usage

The CLI application (`apps/cli/cli.go`) provides a command-line interface for testing different configurations:

### Basic CLI Commands

```bash
# Run with all available vendors (default mode)
go run apps/cli/cli.go

# Vendor mode with default vendor (openai)
go run apps/cli/cli.go --vendor

# Vendor mode with specific vendor override
go run apps/cli/cli.go --vendor --vendor-override anthropic
go run apps/cli/cli.go --vendor --vendor-override openai

# Local mode with Ollama
go run apps/cli/cli.go --local

# Local mode with custom model
go run apps/cli/cli.go --local --model llama2:13b
go run apps/cli/cli.go --local --model mistral:7b

# Local mode with custom server
go run apps/cli/cli.go --local --server http://localhost:11434
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
cp env.example .env

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
        MaxTokens:   1000,
    }

    resp, err := dispatcher.Send(ctx, req)
    if err != nil {
        log.Fatalf("Failed to send request: %v", err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Vendor: %s\n", resp.Vendor)
    fmt.Printf("Usage: %+v\n", resp.Usage)
}
```

### Multi-Vendor Setup

```go
// Create dispatcher with fallback configuration
config := &llmdispatcher.Config{
    DefaultVendor:  "openai",
    FallbackVendor: "anthropic",
    Timeout:        30 * time.Second,
    EnableLogging:  true,
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
        RetryableErrors: []string{"rate_limit", "timeout"},
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)

// Register multiple vendors
openaiConfig := &llmdispatcher.VendorConfig{
    APIKey:  "your-openai-api-key",
    Timeout: 30 * time.Second,
}
openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
dispatcher.RegisterVendor(openaiVendor)

anthropicConfig := &llmdispatcher.VendorConfig{
    APIKey:  "your-anthropic-api-key",
    Timeout: 30 * time.Second,
}
anthropicVendor := llmdispatcher.NewAnthropicVendor(anthropicConfig)
dispatcher.RegisterVendor(anthropicVendor)

// Send request - will use OpenAI by default, fallback to Anthropic if needed
resp, err := dispatcher.Send(ctx, req)
if err != nil {
    log.Fatalf("Failed to send request: %v", err)
}

fmt.Printf("Response from %s: %s\n", resp.Vendor, resp.Content)
```

## Routing Rules

### Basic Routing Rules

**⚠️ Note: Only basic routing conditions are currently implemented**

```go
// Create dispatcher with routing rules
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    Timeout:       30 * time.Second,
    EnableLogging: true,
    RoutingRules: []llmdispatcher.RoutingRule{
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "gpt-4",     // ✅ Implemented: Exact model matching
                MaxTokens:    2000,         // ✅ Implemented: Token limit checking
                Temperature:  0.8,          // ✅ Implemented: Temperature threshold
            },
            Vendor:   "openai",
            Priority: 1,
            Enabled:  true,
        },
        {
            Condition: llmdispatcher.RoutingCondition{
                ModelPattern: "claude-3",   // Route Claude models to Anthropic
                MaxTokens:    1000,
            },
            Vendor:   "anthropic",
            Priority: 2,
            Enabled:  true,
        },
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)

// Register vendors
openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
dispatcher.RegisterVendor(openaiVendor)

anthropicVendor := llmdispatcher.NewAnthropicVendor(anthropicConfig)
dispatcher.RegisterVendor(anthropicVendor)

// Send requests - will be routed based on rules
req1 := &llmdispatcher.Request{
    Model: "gpt-4",
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "Complex analysis request"},
    },
    MaxTokens:   1500,
    Temperature: 0.7,
}

req2 := &llmdispatcher.Request{
    Model: "claude-3-sonnet",
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "Simple question"},
    },
    MaxTokens: 500,
}

resp1, _ := dispatcher.Send(ctx, req1) // Will use OpenAI
resp2, _ := dispatcher.Send(ctx, req2) // Will use Anthropic

fmt.Printf("Request 1 vendor: %s\n", resp1.Vendor)
fmt.Printf("Request 2 vendor: %s\n", resp2.Vendor)
```

### Advanced Routing Rules (Not Yet Implemented)

**❌ The following routing conditions are defined but not yet implemented:**

```go
// These conditions exist in the config but are not used in routing logic
Condition: llmdispatcher.RoutingCondition{
    CostThreshold:    0.10,              // ❌ Not implemented
    LatencyThreshold: 30 * time.Second,  // ❌ Not implemented
    UserID:           "premium_user",     // ❌ Not implemented
    RequestType:      "analysis",         // ❌ Not implemented
    ContentLength:    1000,               // ❌ Not implemented
},
```

## Local Vendor Integration

### Ollama Integration

```go
// Configure local vendor with Ollama
localConfig := &llmdispatcher.VendorConfig{
    APIKey: "dummy", // Not used for local vendor
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:7b",
    },
    Timeout: 60 * time.Second,
}

localVendor := llmdispatcher.NewLocalVendor(localConfig)
dispatcher.RegisterVendor(localVendor)

// Send request to local model
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

fmt.Printf("Local response: %s\n", resp.Content)
```

### Direct llama.cpp Integration

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
dispatcher.RegisterVendor(localVendor)
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
dispatcher.RegisterVendor(localVendor)
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
dispatcher.RegisterVendor(localVendor)
```

## Cost Optimization (Not Yet Implemented)

**⚠️ Cost optimization is defined in configuration but not yet implemented in routing logic**

```go
// This configuration exists but is not used in vendor selection
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

// Note: This configuration is currently ignored by the routing logic
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
        {Role: "user", Content: "Explain quantum computing in simple terms."},
    },
    Temperature: 0.7,
    MaxTokens:   300,
}

streamResp, err := dispatcher.SendStreaming(ctx, req)
if err != nil {
    log.Fatalf("Failed to start streaming: %v", err)
}

fmt.Print("Response: ")
for {
    select {
    case content := <-streamResp.ContentChan:
        fmt.Print(content)
    case <-streamResp.DoneChan:
        fmt.Println("\n[Local streaming completed]")
        return
    case err := <-streamResp.ErrorChan:
        log.Printf("Local streaming error: %v", err)
        return
    }
}
```

## Error Handling

### Comprehensive Error Handling

```go
resp, err := dispatcher.Send(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, llmdispatcher.ErrNoVendorsRegistered):
        log.Fatal("No vendors registered with dispatcher")
    case errors.Is(err, llmdispatcher.ErrVendorUnavailable):
        log.Printf("All vendors are currently unavailable")
    case errors.Is(err, llmdispatcher.ErrInvalidRequest):
        log.Printf("Invalid request: %v", err)
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

### Retry Logic

```go
// Configure retry policy
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
        RetryableErrors: []string{"rate_limit", "timeout", "network_error"},
    },
}

dispatcher := llmdispatcher.NewWithConfig(config)

// Send request with automatic retries
resp, err := dispatcher.Send(ctx, req)
if err != nil {
    log.Printf("Request failed after retries: %v", err)
    return
}
```

## Statistics and Monitoring

### Basic Statistics

```go
// Get dispatcher statistics
stats := dispatcher.GetStats()

fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Successful requests: %d\n", stats.SuccessfulRequests)
fmt.Printf("Failed requests: %d\n", stats.FailedRequests)
fmt.Printf("Success rate: %.2f%%\n", 
    float64(stats.SuccessfulRequests)/float64(stats.TotalRequests)*100)

// Vendor-specific statistics
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d successes, %d failures\n",
        vendorName, vendorStats.Requests, vendorStats.Successes, vendorStats.Failures)
}
```

### Vendor Availability

```go
// Check vendor availability
vendors := dispatcher.GetVendors()
for _, vendorName := range vendors {
    vendor, exists := dispatcher.GetVendor(vendorName)
    if !exists {
        continue
    }
    
    if vendor.IsAvailable(ctx) {
        fmt.Printf("%s: Available\n", vendorName)
    } else {
        fmt.Printf("%s: Unavailable\n", vendorName)
    }
}
```

## Web Service Usage

### Start Web Service

```bash
# Start the web service
make webservice

# Or run directly
go run apps/server/main.go
```

### API Examples

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Get statistics
curl http://localhost:8080/api/v1/stats

# Get vendors list
curl http://localhost:8080/api/v1/vendors

# Send chat completion
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "temperature": 0.7,
    "max_tokens": 100
  }'

# Test specific vendor
curl -X POST http://localhost:8080/api/v1/test/vendor \
  -H "Content-Type: application/json" \
  -d '{
    "vendor": "openai",
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Test message"}
    ]
  }'
```

## Environment Configuration

### Environment Variables

```bash
# Copy template
cp env.example .env

# Edit with your API keys
OPENAI_API_KEY=sk-your-openai-api-key-here
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key-here
GOOGLE_API_KEY=your-google-api-key-here
AZURE_OPENAI_API_KEY=your-azure-openai-api-key-here
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/

# Optional: Rate limiting (not yet implemented)
OPENAI_RATE_LIMIT_REQUESTS_PER_MINUTE=60
ANTHROPIC_RATE_LIMIT_REQUESTS_PER_MINUTE=50
```

### Programmatic Configuration

```go
// Load environment variables
if err := godotenv.Load(); err != nil {
    log.Printf("No .env file found")
}

// Create vendor configs from environment
openaiConfig := &llmdispatcher.VendorConfig{
    APIKey:  os.Getenv("OPENAI_API_KEY"),
    BaseURL: os.Getenv("OPENAI_BASE_URL"),
    Timeout: 30 * time.Second,
}

anthropicConfig := &llmdispatcher.VendorConfig{
    APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
    BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
    Timeout: 30 * time.Second,
}
```

## Complete Example

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
    // Create dispatcher with configuration
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
    
    dispatcher := llmdispatcher.NewWithConfig(config)

    // Register vendors
    openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    dispatcher.RegisterVendor(openaiVendor)

    anthropicVendor := llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    })
    dispatcher.RegisterVendor(anthropicVendor)

    // Create request
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
        Temperature: 0.7,
        MaxTokens:   1000,
    }

    // Send request
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    response, err := dispatcher.Send(ctx, request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
    fmt.Printf("Model: %s\n", response.Model)
    fmt.Printf("Vendor: %s\n", response.Vendor)
    fmt.Printf("Usage: %+v\n", response.Usage)

    // Get statistics
    stats := dispatcher.GetStats()
    fmt.Printf("Total requests: %d\n", stats.TotalRequests)
    fmt.Printf("Success rate: %.2f%%\n", 
        float64(stats.SuccessfulRequests)/float64(stats.TotalRequests)*100)
}
``` 