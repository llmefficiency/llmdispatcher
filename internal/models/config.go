package models

import (
	"time"
)

// Config holds the main dispatcher configuration
type Config struct {
	DefaultVendor  string        `json:"default_vendor"`
	FallbackVendor string        `json:"fallback_vendor,omitempty"`
	RetryPolicy    *RetryPolicy  `json:"retry_policy,omitempty"`
	RoutingRules   []RoutingRule `json:"routing_rules,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	EnableLogging  bool          `json:"enable_logging"`
	EnableMetrics  bool          `json:"enable_metrics"`
	// Advanced routing options
	CostOptimization *CostOptimization `json:"cost_optimization,omitempty"`
	LatencyOptimization *LatencyOptimization `json:"latency_optimization,omitempty"`
}

// CostOptimization defines cost-based routing configuration
type CostOptimization struct {
	Enabled     bool    `json:"enabled"`
	MaxCost     float64 `json:"max_cost_per_request"`
	PreferCheap bool    `json:"prefer_cheap"`
	// Cost per 1K tokens for each vendor
	VendorCosts map[string]float64 `json:"vendor_costs"`
}

// LatencyOptimization defines latency-based routing configuration
type LatencyOptimization struct {
	Enabled        bool          `json:"enabled"`
	MaxLatency     time.Duration `json:"max_latency"`
	PreferFast     bool          `json:"prefer_fast"`
	LatencyWeights map[string]float64 `json:"latency_weights"`
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
	// Advanced conditions
	UserID           string        `json:"user_id,omitempty"`
	RequestType      string        `json:"request_type,omitempty"`
	ContentLength    int           `json:"content_length,omitempty"`
}

// DispatcherStats holds statistics about the dispatcher
type DispatcherStats struct {
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	VendorStats        map[string]VendorStats `json:"vendor_stats"`
	AverageLatency     time.Duration          `json:"average_latency"`
	LastRequestTime    time.Time              `json:"last_request_time"`
	// Advanced metrics
	TotalCost          float64                `json:"total_cost"`
	AverageCost        float64                `json:"average_cost"`
	CostByVendor       map[string]float64     `json:"cost_by_vendor"`
}

// VendorStats holds statistics for a specific vendor
type VendorStats struct {
	Requests       int64         `json:"requests"`
	Successes      int64         `json:"successes"`
	Failures       int64         `json:"failures"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
	// Advanced metrics
	TotalCost      float64 `json:"total_cost"`
	AverageCost    float64 `json:"average_cost"`
	TokenUsage     int64   `json:"token_usage"`
}
