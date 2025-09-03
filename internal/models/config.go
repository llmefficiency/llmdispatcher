package models

import (
	"context"
	"fmt"
	"time"
)

// Strategy represents the predefined optimization strategies
type Strategy string

const (
	// SpeedStrategy prioritizes speed over cost and accuracy
	SpeedStrategy Strategy = "speed"

	// QualityStrategy prioritizes accuracy and intelligence over speed and cost
	QualityStrategy Strategy = "quality"

	// BudgetStrategy prioritizes cost savings over speed and accuracy
	BudgetStrategy Strategy = "budget"

	// BalancedStrategy automatically balances speed, accuracy, and cost
	BalancedStrategy Strategy = "balanced"
)

// StrategyContext represents the context and state for a specific strategy
type StrategyContext struct {
	Strategy         Strategy
	Request          *Request
	AvailableVendors map[string]LLMVendor
	Config           *Config
	Stats            *StrategyStats
	Context          context.Context
}

// StrategyStats tracks strategy-specific performance metrics
type StrategyStats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	AverageCost        float64
	LastRequestTime    time.Time
}

// StrategyImplementation defines the interface for strategy-specific behavior
type StrategyImplementation interface {
	// Name returns the strategy name
	Name() string

	// SelectVendor selects the best vendor for this strategy
	SelectVendor(ctx *StrategyContext) (LLMVendor, error)

	// PreprocessContext applies strategy-specific context preprocessing
	PreprocessContext(ctx *StrategyContext) error

	// OptimizeRequest applies strategy-specific request optimizations
	OptimizeRequest(ctx *StrategyContext) error

	// ValidateContext validates the context for this strategy
	ValidateContext(ctx *StrategyContext) error

	// GetPriority returns the priority level for this strategy (1-10, 10 being highest)
	GetPriority() int
}

// StrategyRegistry manages all available strategies and their implementations
type StrategyRegistry struct {
	strategies map[Strategy]StrategyImplementation
}

// NewStrategyRegistry creates a new strategy registry
func NewStrategyRegistry() *StrategyRegistry {
	registry := &StrategyRegistry{
		strategies: make(map[Strategy]StrategyImplementation),
	}

	// Register default strategies
	registry.RegisterStrategy(SpeedStrategy, NewSpeedStrategyImplementation())
	registry.RegisterStrategy(QualityStrategy, NewQualityStrategyImplementation())
	registry.RegisterStrategy(BudgetStrategy, NewBudgetStrategyImplementation())
	registry.RegisterStrategy(BalancedStrategy, NewBalancedStrategyImplementation())

	return registry
}

// RegisterStrategy registers a new strategy implementation
func (r *StrategyRegistry) RegisterStrategy(strategy Strategy, impl StrategyImplementation) {
	r.strategies[strategy] = impl
}

// GetStrategy returns the implementation for a given strategy
func (r *StrategyRegistry) GetStrategy(strategy Strategy) (StrategyImplementation, error) {
	impl, exists := r.strategies[strategy]
	if !exists {
		return nil, fmt.Errorf("no implementation registered for strategy: %s", strategy)
	}
	return impl, nil
}

