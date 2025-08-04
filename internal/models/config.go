package models

import (
	"context"
	"fmt"
	"time"
)

// Mode represents the predefined optimization modes
type Mode string

const (
	// FastMode prioritizes speed over cost and accuracy
	FastMode Mode = "fast"

	// SophisticatedMode prioritizes accuracy and intelligence over speed and cost
	SophisticatedMode Mode = "sophisticated"

	// CostSavingMode prioritizes cost savings over speed and accuracy
	CostSavingMode Mode = "cost_saving"

	// AutoMode automatically balances speed, accuracy, and cost
	AutoMode Mode = "auto"
)

// ModeContext represents the context and state for a specific mode
type ModeContext struct {
	Mode             Mode
	Request          *Request
	AvailableVendors map[string]LLMVendor
	Config           *Config
	Stats            *ModeStats
	Context          context.Context
}

// ModeStats tracks mode-specific performance metrics
type ModeStats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	AverageCost        float64
	LastRequestTime    time.Time
}

// ModeStrategy defines the interface for mode-specific behavior
type ModeStrategy interface {
	// Name returns the strategy name
	Name() string

	// SelectVendor selects the best vendor for this mode
	SelectVendor(ctx *ModeContext) (LLMVendor, error)

	// PreprocessContext applies mode-specific context preprocessing
	PreprocessContext(ctx *ModeContext) error

	// OptimizeRequest applies mode-specific request optimizations
	OptimizeRequest(ctx *ModeContext) error

	// ValidateContext validates the context for this mode
	ValidateContext(ctx *ModeContext) error

	// GetPriority returns the priority level for this mode (1-10, 10 being highest)
	GetPriority() int
}

// ModeRegistry manages all available modes and their strategies
type ModeRegistry struct {
	strategies map[Mode]ModeStrategy
}

// NewModeRegistry creates a new mode registry
func NewModeRegistry() *ModeRegistry {
	registry := &ModeRegistry{
		strategies: make(map[Mode]ModeStrategy),
	}

	// Register default strategies
	registry.RegisterStrategy(FastMode, NewFastModeStrategy())
	registry.RegisterStrategy(SophisticatedMode, NewSophisticatedModeStrategy())
	registry.RegisterStrategy(CostSavingMode, NewCostSavingModeStrategy())
	registry.RegisterStrategy(AutoMode, NewAutoModeStrategy())

	return registry
}

// RegisterStrategy registers a new mode strategy
func (r *ModeRegistry) RegisterStrategy(mode Mode, strategy ModeStrategy) {
	r.strategies[mode] = strategy
}

// GetStrategy returns the strategy for a given mode
func (r *ModeRegistry) GetStrategy(mode Mode) (ModeStrategy, error) {
	strategy, exists := r.strategies[mode]
	if !exists {
		return nil, fmt.Errorf("no strategy registered for mode: %s", mode)
	}
	return strategy, nil
}

// GetAvailableModes returns all registered modes
func (r *ModeRegistry) GetAvailableModes() []Mode {
	modes := make([]Mode, 0, len(r.strategies))
	for mode := range r.strategies {
		modes = append(modes, mode)
	}
	return modes
}

// Config holds the simplified dispatcher configuration
type Config struct {
	// Mode determines the optimization strategy
	Mode Mode `json:"mode"`

	// Basic configuration
	Timeout       time.Duration `json:"timeout,omitempty"`
	EnableLogging bool          `json:"enable_logging"`
	EnableMetrics bool          `json:"enable_metrics"`

	// Retry configuration
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// Mode-specific overrides (optional)
	ModeOverrides *ModeOverrides `json:"mode_overrides,omitempty"`

	// Context preprocessing configuration
	ContextPreprocessing *ContextPreprocessingConfig `json:"context_preprocessing,omitempty"`
}

