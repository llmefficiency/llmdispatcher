package vendors

import (
	"context"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

func TestNewAnthropic(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.VendorConfig
		wantNil bool
	}{
		{
			name: "with config",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.anthropic.com",
				Timeout: 30 * time.Second,
			},
			wantNil: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := NewAnthropic(tt.config)
			if (vendor == nil) != tt.wantNil {
				t.Errorf("NewAnthropic() = %v, want nil = %v", vendor, tt.wantNil)
			}
		})
	}
}

func TestAnthropicVendor_Name(t *testing.T) {
	vendor := NewAnthropic(nil)
	if vendor.Name() != "anthropic" {
		t.Errorf("Expected name 'anthropic', got %s", vendor.Name())
	}
}

func TestAnthropicVendor_GetCapabilities(t *testing.T) {
	vendor := NewAnthropic(nil)
	capabilities := vendor.GetCapabilities()

	expectedModels := []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming support to be true")
	}

	if capabilities.MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens 4096, got %d", capabilities.MaxTokens)
	}

	if capabilities.MaxInputTokens != 200000 {
		t.Errorf("Expected MaxInputTokens 200000, got %d", capabilities.MaxInputTokens)
	}

	// Check if all expected models are present
	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range capabilities.Models {
			if model == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found in capabilities", expectedModel)
		}
	}
}

func TestAnthropicVendor_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		config   *models.VendorConfig
		expected bool
	}{
		{
			name: "with API key",
			config: &models.VendorConfig{
				APIKey: "test-key",
			},
			expected: true,
		},
		{
			name: "without API key",
			config: &models.VendorConfig{
				APIKey: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := NewAnthropic(tt.config)
			result := vendor.IsAvailable(context.Background())
			if result != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnthropicVendor_ConvertRequest(t *testing.T) {
	vendor := NewAnthropic(nil)
	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	anthropicReq := vendor.convertRequest(req)

	if anthropicReq.Model != req.Model {
		t.Errorf("Expected model %s, got %s", req.Model, anthropicReq.Model)
	}

	if len(anthropicReq.Messages) != len(req.Messages) {
		t.Errorf("Expected %d messages, got %d", len(req.Messages), len(anthropicReq.Messages))
	}

	if anthropicReq.Temperature != req.Temperature {
		t.Errorf("Expected temperature %f, got %f", req.Temperature, anthropicReq.Temperature)
	}

	if anthropicReq.MaxTokens != req.MaxTokens {
		t.Errorf("Expected max tokens %d, got %d", req.MaxTokens, anthropicReq.MaxTokens)
	}
}

func TestAnthropicVendor_ConvertResponse(t *testing.T) {
	vendor := NewAnthropic(nil)
	anthropicResp := &anthropicResponse{
		ID:   "test-id",
		Type: "message",
		Role: "assistant",
		Content: []anthropicContent{
			{Type: "text", Text: "Hello! How can I help you?"},
		},
		Usage: anthropicUsage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	response := vendor.convertResponse(anthropicResp, "claude-3-sonnet-20240229")

	if response.Content != "Hello! How can I help you?" {
		t.Errorf("Expected content 'Hello! How can I help you?', got %s", response.Content)
	}

	if response.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected model 'claude-3-sonnet-20240229', got %s", response.Model)
	}

	if response.Vendor != "anthropic" {
		t.Errorf("Expected vendor 'anthropic', got %s", response.Vendor)
	}

	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", response.Usage.PromptTokens)
	}

	if response.Usage.CompletionTokens != 15 {
		t.Errorf("Expected completion tokens 15, got %d", response.Usage.CompletionTokens)
	}

	if response.Usage.TotalTokens != 25 {
		t.Errorf("Expected total tokens 25, got %d", response.Usage.TotalTokens)
	}
}
