# LLM Dispatcher Examples

This document shows how to use the LLM Dispatcher with the simplified configuration interface.

## Overview

The LLM Dispatcher now supports 4 predefined optimization modes:

- **Fast Mode**: Prioritizes speed over cost and accuracy
- **Sophisticated Mode**: Prioritizes accuracy and intelligence over speed and cost  
- **Cost Saving Mode**: Prioritizes cost savings over speed and accuracy
- **Auto Mode**: Automatically balances speed, accuracy, and cost

## Basic Usage

### Fast Mode - For Quick Responses

```go
package main

import (
    "context"
    "log"
    
    "github.com/llmefficiency/llmdispatcher/internal/models"
    "github.com/llmefficiency/llmdispatcher/internal/dispatcher"
)

func main() {
    // Create dispatcher with fast mode
    config := &models.Config{
        Mode: models.FastMode,
        Timeout: 10 * time.Second,
        EnableLogging: true,
    }
    
    d := dispatcher.NewWithConfig(config)
    
    // Register vendors
    d.RegisterVendor(openaiVendor)
    d.RegisterVendor(anthropicVendor)
    d.RegisterVendor(localVendor)
    
    // Send request - will prioritize fastest available vendor
    req := &models.Request{
        Model: "gpt-3.5-turbo",
        Messages: []models.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
    }
    
    resp, err := d.Send(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", resp.Content)
}
```

### Sophisticated Mode - For High-Quality Responses

```go
// Create dispatcher with sophisticated mode
config := &models.Config{
    Mode: models.SophisticatedMode,
    Timeout: 30 * time.Second,
    EnableLogging: true,
    ModeOverrides: &models.ModeOverrides{
        SophisticatedModels: []string{"claude-3-opus", "gpt-4", "gemini-pro"},
    },
}

d := dispatcher.NewWithConfig(config)
```

### Cost Saving Mode - For Budget-Conscious Usage

```go
// Create dispatcher with cost-saving mode
config := &models.Config{
    Mode: models.CostSavingMode,
    Timeout: 20 * time.Second,
    EnableLogging: true,
    ModeOverrides: &models.ModeOverrides{
        MaxCostPerRequest: 0.01, // $0.01 per request
    },
}

d := dispatcher.NewWithConfig(config)
```

### Auto Mode - For Balanced Optimization

```go
// Create dispatcher with auto mode (default)
config := &models.Config{
    Mode: models.AutoMode,
    Timeout: 15 * time.Second,
    EnableLogging: true,
}

d := dispatcher.NewWithConfig(config)
```

## Advanced Configuration

### Custom Vendor Preferences

You can override the default vendor preferences for each mode:

```go
config := &models.Config{
    Mode: models.FastMode,
    ModeOverrides: &models.ModeOverrides{
        VendorPreferences: map[models.Mode][]string{
            models.FastMode: {"local", "anthropic", "openai"},
            models.SophisticatedMode: {"anthropic", "openai", "google"},
            models.CostSavingMode: {"local", "azure", "google"},
            models.AutoMode: {"local", "anthropic", "openai", "google"},
        },
    },
}
```

### Mode-Specific Limits

```go
config := &models.Config{
    Mode: models.CostSavingMode,
    ModeOverrides: &models.ModeOverrides{
        MaxCostPerRequest: 0.005, // $0.005 per request
    },
}

// For fast mode with latency limits
config := &models.Config{
    Mode: models.FastMode,
    ModeOverrides: &models.ModeOverrides{
        MaxLatency: 2 * time.Second,
    },
}
```

## Mode Selection Strategy

### Fast Mode
- Prioritizes local vendors (fastest)
- Falls back to vendors known for low latency (Anthropic, OpenAI)
- Ignores cost and model sophistication

### Sophisticated Mode  
- Prioritizes vendors with the most capable models
- Preference order: Claude (Anthropic) > GPT-4 (OpenAI) > Google
- Ignores cost and speed considerations

### Cost Saving Mode
- Prioritizes local vendors (free)
- Falls back to cheaper cloud options
- Preference order: local > Azure > Google > OpenAI > Anthropic

### Auto Mode
- Balances all three factors
- Starts with local (good balance of speed and cost)
- Falls back to vendors that are reasonably fast and cost-effective
- Preference order: local > Anthropic > OpenAI > Google > Azure

## Migration from Old Configuration

If you were using the old complex routing strategies, here's how to migrate:

### Old Cascading Strategy
```go
// Old way
config := &models.Config{
    RoutingStrategy: &models.CascadingFailureStrategy{
        VendorOrder: []string{"openai", "anthropic", "google"},
    },
}

// New way - use mode overrides
config := &models.Config{
    Mode: models.AutoMode,
    ModeOverrides: &models.ModeOverrides{
        VendorPreferences: map[models.Mode][]string{
            models.AutoMode: {"openai", "anthropic", "google"},
        },
    },
}
```

### Old Cost Optimization
```go
// Old way
config := &models.Config{
    CostOptimization: &models.CostOptimization{
        Enabled: true,
        MaxCost: 0.01,
        PreferCheap: true,
    },
}

// New way
config := &models.Config{
    Mode: models.CostSavingMode,
    ModeOverrides: &models.ModeOverrides{
        MaxCostPerRequest: 0.01,
    },
}
```

## Best Practices

1. **Start with Auto Mode**: It provides a good balance for most use cases
2. **Use Fast Mode for real-time applications**: Chat interfaces, quick responses
3. **Use Sophisticated Mode for complex tasks**: Analysis, reasoning, creative writing
4. **Use Cost Saving Mode for high-volume usage**: Batch processing, testing
5. **Customize with ModeOverrides**: Fine-tune behavior when needed
6. **Monitor performance**: Use the built-in metrics to track vendor performance

## Error Handling

The dispatcher will automatically fall back to available vendors if the preferred ones are unavailable:

```go
resp, err := d.Send(ctx, req)
if err != nil {
    // Check if it's a vendor availability issue
    if errors.Is(err, models.ErrVendorUnavailable) {
        log.Printf("No vendors available for %s mode", d.config.Mode)
    }
    log.Fatal(err)
}
``` 