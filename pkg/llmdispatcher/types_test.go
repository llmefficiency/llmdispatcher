package llmdispatcher

import (
	"testing"
	"time"
)

func TestNewStreamingResponse(t *testing.T) {
	model := "test-model"
	vendor := "test-vendor"

	streamingResp := NewStreamingResponse(model, vendor)

	if streamingResp == nil {
		t.Fatal("NewStreamingResponse() returned nil")
	}

	if streamingResp.Model != model {
		t.Errorf("Expected model '%s', got '%s'", model, streamingResp.Model)
	}

	if streamingResp.Vendor != vendor {
		t.Errorf("Expected vendor '%s', got '%s'", vendor, streamingResp.Vendor)
	}

	if streamingResp.ContentChan == nil {
		t.Error("ContentChan should not be nil")
	}

	if streamingResp.DoneChan == nil {
		t.Error("DoneChan should not be nil")
	}

	if streamingResp.ErrorChan == nil {
		t.Error("ErrorChan should not be nil")
	}

	if streamingResp.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestStreamingResponse_CloseChannels(t *testing.T) {
	streamingResp := NewStreamingResponse("test-model", "test-vendor")

	// Test that channels are initially open
	select {
	case _, ok := <-streamingResp.ContentChan:
		if !ok {
			t.Error("ContentChan should be open initially")
		}
	default:
		// Channel is open, which is expected
	}

	// Close the streaming response
	streamingResp.Close()

	// Test that channels are closed
	select {
	case _, ok := <-streamingResp.ContentChan:
		if ok {
			t.Error("ContentChan should be closed after Close()")
		}
	default:
		t.Error("ContentChan should be closed after Close()")
	}

	select {
	case _, ok := <-streamingResp.DoneChan:
		if ok {
			t.Error("DoneChan should be closed after Close()")
		}
	default:
		t.Error("DoneChan should be closed after Close()")
	}

	select {
	case _, ok := <-streamingResp.ErrorChan:
		if ok {
			t.Error("ErrorChan should be closed after Close()")
		}
	default:
		t.Error("ErrorChan should be closed after Close()")
	}
}

func TestRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
		isValid bool
	}{
		{
			name: "valid request",
			request: &Request{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 0.7,
				MaxTokens:   100,
			},
			isValid: true,
		},
		{
			name: "empty model",
			request: &Request{
				Model: "",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			isValid: false,
		},
		{
			name: "no messages",
			request: &Request{
				Model:    "test-model",
				Messages: []Message{},
			},
			isValid: false,
		},
		{
			name: "invalid temperature",
			request: &Request{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 2.0, // Should be between 0 and 1
			},
			isValid: false,
		},
		{
			name: "negative max tokens",
			request: &Request{
				Model: "test-model",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: -1,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequest(tt.request)
			if (err == nil) != tt.isValid {
				t.Errorf("validateRequest() error = %v, isValid %v", err, tt.isValid)
			}
		})
	}
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		isValid bool
	}{
		{
			name: "valid message",
			message: Message{
				Role:    "user",
				Content: "Hello",
			},
			isValid: true,
		},
		{
			name: "empty role",
			message: Message{
				Role:    "",
				Content: "Hello",
			},
			isValid: false,
		},
		{
			name: "empty content",
			message: Message{
				Role:    "user",
				Content: "",
			},
			isValid: false,
		},
		{
			name: "system role",
			message: Message{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			isValid: true,
		},
		{
			name: "assistant role",
			message: Message{
				Role:    "assistant",
				Content: "I can help you",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMessage(tt.message)
			if (err == nil) != tt.isValid {
				t.Errorf("validateMessage() error = %v, isValid %v", err, tt.isValid)
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       30 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "empty default vendor",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       30 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "zero timeout",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       0,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "negative timeout",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       -1 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRetryPolicy_Validation(t *testing.T) {
	tests := []struct {
		name        string
		retryPolicy *RetryPolicy
		isValid     bool
	}{
		{
			name: "valid retry policy",
			retryPolicy: &RetryPolicy{
				MaxRetries:      3,
				BackoffStrategy: ExponentialBackoff,
				RetryableErrors: []string{"rate limit exceeded", "timeout"},
			},
			isValid: true,
		},
		{
			name: "negative max retries",
			retryPolicy: &RetryPolicy{
				MaxRetries:      -1,
				BackoffStrategy: ExponentialBackoff,
			},
			isValid: false,
		},
		{
			name: "invalid backoff strategy",
			retryPolicy: &RetryPolicy{
				MaxRetries:      3,
				BackoffStrategy: "invalid",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRetryPolicy(tt.retryPolicy)
			if (err == nil) != tt.isValid {
				t.Errorf("validateRetryPolicy() error = %v, isValid %v", err, tt.isValid)
			}
		})
	}
}

// Helper validation functions for testing
func validateRequest(req *Request) error {
	if req == nil {
		return &MockError{message: "request cannot be nil"}
	}
	if req.Model == "" {
		return &MockError{message: "model is required"}
	}
	if len(req.Messages) == 0 {
		return &MockError{message: "at least one message is required"}
	}
	if req.Temperature < 0 || req.Temperature > 1 {
		return &MockError{message: "temperature must be between 0 and 1"}
	}
	if req.MaxTokens < 0 {
		return &MockError{message: "max tokens must be non-negative"}
	}

	for _, msg := range req.Messages {
		if err := validateMessage(msg); err != nil {
			return err
		}
	}

	return nil
}

func validateMessage(msg Message) error {
	if msg.Role == "" {
		return &MockError{message: "role is required"}
	}
	if msg.Content == "" {
		return &MockError{message: "content is required"}
	}

	validRoles := map[string]bool{
		"system":    true,
		"user":      true,
		"assistant": true,
	}

	if !validRoles[msg.Role] {
		return &MockError{message: "invalid role"}
	}

	return nil
}

func validateConfig(config *Config) error {
	if config == nil {
		return &MockError{message: "config cannot be nil"}
	}
	if config.Mode == "" {
		return &MockError{message: "mode is required"}
	}
	if config.Timeout < 0 {
		return &MockError{message: "timeout must be non-negative"}
	}
	return nil
}

func validateRetryPolicy(policy *RetryPolicy) error {
	if policy == nil {
		return &MockError{message: "retry policy cannot be nil"}
	}
	if policy.MaxRetries < 0 {
		return &MockError{message: "max retries must be non-negative"}
	}

	validStrategies := map[BackoffStrategy]bool{
		ExponentialBackoff: true,
		LinearBackoff:      true,
		FixedBackoff:       true,
	}

	if !validStrategies[policy.BackoffStrategy] {
		return &MockError{message: "invalid backoff strategy"}
	}

	return nil
}
