# Troubleshooting Guide

## Common Issues and Solutions

### Authentication Errors

#### Issue: "authentication failed" or "invalid API key"

**Symptoms:**
- Error messages mentioning authentication
- 401 HTTP status codes
- Requests failing immediately

**Solutions:**

1. **Check API Key Format**
```bash
# OpenAI keys should start with "sk-"
echo $OPENAI_API_KEY | head -c 10

# Anthropic keys should start with "sk-ant-"
echo $ANTHROPIC_API_KEY | head -c 10

# Google keys should be valid API keys
echo $GOOGLE_API_KEY | head -c 10
```

2. **Verify Environment Variables**
```bash
# Check if variables are set
env | grep -i api_key

# Test with a simple request
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

3. **Check API Key Permissions**
- Ensure the API key has the correct permissions
- Verify the key is not expired
- Check if the key has usage limits

**Example Fix:**
```go
// Verify API key is loaded
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    log.Fatal("OPENAI_API_KEY environment variable not set")
}

// Test vendor availability
vendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: apiKey,
})

if !vendor.IsAvailable(context.Background()) {
    log.Fatal("OpenAI vendor is not available - check API key")
}
```

### Rate Limiting Issues

#### Issue: "rate limit exceeded" or "too many requests"

**Symptoms:**
- 429 HTTP status codes
- Intermittent failures
- Requests working sometimes but not others

**Solutions:**

1. **Implement Retry Logic**
```go
config := &llmdispatcher.Config{
    RetryPolicy: &llmdispatcher.RetryPolicy{
        MaxRetries: 5,
        BackoffStrategy: llmdispatcher.ExponentialBackoff,
        RetryableErrors: []string{
            "rate limit exceeded",
            "too many requests",
            "quota exceeded",
        },
        InitialDelay: 1 * time.Second,
        MaxDelay: 60 * time.Second,
    },
}
```

2. **Use Multiple Vendors**
```go
// Register multiple vendors for fallback
dispatcher.RegisterVendor(openaiVendor)
dispatcher.RegisterVendor(anthropicVendor)
dispatcher.RegisterVendor(googleVendor)

// Configure fallback
config := &llmdispatcher.Config{
    DefaultVendor: "openai",
    FallbackVendor: "anthropic",
}
```

3. **Monitor Rate Limits**
```go
stats := dispatcher.GetStats()
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d failures\n", 
        vendorName, vendorStats.Requests, vendorStats.Failures)
}
```

### Timeout Issues

#### Issue: "context deadline exceeded" or "request timeout"

**Symptoms:**
- Requests taking too long
- Context cancellation errors
- Hanging requests

**Solutions:**

1. **Increase Timeout Settings**
```go
config := &llmdispatcher.Config{
    Timeout: 60 * time.Second, // Increase from default 30s
}

// Vendor-specific timeouts
openaiVendor := llmdispatcher.NewOpenAIVendor(&llmdispatcher.VendorConfig{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Timeout: 45 * time.Second,
})
```

2. **Use Context with Timeout**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := dispatcher.Send(ctx, request)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Println("Request timed out - consider increasing timeout")
    }
    return err
}
```

3. **Check Network Connectivity**
```bash
# Test connectivity to vendor APIs
curl -I https://api.openai.com/v1/models
curl -I https://api.anthropic.com/v1/messages
curl -I https://generativelanguage.googleapis.com/v1/models
```

### Model Not Found Errors

#### Issue: "model not found" or "unsupported model"

**Symptoms:**
- 400 HTTP status codes
- Model name errors
- Vendor rejecting requests

**Solutions:**

1. **Check Model Names**
```go
// Verify model is supported by vendor
capabilities := vendor.GetCapabilities()
for _, model := range capabilities.Models {
    fmt.Printf("Supported model: %s\n", model)
}

// Use correct model names
request := &llmdispatcher.Request{
    Model: "gpt-3.5-turbo", // Not "gpt-3.5"
    Messages: []llmdispatcher.Message{
        {Role: "user", Content: "Hello"},
    },
}
```

