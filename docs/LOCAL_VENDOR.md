# Local Model Vendor

The local vendor allows you to integrate local models like Ollama or llama.cpp into your LLM dispatcher as the cheapest option for cost optimization.

## Overview

The local vendor supports two main integration methods:

1. **HTTP API** (Ollama) - Easy to set up, good for development
2. **Direct Process Execution** (llama.cpp) - Maximum performance, production-ready

## Features

- ✅ **Cost Optimization** - Set as the cheapest option ($0.0001 per 1K tokens)
- ✅ **Streaming Support** - Real-time text generation
- ✅ **Multiple Model Formats** - Ollama, llama.cpp, custom HTTP servers
- ✅ **Resource Management** - CPU, memory, and GPU layer limits
- ✅ **Error Handling** - Graceful fallbacks and retries
- ✅ **Health Checks** - Availability monitoring

## Quick Start

### 1. Install Ollama (Recommended)

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama
ollama serve

# Pull a model
ollama pull llama2:7b
```

### 2. Configure Local Vendor

```go
package main

import (
    "github.com/llmefficiency/llmdispatcher/internal/dispatcher"
    "github.com/llmefficiency/llmdispatcher/internal/models"
    "github.com/llmefficiency/llmdispatcher/internal/vendors"
)

func main() {
    // Create dispatcher with cost optimization
    config := &models.Config{
        DefaultVendor: "local",
        CostOptimization: &models.CostOptimization{
            Enabled:     true,
            PreferCheap: true,
            VendorCosts: map[string]float64{
                "local":    0.0001, // Cheapest option
                "openai":   0.0020,
                "anthropic": 0.0015,
            },
        },
    }

    disp := dispatcher.NewWithConfig(config)

    // Register local vendor
    localConfig := &models.VendorConfig{
        APIKey: "dummy", // Not used for local models
        Headers: map[string]string{
            "server_url": "http://localhost:11434", // Ollama default
            "model_path": "llama2:7b",             // Model name
        },
        Timeout: 60 * time.Second,
    }

    localVendor := vendors.NewLocal(localConfig)
    disp.RegisterVendor(localVendor)

    // Use the dispatcher
    req := &models.Request{
        Model: "llama2:7b",
        Messages: []models.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
    }

    resp, err := disp.Send(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
}
```

## Configuration Options

### HTTP API Configuration (Ollama)

```go
localConfig := &models.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:11434", // Ollama server
        "model_path": "llama2:7b",             // Model name
    },
    Timeout: 60 * time.Second,
}
```

### Direct Process Configuration (llama.cpp)

```go
localConfig := &models.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "executable": "/usr/local/bin/llama",     // llama.cpp binary
        "model_path": "/path/to/llama2-7b.gguf", // Model file
    },
    Timeout: 120 * time.Second, // Longer timeout for direct execution
}
```

### GPU-Optimized Configuration

```go
localConfig := &models.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:13b",
        "max_gpu_layers": "32",    // GPU layers
        "max_memory_mb": "8192",   // Memory limit
    },
    Timeout: 90 * time.Second,
}
```

### Custom HTTP Server

```go
localConfig := &models.VendorConfig{
    APIKey: "dummy",
    Headers: map[string]string{
        "server_url": "http://localhost:8080", // Custom server
        "model_path": "mistral:7b",            // Model name
    },
    Timeout: 60 * time.Second,
}
```

## Supported Models

The local vendor supports various model formats:

### Ollama Models
- `llama2:7b` - Llama 2 7B parameter model
- `llama2:13b` - Llama 2 13B parameter model
- `llama2:70b` - Llama 2 70B parameter model
- `codellama` - Code-focused Llama model
- `mistral` - Mistral 7B model
- `mixtral` - Mixtral 8x7B model

### llama.cpp Models
- Any GGUF format model file
- Supports various quantization levels (Q4_K_M, Q5_K_M, etc.)

## Resource Management

### Memory Limits

```go
// Limit memory usage to 4GB
Headers: map[string]string{
    "max_memory_mb": "4096",
}
```

### CPU Threads

```go
// Use 8 CPU threads
Headers: map[string]string{
    "max_threads": "8",
}
```

### GPU Layers

```go
// Use 32 GPU layers for acceleration
Headers: map[string]string{
    "max_gpu_layers": "32",
}
```

## Cost Optimization

The local vendor is designed to be the cheapest option in your cost optimization strategy:

```go
config := &models.Config{
    CostOptimization: &models.CostOptimization{
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
```

## Streaming Support

The local vendor supports streaming responses for real-time text generation:

```go
req := &models.Request{
    Model: "llama2:7b",
    Messages: []models.Message{
        {Role: "user", Content: "Write a poem about AI."},
    },
    Temperature: 0.8,
    MaxTokens:   200,
}

streamResp, err := disp.SendStreaming(context.Background(), req)
if err != nil {
    log.Fatal(err)
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

The local vendor includes comprehensive error handling:

### Common Errors

1. **Server Unavailable** - Ollama server not running
2. **Model Not Found** - Model not downloaded in Ollama
3. **Executable Not Found** - llama.cpp binary not found
4. **Model File Missing** - GGUF file not found
5. **Resource Limits** - Memory or GPU constraints

### Fallback Strategy

```go
config := &models.Config{
    DefaultVendor:  "local",
    FallbackVendor: "openai", // Fallback to OpenAI if local fails
    RetryPolicy: &models.RetryPolicy{
        MaxRetries:      3,
        BackoffStrategy: models.ExponentialBackoff,
        RetryableErrors: []string{
            "server unavailable",
            "timeout",
            "model not found",
        },
    },
}
```

## Performance Considerations

### HTTP API (Ollama)
- **Pros**: Easy setup, good for development
- **Cons**: Network overhead, server dependency
- **Best for**: Development, testing, small-scale production

### Direct Process (llama.cpp)
- **Pros**: Maximum performance, no network overhead
- **Cons**: More complex setup, resource intensive
- **Best for**: Production, high-performance requirements

### Resource Optimization

```go
// For development (lower resources)
localConfig := &models.VendorConfig{
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:7b",
        "max_memory_mb": "2048", // 2GB
        "max_threads": "2",       // 2 threads
    },
    Timeout: 60 * time.Second,
}

// For production (higher resources)
localConfig := &models.VendorConfig{
    Headers: map[string]string{
        "server_url": "http://localhost:11434",
        "model_path": "llama2:13b",
        "max_memory_mb": "8192", // 8GB
        "max_threads": "8",       // 8 threads
        "max_gpu_layers": "32",   // GPU acceleration
    },
    Timeout: 120 * time.Second,
}
```

## Monitoring and Metrics

The local vendor integrates with the dispatcher's metrics system:

```go
stats := dispatcher.GetStats()
fmt.Printf("Local vendor stats:\n")
fmt.Printf("  Requests: %d\n", stats.VendorStats["local"].Requests)
fmt.Printf("  Successes: %d\n", stats.VendorStats["local"].Successes)
fmt.Printf("  Failures: %d\n", stats.VendorStats["local"].Failures)
fmt.Printf("  Average latency: %v\n", stats.VendorStats["local"].AverageLatency)
fmt.Printf("  Total cost: $%.4f\n", stats.VendorStats["local"].TotalCost)
```

## Troubleshooting

### Common Issues

1. **Ollama server not running**
   ```bash
   # Start Ollama
   ollama serve
   ```

2. **Model not downloaded**
   ```bash
   # Pull the model
   ollama pull llama2:7b
   ```

3. **llama.cpp not installed**
   ```bash
   # Install llama.cpp
   git clone https://github.com/ggerganov/llama.cpp
   cd llama.cpp
   make
   ```

4. **Model file not found**
   ```bash
   # Download GGUF model
   wget https://huggingface.co/TheBloke/Llama-2-7B-GGUF/resolve/main/llama-2-7b.Q4_K_M.gguf
   ```

### Health Checks

```go
// Check if local vendor is available
if localVendor.IsAvailable(context.Background()) {
    fmt.Println("Local vendor is available")
} else {
    fmt.Println("Local vendor is not available")
}
```

## Best Practices

1. **Start with Ollama** - Easy to set up and test
2. **Use appropriate model sizes** - 7B for development, 13B+ for production
3. **Monitor resource usage** - Adjust memory and thread limits
4. **Implement fallbacks** - Always have a cloud vendor as backup
5. **Test thoroughly** - Verify model quality and performance
6. **Use cost optimization** - Let the dispatcher choose the cheapest option

## Example Integration

See `cmd/example/local_demo.go` for a complete working example that demonstrates:

- Multiple local vendor configurations
- Cost optimization with local models
- Streaming with local models
- Error handling and fallbacks
- Performance monitoring

## Conclusion

The local vendor provides a cost-effective way to integrate local models into your LLM dispatcher. By setting it as the cheapest option in your cost optimization strategy, you can significantly reduce costs while maintaining high-quality responses for appropriate use cases. 