package models

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// LLMVendor defines the interface that all LLM vendors must implement
type LLMVendor interface {
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
	Mode        string    `json:"mode,omitempty"`
}

// Validate checks if the request is valid
func (r *Request) Validate() error {
	// Debug logging
	fmt.Printf("DEBUG: Validate called with Model='%s', Mode='%s'\n", r.Model, r.Mode)

	// For mode-based requests, model is optional as it will be auto-selected
	if r.Model == "" && r.Mode == "" {
		return fmt.Errorf("%w: either model or mode must be specified", ErrInvalidRequest)
	}

	if len(r.Messages) == 0 {
		return fmt.Errorf("%w: at least one message is required", ErrInvalidRequest)
	}

	// Validate temperature range
	if r.Temperature < 0 || r.Temperature > 2 {
		return fmt.Errorf("%w: temperature must be between 0 and 2", ErrInvalidRequest)
	}

	// Validate top_p range
	if r.TopP < 0 || r.TopP > 1 {
		return fmt.Errorf("%w: top_p must be between 0 and 1", ErrInvalidRequest)
	}

	// Validate max tokens
	if r.MaxTokens < 0 {
		return fmt.Errorf("%w: max_tokens cannot be negative", ErrInvalidRequest)
	}

	// Validate mode if specified
	if r.Mode != "" {
		validModes := map[string]bool{
			"auto":          true,
			"fast":          true,
			"sophisticated": true,
			"cost_saving":   true,
		}
		if !validModes[r.Mode] {
			return fmt.Errorf("%w: invalid mode: %s", ErrInvalidRequest, r.Mode)
		}
	}

	// Validate messages
	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("%w: message %d: %v", ErrInvalidRequest, i, err)
		}
	}

	return nil
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Validate checks if the message is valid
func (m *Message) Validate() error {
	if m.Role == "" {
		return errors.New("role cannot be empty")
	}

	if m.Content == "" {
		return errors.New("content cannot be empty")
	}

	// Validate role values
	validRoles := map[string]bool{
		"system":    true,
		"user":      true,
		"assistant": true,
	}

	if !validRoles[m.Role] {
		return fmt.Errorf("invalid role: %s", m.Role)
	}

	return nil
}

// Response represents a standardized LLM response
type Response struct {
	Content       string    `json:"content"`
	Usage         Usage     `json:"usage"`
	Model         string    `json:"model"`
	Vendor        string    `json:"vendor"`
	FinishReason  string    `json:"finish_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	EstimatedCost float64   `json:"estimated_cost,omitempty"`
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
func (sr *StreamingResponse) Close() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.closed {
		return
	}
	sr.closed = true
	close(sr.ContentChan)
	close(sr.DoneChan)
	close(sr.ErrorChan)
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

// VendorConfig holds configuration for a specific vendor
type VendorConfig struct {
	APIKey    string            `json:"api_key"`
	BaseURL   string            `json:"base_url,omitempty"`
	Timeout   time.Duration     `json:"timeout,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	RateLimit RateLimit         `json:"rate_limit,omitempty"`
}

// Validate checks if the vendor config is valid
func (vc *VendorConfig) Validate() error {
	if vc.APIKey == "" {
		return fmt.Errorf("%w: API key cannot be empty", ErrInvalidConfig)
	}

	if vc.Timeout < 0 {
		return fmt.Errorf("%w: timeout cannot be negative", ErrInvalidConfig)
	}

	return nil
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute"`
}