// ContextPreprocessingConfig defines how context should be preprocessed for each mode
type ContextPreprocessingConfig struct {
	// Enable context preprocessing for each mode
	EnabledModes map[Mode]bool `json:"enabled_modes,omitempty"`

	// Mode-specific preprocessing rules
	PreprocessingRules map[Mode][]PreprocessingRule `json:"preprocessing_rules,omitempty"`

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

// ModeOverrides allows fine-tuning of mode behavior
type ModeOverrides struct {
	// Vendor preferences for each mode (ordered by preference)
	VendorPreferences map[Mode][]string `json:"vendor_preferences,omitempty"`

	// Cost limits for cost-saving mode
	MaxCostPerRequest float64 `json:"max_cost_per_request,omitempty"`

	// Latency limits for fast mode
	MaxLatency time.Duration `json:"max_latency,omitempty"`

	// Model preferences for sophisticated mode
	SophisticatedModels []string `json:"sophisticated_models,omitempty"`

	// Context preprocessing overrides
	ContextPreprocessing map[Mode]*ContextPreprocessingConfig `json:"context_preprocessing,omitempty"`
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
	// Mode-specific stats
	ModeStats map[Mode]*ModeStats `json:"mode_stats"`
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

// BaseModeStrategy provides common functionality for all mode strategies
type BaseModeStrategy struct {
	mode     Mode
	priority int
}

// NewBaseModeStrategy creates a new base mode strategy
func NewBaseModeStrategy(mode Mode, priority int) *BaseModeStrategy {
	return &BaseModeStrategy{
		mode:     mode,
		priority: priority,
	}
}

// Name returns the strategy name
func (b *BaseModeStrategy) Name() string {
	return string(b.mode)
}

// GetPriority returns the priority level
func (b *BaseModeStrategy) GetPriority() int {
	return b.priority
}

// ValidateContext provides basic context validation
func (b *BaseModeStrategy) ValidateContext(ctx *ModeContext) error {
	if ctx == nil {
		return fmt.Errorf("mode context cannot be nil")
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
func (b *BaseModeStrategy) PreprocessContext(ctx *ModeContext) error {
	// Create mode-specific preprocessing pipeline
	pipeline := CreateModeSpecificPipeline(ctx.Mode, ctx.Config)

	// Execute the preprocessing pipeline
	return pipeline.Execute(ctx)
}

// OptimizeRequest provides basic request optimization
func (b *BaseModeStrategy) OptimizeRequest(ctx *ModeContext) error {
	// TODO: Implement basic request optimization
	// - Set default parameters if not specified
	// - Apply mode-specific optimizations
	return nil
}

// estimateRequestCost estimates the cost of a request based on token count and vendor cost
func (b *BaseModeStrategy) estimateRequestCost(req *Request, costPer1KTokens float64) float64 {
	// Rough estimation based on input length and max tokens
	inputTokens := b.estimateInputTokens(req)
	outputTokens := req.MaxTokens
	if outputTokens == 0 {
		outputTokens = 500 // Default estimate
	}

	totalTokens := inputTokens + outputTokens
	return (float64(totalTokens) / 1000.0) * costPer1KTokens
}

// estimateInputTokens roughly estimates the number of tokens in the input
func (b *BaseModeStrategy) estimateInputTokens(req *Request) int {
	totalChars := 0
	for _, msg := range req.Messages {
		totalChars += len(msg.Content)
	}

	// Rough estimation: 1 token ≈ 4 characters
	return totalChars / 4
}

// FastModeStrategy implements fast mode behavior
type FastModeStrategy struct {
	*BaseModeStrategy
}

// NewFastModeStrategy creates a new fast mode strategy
func NewFastModeStrategy() *FastModeStrategy {
	return &FastModeStrategy{
		BaseModeStrategy: NewBaseModeStrategy(FastMode, 8),
	}
}

// SelectVendor selects the best vendor for fast mode
func (f *FastModeStrategy) SelectVendor(ctx *ModeContext) (LLMVendor, error) {
	// Check mode overrides first
	if ctx.Config.ModeOverrides != nil {
		if preferences, exists := ctx.Config.ModeOverrides.VendorPreferences[FastMode]; exists {
			for _, vendorName := range preferences {
				if vendor, exists := ctx.AvailableVendors[vendorName]; exists && vendor.IsAvailable(ctx.Context) {
					return vendor, nil
				}
			}
		}
	}

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

// PreprocessContext applies fast mode context preprocessing
func (f *FastModeStrategy) PreprocessContext(ctx *ModeContext) error {
	// TODO: Implement fast mode context preprocessing
	// - Truncate long contexts
	// - Remove unnecessary messages
	// - Optimize for speed
	return nil
}

// OptimizeRequest applies fast mode request optimizations
func (f *FastModeStrategy) OptimizeRequest(ctx *ModeContext) error {
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

// SophisticatedModeStrategy implements sophisticated mode behavior
type SophisticatedModeStrategy struct {
	*BaseModeStrategy
}

// NewSophisticatedModeStrategy creates a new sophisticated mode strategy
func NewSophisticatedModeStrategy() *SophisticatedModeStrategy {
	return &SophisticatedModeStrategy{
		BaseModeStrategy: NewBaseModeStrategy(SophisticatedMode, 10),
	}
}

// SelectVendor selects the best vendor for sophisticated mode
func (s *SophisticatedModeStrategy) SelectVendor(ctx *ModeContext) (LLMVendor, error) {
	// Check mode overrides first
	if ctx.Config.ModeOverrides != nil {
		if preferences, exists := ctx.Config.ModeOverrides.VendorPreferences[SophisticatedMode]; exists {
			for _, vendorName := range preferences {
				if vendor, exists := ctx.AvailableVendors[vendorName]; exists && vendor.IsAvailable(ctx.Context) {
					return vendor, nil
				}
			}
		}
	}

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

// PreprocessContext applies sophisticated mode context preprocessing
func (s *SophisticatedModeStrategy) PreprocessContext(ctx *ModeContext) error {
	// TODO: Implement sophisticated mode context preprocessing
	// - Enhance context with additional information
	// - Add relevant system prompts
	// - Optimize for quality
	return nil
}

// OptimizeRequest applies sophisticated mode request optimizations
func (s *SophisticatedModeStrategy) OptimizeRequest(ctx *ModeContext) error {
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

// CostSavingModeStrategy implements cost-saving mode behavior
type CostSavingModeStrategy struct {
	*BaseModeStrategy
}

// NewCostSavingModeStrategy creates a new cost-saving mode strategy
func NewCostSavingModeStrategy() *CostSavingModeStrategy {
	return &CostSavingModeStrategy{
		BaseModeStrategy: NewBaseModeStrategy(CostSavingMode, 6),
	}
}

// SelectVendor selects the best vendor for cost-saving mode
func (c *CostSavingModeStrategy) SelectVendor(ctx *ModeContext) (LLMVendor, error) {
	// Check mode overrides first
	if ctx.Config.ModeOverrides != nil {
		if preferences, exists := ctx.Config.ModeOverrides.VendorPreferences[CostSavingMode]; exists {
			for _, vendorName := range preferences {
				if vendor, exists := ctx.AvailableVendors[vendorName]; exists && vendor.IsAvailable(ctx.Context) {
					return vendor, nil
				}
			}
		}
	}

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
			// Check cost limits if specified
			if ctx.Config.ModeOverrides != nil && ctx.Config.ModeOverrides.MaxCostPerRequest > 0 {
				estimatedCost := c.estimateRequestCost(ctx.Request, costVendor.cost)
				if estimatedCost > ctx.Config.ModeOverrides.MaxCostPerRequest {
					continue // Skip if too expensive
				}
			}
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

// PreprocessContext applies cost-saving mode context preprocessing
func (c *CostSavingModeStrategy) PreprocessContext(ctx *ModeContext) error {
	// TODO: Implement cost-saving mode context preprocessing
	// - Compress context
	// - Remove redundant information
	// - Optimize for cost
	return nil
}

// OptimizeRequest applies cost-saving mode request optimizations
func (c *CostSavingModeStrategy) OptimizeRequest(ctx *ModeContext) error {
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

// AutoModeStrategy implements auto mode behavior
type AutoModeStrategy struct {
	*BaseModeStrategy
}

// NewAutoModeStrategy creates a new auto mode strategy
func NewAutoModeStrategy() *AutoModeStrategy {
	return &AutoModeStrategy{
		BaseModeStrategy: NewBaseModeStrategy(AutoMode, 5),
	}
}

// SelectVendor selects the best vendor for auto mode
func (a *AutoModeStrategy) SelectVendor(ctx *ModeContext) (LLMVendor, error) {
	// Check mode overrides first
	if ctx.Config.ModeOverrides != nil {
		if preferences, exists := ctx.Config.ModeOverrides.VendorPreferences[AutoMode]; exists {
			for _, vendorName := range preferences {
				if vendor, exists := ctx.AvailableVendors[vendorName]; exists && vendor.IsAvailable(ctx.Context) {
					return vendor, nil
				}
			}
		}
	}

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

// PreprocessContext applies auto mode context preprocessing
func (a *AutoModeStrategy) PreprocessContext(ctx *ModeContext) error {
	// TODO: Implement auto mode context preprocessing
	// - Analyze context complexity
	// - Apply appropriate preprocessing based on analysis
	// - Balance preprocessing cost vs benefit
	return nil
}

// OptimizeRequest applies auto mode request optimizations
func (a *AutoModeStrategy) OptimizeRequest(ctx *ModeContext) error {
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
