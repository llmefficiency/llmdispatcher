# LLM Dispatcher Examples

This document shows how to use the LLM Dispatcher with the simplified configuration interface.

## Overview

The LLM Dispatcher now supports 4 predefined optimization strategies:

- **Fast Strategy**: Prioritizes speed over cost and accuracy
- **Sophisticated Strategy**: Prioritizes accuracy and intelligence over speed and cost  
- **Cost Saving Strategy**: Prioritizes cost savings over speed and accuracy
- **Auto Strategy**: Automatically balances speed, accuracy, and cost

## Basic Usage

### Fast Strategy - For Quick Responses

```go
package main

import (
    "context"
    "log"
    
    "github.com/llmefficiency/llmdispatcher/internal/models"
    "github.com/llmefficiency/llmdispatcher/internal/dispatcher"
)

func main() {
    // Create dispatcher with speed strategy
    config := &models.Config{
        Strategy: models.SpeedStrategy,
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

### Quality Strategy - For High-Quality Responses

```go
// Create dispatcher with quality strategy
config := &models.Config{
    Strategy: models.QualityStrategy,
    Timeout: 30 * time.Second,
    EnableLogging: true,
}

d := dispatcher.NewWithConfig(config)
```

### Budget Strategy - For Budget-Conscious Usage

```go
// Create dispatcher with budget strategy
config := &models.Config{
    Strategy: models.BudgetStrategy,
    Timeout: 20 * time.Second,
    EnableLogging: true,
}

d := dispatcher.NewWithConfig(config)
```

### Balanced Strategy - For Balanced Optimization

```go
// Create dispatcher with balanced strategy (default)
config := &models.Config{
    Strategy: models.BalancedStrategy,
    Timeout: 15 * time.Second,
    EnableLogging: true,
}

d := dispatcher.NewWithConfig(config)
```

## Advanced Configuration





## Strategy Selection Logic

### Fast Strategy
- Prioritizes local vendors (fastest)
- Falls back to vendors known for low latency (Anthropic, OpenAI)
- Ignores cost and model sophistication

### Sophisticated Strategy  
- Prioritizes vendors with the most capable models
- Preference order: Claude (Anthropic) > GPT-4 (OpenAI) > Google
- Ignores cost and speed considerations

### Cost Saving Strategy
- Prioritizes local vendors (free)
- Falls back to cheaper cloud options
- Preference order: local > Azure > Google > OpenAI > Anthropic

### Auto Strategy
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

// New way - use strategies
config := &models.Config{
    Strategy: models.BalancedStrategy,
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
    Strategy: models.BudgetStrategy,
}
```

## Best Practices

1. **Start with Auto Strategy**: It provides a good balance for most use cases
2. **Use Fast Strategy for real-time applications**: Chat interfaces, quick responses
3. **Use Sophisticated Strategy for complex tasks**: Analysis, reasoning, creative writing
4. **Use Cost Saving Strategy for high-volume usage**: Batch processing, testing
5. **Choose the right strategy**: Select the strategy that best fits your use case
6. **Monitor performance**: Use the built-in metrics to track vendor performance

## Error Handling

The dispatcher will automatically fall back to available vendors if the preferred ones are unavailable:

```go
resp, err := d.Send(ctx, req)
if err != nil {
    // Check if it's a vendor availability issue
    if errors.Is(err, models.ErrVendorUnavailable) {
        log.Printf("No vendors available for %s strategy", d.config.Strategy)
    }
    log.Fatal(err)
}
``` 