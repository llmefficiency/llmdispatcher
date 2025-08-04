# API Reference

## Package Overview

The `llmdispatcher` package provides a unified interface for dispatching LLM requests to multiple vendors with intelligent routing, retry logic, and streaming support.

```go
import "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
```

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
    case <-streamingResp.DoneChan:
        fmt.Println("\n[Stream completed]")
        return
    case err := <-streamingResp.ErrorChan:
        log.Printf("Streaming error: %v", err)
        return
    }
}
```

### Usage

Represents token usage statistics.

```go
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`     // Input tokens
    CompletionTokens int `json:"completion_tokens"` // Output tokens
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
    // Note: CostOptimization and LatencyOptimization are defined but not yet implemented
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
}
```

**Backoff Strategies:**
- `LinearBackoff`: Fixed delay between retries
- `ExponentialBackoff`: Exponential delay increase
- `FixedBackoff`: Fixed delay between retries

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

Conditions for routing rule matching. **Currently implemented conditions:**

```go
type RoutingCondition struct {
    ModelPattern string `json:"model_pattern,omitempty"` // ✅ Implemented: Exact model name matching
    MaxTokens    int    `json:"max_tokens,omitempty"`   // ✅ Implemented: Token limit checking
    Temperature  float64 `json:"temperature,omitempty"`  // ✅ Implemented: Temperature threshold
    // ❌ Not yet implemented:
    CostThreshold    float64       `json:"cost_threshold,omitempty"`
    LatencyThreshold time.Duration `json:"latency_threshold,omitempty"`
    UserID           string        `json:"user_id,omitempty"`
    RequestType      string        `json:"request_type,omitempty"`
    ContentLength    int           `json:"content_length,omitempty"`
}
```

### CostOptimization

**⚠️ Defined but not yet implemented**

```go
type CostOptimization struct {
    Enabled     bool              `json:"enabled"`
    MaxCost     float64           `json:"max_cost"`
    PreferCheap bool              `json:"prefer_cheap"`
    VendorCosts map[string]float64 `json:"vendor_costs"`
}
```

### LatencyOptimization

**⚠️ Defined but not yet implemented**

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

#### NewWithConfig(config)
Creates a new dispatcher with custom configuration.

```go
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    Timeout: 30 * time.Second,
}
dispatcher := llmdispatcher.NewWithConfig(config)
```

### Core Methods

#### RegisterVendor(vendor)
Registers a vendor with the dispatcher.

```go
openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
err := dispatcher.RegisterVendor(openaiVendor)
```

#### Send(ctx, request)
Sends a request to the appropriate vendor.

```go
response, err := dispatcher.Send(ctx, request)
if err != nil {
    return err
}
fmt.Printf("Response: %s\n", response.Content)
```

#### SendStreaming(ctx, request)
Sends a streaming request to the appropriate vendor.

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
    case <-streamingResp.DoneChan:
        return
    case err := <-streamingResp.ErrorChan:
        return err
    }
}
```

#### SendToVendor(ctx, vendorName, request)
Sends a request to a specific vendor.

```go
response, err := dispatcher.SendToVendor(ctx, "openai", request)
```

#### GetStats()
Returns dispatcher statistics.

```go
stats := dispatcher.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n", float64(stats.SuccessfulRequests)/float64(stats.TotalRequests)*100)
```

#### GetVendors()
Returns a list of registered vendor names.

```go
vendors := dispatcher.GetVendors()
fmt.Printf("Available vendors: %v\n", vendors)
```

## Vendor Implementations

### OpenAI Vendor

```go
openaiConfig := &llmdispatcher.VendorConfig{
    APIKey:  "sk-your-openai-api-key",
    BaseURL: "https://api.openai.com/v1",
    Timeout: 30 * time.Second,
}

openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
dispatcher.RegisterVendor(openaiVendor)
```

### Anthropic Vendor

```go
anthropicConfig := &llmdispatcher.VendorConfig{
    APIKey:  "sk-ant-your-anthropic-api-key",
    BaseURL: "https://api.anthropic.com",
    Timeout: 30 * time.Second,
}

anthropicVendor := llmdispatcher.NewAnthropicVendor(anthropicConfig)
dispatcher.RegisterVendor(anthropicVendor)
```

### Google Vendor

```go
googleConfig := &llmdispatcher.VendorConfig{
    APIKey:  "your-google-api-key",
    BaseURL: "https://generativelanguage.googleapis.com",
    Timeout: 30 * time.Second,
}

googleVendor := llmdispatcher.NewGoogleVendor(googleConfig)
dispatcher.RegisterVendor(googleVendor)
```

### Azure OpenAI Vendor

```go
azureConfig := &llmdispatcher.VendorConfig{
    APIKey:  "your-azure-api-key",
    BaseURL: "https://your-resource.openai.azure.com/",
    Timeout: 30 * time.Second,
}

azureVendor := llmdispatcher.NewAzureOpenAIVendor(azureConfig)
dispatcher.RegisterVendor(azureVendor)
```

### Local Vendor (Ollama)

```go
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
```

## Error Handling

The dispatcher provides comprehensive error handling:

```go
response, err := dispatcher.Send(ctx, request)
if err != nil {
    switch {
    case errors.Is(err, llmdispatcher.ErrNoVendorsRegistered):
        log.Fatal("No vendors registered")
    case errors.Is(err, llmdispatcher.ErrVendorUnavailable):
        log.Printf("All vendors unavailable")
    case errors.Is(err, llmdispatcher.ErrInvalidRequest):
        log.Printf("Invalid request: %v", err)
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Web Service API

The dispatcher includes a web service with REST API endpoints:

### Health Check
```bash
GET /api/v1/health
```

### Chat Completions
```bash
POST /api/v1/chat/completions
```

### Streaming Chat
```bash
POST /api/v1/chat/completions/stream
```

### Vendor Testing
```bash
POST /api/v1/test/vendor
```

### Statistics
```bash
GET /api/v1/stats
```

### Vendors List
```bash
GET /api/v1/vendors
```

## Environment Variables

Configure vendors using environment variables:

```bash
# OpenAI
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_TIMEOUT=30s

# Anthropic
ANTHROPIC_API_KEY=sk-ant-your-anthropic-api-key
ANTHROPIC_BASE_URL=https://api.anthropic.com
ANTHROPIC_TIMEOUT=30s

# Google
GOOGLE_API_KEY=your-google-api-key
GOOGLE_BASE_URL=https://generativelanguage.googleapis.com
GOOGLE_TIMEOUT=30s

# Azure OpenAI
AZURE_OPENAI_API_KEY=your-azure-api-key
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_TIMEOUT=30s
``` 