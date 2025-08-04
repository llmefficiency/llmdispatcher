package models

import (
	"context"
	"fmt"
	"time"
)

// Config holds the main dispatcher configuration
// This is identical to pkg/llmdispatcher/types.go Config to avoid import cycles
type Config struct {
	DefaultVendor   string          `json:"default_vendor"`
	FallbackVendor  string          `json:"fallback_vendor,omitempty"`
	RetryPolicy     *RetryPolicy    `json:"retry_policy,omitempty"`
	Timeout         time.Duration   `json:"timeout,omitempty"`
	EnableLogging   bool            `json:"enable_logging"`
	EnableMetrics   bool            `json:"enable_metrics"`
	RoutingStrategy RoutingStrategy `json:"routing_strategy,omitempty"`
	// Advanced routing options
	CostOptimization    *CostOptimization    `json:"cost_optimization,omitempty"`
	LatencyOptimization *LatencyOptimization `json:"latency_optimization,omitempty"`
}

// RoutingStrategy defines how requests should be routed to vendors
type RoutingStrategy interface {
	// SelectVendor selects the next vendor to try based on the request and available vendors
	SelectVendor(ctx context.Context, req *Request, vendors map[string]LLMVendor) (LLMVendor, error)

	// Name returns the name of the routing strategy
	Name() string
}

// CascadingFailureStrategy implements a simple fallback strategy
// It tries vendors in order until one succeeds
type CascadingFailureStrategy struct {
	VendorOrder []string `json:"vendor_order"`
}

// NewCascadingFailureStrategy creates a new cascading failure strategy
func NewCascadingFailureStrategy(vendorOrder []string) *CascadingFailureStrategy {
	return &CascadingFailureStrategy{
		VendorOrder: vendorOrder,
	}
}

// Name returns the strategy name
func (c *CascadingFailureStrategy) Name() string {
	return "cascading_failure"
}

// SelectVendor selects the first available vendor in the order
func (c *CascadingFailureStrategy) SelectVendor(ctx context.Context, req *Request, vendors map[string]LLMVendor) (LLMVendor, error) {
	for _, vendorName := range c.VendorOrder {
		if vendor, exists := vendors[vendorName]; exists {
			if vendor.IsAvailable(ctx) {
				return vendor, nil
			}
		}
	}
	return nil, fmt.Errorf("no available vendors in cascading strategy")
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
	Enabled        bool               `json:"enabled"`
	MaxLatency     time.Duration      `json:"max_latency"`
	PreferFast     bool               `json:"prefer_fast"`
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
