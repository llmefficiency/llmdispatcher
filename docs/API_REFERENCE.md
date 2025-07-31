# API Reference

## Package Overview

The `llmdispatcher` package provides a unified interface for dispatching LLM requests to multiple vendors with intelligent routing, retry logic, and streaming support.

```go
import "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
```

## Core Types

### Request

Represents a chat completion request to an LLM vendor.

```go
type Request struct {
    Model       string    `json:"model"`           // Model name (e.g., "gpt-3.5-turbo")
    Messages    []Message `json:"messages"`        // Conversation messages
    Temperature float64   `json:"temperature,omitempty"` // Creativity (0.0-2.0)
    MaxTokens   int       `json:"max_tokens,omitempty"` // Maximum tokens to generate
    TopP        float64   `json:"top_p,omitempty"`      // Nucleus sampling (0.0-1.0)
    Stream      bool      `json:"stream,omitempty"`     // Enable streaming
    Stop        []string  `json:"stop,omitempty"`       // Stop sequences
    User        string    `json:"user,omitempty"`       // User identifier
}
```

**Example:**
```go
request := &llmdispatcher.Request{
    Model: "gpt-3.5-turbo",
    Messages: []llmdispatcher.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Hello, how are you?"},
    },
    Temperature: 0.7,
    MaxTokens: 1000,
}
```

### Message

Represents a single message in a conversation.

```go
type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"` // Message content
}
```

**Roles:**
- `"system"`: System instructions or context
- `"user"`: User input
- `"assistant"`: Assistant responses

### Response

Represents a completed LLM response.

```go
type Response struct {
    Content      string    `json:"content"`       // Generated text
    Usage        Usage     `json:"usage"`         // Token usage statistics
    Model        string    `json:"model"`         // Model used
    Vendor       string    `json:"vendor"`        // Vendor that processed request
    FinishReason string    `json:"finish_reason,omitempty"` // Why generation stopped
    CreatedAt    time.Time `json:"created_at"`    // Response timestamp
}
```

### StreamingResponse

Represents a streaming response with channels for real-time communication.

```go
type StreamingResponse struct {
    ContentChan chan string `json:"-"` // Channel for content chunks
    DoneChan    chan bool   `json:"-"` // Channel for completion signal
    ErrorChan   chan error  `json:"-"` // Channel for errors
    Usage       Usage       `json:"usage"`        // Token usage statistics
    Model       string      `json:"model"`        // Model used
    Vendor      string      `json:"vendor"`       // Vendor that processed request
    CreatedAt   time.Time   `json:"created_at"`   // Response timestamp
}
```

**Usage:**
```go
streamingResp, err := dispatcher.SendStreaming(ctx, request)
if err != nil {
    return err
}
defer streamingResp.Close()

for {
    select {
    case content := <-streamingResp.ContentChan:
        fmt.Print(content)
    case done := <-streamingResp.DoneChan:
        if done {
            break
        }
    case err := <-streamingResp.ErrorChan:
        return err
    }
}
```

### Usage

Token usage statistics.

```go
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`     // Input tokens
    CompletionTokens int `json:"completion_tokens"` // Generated tokens
    TotalTokens      int `json:"total_tokens"`      // Total tokens
}
```

## Configuration Types

### Config

Main dispatcher configuration.

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

**Example:**
```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    FallbackVendor: "anthropic",
    Timeout: 30 * time.Second,
    EnableLogging: true,
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries: 3,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
    },
}
```

### VendorConfig

Configuration for individual vendors.

```go
type VendorConfig struct {
    APIKey     string        `json:"api_key"`
    BaseURL    string        `json:"base_url,omitempty"`
    Timeout    time.Duration `json:"timeout,omitempty"`
    Headers    map[string]string `json:"headers,omitempty"`
}
```

### RetryPolicy

Configures retry behavior for failed requests.

