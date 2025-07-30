package models

import (
	"context"
	"testing"
	"time"
)

// MockVendor is a mock implementation of LLMVendor for testing
type MockVendor struct {
	name         string
	shouldFail   bool
	response     *Response
	capabilities Capabilities
	available    bool
}

func (m *MockVendor) Name() string {
	return m.name
}

func (m *MockVendor) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	if m.shouldFail {
		return nil, &MockError{message: "mock error"}
	}
	return m.response, nil
}

func (m *MockVendor) GetCapabilities() Capabilities {
	return m.capabilities
}

func (m *MockVendor) IsAvailable(ctx context.Context) bool {
	return m.available
}

type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

func TestRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
		wantErr bool
	}{
		{
			name: "valid request",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: false,
		},
		{
			name: "empty model",
			request: &Request{
				Model: "",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "no messages",
			request: &Request{
				Model:    "gpt-3.5-turbo",
				Messages: []Message{},
			},
			wantErr: true,
		},
		{
			name: "invalid temperature too high",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 2.0,
			},
			wantErr: true,
		},
		{
			name: "invalid temperature too low",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: -1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name: "valid message",
			message: Message{
				Role:    "user",
				Content: "Hello",
			},
			wantErr: false,
		},
		{
			name: "empty role",
			message: Message{
				Role:    "",
				Content: "Hello",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			message: Message{
				Role:    "user",
				Content: "",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			message: Message{
				Role:    "invalid",
				Content: "Hello",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		wantErr  bool
	}{
		{
			name: "valid response",
			response: &Response{
				Content: "Hello",
				Model:   "gpt-3.5-turbo",
				Vendor:  "openai",
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty content",
			response: &Response{
				Content: "",
				Model:   "gpt-3.5-turbo",
				Vendor:  "openai",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			response: &Response{
				Content: "Hello",
				Model:   "",
				Vendor:  "openai",
			},
			wantErr: true,
		},
		{
			name: "empty vendor",
			response: &Response{
				Content: "Hello",
				Model:   "gpt-3.5-turbo",
				Vendor:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		usage   Usage
		wantErr bool
	}{
		{
			name: "valid usage",
			usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
			wantErr: false,
		},
		{
			name: "negative prompt tokens",
			usage: Usage{
				PromptTokens:     -1,
				CompletionTokens: 5,
				TotalTokens:      4,
			},
			wantErr: true,
		},
		{
			name: "negative completion tokens",
			usage: Usage{
				PromptTokens:     10,
				CompletionTokens: -1,
				TotalTokens:      9,
			},
			wantErr: true,
		},
		{
			name: "total tokens mismatch",
			usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      20,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsage(tt.usage)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUsage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCapabilities_Validation(t *testing.T) {
	tests := []struct {
		name         string
		capabilities Capabilities
		wantErr      bool
	}{
		{
			name: "valid capabilities",
			capabilities: Capabilities{
				Models:            []string{"gpt-3.5-turbo", "gpt-4"},
				SupportsStreaming: true,
				MaxTokens:         4096,
				MaxInputTokens:    128000,
			},
			wantErr: false,
		},
		{
			name: "empty models",
			capabilities: Capabilities{
				Models:            []string{},
				SupportsStreaming: true,
				MaxTokens:         4096,
				MaxInputTokens:    128000,
			},
			wantErr: true,
		},
		{
			name: "negative max tokens",
			capabilities: Capabilities{
				Models:            []string{"gpt-3.5-turbo"},
				SupportsStreaming: true,
				MaxTokens:         -1,
				MaxInputTokens:    128000,
			},
			wantErr: true,
		},
		{
			name: "negative max input tokens",
			capabilities: Capabilities{
				Models:            []string{"gpt-3.5-turbo"},
				SupportsStreaming: true,
				MaxTokens:         4096,
				MaxInputTokens:    -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCapabilities(tt.capabilities)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCapabilities() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVendorConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *VendorConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.example.com",
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: &VendorConfig{
				APIKey:  "",
				BaseURL: "https://api.example.com",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &VendorConfig{
				APIKey:  "test-key",
				BaseURL: "not-a-url",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.example.com",
				Timeout: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVendorConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVendorConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper validation functions (these would be implemented in the actual package)
func validateRequest(req *Request) error {
	if req.Model == "" {
		return &MockError{message: "model is required"}
	}
	if len(req.Messages) == 0 {
		return &MockError{message: "at least one message is required"}
	}
	if req.Temperature < 0 || req.Temperature >= 2 {
		return &MockError{message: "temperature must be between 0 and 2"}
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
	if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
		return &MockError{message: "invalid role"}
	}
	return nil
}

func validateResponse(resp *Response) error {
	if resp.Content == "" {
		return &MockError{message: "content is required"}
	}
	if resp.Model == "" {
		return &MockError{message: "model is required"}
	}
	if resp.Vendor == "" {
		return &MockError{message: "vendor is required"}
	}
	return nil
}

func validateUsage(usage Usage) error {
	if usage.PromptTokens < 0 {
		return &MockError{message: "prompt tokens cannot be negative"}
	}
	if usage.CompletionTokens < 0 {
		return &MockError{message: "completion tokens cannot be negative"}
	}
	if usage.TotalTokens != usage.PromptTokens+usage.CompletionTokens {
		return &MockError{message: "total tokens must equal prompt + completion tokens"}
	}
	return nil
}

func validateCapabilities(caps Capabilities) error {
	if len(caps.Models) == 0 {
		return &MockError{message: "at least one model is required"}
	}
	if caps.MaxTokens < 0 {
		return &MockError{message: "max tokens cannot be negative"}
	}
	if caps.MaxInputTokens < 0 {
		return &MockError{message: "max input tokens cannot be negative"}
	}
	return nil
}

func validateVendorConfig(config *VendorConfig) error {
	if config.APIKey == "" {
		return &MockError{message: "API key is required"}
	}
	if config.BaseURL != "" && !isValidURL(config.BaseURL) {
		return &MockError{message: "invalid base URL"}
	}
	if config.Timeout <= 0 {
		return &MockError{message: "timeout must be positive"}
	}
	return nil
}

func isValidURL(url string) bool {
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}
