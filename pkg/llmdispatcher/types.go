package llmdispatcher

import (
	"context"
	"fmt"
	"sync"
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
}

// RoutingStrategy defines how requests should be routed to vendors
type RoutingStrategy interface {
	// SelectVendor selects the next vendor to try based on the request and available vendors
	SelectVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error)

	// Name returns the name of the routing strategy
	Name() string
}

// ModeStrategy implements the routing strategy for each mode
type ModeStrategy struct {
	mode    Mode
	config  *Config
	vendors map[string]Vendor
}

// NewModeStrategy creates a new mode-based routing strategy
func NewModeStrategy(mode Mode, config *Config, vendors map[string]Vendor) *ModeStrategy {
	return &ModeStrategy{
		mode:    mode,
		config:  config,
		vendors: vendors,
	}
}

// Name returns the strategy name
func (m *ModeStrategy) Name() string {
	return string(m.mode)
}

// SelectVendor selects the best vendor based on the current mode
func (m *ModeStrategy) SelectVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
	switch m.mode {
	case FastMode:
		return m.selectFastVendor(ctx, req, vendors)
	case SophisticatedMode:
		return m.selectSophisticatedVendor(ctx, req, vendors)
	case CostSavingMode:
		return m.selectCostSavingVendor(ctx, req, vendors)
	case AutoMode:
		return m.selectAutoVendor(ctx, req, vendors)
	default:
		return nil, fmt.Errorf("unknown mode: %s", m.mode)
	}
}

