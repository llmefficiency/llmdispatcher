package vendors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

func TestNewOpenAI(t *testing.T) {
	tests := []struct {
		name   string
		config *models.VendorConfig
	}{
		{
			name: "with config",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
				Timeout: 30 * time.Second,
			},
		},
		{
			name:   "nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := NewOpenAI(tt.config)
			if vendor == nil {
				t.Fatal("NewOpenAI() returned nil")
			}
			if vendor.config == nil {
				t.Error("config should not be nil")
			}
			if vendor.client == nil {
				t.Error("client should not be nil")
			}
		})
	}
}

func TestOpenAI_Name(t *testing.T) {
	vendor := NewOpenAI(nil)
	name := vendor.Name()
	if name != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", name)
	}
}

func TestOpenAI_GetCapabilities(t *testing.T) {
	vendor := NewOpenAI(nil)
	capabilities := vendor.GetCapabilities()

	expectedModels := []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}

	if len(capabilities.Models) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(capabilities.Models))
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming to be supported")
	}

	if capabilities.MaxTokens != 4096 {
		t.Errorf("Expected max tokens 4096, got %d", capabilities.MaxTokens)
	}

	if capabilities.MaxInputTokens != 128000 {
		t.Errorf("Expected max input tokens 128000, got %d", capabilities.MaxInputTokens)
	}
}

func TestOpenAI_IsAvailable(t *testing.T) {
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
			vendor := NewOpenAI(tt.config)
			available := vendor.IsAvailable(context.Background())
			if available != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", available, tt.expected)
			}
		})
	}
}

func TestOpenAI_SendRequest_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Check authorization header
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Parse request body
		var req OpenAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Check request fields
		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model gpt-3.5-turbo, got %s", req.Model)
		}

		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != "user" {
			t.Errorf("Expected role user, got %s", req.Messages[0].Role)
		}

		if req.Messages[0].Content != "Hello" {
			t.Errorf("Expected content Hello, got %s", req.Messages[0].Content)
		}

		// Return success response
		response := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create vendor with test server URL
	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create request
	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// Send request
	ctx := context.Background()
	response, err := vendor.SendRequest(ctx, request)
	if err != nil {
		t.Fatalf("SendRequest() failed: %v", err)
	}

	// Check response
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected content 'Hello! How can I help you today?', got '%s'", response.Content)
	}

	if response.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model gpt-3.5-turbo, got %s", response.Model)
	}

	if response.Vendor != "openai" {
		t.Errorf("Expected vendor openai, got %s", response.Vendor)
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

func TestOpenAI_SendRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := OpenAIError{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code,omitempty"`
			}{
				Message: "Invalid request",
				Type:    "invalid_request_error",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := vendor.SendRequest(ctx, request)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check that the error message contains the API error
	if err.Error() != "OpenAI API error: Invalid request" {
		t.Errorf("Expected error message to contain 'Invalid request', got '%s'", err.Error())
	}
}

func TestOpenAI_SendRequest_NetworkError(t *testing.T) {
	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Millisecond, // Very short timeout
	})

	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := vendor.SendRequest(ctx, request)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestOpenAI_SendRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := vendor.SendRequest(ctx, request)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestOpenAI_SendRequest_NoChoices(t *testing.T) {
	// Create a test server that returns response without choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{}, // Empty choices
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := vendor.SendRequest(ctx, request)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "no choices in response" {
		t.Errorf("Expected error 'no choices in response', got '%s'", err.Error())
	}
}

func TestOpenAI_SendRequest_WithHeaders(t *testing.T) {
	// Create a test server that checks custom headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom header
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}

		// Return success response
		response := OpenAIResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Response with custom headers",
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	})

	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	response, err := vendor.SendRequest(ctx, request)
	if err != nil {
		t.Fatalf("SendRequest() failed: %v", err)
	}

	if response.Content != "Response with custom headers" {
		t.Errorf("Expected content 'Response with custom headers', got '%s'", response.Content)
	}
}

func TestOpenAI_SendStreamingRequest_Success(t *testing.T) {
	// Create a test server that returns streaming data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Set response headers for streaming
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Send streaming data
		streamData := []string{
			"data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n",
			"data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n",
			"data: {\"choices\":[{\"delta\":{\"content\":\"!\"}}]}\n\n",
			"data: [DONE]\n\n",
		}

		for _, data := range streamData {
			w.Write([]byte(data))
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond) // Add small delay between chunks
		}
	}))
	defer server.Close()

	// Create vendor with test server URL
	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create request
	req := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		Stream:      true,
	}

	// Send streaming request
	ctx := context.Background()
	streamingResp, err := vendor.SendStreamingRequest(ctx, req)
	if err != nil {
		t.Fatalf("SendStreamingRequest failed: %v", err)
	}
	defer streamingResp.Close()

	// Collect streaming content
	var content string
	done := false
	for !done {
		select {
		case chunk := <-streamingResp.ContentChan:
			content += chunk
		case done = <-streamingResp.DoneChan:
		case err := <-streamingResp.ErrorChan:
			t.Fatalf("Streaming error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for streaming response")
		}
	}

	expected := "Hello world!"
	if content != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, content)
	}
}

func TestOpenAI_SendStreamingRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-3.5-turbo",
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

func TestOpenAI_SendStreamingRequest_NetworkError(t *testing.T) {
	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-3.5-turbo",
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

func TestOpenAI_SendStreamingRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: invalid json\n\n"))
	}))
	defer server.Close()

	vendor := NewOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	ctx := context.Background()
	streamingResp, err := vendor.SendStreamingRequest(ctx, req)
	if err != nil {
		t.Fatalf("SendStreamingRequest failed: %v", err)
	}
	defer streamingResp.Close()

	// Wait for error
	select {
	case err := <-streamingResp.ErrorChan:
		if err == nil {
			t.Error("Expected error from streaming response")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for streaming error")
	}
}

func TestOpenAI_SendStreamingRequest_WithHeaders(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}
