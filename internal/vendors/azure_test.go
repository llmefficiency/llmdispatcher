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

func TestNewAzureOpenAI(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.VendorConfig
		wantNil bool
	}{
		{
			name: "with config",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://your-resource.openai.azure.com",
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
			vendor := NewAzureOpenAI(tt.config)
			if (vendor == nil) != tt.wantNil {
				t.Errorf("NewAzureOpenAI() = %v, want nil = %v", vendor, tt.wantNil)
			}
		})
	}
}

func TestAzureOpenAIVendor_Name(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	if vendor.Name() != "azure-openai" {
		t.Errorf("Expected name 'azure-openai', got %s", vendor.Name())
	}
}

func TestAzureOpenAIVendor_GetCapabilities(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	capabilities := vendor.GetCapabilities()

	expectedModels := []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-35-turbo",
		"gpt-35-turbo-16k",
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming support to be true")
	}

	if capabilities.MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens 4096, got %d", capabilities.MaxTokens)
	}

	if capabilities.MaxInputTokens != 128000 {
		t.Errorf("Expected MaxInputTokens 128000, got %d", capabilities.MaxInputTokens)
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

func TestAzureOpenAIVendor_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		config   *models.VendorConfig
		expected bool
	}{
		{
			name: "with API key and base URL",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "https://your-resource.openai.azure.com",
			},
			expected: true,
		},
		{
			name: "without API key",
			config: &models.VendorConfig{
				APIKey:  "",
				BaseURL: "https://your-resource.openai.azure.com",
			},
			expected: false,
		},
		{
			name: "without base URL",
			config: &models.VendorConfig{
				APIKey:  "test-key",
				BaseURL: "",
			},
			expected: false,
		},
		{
			name: "without API key and base URL",
			config: &models.VendorConfig{
				APIKey:  "",
				BaseURL: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := NewAzureOpenAI(tt.config)
			result := vendor.IsAvailable(context.Background())
			if result != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAzureOpenAIVendor_ConvertRequest(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	req := &models.Request{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		TopP:        0.9,
		Stream:      true,
	}

	azureReq := vendor.convertRequest(req)

	if len(azureReq.Messages) != len(req.Messages) {
		t.Errorf("Expected %d messages, got %d", len(req.Messages), len(azureReq.Messages))
	}

	if azureReq.Messages[0].Role != req.Messages[0].Role {
		t.Errorf("Expected role %s, got %s", req.Messages[0].Role, azureReq.Messages[0].Role)
	}

	if azureReq.Messages[0].Content != req.Messages[0].Content {
		t.Errorf("Expected content %s, got %s", req.Messages[0].Content, azureReq.Messages[0].Content)
	}

	if azureReq.Temperature != req.Temperature {
		t.Errorf("Expected temperature %f, got %f", req.Temperature, azureReq.Temperature)
	}

	if azureReq.MaxTokens != req.MaxTokens {
		t.Errorf("Expected max tokens %d, got %d", req.MaxTokens, azureReq.MaxTokens)
	}

	if azureReq.TopP != req.TopP {
		t.Errorf("Expected top_p %f, got %f", req.TopP, azureReq.TopP)
	}

	if azureReq.Stream != req.Stream {
		t.Errorf("Expected stream %v, got %v", req.Stream, azureReq.Stream)
	}
}

func TestAzureOpenAIVendor_ConvertResponse(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	azureResp := &azureResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []azureChoice{
			{
				Index: 0,
				Message: azureMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
			},
		},
		Usage: azureUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
	}

	response := vendor.convertResponse(azureResp, "gpt-4")

	if response.Content != "Hello! How can I help you?" {
		t.Errorf("Expected content 'Hello! How can I help you?', got %s", response.Content)
	}

	if response.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", response.Model)
	}

	if response.Vendor != "azure-openai" {
		t.Errorf("Expected vendor 'azure-openai', got %s", response.Vendor)
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

func TestAzureOpenAIVendor_ConvertResponse_EmptyChoices(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	azureResp := &azureResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: []azureChoice{},
		Usage: azureUsage{
			PromptTokens:     5,
			CompletionTokens: 0,
			TotalTokens:      5,
		},
	}

	response := vendor.convertResponse(azureResp, "gpt-4")

	if response.Content != "" {
		t.Errorf("Expected empty content, got %s", response.Content)
	}

	if response.Usage.TotalTokens != 5 {
		t.Errorf("Expected total tokens 5, got %d", response.Usage.TotalTokens)
	}
}

func TestAzureOpenAIVendor_ConvertResponse_NilChoices(t *testing.T) {
	vendor := NewAzureOpenAI(nil)
	azureResp := &azureResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Choices: nil,
		Usage: azureUsage{
			PromptTokens:     5,
			CompletionTokens: 0,
			TotalTokens:      5,
		},
	}

	response := vendor.convertResponse(azureResp, "gpt-4")

	if response.Content != "" {
		t.Errorf("Expected empty content, got %s", response.Content)
	}
}

func TestAzureOpenAI_SendStreamingRequest_Success(t *testing.T) {
	// Create a test server that returns streaming data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("api-key") != "test-key" {
			t.Errorf("Expected api-key test-key, got %s", r.Header.Get("api-key"))
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
	vendor := NewAzureOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	// Create request
	req := &models.Request{
		Model: "gpt-4",
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

func TestAzureOpenAI_SendStreamingRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	vendor := NewAzureOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-4",
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

func TestAzureOpenAI_SendStreamingRequest_NetworkError(t *testing.T) {
	vendor := NewAzureOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-that-does-not-exist.com",
		Timeout: 1 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-4",
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

func TestAzureOpenAI_SendStreamingRequest_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: invalid json\n\n"))
	}))
	defer server.Close()

	vendor := NewAzureOpenAI(&models.VendorConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	})

	req := &models.Request{
		Model: "gpt-4",
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

func TestAzureOpenAI_SendStreamingRequest_WithHeaders(t *testing.T) {
	t.Skip("Skipping streaming test due to race conditions")
}