// selectFastVendor prioritizes vendors with lowest latency and fastest models
func (m *ModeStrategy) selectFastVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
	// Fast mode intelligence: prioritize vendors and models known for speed
	fastVendors := []struct {
		name     string
		priority int
		models   []string
	}{
		{"local", 1, []string{"llama2:7b", "mistral:7b", "gemma:2b"}},   // Local is fastest
		{"anthropic", 2, []string{"claude-3-haiku", "claude-3-sonnet"}}, // Haiku is very fast
		{"openai", 3, []string{"gpt-3.5-turbo", "gpt-4o-mini"}},         // GPT-3.5 is fast
		{"google", 4, []string{"gemini-1.5-flash", "gemini-1.5-pro"}},   // Flash is fast
		{"azure", 5, []string{"gpt-35-turbo", "gpt-4"}},                 // Azure OpenAI
	}

	for _, fastVendor := range fastVendors {
		if vendor, exists := vendors[fastVendor.name]; exists && vendor.IsAvailable(ctx) {
			// Optimize the request for speed
			m.optimizeRequestForSpeed(req)
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range vendors {
		if vendor.IsAvailable(ctx) {
			m.optimizeRequestForSpeed(req)
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for fast mode")
}

// selectSophisticatedVendor prioritizes the most capable models and vendors
func (m *ModeStrategy) selectSophisticatedVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
	// Sophisticated mode intelligence: prioritize vendors with the most capable models
	sophisticatedVendors := []struct {
		name     string
		priority int
		models   []string
	}{
		{"anthropic", 1, []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}}, // Claude is most sophisticated
		{"openai", 2, []string{"gpt-4o", "gpt-4-turbo", "gpt-4"}},                        // GPT-4 is very capable
		{"google", 3, []string{"gemini-1.5-pro", "gemini-1.5-flash"}},                    // Gemini Pro is capable
		{"azure", 4, []string{"gpt-4", "gpt-35-turbo"}},                                  // Azure OpenAI
		{"local", 5, []string{"llama2:70b", "llama2:13b", "mistral:7b"}},                 // Large local models
	}

	for _, sophisticatedVendor := range sophisticatedVendors {
		if vendor, exists := vendors[sophisticatedVendor.name]; exists && vendor.IsAvailable(ctx) {
			// Optimize the request for sophistication
			m.optimizeRequestForSophistication(req)
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range vendors {
		if vendor.IsAvailable(ctx) {
			m.optimizeRequestForSophistication(req)
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for sophisticated mode")
}

// selectCostSavingVendor prioritizes the cheapest options
func (m *ModeStrategy) selectCostSavingVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
	// Cost-saving mode intelligence: prioritize cheapest vendors and models
	costSavingVendors := []struct {
		name     string
		priority int
		models   []string
		cost     float64 // Cost per 1K tokens (approximate)
	}{
		{"local", 1, []string{"llama2:7b", "mistral:7b", "gemma:2b"}, 0.0001},  // Local is cheapest
		{"google", 2, []string{"gemini-1.5-flash", "gemini-1.5-pro"}, 0.0005},  // Google is cheap
		{"azure", 3, []string{"gpt-35-turbo", "gpt-4"}, 0.002},                 // Azure is reasonable
		{"openai", 4, []string{"gpt-3.5-turbo", "gpt-4o-mini"}, 0.002},         // OpenAI is moderate
		{"anthropic", 5, []string{"claude-3-haiku", "claude-3-sonnet"}, 0.003}, // Anthropic is pricier
	}

	for _, costVendor := range costSavingVendors {
		if vendor, exists := vendors[costVendor.name]; exists && vendor.IsAvailable(ctx) {
			// Optimize the request for cost saving
			m.optimizeRequestForCostSaving(req)
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range vendors {
		if vendor.IsAvailable(ctx) {
			m.optimizeRequestForCostSaving(req)
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for cost-saving mode")
}

// selectAutoVendor balances all factors intelligently
func (m *ModeStrategy) selectAutoVendor(ctx context.Context, req *Request, vendors map[string]Vendor) (Vendor, error) {
	// Auto mode intelligence: balance speed, cost, and capability
	balancedVendors := []struct {
		name     string
		priority int
		models   []string
		speed    int // 1-5 scale
		cost     int // 1-5 scale (1=cheap, 5=expensive)
		quality  int // 1-5 scale
	}{
		{"local", 1, []string{"llama2:13b", "mistral:7b"}, 5, 1, 3},              // Fast, cheap, decent quality
		{"anthropic", 2, []string{"claude-3-sonnet", "claude-3-haiku"}, 4, 4, 5}, // Good speed, high quality
		{"openai", 3, []string{"gpt-4o-mini", "gpt-3.5-turbo"}, 4, 3, 4},         // Good balance
		{"google", 4, []string{"gemini-1.5-flash", "gemini-1.5-pro"}, 3, 2, 4},   // Cheap, good quality
		{"azure", 5, []string{"gpt-35-turbo", "gpt-4"}, 3, 3, 4},                 // Moderate across all
	}

	for _, balancedVendor := range balancedVendors {
		if vendor, exists := vendors[balancedVendor.name]; exists && vendor.IsAvailable(ctx) {
			// Optimize the request for balance
			m.optimizeRequestForBalance(req)
			return vendor, nil
		}
	}

	// Fallback to any available vendor
	for _, vendor := range vendors {
		if vendor.IsAvailable(ctx) {
			m.optimizeRequestForBalance(req)
			return vendor, nil
		}
	}

	return nil, fmt.Errorf("no available vendors for auto mode")
}

// optimizeRequestForSpeed tunes request parameters for maximum speed
func (m *ModeStrategy) optimizeRequestForSpeed(req *Request) {
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

	// Prefer smaller, faster models if not specified
	// Let vendor choose the fastest available model
}

// optimizeRequestForSophistication tunes request parameters for maximum quality
func (m *ModeStrategy) optimizeRequestForSophistication(req *Request) {
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

	// Prefer larger, more capable models if not specified
	// Let vendor choose the most capable available model
}

// optimizeRequestForCostSaving tunes request parameters for minimum cost
func (m *ModeStrategy) optimizeRequestForCostSaving(req *Request) {
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

	// Prefer smaller, cheaper models if not specified
	// Let vendor choose the cheapest available model
}

// optimizeRequestForBalance tunes request parameters for balanced performance
func (m *ModeStrategy) optimizeRequestForBalance(req *Request) {
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

	// Let vendor choose a balanced model
	// Let vendor choose a balanced available model
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
