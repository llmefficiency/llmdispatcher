package vendors

import (
	"context"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

func TestNewGoogle(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.VendorConfig
		wantNil bool
	}{
		{
			name: "with config",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://generativelanguage.googleapis.com",
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
			vendor := NewGoogle(tt.config)
			if (vendor == nil) != tt.wantNil {
				t.Errorf("NewGoogle() = %v, want nil = %v", vendor, tt.wantNil)
			}
		})
	}
}

func TestGoogleVendor_Name(t *testing.T) {
	vendor := NewGoogle(nil)
	if vendor.Name() != "google" {
		t.Errorf("Expected name 'google', got %s", vendor.Name())
	}
}

func TestGoogleVendor_GetCapabilities(t *testing.T) {
	vendor := NewGoogle(nil)
	capabilities := vendor.GetCapabilities()

	expectedModels := []string{
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-1.0-pro",
		"gemini-pro",
		"gemini-pro-vision",
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming support to be true")
	}

	if capabilities.MaxTokens != 8192 {
		t.Errorf("Expected MaxTokens 8192, got %d", capabilities.MaxTokens)
	}

	if capabilities.MaxInputTokens != 1000000 {
		t.Errorf("Expected MaxInputTokens 1000000, got %d", capabilities.MaxInputTokens)
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

func TestGoogleVendor_IsAvailable(t *testing.T) {
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
			vendor := NewGoogle(tt.config)
			result := vendor.IsAvailable(context.Background())
			if result != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGoogleVendor_ConvertRequest(t *testing.T) {
	vendor := NewGoogle(nil)
	req := &models.Request{
		Model: "gemini-1.5-pro",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	googleReq := vendor.convertRequest(req)

	if len(googleReq.Contents) != len(req.Messages) {
		t.Errorf("Expected %d contents, got %d", len(req.Messages), len(googleReq.Contents))
	}

	if googleReq.GenerationConfig.Temperature != req.Temperature {
		t.Errorf("Expected temperature %f, got %f", req.Temperature, googleReq.GenerationConfig.Temperature)
	}

	if googleReq.GenerationConfig.MaxOutputTokens != req.MaxTokens {
		t.Errorf("Expected max tokens %d, got %d", req.MaxTokens, googleReq.GenerationConfig.MaxOutputTokens)
	}

	if googleReq.GenerationConfig.TopP != req.TopP {
		t.Errorf("Expected top_p %f, got %f", req.TopP, googleReq.GenerationConfig.TopP)
	}
}

func TestGoogleVendor_ConvertResponse(t *testing.T) {
	vendor := NewGoogle(nil)
	googleResp := &googleResponse{
		Candidates: []googleCandidate{
			{
				Content: googleContent{
					Parts: []googlePart{
						{Text: "Hello! How can I help you?"},
					},
				},
			},
		},
		UsageMetadata: googleUsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 15,
		},
	}

	response := vendor.convertResponse(googleResp, "gemini-1.5-pro")

	if response.Content != "Hello! How can I help you?" {
		t.Errorf("Expected content 'Hello! How can I help you?', got %s", response.Content)
	}

	if response.Model != "gemini-1.5-pro" {
		t.Errorf("Expected model 'gemini-1.5-pro', got %s", response.Model)
	}

	if response.Vendor != "google" {
		t.Errorf("Expected vendor 'google', got %s", response.Vendor)
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

func TestGoogleVendor_ConvertResponse_EmptyCandidates(t *testing.T) {
	vendor := NewGoogle(nil)
	googleResp := &googleResponse{
		Candidates: []googleCandidate{},
		UsageMetadata: googleUsageMetadata{
			PromptTokenCount:     5,
			CandidatesTokenCount: 0,
		},
	}

	response := vendor.convertResponse(googleResp, "gemini-1.5-pro")

	if response.Content != "" {
		t.Errorf("Expected empty content, got %s", response.Content)
	}

	if response.Usage.TotalTokens != 5 {
		t.Errorf("Expected total tokens 5, got %d", response.Usage.TotalTokens)
	}
}

func TestGoogleVendor_ConvertResponse_EmptyParts(t *testing.T) {
	vendor := NewGoogle(nil)
	googleResp := &googleResponse{
		Candidates: []googleCandidate{
			{
				Content: googleContent{
					Parts: []googlePart{},
				},
			},
		},
		UsageMetadata: googleUsageMetadata{
			PromptTokenCount:     5,
			CandidatesTokenCount: 0,
		},
	}

	response := vendor.convertResponse(googleResp, "gemini-1.5-pro")

	if response.Content != "" {
		t.Errorf("Expected empty content, got %s", response.Content)
	}
} 