```go
type RetryPolicy struct {
    MaxRetries      int           `json:"max_retries"`
    BackoffStrategy BackoffStrategy `json:"backoff_strategy"`
    RetryableErrors []string      `json:"retryable_errors,omitempty"`
    InitialDelay    time.Duration `json:"initial_delay,omitempty"`
    MaxDelay        time.Duration `json:"max_delay,omitempty"`
}
```

**Backoff Strategies:**
- `LinearBackoff`: Fixed delay between retries
- `ExponentialBackoff`: Exponential delay increase
- `JitterBackoff`: Exponential backoff with jitter

### RoutingRule

Defines routing logic for request distribution.

```go
type RoutingRule struct {
    Condition RoutingCondition `json:"condition"`
    Vendor    string          `json:"vendor"`
    Priority  int             `json:"priority"`
    Enabled   bool            `json:"enabled"`
}
```

### RoutingCondition

Conditions for routing rule matching.

```go
type RoutingCondition struct {
    ModelPattern string `json:"model_pattern,omitempty"`
    MaxTokens    int    `json:"max_tokens,omitempty"`
    Temperature  float64 `json:"temperature,omitempty"`
    User         string `json:"user,omitempty"`
}
```

### CostOptimization

Configures cost-based routing.

```go
type CostOptimization struct {
    Enabled     bool              `json:"enabled"`
    MaxCost     float64           `json:"max_cost"`
    PreferCheap bool              `json:"prefer_cheap"`
    VendorCosts map[string]float64 `json:"vendor_costs"`
}
```

### LatencyOptimization

Configures latency-based routing.

```go
type LatencyOptimization struct {
    Enabled       bool                `json:"enabled"`
    MaxLatency    time.Duration       `json:"max_latency"`
    PreferFast    bool                `json:"prefer_fast"`
    LatencyWeights map[string]float64 `json:"latency_weights"`
}
```

## Core Interfaces

### Vendor

Interface that all vendor implementations must satisfy.

```go
type Vendor interface {
    Name() string
    SendRequest(ctx context.Context, req *Request) (*Response, error)
    SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error)
    GetCapabilities() Capabilities
    IsAvailable(ctx context.Context) bool
}
```

### Capabilities

Describes vendor capabilities and limitations.

```go
type Capabilities struct {
    Models            []string `json:"models"`
    MaxTokens         int      `json:"max_tokens"`
    SupportsStreaming bool     `json:"supports_streaming"`
    SupportsVision    bool     `json:"supports_vision"`
    ContextWindow     int      `json:"context_window"`
}
```

## Dispatcher Methods

### Constructor Functions

#### New()
Creates a new dispatcher with default configuration.

```go
dispatcher := llmdispatcher.New()
```

#### NewWithConfig(config *Config)
Creates a new dispatcher with custom configuration.

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    Timeout: 30 * time.Second,
}
dispatcher := llmdispatcher.NewWithConfig(config)
```

### Core Methods

#### Send(ctx context.Context, req *Request) (*Response, error)
Sends a request and returns the response.

```go
response, err := dispatcher.Send(ctx, request)
if err != nil {
    return err
}
fmt.Printf("Response: %s\n", response.Content)
```

#### SendStreaming(ctx context.Context, req *Request) (*StreamingResponse, error)
Sends a streaming request and returns a streaming response.

```go
streamingResp, err := dispatcher.SendStreaming(ctx, request)
if err != nil {
    return err
}
defer streamingResp.Close()

