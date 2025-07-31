# Examples Guide

This guide provides comprehensive examples for using the LLM Dispatcher library in various scenarios.

## Quick Start Examples

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

    // Register OpenAI vendor
    openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    dispatcher.RegisterVendor(openaiVendor)

    // Create request
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
    }

    // Send request
    ctx := context.Background()
    response, err := dispatcher.Send(ctx, request)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Content)
}
```

### Multiple Vendors

```go
func setupMultipleVendors() *llmdispatcher.Dispatcher {
    dispatcher := llmdispatcher.New()

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

    return dispatcher
}
```

## Configuration Examples

### Advanced Configuration

```go
func createAdvancedDispatcher() *llmdispatcher.Dispatcher {
    config := &llmdispatcher.Config{
        DefaultVendor: "openai",
        FallbackVendor: "anthropic",
        Timeout: 30 * time.Second,
        EnableLogging: true,
        EnableMetrics: true,
        RetryPolicy: &llmdispatcher.RetryPolicy{
            MaxRetries: 3,
            BackoffStrategy: llmdispatcher.ExponentialBackoff,
            RetryableErrors: []string{
                "rate limit exceeded",
                "timeout",
                "connection error",
            },
            InitialDelay: 1 * time.Second,
            MaxDelay: 60 * time.Second,
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
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "claude",
                },
                Vendor: "anthropic",
                Priority: 2,
                Enabled: true,
            },
        },
        CostOptimization: &llmdispatcher.CostOptimization{
            Enabled: true,
            MaxCost: 0.10,
            PreferCheap: true,
            VendorCosts: map[string]float64{
                "openai":   0.002,
                "anthropic": 0.003,
                "google":    0.001,
            },
        },
        LatencyOptimization: &llmdispatcher.LatencyOptimization{
            Enabled: true,
            MaxLatency: 30 * time.Second,
            PreferFast: true,
            LatencyWeights: map[string]float64{
                "openai":   1.0,
                "anthropic": 1.2,
                "google":    0.8,
            },
        },
    }

    return llmdispatcher.NewWithConfig(config)
}
```

### Environment-Based Configuration

```go
func loadConfigFromEnv() *llmdispatcher.Config {
    config := &llmdispatcher.Config{
        DefaultVendor: getEnvOrDefault("DEFAULT_VENDOR", "openai"),
        FallbackVendor: getEnvOrDefault("FALLBACK_VENDOR", "anthropic"),
        Timeout: parseDuration(getEnvOrDefault("TIMEOUT", "30s")),
        EnableLogging: parseBool(getEnvOrDefault("ENABLE_LOGGING", "true")),
        EnableMetrics: parseBool(getEnvOrDefault("ENABLE_METRICS", "true")),
    }

    // Load retry policy from environment
    if maxRetries := getEnvOrDefault("MAX_RETRIES", "3"); maxRetries != "" {
        config.RetryPolicy = &llmdispatcher.RetryPolicy{
            MaxRetries: parseInt(maxRetries),
            BackoffStrategy: llmdispatcher.ExponentialBackoff,
        }
    }

    return config
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Streaming Examples

### Basic Streaming

```go
func basicStreaming(dispatcher *llmdispatcher.Dispatcher) error {
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Write a short story about a robot."},
        },
        Stream: true,
    }

    ctx := context.Background()
    streamingResp, err := dispatcher.SendStreaming(ctx, request)
    if err != nil {
        return err
    }
    defer streamingResp.Close()

    for {
        select {
        case content := <-streamingResp.ContentChan:
            if content == "" {
                continue
            }
            fmt.Print(content)
        case done := <-streamingResp.DoneChan:
            if done {
                fmt.Println("\nStreaming completed")
                return nil
            }
        case err := <-streamingResp.ErrorChan:
            if err != nil {
                return fmt.Errorf("streaming error: %w", err)
            }
        }
    }
}
```

### Streaming with Context

```go
func streamingWithContext(dispatcher *llmdispatcher.Dispatcher) error {
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Explain quantum computing in simple terms."},
        },
        Stream: true,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

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
                return nil
            }
        case err := <-streamingResp.ErrorChan:
            return err
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### Streaming with Progress Tracking

```go
func streamingWithProgress(dispatcher *llmdispatcher.Dispatcher) error {
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Write a detailed analysis of AI trends."},
        },
        Stream: true,
    }

    ctx := context.Background()
    streamingResp, err := dispatcher.SendStreaming(ctx, request)
    if err != nil {
        return err
    }
    defer streamingResp.Close()

    var totalContent strings.Builder
    chunkCount := 0

    for {
        select {
        case content := <-streamingResp.ContentChan:
            if content == "" {
                continue
            }
            totalContent.WriteString(content)
            chunkCount++
            
            // Print progress every 10 chunks
            if chunkCount%10 == 0 {
                fmt.Printf("\rReceived %d chunks, %d characters", 
                    chunkCount, totalContent.Len())
            }
            
            fmt.Print(content)
        case done := <-streamingResp.DoneChan:
            if done {
                fmt.Printf("\nCompleted: %d chunks, %d characters\n", 
                    chunkCount, totalContent.Len())
                return nil
            }
        case err := <-streamingResp.ErrorChan:
            return err
        }
    }
}
```

## Error Handling Examples

### Comprehensive Error Handling

```go
func handleErrors(dispatcher *llmdispatcher.Dispatcher) error {
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello"},
        },
    }

    ctx := context.Background()
    response, err := dispatcher.Send(ctx, request)
    if err != nil {
        // Check for specific error types
        switch {
        case isRateLimitError(err):
            log.Println("Rate limit exceeded, implementing backoff...")
            return implementBackoff(dispatcher, request)
        case isAuthenticationError(err):
            log.Println("Authentication failed, check API keys")
            return fmt.Errorf("authentication error: %w", err)
        case isTimeoutError(err):
            log.Println("Request timed out, retrying with longer timeout...")
            return retryWithLongerTimeout(dispatcher, request)
        case isVendorUnavailableError(err):
            log.Println("Vendor unavailable, trying fallback...")
            return tryFallback(dispatcher, request)
        default:
            return fmt.Errorf("unexpected error: %w", err)
        }
    }

    fmt.Printf("Success: %s\n", response.Content)
    return nil
}