// GetAvailableStrategies returns all registered strategies
func (r *StrategyRegistry) GetAvailableStrategies() []Strategy {
	strategies := make([]Strategy, 0, len(r.strategies))
	for strategy := range r.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// Config holds the simplified dispatcher configuration
type Config struct {
	// Strategy determines the optimization strategy
	Strategy Strategy `json:"strategy"`

	// Basic configuration
	Timeout       time.Duration `json:"timeout,omitempty"`
	EnableLogging bool          `json:"enable_logging"`
	EnableMetrics bool          `json:"enable_metrics"`

	// Retry configuration
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// Context preprocessing configuration
	ContextPreprocessing *ContextPreprocessingConfig `json:"context_preprocessing,omitempty"`
}

// ContextPreprocessingConfig defines how context should be preprocessed for each strategy
type ContextPreprocessingConfig struct {
	// Enable context preprocessing for each strategy
	EnabledStrategies map[Strategy]bool `json:"enabled_strategies,omitempty"`

	// Strategy-specific preprocessing rules
	PreprocessingRules map[Strategy][]PreprocessingRule `json:"preprocessing_rules,omitempty"`

	// Global preprocessing settings
	MaxContextLength    int  `json:"max_context_length,omitempty"`
	EnableSummarization bool `json:"enable_summarization,omitempty"`
	EnableCompression   bool `json:"enable_compression,omitempty"`
}

// PreprocessingRule defines a single preprocessing rule
type PreprocessingRule struct {
	Type       string                 `json:"type"`       // "summarize", "compress", "filter", "enhance"
	Condition  string                 `json:"condition"`  // When to apply this rule
	Parameters map[string]interface{} `json:"parameters"` // Rule-specific parameters
	Priority   int                    `json:"priority"`   // Execution priority (1-10)
}

// RetryPolicy defines how retries should be handled
type RetryPolicy struct {
	MaxRetries      int             `json:"max_retries"`
	BackoffStrategy BackoffStrategy `json:"backoff_strategy"`
	RetryableErrors []string        `json:"retryable_errors,omitempty"`
}

// BackoffStrategy defines the retry backoff strategy
type BackoffStrategy string

const (
	ExponentialBackoff BackoffStrategy = "exponential"
	LinearBackoff      BackoffStrategy = "linear"
	FixedBackoff       BackoffStrategy = "fixed"
)

// DispatcherStats holds statistics about the dispatcher
type DispatcherStats struct {
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	VendorStats        map[string]VendorStats `json:"vendor_stats"`
	AverageLatency     time.Duration          `json:"average_latency"`
	LastRequestTime    time.Time              `json:"last_request_time"`
	// Advanced metrics
	TotalCost    float64            `json:"total_cost"`
	AverageCost  float64            `json:"average_cost"`
	CostByVendor map[string]float64 `json:"cost_by_vendor"`
	// Strategy-specific stats
	StrategyStats map[Strategy]*StrategyStats `json:"strategy_stats"`
}

// VendorStats holds statistics for a specific vendor
type VendorStats struct {
	Requests       int64         `json:"requests"`
	Successes      int64         `json:"successes"`
	Failures       int64         `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
	// Advanced metrics
	TotalCost   float64 `json:"total_cost"`
	AverageCost float64 `json:"average_cost"`
	TokenUsage  int64   `json:"token_usage"`
}

// BaseStrategyImplementation provides common functionality for all strategy implementations
type BaseStrategyImplementation struct {
	strategy Strategy
	priority int
}

// NewBaseStrategyImplementation creates a new base strategy implementation
func NewBaseStrategyImplementation(strategy Strategy, priority int) *BaseStrategyImplementation {
	return &BaseStrategyImplementation{
		strategy: strategy,
		priority: priority,
	}
}

// Name returns the strategy name
func (b *BaseStrategyImplementation) Name() string {
	return string(b.strategy)
}

// GetPriority returns the priority level
func (b *BaseStrategyImplementation) GetPriority() int {
	return b.priority
}

// ValidateContext provides basic context validation
func (b *BaseStrategyImplementation) ValidateContext(ctx *StrategyContext) error {
	if ctx == nil {
		return fmt.Errorf("strategy context cannot be nil")
	}
	if ctx.Request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if len(ctx.AvailableVendors) == 0 {
		return fmt.Errorf("no available vendors")
	}
	return nil
}

// PreprocessContext provides basic context preprocessing
func (b *BaseStrategyImplementation) PreprocessContext(ctx *StrategyContext) error {
	// Create strategy-specific preprocessing pipeline
	pipeline := CreateStrategySpecificPipeline(ctx.Strategy, ctx.Config)

	// Execute the preprocessing pipeline
	return pipeline.Execute(ctx)
}

// OptimizeRequest provides basic request optimization
func (b *BaseStrategyImplementation) OptimizeRequest(ctx *StrategyContext) error {
	// TODO: Implement basic request optimization
	// - Set default parameters if not specified
	// - Apply mode-specific optimizations
	return nil
}

// SpeedStrategyImpl implements speed strategy behavior
type SpeedStrategyImpl struct {
	*BaseStrategyImplementation
}

// NewSpeedStrategyImplementation creates a new speed strategy implementation
func NewSpeedStrategyImplementation() *SpeedStrategyImpl {
	return &SpeedStrategyImpl{
		BaseStrategyImplementation: NewBaseStrategyImplementation(SpeedStrategy, 8),
	}
}

// SelectVendor selects the best vendor for speed strategy
func (f *SpeedStrategyImpl) SelectVendor(ctx *StrategyContext) (LLMVendor, error) {
	// Fast mode intelligence: prioritize vendors known for speed
	fastVendors := []struct {
		name     string
		priority int
	}{
		{"local", 1},     // Local is fastest (if available)
		{"anthropic", 2}, // Haiku is very fast
		{"openai", 3},    // GPT-3.5 is fast
		{"google", 4},    // Flash is fast
		{"azure", 5},     // Azure OpenAI
	}

	for _, fastVendor := range fastVendors {
		if vendor, exists := ctx.AvailableVendors[fastVendor.name]; exists && vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range ctx.AvailableVendors {
		if vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for fast mode")
}

// PreprocessContext applies speed strategy context preprocessing
func (f *SpeedStrategyImpl) PreprocessContext(ctx *StrategyContext) error {
	// TODO: Implement fast mode context preprocessing
	// - Truncate long contexts
	// - Remove unnecessary messages
	// - Optimize for speed
	return nil
}

// OptimizeRequest applies speed strategy request optimizations
func (f *SpeedStrategyImpl) OptimizeRequest(ctx *StrategyContext) error {
	req := ctx.Request

	// Speed optimizations
	if req.Temperature == 0 {
		req.Temperature = 0.3 // Lower temperature for faster, more deterministic responses
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 150 // Shorter responses for speed
	}
	if req.TopP == 0 {
		req.TopP = 0.8 // Slightly lower for faster generation
	}

	return nil
}

// QualityStrategyImpl implements quality strategy behavior
type QualityStrategyImpl struct {
	*BaseStrategyImplementation
}

// NewQualityStrategyImplementation creates a new quality strategy implementation
func NewQualityStrategyImplementation() *QualityStrategyImpl {
	return &QualityStrategyImpl{
		BaseStrategyImplementation: NewBaseStrategyImplementation(QualityStrategy, 10),
	}
}

// SelectVendor selects the best vendor for quality mode
func (s *QualityStrategyImpl) SelectVendor(ctx *StrategyContext) (LLMVendor, error) {
	// Sophisticated mode intelligence: prioritize vendors with most capable models
	sophisticatedVendors := []struct {
		name     string
		priority int
	}{
		{"anthropic", 1}, // Claude is most sophisticated
		{"openai", 2},    // GPT-4 is very capable
		{"google", 3},    // Gemini Pro is capable
		{"azure", 4},     // Azure OpenAI
		{"local", 5},     // Large local models (if available)
	}

	for _, sophisticatedVendor := range sophisticatedVendors {
		if vendor, exists := ctx.AvailableVendors[sophisticatedVendor.name]; exists && vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range ctx.AvailableVendors {
		if vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for sophisticated mode")
}

// PreprocessContext applies quality mode context preprocessing
func (s *QualityStrategyImpl) PreprocessContext(ctx *StrategyContext) error {
	// TODO: Implement sophisticated mode context preprocessing
	// - Enhance context with additional information
	// - Add relevant system prompts
	// - Optimize for quality
	return nil
}

// OptimizeRequest applies quality mode request optimizations
func (s *QualityStrategyImpl) OptimizeRequest(ctx *StrategyContext) error {
	req := ctx.Request

	// Sophistication optimizations
	if req.Temperature == 0 {
		req.Temperature = 0.7 // Higher temperature for more creative responses
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1000 // Longer responses for detailed answers
	}
	if req.TopP == 0 {
		req.TopP = 0.9 // Higher for more diverse responses
	}

	return nil
}

// BudgetStrategyImpl implements budget strategy behavior
type BudgetStrategyImpl struct {
	*BaseStrategyImplementation
}

// NewBudgetStrategyImplementation creates a new budget strategy implementation
func NewBudgetStrategyImplementation() *BudgetStrategyImpl {
	return &BudgetStrategyImpl{
		BaseStrategyImplementation: NewBaseStrategyImplementation(BudgetStrategy, 6),
	}
}

// SelectVendor selects the best vendor for budget mode
func (c *BudgetStrategyImpl) SelectVendor(ctx *StrategyContext) (LLMVendor, error) {
	// Cost-saving mode intelligence: prioritize cheapest vendors
	costSavingVendors := []struct {
		name     string
		priority int
		cost     float64 // Cost per 1K tokens (approximate)
	}{
		{"local", 1, 0.0001},    // Local is cheapest (if available)
		{"google", 2, 0.0005},   // Google is cheap
		{"openai", 3, 0.002},    // OpenAI is moderate
		{"anthropic", 4, 0.003}, // Anthropic is pricier
		{"azure", 5, 0.002},     // Azure is reasonable
	}

	for _, costVendor := range costSavingVendors {
		if vendor, exists := ctx.AvailableVendors[costVendor.name]; exists && vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range ctx.AvailableVendors {
		if vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for cost-saving mode")
}

// PreprocessContext applies budget mode context preprocessing
func (c *BudgetStrategyImpl) PreprocessContext(ctx *StrategyContext) error {
	// TODO: Implement cost-saving mode context preprocessing
	// - Compress context
	// - Remove redundant information
	// - Optimize for cost
	return nil
}

// OptimizeRequest applies budget mode request optimizations
func (c *BudgetStrategyImpl) OptimizeRequest(ctx *StrategyContext) error {
	req := ctx.Request

	// Cost-saving optimizations
	if req.Temperature == 0 {
		req.Temperature = 0.1 // Very low temperature for deterministic, shorter responses
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 100 // Very short responses to minimize tokens
	}
	if req.TopP == 0 {
		req.TopP = 0.7 // Lower for more focused, shorter responses
	}

	return nil
}

// BalancedStrategyImpl implements balanced strategy behavior
type BalancedStrategyImpl struct {
	*BaseStrategyImplementation
}

// NewBalancedStrategyImplementation creates a new balanced strategy implementation
func NewBalancedStrategyImplementation() *BalancedStrategyImpl {
	return &BalancedStrategyImpl{
		BaseStrategyImplementation: NewBaseStrategyImplementation(BalancedStrategy, 5),
	}
}

// SelectVendor selects the best vendor for balanced mode
func (a *BalancedStrategyImpl) SelectVendor(ctx *StrategyContext) (LLMVendor, error) {
	// Auto mode intelligence: balance speed, cost, and capability
	balancedVendors := []struct {
		name     string
		priority int
		speed    int // 1-5 scale
		cost     int // 1-5 scale (1=cheap, 5=expensive)
		quality  int // 1-5 scale
	}{
		{"anthropic", 1, 4, 4, 5}, // Good speed, high quality
		{"openai", 2, 4, 3, 4},    // Good balance
		{"google", 3, 3, 2, 4},    // Cheap, good quality
		{"local", 4, 5, 1, 3},     // Fast, cheap, decent quality (if available)
		{"azure", 5, 3, 3, 4},     // Moderate across all
	}

	for _, balancedVendor := range balancedVendors {
		if vendor, exists := ctx.AvailableVendors[balancedVendor.name]; exists && vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range ctx.AvailableVendors {
		if vendor.IsAvailable(ctx.Context) {
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for auto mode")
}

// PreprocessContext applies balanced mode context preprocessing
func (a *BalancedStrategyImpl) PreprocessContext(ctx *StrategyContext) error {
	// TODO: Implement auto mode context preprocessing
	// - Analyze context complexity
	// - Apply appropriate preprocessing based on analysis
	// - Balance preprocessing cost vs benefit
	return nil
}

// OptimizeRequest applies balanced mode request optimizations
func (a *BalancedStrategyImpl) OptimizeRequest(ctx *StrategyContext) error {
	req := ctx.Request

	// Balanced optimizations
	if req.Temperature == 0 {
		req.Temperature = 0.5 // Moderate temperature for balanced creativity
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 500 // Moderate length responses
	}
	if req.TopP == 0 {
		req.TopP = 0.85 // Moderate diversity
	}

	return nil
}