for {
    select {
    case content := <-streamingResp.ContentChan:
        fmt.Print(content)
    case done := <-streamingResp.DoneChan:
        if done {
            break
        }
    case err := <-streamingResp.ErrorChan:
        return err
    }
}
```

### Vendor Management

#### RegisterVendor(vendor Vendor) error
Registers a vendor with the dispatcher.

```go
openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
err := dispatcher.RegisterVendor(openaiVendor)
if err != nil {
    return err
}
```

#### GetVendors() []string
Returns a list of registered vendor names.

```go
vendors := dispatcher.GetVendors()
fmt.Printf("Registered vendors: %v\n", vendors)
```

#### GetVendor(name string) (Vendor, bool)
Retrieves a specific vendor by name.

```go
vendor, exists := dispatcher.GetVendor("openai")
if !exists {
    return fmt.Errorf("vendor not found")
}
```

### Statistics and Monitoring

#### GetStats() *Stats
Returns comprehensive dispatcher statistics.

```go
stats := dispatcher.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n", stats.SuccessRate())
fmt.Printf("Average latency: %v\n", stats.AverageLatency)
fmt.Printf("Total cost: $%.4f\n", stats.TotalCost)
```

## Vendor Factory Functions

### OpenAI

#### NewOpenAIVendor(config *VendorConfig) Vendor
Creates an OpenAI vendor instance.

```go
openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    BaseURL: "https://api.openai.com/v1",
    Timeout: 30 * time.Second,
})
```

**Supported Models:**
- `gpt-3.5-turbo`
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`

### Anthropic

#### NewAnthropicVendor(config *VendorConfig) Vendor
Creates an Anthropic vendor instance.

```go
anthropicVendor := llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    Timeout: 30 * time.Second,
})
```

**Supported Models:**
- `claude-3-opus`
- `claude-3-sonnet`
- `claude-3-haiku`
- `claude-3-5-sonnet`
- `claude-3-5-haiku`

### Google

#### NewGoogleVendor(config *VendorConfig) Vendor
Creates a Google vendor instance.

```go
googleVendor := llmdispatcher.NewGoogleVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("GOOGLE_API_KEY"),
    Timeout: 30 * time.Second,
})
```

**Supported Models:**
- `gemini-1.5-pro`
- `gemini-1.5-flash`
- `gemini-pro`
- `gemini-pro-vision`

### Azure OpenAI

#### NewAzureOpenAIVendor(config *VendorConfig) Vendor
Creates an Azure OpenAI vendor instance.

```go
azureVendor := llmdispatcher.NewAzureOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("AZURE_OPENAI_API_KEY"),
    BaseURL: os.Getenv("AZURE_OPENAI_ENDPOINT"),
    Timeout: 30 * time.Second,
})
```

**Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-35-turbo`
- `gpt-35-turbo-16k`

## Error Types

### RateLimitError
Error returned when rate limits are exceeded.

```go
type RateLimitError struct {
    Vendor string
    Err    error
}
```

### AuthenticationError
Error returned for authentication failures.

```go
type AuthenticationError struct {
    Vendor string
    Err    error
}
```

### TimeoutError
Error returned when requests timeout.

```go
type TimeoutError struct {
    Vendor string
    Err    error
}
```

### VendorUnavailableError
Error returned when a vendor is unavailable.

```go
type VendorUnavailableError struct {
    Vendor string
    Err    error
}
```

## Statistics Types

### Stats
Comprehensive dispatcher statistics.

```go
type Stats struct {
    TotalRequests    int64                    `json:"total_requests"`
    SuccessfulRequests int64                  `json:"successful_requests"`
    FailedRequests   int64                    `json:"failed_requests"`
    AverageLatency   time.Duration           `json:"average_latency"`
    TotalCost        float64                  `json:"total_cost"`
    AverageCost      float64                  `json:"average_cost"`
    VendorStats      map[string]*VendorStats `json:"vendor_stats"`
}
```

### VendorStats
Statistics for individual vendors.

```go
type VendorStats struct {
    Requests   int64         `json:"requests"`
    Successes  int64         `json:"successes"`
    Failures   int64         `json:"failures"`
    TotalCost  float64       `json:"total_cost"`
    AvgLatency time.Duration `json:"avg_latency"`
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
        DefaultVendor: "openai",
        FallbackVendor: "anthropic",
        Timeout: 30 * time.Second,
        EnableLogging: true,
        RetryPolicy: &llmdispatcher.RetryPolicy{
            MaxRetries: 3,
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
        MaxTokens: 1000,
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
    fmt.Printf("Success rate: %.2f%%\n", stats.SuccessRate())
}
``` 