func isRateLimitError(err error) bool {
    return strings.Contains(err.Error(), "rate limit")
}

func isAuthenticationError(err error) bool {
    return strings.Contains(err.Error(), "authentication")
}

func isTimeoutError(err error) bool {
    return strings.Contains(err.Error(), "timeout")
}

func isVendorUnavailableError(err error) bool {
    return strings.Contains(err.Error(), "unavailable")
}
```

### Retry with Exponential Backoff

```go
func implementBackoff(dispatcher *llmdispatcher.Dispatcher, request *llmdispatcher.Request) error {
    maxRetries := 5
    baseDelay := time.Second

    for attempt := 0; attempt < maxRetries; attempt++ {
        delay := time.Duration(attempt+1) * baseDelay
        log.Printf("Retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
        
        time.Sleep(delay)

        ctx := context.Background()
        response, err := dispatcher.Send(ctx, request)
        if err == nil {
            fmt.Printf("Success after %d attempts: %s\n", attempt+1, response.Content)
            return nil
        }

        if !isRateLimitError(err) {
            return err
        }
    }

    return fmt.Errorf("failed after %d retries", maxRetries)
}
```

## Statistics and Monitoring Examples

### Basic Statistics

```go
func printStatistics(dispatcher *llmdispatcher.Dispatcher) {
    stats := dispatcher.GetStats()
    
    fmt.Printf("=== Dispatcher Statistics ===\n")
    fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
    fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
    fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
    fmt.Printf("Success Rate: %.2f%%\n", stats.SuccessRate())
    fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
    fmt.Printf("Total Cost: $%.4f\n", stats.TotalCost)
    fmt.Printf("Average Cost: $%.4f\n", stats.AverageCost)
    
    fmt.Printf("\n=== Vendor Statistics ===\n")
    for vendorName, vendorStats := range stats.VendorStats {
        fmt.Printf("%s:\n", vendorName)
        fmt.Printf("  Requests: %d\n", vendorStats.Requests)
        fmt.Printf("  Successes: %d\n", vendorStats.Successes)
        fmt.Printf("  Failures: %d\n", vendorStats.Failures)
        fmt.Printf("  Success Rate: %.2f%%\n", 
            float64(vendorStats.Successes)/float64(vendorStats.Requests)*100)
        fmt.Printf("  Total Cost: $%.4f\n", vendorStats.TotalCost)
        fmt.Printf("  Average Latency: %v\n", vendorStats.AvgLatency)
    }
}
```

### Real-time Monitoring

```go
func monitorInRealTime(dispatcher *llmdispatcher.Dispatcher) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            stats := dispatcher.GetStats()
            fmt.Printf("[%s] Requests: %d, Success Rate: %.2f%%, Cost: $%.4f\n",
                time.Now().Format("15:04:05"),
                stats.TotalRequests,
                stats.SuccessRate(),
                stats.TotalCost)
        }
    }
}
```

## Advanced Routing Examples

### Model-Based Routing

```go
func setupModelBasedRouting() *llmdispatcher.Dispatcher {
    config := &llmdispatcher.Config{
        RoutingRules: []llmdispatcher.RoutingRule{
            // GPT-4 models go to OpenAI
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "gpt-4",
                },
                Vendor: "openai",
                Priority: 1,
                Enabled: true,
            },
            // Claude models go to Anthropic
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "claude",
                },
                Vendor: "anthropic",
                Priority: 1,
                Enabled: true,
            },
            // Gemini models go to Google
            {
                Condition: llmdispatcher.RoutingCondition{
                    ModelPattern: "gemini",
                },
                Vendor: "google",
                Priority: 1,
                Enabled: true,
            },
            // Default to OpenAI for other models
            {
                Condition: llmdispatcher.RoutingCondition{},
                Vendor: "openai",
                Priority: 10,
                Enabled: true,
            },
        },
    }

    dispatcher := llmdispatcher.NewWithConfig(config)
    
    // Register vendors
    dispatcher.RegisterVendor(llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    }))
    dispatcher.RegisterVendor(llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    }))
    dispatcher.RegisterVendor(llmdispatcher.NewGoogleVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("GOOGLE_API_KEY"),
    }))

    return dispatcher
}
```

### Cost-Based Routing

```go
func setupCostBasedRouting() *llmdispatcher.Dispatcher {
    config := &llmdispatcher.Config{
        CostOptimization: &llmdispatcher.CostOptimization{
            Enabled: true,
            MaxCost: 0.05, // $0.05 per request
            PreferCheap: true,
            VendorCosts: map[string]float64{
                "openai":   0.002, // $0.002 per 1K tokens
                "anthropic": 0.003, // $0.003 per 1K tokens
                "google":    0.001, // $0.001 per 1K tokens
            },
        },
    }

    dispatcher := llmdispatcher.NewWithConfig(config)
    
    // Register vendors
    dispatcher.RegisterVendor(llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    }))
    dispatcher.RegisterVendor(llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    }))
    dispatcher.RegisterVendor(llmdispatcher.NewGoogleVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("GOOGLE_API_KEY"),
    }))

    return dispatcher
}
```

## Performance Examples

### Concurrent Requests

```go
func concurrentRequests(dispatcher *llmdispatcher.Dispatcher) error {
    requests := []*llmdispatcher.Request{
        {
            Model: "gpt-3.5-turbo",
            Messages: []llmdispatcher.Message{
                {Role: "user", Content: "Explain quantum physics"},
            },
        },
        {
            Model: "gpt-3.5-turbo",
            Messages: []llmdispatcher.Message{
                {Role: "user", Content: "Write a poem about AI"},
            },
        },
        {
            Model: "gpt-3.5-turbo",
            Messages: []llmdispatcher.Message{
                {Role: "user", Content: "Summarize machine learning"},
            },
        },
    }

    // Use semaphore to limit concurrent requests
    semaphore := make(chan struct{}, 3)
    results := make(chan string, len(requests))
    errors := make(chan error, len(requests))

    for i, request := range requests {
        go func(req *llmdispatcher.Request, index int) {
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release

            ctx := context.Background()
            response, err := dispatcher.Send(ctx, req)
            if err != nil {
                errors <- fmt.Errorf("request %d failed: %w", index, err)
                return
            }

            results <- fmt.Sprintf("Request %d: %s", index, response.Content)
        }(request, i)
    }

    // Collect results
    for i := 0; i < len(requests); i++ {
        select {
        case result := <-results:
            fmt.Println(result)
        case err := <-errors:
            return err
        }
    }

    return nil
}
```

### Benchmarking

```go
func benchmarkVendors(dispatcher *llmdispatcher.Dispatcher) {
    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello"},
        },
    }

    vendors := dispatcher.GetVendors()
    results := make(map[string][]time.Duration)

    for _, vendorName := range vendors {
        fmt.Printf("Benchmarking %s...\n", vendorName)
        
        for i := 0; i < 5; i++ {
            start := time.Now()
            
            ctx := context.Background()
            _, err := dispatcher.Send(ctx, request)
            
            duration := time.Since(start)
            results[vendorName] = append(results[vendorName], duration)
            
            if err != nil {
                fmt.Printf("  Request %d failed: %v\n", i+1, err)
            } else {
                fmt.Printf("  Request %d: %v\n", i+1, duration)
            }
            
            time.Sleep(1 * time.Second) // Rate limiting
        }
    }

    // Print summary
    fmt.Printf("\n=== Benchmark Results ===\n")
    for vendorName, durations := range results {
        if len(durations) == 0 {
            continue
        }
        
        var total time.Duration
        for _, d := range durations {
            total += d
        }
        avg := total / time.Duration(len(durations))
        
        fmt.Printf("%s: Average %v (%d successful requests)\n", 
            vendorName, avg, len(durations))
    }
}
```

## Integration Examples

### HTTP Server Integration

```go
func createHTTPServer(dispatcher *llmdispatcher.Dispatcher) {
    http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var requestBody struct {
            Model    string                    `json:"model"`
            Messages []llmdispatcher.Message   `json:"messages"`
            Stream   bool                      `json:"stream"`
        }

        if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        request := &llmdispatcher.Request{
            Model:    requestBody.Model,
            Messages: requestBody.Messages,
            Stream:   requestBody.Stream,
        }

        if requestBody.Stream {
            handleStreamingRequest(w, r, dispatcher, request)
        } else {
            handleRegularRequest(w, r, dispatcher, request)
        }
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRegularRequest(w http.ResponseWriter, r *http.Request, 
    dispatcher *llmdispatcher.Dispatcher, request *llmdispatcher.Request) {
    
    ctx := r.Context()
    response, err := dispatcher.Send(ctx, request)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func handleStreamingRequest(w http.ResponseWriter, r *http.Request, 
    dispatcher *llmdispatcher.Dispatcher, request *llmdispatcher.Request) {
    
    ctx := r.Context()
    streamingResp, err := dispatcher.SendStreaming(ctx, request)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer streamingResp.Close()

    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Transfer-Encoding", "chunked")
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    for {
        select {
        case content := <-streamingResp.ContentChan:
            if content == "" {
                continue
            }
            fmt.Fprint(w, content)
            flusher.Flush()
        case done := <-streamingResp.DoneChan:
            if done {
                return
            }
        case err := <-streamingResp.ErrorChan:
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### Database Integration

```go
type RequestLog struct {
    ID        int64     `json:"id"`
    Model     string    `json:"model"`
    Vendor    string    `json:"vendor"`
    Content   string    `json:"content"`
    Cost      float64   `json:"cost"`
    Latency   time.Duration `json:"latency"`
    CreatedAt time.Time `json:"created_at"`
}

func logRequest(db *sql.DB, response *llmdispatcher.Response, 
    request *llmdispatcher.Request, duration time.Duration) error {
    
    query := `
        INSERT INTO request_logs (model, vendor, content, cost, latency, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    
    _, err := db.Exec(query,
        request.Model,
        response.Vendor,
        response.Content,
        response.Usage.TotalTokens * 0.002 / 1000, // Approximate cost
        duration,
        time.Now(),
    )
    
    return err
}
```

## Testing Examples

### Unit Testing

```go
func TestDispatcher(t *testing.T) {
    // Create mock vendor
    mockVendor := &MockVendor{
        NameFunc: func() string { return "mock" },
        SendRequestFunc: func(ctx context.Context, req *llmdispatcher.Request) (*llmdispatcher.Response, error) {
            return &llmdispatcher.Response{
                Content: "Mock response",
                Model:   req.Model,
                Vendor:  "mock",
                Usage: llmdispatcher.Usage{
                    PromptTokens:     10,
                    CompletionTokens: 20,
                    TotalTokens:      30,
                },
            }, nil
        },
    }

    // Create dispatcher with mock vendor
    dispatcher := llmdispatcher.New()
    dispatcher.RegisterVendor(mockVendor)

    // Test request
    request := &llmdispatcher.Request{
        Model: "mock-model",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Hello"},
        },
    }

    ctx := context.Background()
    response, err := dispatcher.Send(ctx, request)
    
    assert.NoError(t, err)
    assert.Equal(t, "Mock response", response.Content)
    assert.Equal(t, "mock", response.Vendor)
    assert.Equal(t, "mock-model", response.Model)
}
```

### Integration Testing

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Only run if API keys are available
    if os.Getenv("OPENAI_API_KEY") == "" {
        t.Skip("OPENAI_API_KEY not set")
    }

    dispatcher := llmdispatcher.New()
    openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    dispatcher.RegisterVendor(openaiVendor)

    request := &llmdispatcher.Request{
        Model: "gpt-3.5-turbo",
        Messages: []llmdispatcher.Message{
            {Role: "user", Content: "Say hello"},
        },
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    response, err := dispatcher.Send(ctx, request)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    assert.Equal(t, "gpt-3.5-turbo", response.Model)
    assert.Equal(t, "openai", response.Vendor)
}
```

## Complete Application Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

type ChatServer struct {
    dispatcher *llmdispatcher.Dispatcher
}

func NewChatServer() *ChatServer {
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
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
            APIKey: apiKey,
        })
        dispatcher.RegisterVendor(openaiVendor)
    }

    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        anthropicVendor := llmdispatcher.NewAnthropicVendor(&llmdispatcher.VendorConfig{
            APIKey: apiKey,
        })
        dispatcher.RegisterVendor(anthropicVendor)
    }

    return &ChatServer{dispatcher: dispatcher}
}

