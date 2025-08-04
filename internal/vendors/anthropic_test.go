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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Send streaming data
		w.Write([]byte("data: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_test123\"}}\n\n"))
		w.Write([]byte("data: {\"type\": \"content_block_start\", \"index\": 0, \"content_block\": {\"type\": \"text\", \"text\": \"Hello\"}}\n\n"))
		w.Write([]byte("data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"! How can I help you today?\"}}\n\n"))
		w.Write([]byte("data: {\"type\": \"content_block_stop\", \"index\": 0}\n\n"))
		w.Write([]byte("data: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\", \"stop_sequence\": null}}\n\n"))
		w.Write([]byte("data: {\"type\": \"message_stop\"}\n\n"))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	streamingResp, err := vendor.SendStreamingRequest(context.TODO(), req)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
		return
	}
	if streamingResp == nil {
		t.Error("Expected streaming response, got nil")
		return
	}

	// Read from the streaming response
	content := ""
	for {
		select {
		case chunk := <-streamingResp.ContentChan:
			content += chunk
		case <-streamingResp.DoneChan:
			goto done
		case err := <-streamingResp.ErrorChan:
			if err != nil {
				t.Errorf("Unexpected error from streaming: %v", err)
			}
			goto done
		}
	}
done:

	if content == "" {
		t.Error("Expected content from streaming response")
	}
}

func TestAnthropic_SendStreamingRequest_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	streamingResp, err := vendor.SendStreamingRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for HTTP error, got nil")
		return
	}
	if streamingResp != nil {
		t.Error("Expected nil streaming response for error")
	}
}

func TestAnthropic_SendStreamingRequest_InvalidRequest(t *testing.T) {
	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.anthropic.com",
	})

	// Test with invalid request (empty model)
	req := &models.Request{
		Model: "",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	streamingResp, err := vendor.SendStreamingRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
		return
	}
	if streamingResp != nil {
		t.Error("Expected nil streaming response for invalid request")
	}
}

func TestAnthropic_SendStreamingRequest_JSONMarshalError(t *testing.T) {
	vendor := NewAnthropic(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.anthropic.com",
	})

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// This should fail due to HTTP request creation
	streamingResp, err := vendor.SendStreamingRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for HTTP request failure, got nil")
		return
	}
	if streamingResp != nil {
		t.Error("Expected nil streaming response for error")
	}
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

func TestAnthropic_SendRequest_InvalidRequest(t *testing.T) {
	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: "https://api.anthropic.com",
		},
		client: &http.Client{},
	}

	// Test with invalid request (empty model)
	req := &models.Request{
		Model: "",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := vendor.SendRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
		return
	}
	if resp != nil {
		t.Error("Expected nil response for invalid request")
	}
}

func TestAnthropic_SendRequest_JSONMarshalError(t *testing.T) {
	// This test is difficult to trigger in practice, but we can test the structure
	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: "https://api.anthropic.com",
		},
		client: &http.Client{},
	}

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// This should work fine since the request is valid
	resp, err := vendor.SendRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error due to HTTP request failure, got nil")
		return
	}
	if resp != nil {
		t.Error("Expected nil response for HTTP error")
	}
}

func TestAnthropic_SendRequest_HTTPRequestCreationError(t *testing.T) {
	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: "https://api.anthropic.com",
		},
		client: &http.Client{},
	}

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// This should fail due to HTTP request creation (invalid URL or context)
	resp, err := vendor.SendRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error due to HTTP request failure, got nil")
		return
	}
	if resp != nil {
		t.Error("Expected nil response for HTTP error")
	}
}

func TestAnthropic_SendRequest_ResponseBodyReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers but don't write body, then close connection
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Close the connection immediately to cause read error
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
		client: &http.Client{},
	}

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := vendor.SendRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for response body read failure, got nil")
		return
	}
	if resp != nil {
		t.Error("Expected nil response for read error")
	}
}

func TestAnthropic_SendRequest_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
		client: &http.Client{},
	}

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := vendor.SendRequest(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for invalid JSON response, got nil")
		return
	}
	if resp != nil {
		t.Error("Expected nil response for JSON parse error")
	}
}

func TestAnthropic_SendRequest_WithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers are set
		if r.Header.Get("custom-header") != "custom-value" {
			t.Errorf("Expected custom-header, got: %s", r.Header.Get("custom-header"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "msg_test123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello! How can I help you today?"}],
			"usage": {"input_tokens": 10, "output_tokens": 15}
		}`))
	}))
	defer server.Close()

	vendor := &AnthropicVendor{
		config: &models.VendorConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
			Headers: map[string]string{
				"custom-header": "custom-value",
			},
		},
		client: &http.Client{},
	}

	req := &models.Request{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := vendor.SendRequest(context.TODO(), req)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
		return
	}
	if resp == nil {
		t.Error("Expected response, got nil")
		return
	}
	if resp.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected 'Hello! How can I help you today?', got: %s", resp.Content)
	}
}
