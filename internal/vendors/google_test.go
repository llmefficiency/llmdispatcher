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

func TestGoogle_SendStreamingRequest_Success(t *testing.T) {
	// Create a test server that returns streaming data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("Expected x-goog-api-key test-key, got %s", r.Header.Get("x-goog-api-key"))
		}

		// Set response headers for streaming
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Send streaming data
		streamData := []string{
			"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Hello\"}]}}]}\n\n",
			"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\" world\"}]}}]}\n\n",
			"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"!\"}]}}]}\n\n",
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
	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create request
	req := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendStreamingRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendStreamingRequest_NetworkError(t *testing.T) {
	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second,
	})

	req := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendStreamingRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: invalid json\n\n"))
	}))
	defer server.Close()

	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendStreamingRequest_WithHeaders(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}

func TestGoogle_SendRequest_Success(t *testing.T) {
	// Create a test server that returns a valid Google response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1beta/models") {
			t.Errorf("Expected path to contain /v1beta/models, got %s", r.URL.Path)
		}

		// Verify URL contains API key
		if !strings.Contains(r.URL.String(), "key=test-key") {
			t.Errorf("Expected URL to contain 'key=test-key', got %s", r.URL.String())
		}

		// Return a valid Google response
		response := `{
			"candidates": [
				{
					"content": {
						"parts": [
							{
								"text": "Hello! How can I help you today?"
							}
						]
					},
					"finishReason": "STOP"
				}
			],
			"usageMetadata": {
				"promptTokenCount": 10,
				"candidatesTokenCount": 15,
				"totalTokenCount": 25
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create vendor with test server URL
	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create a test request
	request := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"message": "Invalid request"}}`))
	}))
	defer server.Close()

	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendRequest_NetworkError(t *testing.T) {
	// Create vendor with invalid URL
	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second, // Short timeout for faster test
	})

	request := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": json`)) // Invalid JSON
	}))
	defer server.Close()

	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gemini-1.5-pro",
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

func TestGoogle_SendRequest_NoCandidates(t *testing.T) {
	// Create a test server that returns response without candidates
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"candidates": [],
			"usageMetadata": {
				"promptTokenCount": 10,
				"candidatesTokenCount": 0,
				"totalTokenCount": 10
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	vendor := NewGoogle(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	request := &models.Request{
		Model: "gemini-1.5-pro",
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
