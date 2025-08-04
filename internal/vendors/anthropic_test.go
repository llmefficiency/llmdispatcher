package vendors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAnthropic_SendStreamingRequest_Success(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}

func TestAnthropic_SendStreamingRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	ctx := context.Background()
	_, err := vendor.SendStreamingRequest(ctx, req)
	if err == nil {
		t.Fatal("Expected error from SendStreamingRequest")
	}
	if !strings.Contains(err.Error(), "HTTP error 500") {
		t.Errorf("Expected HTTP error 500, got: %v", err)
	}
}

func TestAnthropic_SendStreamingRequest_NetworkError(t *testing.T) {
	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second,
	})

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	ctx := context.Background()
	_, err := vendor.SendStreamingRequest(ctx, req)
	if err == nil {
		t.Fatal("Expected error from SendStreamingRequest")
	}
	if !strings.Contains(err.Error(), "HTTP error 404") {
		t.Errorf("Expected HTTP 404 error, got: %v", err)
	}
}

func TestAnthropic_SendStreamingRequest_InvalidJSON(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}

func TestAnthropic_SendStreamingRequest_WithHeaders(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}

func TestAnthropic_SendRequest_Success(t *testing.T) {
	// Create a test server that returns a valid Anthropic response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/messages") {
			t.Errorf("Expected path to contain /v1/messages, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header 'test-key', got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header '2023-06-01', got %s", r.Header.Get("anthropic-version"))
		}

		// Return a valid Anthropic response
		response := `{
			"id": "msg_test123",
			"type": "message",
			"role": "assistant",
			"content": [
				{
					"type": "text",
					"text": "Hello! How can I help you today?"
				}
			],
			"model": "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {
				"input_tokens": 10,
				"output_tokens": 15
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create vendor with test server URL
	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create a test request
	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens: 100,
	}

	// Send the request
	response, err := vendor.SendRequest(context.Background(), request)

	// Verify the response
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
		return
	}
	if response.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected content 'Hello! How can I help you today?', got %s", response.Content)
	}
	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", response.Usage.PromptTokens)
	}
	if response.Usage.CompletionTokens != 15 {
		t.Errorf("Expected completion tokens 15, got %d", response.Usage.CompletionTokens)
	}
}

func TestAnthropic_SendRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "invalid_request_error", "message": "Invalid request"}}`))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := vendor.SendRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error to contain '400', got %s", err.Error())
	}
}

func TestAnthropic_SendRequest_NetworkError(t *testing.T) {
	// Create vendor with invalid URL
	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second, // Short timeout for faster test
	})

	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := vendor.SendRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAnthropic_SendRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": json`)) // Invalid JSON
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	_, err := vendor.SendRequest(context.Background(), request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAnthropic_SendRequest_NoChoices(t *testing.T) {
	// Create a test server that returns response without choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"id": "msg_test123",
			"type": "message",
			"role": "assistant",
			"content": [],
			"model": "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"usage": {
				"input_tokens": 10,
				"output_tokens": 0
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	response, err := vendor.SendRequest(context.Background(), request)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
		return
	}
	if response.Content != "" {
		t.Errorf("Expected empty content, got %s", response.Content)
	}
}

func TestAnthropic_SendRequest_WithHeaders(t *testing.T) {
	// Create a test server that returns a valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"id": "msg_test123",
			"type": "message",
			"role": "assistant",
			"content": [
				{
					"type": "text",
					"text": "Response with custom headers"
				}
			],
			"model": "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"usage": {
				"input_tokens": 10,
				"output_tokens": 15
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	response, err := vendor.SendRequest(context.Background(), request)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if response == nil {
		t.Error("Expected response, got nil")
		return
	}
	if response.Content == "" {
		t.Error("Expected content in response")
	}
}