2. **Vendor-Specific Model Names**
```go
// OpenAI models
openaiModels := []string{
    "gpt-3.5-turbo",
    "gpt-4",
    "gpt-4-turbo",
    "gpt-4o",
}

// Anthropic models
anthropicModels := []string{
    "claude-3-opus",
    "claude-3-sonnet",
    "claude-3-haiku",
    "claude-3-5-sonnet",
    "claude-3-5-haiku",
}

// Google models
googleModels := []string{
    "gemini-1.5-pro",
    "gemini-1.5-flash",
    "gemini-pro",
    "gemini-pro-vision",
}
```

### Streaming Issues

#### Issue: Streaming not working or hanging

**Symptoms:**
- No content received
- Channels not closing
- Memory leaks

**Solutions:**

1. **Proper Channel Handling**
```go
streamingResp, err := dispatcher.SendStreaming(ctx, request)
if err != nil {
    return err
}
defer streamingResp.Close() // Always close

for {
    select {
    case content := <-streamingResp.ContentChan:
        if content == "" {
            continue // Skip empty chunks
        }
        fmt.Print(content)
    case done := <-streamingResp.DoneChan:
        if done {
            fmt.Println("\nStreaming completed")
            return
        }
    case err := <-streamingResp.ErrorChan:
        if err != nil {
            return fmt.Errorf("streaming error: %w", err)
        }
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

2. **Check Streaming Support**
```go
capabilities := vendor.GetCapabilities()
if !capabilities.SupportsStreaming {
    log.Println("Vendor does not support streaming")
    // Fall back to non-streaming
    response, err := dispatcher.Send(ctx, request)
    return err
}
```

### Memory Issues

#### Issue: High memory usage or leaks

**Symptoms:**
- Increasing memory usage
- Out of memory errors
- Slow performance

**Solutions:**

1. **Close Streaming Responses**
```go
// Always close streaming responses
defer streamingResp.Close()

// Or use a function to ensure cleanup
func handleStreaming(streamingResp *llmdispatcher.StreamingResponse) {
    defer streamingResp.Close()
    // Handle streaming...
}
```

2. **Limit Concurrent Requests**
```go
// Use semaphore to limit concurrent requests
semaphore := make(chan struct{}, 10) // Max 10 concurrent

for i := 0; i < 100; i++ {
    semaphore <- struct{}{} // Acquire
    go func() {
        defer func() { <-semaphore }() // Release
        response, err := dispatcher.Send(ctx, request)
        // Handle response...
    }()
}
```

3. **Monitor Memory Usage**
```go
import "runtime"

// Log memory usage
var m runtime.MemStats
runtime.ReadMemStats(&m)
log.Printf("Memory usage: %d MB", m.Alloc/1024/1024)
```

### Configuration Issues

#### Issue: Configuration not loading or incorrect

**Symptoms:**
- Default values being used
- Environment variables not read
- Configuration file errors

**Solutions:**

1. **Check Environment Variable Loading**
```go
// Debug environment variables
for _, env := range []string{
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
    "GOOGLE_API_KEY",
} {
    value := os.Getenv(env)
    if value == "" {
        log.Printf("Warning: %s not set", env)
    } else {
        log.Printf("%s: %s...", env, value[:10])
    }
}
```

2. **Load Configuration from File**
```go
import "gopkg.in/yaml.v3"

type Config struct {
    Vendors map[string]VendorConfig `yaml:"vendors"`
}

func loadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

3. **Validate Configuration**
```go
func validateConfig(config *llmdispatcher.Config) error {
    if config.DefaultVendor == "" {
        return fmt.Errorf("default vendor not specified")
    }
    
    if config.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    
    return nil
}
```

### Network Issues

#### Issue: Network connectivity problems

**Symptoms:**
- Connection refused errors
- DNS resolution failures
- Intermittent connectivity

**Solutions:**

