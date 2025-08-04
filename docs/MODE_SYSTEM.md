# Mode System Architecture

## Overview

The LLM Dispatcher implements a sophisticated mode-based system that provides different optimization strategies for vendor and model selection, along with context preprocessing capabilities. This system is designed to be extensible, future-proof, and highly configurable.

## Core Concepts

### Modes

The system supports four predefined modes:

- **Fast Mode**: Prioritizes speed over cost and accuracy
- **Sophisticated Mode**: Prioritizes accuracy and intelligence over speed and cost  
- **Cost Saving Mode**: Prioritizes cost savings over speed and accuracy
- **Auto Mode**: Automatically balances speed, accuracy, and cost

### Mode Context

Each mode operates within a `ModeContext` that contains:

- The current mode
- The request being processed
- Available vendors
- Configuration settings
- Mode-specific statistics
- Context information

### Mode Strategies

Each mode implements a `ModeStrategy` interface that provides:

- Vendor selection logic
- Context preprocessing
- Request optimization
- Context validation
- Priority-based execution

## Architecture Components

### 1. Mode Registry

The `ModeRegistry` manages all available modes and their strategies:

```go
registry := models.NewModeRegistry()
registry.RegisterStrategy(models.FastMode, customFastStrategy)
```

### 2. Context Preprocessing

The system includes a comprehensive context preprocessing pipeline:

#### Preprocessor Types

- **ContextLengthPreprocessor**: Truncates context if it exceeds maximum length
- **ContextCompressionPreprocessor**: Compresses context to reduce token usage
- **ContextEnhancementPreprocessor**: Enhances context with additional information
- **ContextFilterPreprocessor**: Filters context based on rules
- **ContextSummarizationPreprocessor**: Summarizes long contexts

#### Preprocessing Pipeline

```go
pipeline := models.NewPreprocessingPipeline()
pipeline.AddPreprocessor(models.NewContextLengthPreprocessor(1000))
pipeline.AddPreprocessor(models.NewContextCompressionPreprocessor(0.8))
pipeline.Execute(modeContext)
```

### 3. Mode-Specific Behavior

Each mode has specialized behavior:

#### Fast Mode
- Prioritizes vendors known for speed (local, anthropic haiku, openai gpt-3.5)
- Truncates context to reduce processing time
- Uses lower temperature and shorter responses
- Applies aggressive filtering to remove unnecessary content

#### Sophisticated Mode
- Prioritizes vendors with most capable models (anthropic claude, openai gpt-4)
- Enhances context with additional system prompts
- Uses higher temperature for more creative responses
- Applies summarization for long contexts

#### Cost Saving Mode
- Prioritizes cheapest vendors (local, google, openai)
- Compresses context to reduce token usage
- Uses very low temperature for deterministic responses
- Applies strict cost limits

#### Auto Mode
- Balances all factors intelligently
- Uses moderate settings across all parameters
- Adapts preprocessing based on context analysis
- Considers current metrics and availability

## Configuration

### Basic Configuration

```go
config := &models.Config{
    Mode: models.FastMode,
    Timeout: 30 * time.Second,
    EnableLogging: true,
    EnableMetrics: true,
    ContextPreprocessing: &models.ContextPreprocessingConfig{
        EnabledModes: map[models.Mode]bool{
            models.FastMode: true,
            models.SophisticatedMode: true,
        },
        MaxContextLength: 2000,
        EnableSummarization: true,
        EnableCompression: true,
    },
}
```

### Mode Overrides

```go
config.ModeOverrides = &models.ModeOverrides{
    VendorPreferences: map[models.Mode][]string{
        models.FastMode: {"local", "anthropic", "openai"},
        models.SophisticatedMode: {"anthropic", "openai", "google"},
    },
    MaxCostPerRequest: 0.10,
    MaxLatency: 5 * time.Second,
    ContextPreprocessing: map[models.Mode]*models.ContextPreprocessingConfig{
        models.FastMode: {
            MaxContextLength: 1000,
            EnableCompression: true,
        },
    },
}
```

## Usage Examples

### Basic Usage

