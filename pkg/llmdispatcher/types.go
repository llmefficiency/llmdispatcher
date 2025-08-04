package llmdispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Vendor defines the interface that all LLM vendors must implement
type Vendor interface {
	// Name returns the vendor name (e.g., "openai", "anthropic")
	Name() string

	// SendRequest sends a request to the vendor and returns the response
	SendRequest(ctx context.Context, req *Request) (*Response, error)

	// SendStreamingRequest sends a streaming request to the vendor
	SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error)

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

// StreamingResponse represents a streaming LLM response
type StreamingResponse struct {
	ContentChan chan string `json:"-"`
	DoneChan    chan bool   `json:"-"`
	ErrorChan   chan error  `json:"-"`
	Usage       Usage       `json:"usage"`
	Model       string      `json:"model"`
	Vendor      string      `json:"vendor"`
	CreatedAt   time.Time   `json:"created_at"`
	closed      bool        `json:"-"`
	mu          sync.Mutex  `json:"-"`
}

// NewStreamingResponse creates a new streaming response
func NewStreamingResponse(model, vendor string) *StreamingResponse {
	return &StreamingResponse{
		ContentChan: make(chan string, 100),
		DoneChan:    make(chan bool, 1),
		ErrorChan:   make(chan error, 1),
		Model:       model,
		Vendor:      vendor,
		CreatedAt:   time.Now(),
	}
}

// Close closes all channels in the streaming response
func (s *StreamingResponse) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true
	close(s.ContentChan)
	close(s.DoneChan)
	close(s.ErrorChan)
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
	SelectVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error)

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
func (c *CascadingFailureStrategy) SelectVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
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