1. **Check Network Connectivity**
```bash
# Test basic connectivity
ping api.openai.com
ping api.anthropic.com
ping generativelanguage.googleapis.com

# Test HTTP connectivity
curl -I https://api.openai.com/v1/models
```

2. **Configure HTTP Client**
```go
import "net/http"

// Custom HTTP client with timeouts
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

// Use custom client in vendor
vendor := &CustomVendor{
    client: httpClient,
}
```

3. **Handle Network Errors**
```go
func isNetworkError(err error) bool {
    if err == nil {
        return false
    }
    
    // Check for network-related errors
    return strings.Contains(err.Error(), "connection refused") ||
           strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "no route to host")
}

// Retry on network errors
if isNetworkError(err) {
    log.Println("Network error detected, retrying...")
    time.Sleep(1 * time.Second)
    // Retry logic...
}
```

### Debugging Techniques

#### Enable Logging
```go
config := &llmdispatcher.Config{
    EnableLogging: true,
    EnableMetrics: true,
}

dispatcher := llmdispatcher.NewWithConfig(config)
```

#### Add Custom Logging
```go
import "log"

func (v *CustomVendor) SendRequest(ctx context.Context, req *Request) (*Response, error) {
    log.Printf("Sending request to %s with model %s", v.Name(), req.Model)
    
    start := time.Now()
    response, err := v.sendRequest(ctx, req)
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("Request failed after %v: %v", duration, err)
    } else {
        log.Printf("Request completed in %v", duration)
    }
    
    return response, err
}
```

#### Monitor Statistics
```go
// Get detailed statistics
stats := dispatcher.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n", stats.SuccessRate())
fmt.Printf("Average latency: %v\n", stats.AverageLatency)

// Vendor-specific stats
for vendorName, vendorStats := range stats.VendorStats {
    fmt.Printf("%s: %d requests, %d successes, %d failures\n",
        vendorName, vendorStats.Requests, vendorStats.Successes, vendorStats.Failures)
}
```

### Performance Issues

#### Issue: Slow response times or high latency

**Symptoms:**
- Long response times
- Timeout errors
- Poor user experience

**Solutions:**

1. **Use Latency Optimization**
```go
config := &llmdispatcher.Config{
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
```

2. **Implement Caching**
```go
type ResponseCache struct {
    cache map[string]*llmdispatcher.Response
    mu    sync.RWMutex
}

func (c *ResponseCache) Get(key string) (*llmdispatcher.Response, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    response, exists := c.cache[key]
    return response, exists
}

func (c *ResponseCache) Set(key string, response *llmdispatcher.Response) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[key] = response
}
```

3. **Use Connection Pooling**
```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
    },
    Timeout: 30 * time.Second,
}
```

## Getting Help

### Debug Information
When reporting issues, include:

1. **Environment Information**
```bash
go version
go env
uname -a
```

2. **Configuration**
```go
// Log your configuration (without sensitive data)
fmt.Printf("Default vendor: %s\n", config.DefaultVendor)
fmt.Printf("Timeout: %v\n", config.Timeout)
fmt.Printf("Retry policy: %+v\n", config.RetryPolicy)
```

3. **Error Details**
```go
// Capture full error context
if err != nil {
    log.Printf("Error type: %T", err)
    log.Printf("Error message: %v", err)
    log.Printf("Stack trace: %+v", err)
}
```

### Common Error Messages

| Error Message | Cause | Solution |
|---------------|-------|----------|
| "authentication failed" | Invalid API key | Check API key format and permissions |
| "rate limit exceeded" | Too many requests | Implement retry logic or use multiple vendors |
| "model not found" | Invalid model name | Use correct vendor-specific model names |
| "context deadline exceeded" | Request timeout | Increase timeout settings |
| "connection refused" | Network issue | Check network connectivity |
| "invalid request" | Malformed request | Validate request parameters |

### Support Resources

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check README.md and API_REFERENCE.md
- **Examples**: See cmd/example/ for usage examples
- **Tests**: Run tests to verify functionality 