```go
dispatcher := dispatcher.NewWithConfig(&models.Config{
    Mode: models.FastMode,
})

// Register vendors
dispatcher.RegisterVendor(openaiVendor)
dispatcher.RegisterVendor(anthropicVendor)

// Send request
response, err := dispatcher.Send(ctx, &models.Request{
    Messages: []models.Message{
        {Role: "user", Content: "Hello, how are you?"},
    },
})
```

### Custom Mode Strategy

```go
type CustomFastStrategy struct {
    *models.BaseModeStrategy
}

func NewCustomFastStrategy() *CustomFastStrategy {
    return &CustomFastStrategy{
        BaseModeStrategy: models.NewBaseModeStrategy(models.FastMode, 9),
    }
}

func (c *CustomFastStrategy) SelectVendor(ctx *models.ModeContext) (models.LLMVendor, error) {
    // Custom vendor selection logic
    return ctx.AvailableVendors["custom_vendor"], nil
}

// Register custom strategy
dispatcher.RegisterModeStrategy(models.FastMode, NewCustomFastStrategy())
```

### Custom Context Preprocessor

```go
type CustomPreprocessor struct {
    *models.BaseContextPreprocessor
}

func NewCustomPreprocessor() *CustomPreprocessor {
    return &CustomPreprocessor{
        BaseContextPreprocessor: models.NewBaseContextPreprocessor("custom", 5),
    }
}

func (c *CustomPreprocessor) Preprocess(ctx *models.ModeContext) error {
    // Custom preprocessing logic
    return nil
}

// Use in pipeline
pipeline := models.NewPreprocessingPipeline()
pipeline.AddPreprocessor(NewCustomPreprocessor())
```

## Future Extensibility

### Adding New Modes

1. Define the new mode constant
2. Create a mode strategy implementing `ModeStrategy`
3. Register the strategy with the registry
4. Add mode-specific preprocessing rules

### Adding New Preprocessors

1. Implement the `ContextPreprocessor` interface
2. Add to the preprocessing pipeline
3. Configure mode-specific usage

### Vendor Selection Algorithms

The system supports pluggable vendor selection algorithms:

- Priority-based selection
- Cost-aware selection
- Latency-aware selection
- Quality-aware selection
- Hybrid selection algorithms

## Performance Considerations

### Context Preprocessing

- Preprocessors are executed in priority order
- Failed preprocessing doesn't stop the request
- Preprocessing results are cached when possible
- Mode-specific preprocessing rules reduce unnecessary processing

### Vendor Selection

- Vendor availability is checked before selection
- Mode overrides provide fast-path selection
- Fallback mechanisms ensure request completion
- Selection algorithms are optimized for speed

### Statistics and Metrics

- Mode-specific statistics are tracked
- Performance metrics help optimize selection
- Cost tracking enables budget management
- Latency tracking enables performance optimization

## Best Practices

### Mode Selection

1. Use Fast Mode for real-time applications
2. Use Sophisticated Mode for quality-critical tasks
3. Use Cost Saving Mode for budget-constrained scenarios
4. Use Auto Mode for general-purpose applications

### Configuration

1. Set appropriate timeouts for each mode
2. Configure cost limits for cost-sensitive applications
3. Enable preprocessing for long-context scenarios
4. Use mode overrides for specific requirements

### Customization

1. Extend base strategies rather than implementing from scratch
2. Use the preprocessing pipeline for context manipulation
3. Implement vendor selection based on your specific needs
4. Add custom metrics for better optimization

## Troubleshooting

### Common Issues

1. **No vendors available**: Check vendor registration and availability
2. **Context too long**: Adjust preprocessing configuration
3. **High costs**: Use cost-saving mode or set cost limits
4. **Slow responses**: Use fast mode or optimize preprocessing

### Debugging

1. Enable logging to see mode selection decisions
2. Check mode statistics for performance insights
3. Monitor preprocessing pipeline execution
4. Verify vendor availability and capabilities

## Migration Guide

### From Old System

1. Update configuration to use new mode system
2. Replace direct vendor selection with mode-based selection
3. Add context preprocessing configuration
4. Update error handling for new mode context

### Configuration Changes

- `Mode` field now uses `models.Mode` type
- Add `ContextPreprocessing` configuration
- Use `ModeOverrides` for fine-tuning
- Update vendor preferences in mode overrides 