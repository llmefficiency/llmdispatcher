package llmdispatcher

import (
	"context"
	"time"
)

// Vendor defines the interface that all LLM vendors must implement
type Vendor interface {
	// Name returns the vendor name (e.g., "openai", "anthropic")
	Name() string

	// SendRequest sends a request to the vendor and returns the response
	SendRequest(ctx context.Context, req *Request) (*Response, error)

	// GetCapabilities returns the vendor's capabilities
	GetCapabilities() Capabilities

	// IsAvailable checks if the vendor is currently available
	IsAvailable(ctx context.Context) bool
}

// Request represents a standardized LLM request
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	User        string    `json:"user,omitempty"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents a standardized LLM response
type Response struct {
	Content      string    `json:"content"`
	Usage        Usage     `json:"usage"`
	Model        string    `json:"model"`
	Vendor       string    `json:"vendor"`
	FinishReason string    `json:"finish_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Capabilities represents what a vendor can do
type Capabilities struct {
	Models            []string `json:"models"`
	SupportsStreaming bool     `json:"supports_streaming"`
	MaxTokens         int      `json:"max_tokens"`
	MaxInputTokens    int      `json:"max_input_tokens"`
}

// Config holds the main dispatcher configuration
type Config struct {
	DefaultVendor  string        `json:"default_vendor"`
	FallbackVendor string        `json:"fallback_vendor,omitempty"`
	RetryPolicy    *RetryPolicy  `json:"retry_policy,omitempty"`
	RoutingRules   []RoutingRule `json:"routing_rules,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	EnableLogging  bool          `json:"enable_logging"`
	EnableMetrics  bool          `json:"enable_metrics"`
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

// RoutingRule defines how requests should be routed to vendors
type RoutingRule struct {
	Condition RoutingCondition `json:"condition"`
	Vendor    string           `json:"vendor"`
	Priority  int              `json:"priority"`
	Enabled   bool             `json:"enabled"`
}

// RoutingCondition defines when a routing rule should be applied
type RoutingCondition struct {
	ModelPattern     string        `json:"model_pattern,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"`
	CostThreshold    float64       `json:"cost_threshold,omitempty"`
	LatencyThreshold time.Duration `json:"latency_threshold,omitempty"`
}

// Stats holds statistics about the dispatcher
type Stats struct {
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	VendorStats        map[string]VendorStats `json:"vendor_stats"`
	AverageLatency     time.Duration          `json:"average_latency"`
	LastRequestTime    time.Time              `json:"last_request_time"`
}

// VendorStats holds statistics for a specific vendor
type VendorStats struct {
	Requests       int64         `json:"requests"`
	Successes      int64         `json:"successes"`
	Failures       int64         `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
}

// VendorConfig holds configuration for a specific vendor
type VendorConfig struct {
	APIKey    string            `json:"api_key"`
	BaseURL   string            `json:"base_url,omitempty"`
	Timeout   time.Duration     `json:"timeout,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	RateLimit RateLimit         `json:"rate_limit,omitempty"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`
}
