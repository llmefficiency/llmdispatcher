package llmdispatcher

import (
	"context"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

func TestVendorAdapter_Name(t *testing.T) {
	// Create a mock internal vendor
	mockInternalVendor := &MockInternalVendor{
		name: "test-vendor",
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	name := adapter.Name()
	if name != "test-vendor" {
		t.Errorf("Expected name 'test-vendor', got %s", name)
	}
}

func TestVendorAdapter_SendRequest(t *testing.T) {
	// Create a mock internal vendor
	mockInternalVendor := &MockInternalVendor{
		name: "test-vendor",
		response: &models.Response{
			Content: "Test response",
			Model:   "test-model",
			Vendor:  "test-vendor",
			Usage: models.Usage{
				PromptTokens:     10,
				CompletionTokens: 15,
				TotalTokens:      25,
			},
		},
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	response, err := adapter.SendRequest(context.Background(), request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
		return
	}
	if response.Content != "Test response" {
		t.Errorf("Expected content 'Test response', got %s", response.Content)
	}
	if response.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", response.Model)
	}
	if response.Vendor != "test-vendor" {
		t.Errorf("Expected vendor 'test-vendor', got %s", response.Vendor)
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

func TestVendorAdapter_SendRequest_Error(t *testing.T) {
	// Create a mock internal vendor that fails
	mockInternalVendor := &MockInternalVendor{
		name:       "test-vendor",
		shouldFail: true,
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := adapter.SendRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestVendorAdapter_GetCapabilities(t *testing.T) {
	// Create a mock internal vendor
	mockInternalVendor := &MockInternalVendor{
		name:              "test-vendor",
		supportsStreaming: true,
		capabilities: models.Capabilities{
			Models:            []string{"test-model"},
			SupportsStreaming: true,
			MaxTokens:         1000,
			MaxInputTokens:    4000,
		},
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	capabilities := adapter.GetCapabilities()

	if len(capabilities.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(capabilities.Models))
	}
	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming support to be true")
	}
	if capabilities.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens 1000, got %d", capabilities.MaxTokens)
	}
	if capabilities.MaxInputTokens != 4000 {
		t.Errorf("Expected MaxInputTokens 4000, got %d", capabilities.MaxInputTokens)
	}
}

func TestVendorAdapter_IsAvailable(t *testing.T) {
	// Create a mock internal vendor
	mockInternalVendor := &MockInternalVendor{
		name:      "test-vendor",
		available: true,
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	available := adapter.IsAvailable(context.Background())
	if !available {
		t.Error("Expected vendor to be available")
	}
}

func TestVendorAdapter_SendStreamingRequest(t *testing.T) {
	// Create a mock internal vendor
	mockInternalVendor := &MockInternalVendor{
		name:              "test-vendor",
		available:         true,
		supportsStreaming: true,
		capabilities: models.Capabilities{
			Models:            []string{"test-model"},
			SupportsStreaming: true,
		},
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	response, err := adapter.SendStreamingRequest(context.Background(), request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
		return
	}
	if response.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", response.Model)
	}
	if response.Vendor != "test-vendor" {
		t.Errorf("Expected vendor 'test-vendor', got %s", response.Vendor)
	}
}

func TestVendorAdapter_SendStreamingRequest_Error(t *testing.T) {
	// Create a mock internal vendor that fails
	mockInternalVendor := &MockInternalVendor{
		name:              "test-vendor",
		available:         true,
		supportsStreaming: true,
		shouldFail:        true,
	}

	adapter := &vendorAdapter{vendor: mockInternalVendor}

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := adapter.SendStreamingRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestNewOpenAIVendor(t *testing.T) {
	config := &VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com",
		Timeout: 30 * time.Second,
	}

	vendor := NewOpenAIVendor(config)
	if vendor == nil {
		t.Error("Expected vendor, got nil")
	}

	name := vendor.Name()
	if name != "openai" {
		t.Errorf("Expected name 'openai', got %s", name)
	}
}

func TestNewAnthropicVendor(t *testing.T) {
	config := &VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.anthropic.com",
		Timeout: 30 * time.Second,
	}

	vendor := NewAnthropicVendor(config)
	if vendor == nil {
		t.Error("Expected vendor, got nil")
	}

	name := vendor.Name()
	if name != "anthropic" {
		t.Errorf("Expected name 'anthropic', got %s", name)
	}
}

func TestNewGoogleVendor(t *testing.T) {
	config := &VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://generativelanguage.googleapis.com",
		Timeout: 30 * time.Second,
	}

	vendor := NewGoogleVendor(config)
	if vendor == nil {
		t.Error("Expected vendor, got nil")
	}

	name := vendor.Name()
	if name != "google" {
		t.Errorf("Expected name 'google', got %s", name)
	}
}

func TestNewAzureOpenAIVendor(t *testing.T) {
	config := &VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://your-resource.openai.azure.com",
		Timeout: 30 * time.Second,
	}

	vendor := NewAzureOpenAIVendor(config)
	if vendor == nil {
		t.Error("Expected vendor, got nil")
	}

	name := vendor.Name()
	if name != "azure-openai" {
		t.Errorf("Expected name 'azure-openai', got %s", name)
	}
}
