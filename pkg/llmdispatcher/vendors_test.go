package llmdispatcher

import (
	"context"
	"testing"
	"time"
)

func TestNewOpenAIVendor(t *testing.T) {
	tests := []struct {
		name   string
		config *VendorConfig
	}{
		{
			name: "with config",
			config: &VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
				Timeout: 30 * time.Second,
				Headers: map[string]string{
					"User-Agent": "test-agent",
				},
			},
		},
		{
			name:   "nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := NewOpenAIVendor(tt.config)
			if vendor == nil {
				t.Fatal("NewOpenAIVendor() returned nil")
			}

			// Test that the vendor implements the Vendor interface
			if vendor.Name() == "" {
				t.Error("vendor name should not be empty")
			}

			// Test capabilities
			capabilities := vendor.GetCapabilities()
			if len(capabilities.Models) == 0 {
				t.Error("capabilities should have at least one model")
			}

			// Test availability
			available := vendor.IsAvailable(context.Background())
			// Availability depends on config, so we just test that it doesn't panic
			_ = available
		})
	}
}

func TestVendorIntegration(t *testing.T) {
	// Test vendor integration through the public API
	dispatcher := New()

	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &Response{
			Content: "Test response",
			Model:   "test-model",
			Vendor:  "test-vendor",
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
			CreatedAt: time.Now(),
		},
		capabilities: Capabilities{
			Models:            []string{"test-model"},
			SupportsStreaming: true,
			MaxTokens:         1000,
			MaxInputTokens:    10000,
		},
		available: true,
	}

	err := dispatcher.RegisterVendor(mockVendor)
	if err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	ctx := context.Background()
	response, err := dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Content != "Test response" {
		t.Errorf("Expected content 'Test response', got '%s'", response.Content)
	}

	if response.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", response.Model)
	}

	if response.Vendor != "test-vendor" {
		t.Errorf("Expected vendor 'test-vendor', got '%s'", response.Vendor)
	}

	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected 10 prompt tokens, got %d", response.Usage.PromptTokens)
	}

	if response.Usage.CompletionTokens != 5 {
		t.Errorf("Expected 5 completion tokens, got %d", response.Usage.CompletionTokens)
	}

	if response.Usage.TotalTokens != 15 {
		t.Errorf("Expected 15 total tokens, got %d", response.Usage.TotalTokens)
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

// Helper validation function for testing
func validateVendorConfig(config *VendorConfig) error {
	if config == nil {
		return &MockError{message: "config cannot be nil"}
	}
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