func (s *ChatServer) handleChat(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var requestBody struct {
        Model    string                    `json:"model"`
        Messages []llmdispatcher.Message   `json:"messages"`
        Stream   bool                      `json:"stream"`
    }

    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    request := &llmdispatcher.Request{
        Model:    requestBody.Model,
        Messages: requestBody.Messages,
        Stream:   requestBody.Stream,
    }

    if requestBody.Stream {
        s.handleStreamingChat(w, r, request)
    } else {
        s.handleRegularChat(w, r, request)
    }
}

func (s *ChatServer) handleRegularChat(w http.ResponseWriter, r *http.Request, 
    request *llmdispatcher.Request) {
    
    ctx := r.Context()
    start := time.Now()
    
    response, err := s.dispatcher.Send(ctx, request)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("Chat request failed: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Log statistics
    stats := s.dispatcher.GetStats()
    log.Printf("Request completed in %v, total requests: %d, success rate: %.2f%%",
        duration, stats.TotalRequests, stats.SuccessRate())

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *ChatServer) handleStreamingChat(w http.ResponseWriter, r *http.Request, 
    request *llmdispatcher.Request) {
    
    ctx := r.Context()
    streamingResp, err := s.dispatcher.SendStreaming(ctx, request)
    if err != nil {
        log.Printf("Streaming request failed: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer streamingResp.Close()

    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Transfer-Encoding", "chunked")
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    for {
        select {
        case content := <-streamingResp.ContentChan:
            if content == "" {
                continue
            }
            fmt.Fprint(w, content)
            flusher.Flush()
        case done := <-streamingResp.DoneChan:
            if done {
                return
            }
        case err := <-streamingResp.ErrorChan:
            if err != nil {
                log.Printf("Streaming error: %v", err)
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        case <-ctx.Done():
            return
        }
    }
}

func (s *ChatServer) handleStats(w http.ResponseWriter, r *http.Request) {
    stats := s.dispatcher.GetStats()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func main() {
    server := NewChatServer()

    http.HandleFunc("/chat", server.handleChat)
    http.HandleFunc("/stats", server.handleStats)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Starting chat server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

This comprehensive examples guide covers all major use cases and provides working code that can be used as a reference for implementing the LLM Dispatcher in various scenarios. 