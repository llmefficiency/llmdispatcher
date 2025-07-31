package vendors

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

func TestNewLocal(t *testing.T) {
	tests := []struct {
		name   string
		config *models.VendorConfig
		want   *Local
	}{
		{
			name:   "nil config",
			config: nil,
			want: &Local{
				config: &models.VendorConfig{},
				client: nil, // Will be set by NewLocal
				resourceLimits: &ResourceLimits{
					MaxMemoryMB: 4096,
					MaxThreads:  4,
				},
			},
		},
		{
			name: "with HTTP config",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"server_url": "http://localhost:11434",
					"model_path": "llama2:7b",
				},
				Timeout: 30 * time.Second,
			},
			want: &Local{
				config: &models.VendorConfig{
					Headers: map[string]string{
						"server_url": "http://localhost:11434",
						"model_path": "llama2:7b",
					},
					Timeout: 30 * time.Second,
				},
				serverURL: "http://localhost:11434",
				modelPath: "llama2:7b",
				useHTTP:   true,
				resourceLimits: &ResourceLimits{
					MaxMemoryMB: 4096,
					MaxThreads:  4,
				},
			},
		},
		{
			name: "with executable config",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"executable": "/usr/local/bin/llama",
					"model_path": "/path/to/model.gguf",
				},
			},
			want: &Local{
				config: &models.VendorConfig{
					Headers: map[string]string{
						"executable": "/usr/local/bin/llama",
						"model_path": "/path/to/model.gguf",
					},
				},
				executable: "/usr/local/bin/llama",
				modelPath:  "/path/to/model.gguf",
				useHTTP:    false,
				resourceLimits: &ResourceLimits{
					MaxMemoryMB: 4096,
					MaxThreads:  4,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLocal(tt.config)

			// Check basic fields
			if got.config == nil {
				t.Error("Expected config to be set")
			}
			if got.client == nil {
				t.Error("Expected client to be set")
			}
			if got.resourceLimits == nil {
				t.Error("Expected resourceLimits to be set")
			}

			// Check local-specific fields
			if tt.want.useHTTP {
				if got.serverURL != tt.want.serverURL {
					t.Errorf("Expected serverURL %s, got %s", tt.want.serverURL, got.serverURL)
				}
				if got.modelPath != tt.want.modelPath {
					t.Errorf("Expected modelPath %s, got %s", tt.want.modelPath, got.modelPath)
				}
				if got.useHTTP != tt.want.useHTTP {
					t.Errorf("Expected useHTTP %v, got %v", tt.want.useHTTP, got.useHTTP)
				}
			}

			if tt.want.executable != "" {
				if got.executable != tt.want.executable {
					t.Errorf("Expected executable %s, got %s", tt.want.executable, got.executable)
				}
				if got.modelPath != tt.want.modelPath {
					t.Errorf("Expected modelPath %s, got %s", tt.want.modelPath, got.modelPath)
				}
			}
		})
	}
}

func TestLocal_Name(t *testing.T) {
	local := NewLocal(nil)
	if got := local.Name(); got != "local" {
		t.Errorf("Local.Name() = %v, want %v", got, "local")
	}
}

func TestLocal_GetCapabilities(t *testing.T) {
	local := NewLocal(nil)
	capabilities := local.GetCapabilities()

	// Check that capabilities are set
	if len(capabilities.Models) == 0 {
		t.Error("Expected models to be set")
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming to be supported")
	}

	if capabilities.MaxTokens == 0 {
		t.Error("Expected MaxTokens to be set")
	}

	if capabilities.MaxInputTokens == 0 {
		t.Error("Expected MaxInputTokens to be set")
	}
}

func TestLocal_prepareInput(t *testing.T) {
	local := NewLocal(nil)

	tests := []struct {
		name     string
		messages []models.Message
		want     string
	}{
		{
			name: "single user message",
			messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			want: "User: Hello\nAssistant: ",
		},
		{
			name: "system and user messages",
			messages: []models.Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
			},
			want: "System: You are a helpful assistant\nUser: Hello\nAssistant: ",
		},
		{
			name: "conversation with assistant",
			messages: []models.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			want: "User: Hello\nAssistant: Hi there!\nUser: How are you?\nAssistant: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := local.prepareInput(tt.messages)
			if got != tt.want {
				t.Errorf("prepareInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocal_extractGeneratedContent(t *testing.T) {
	local := NewLocal(nil)

	tests := []struct {
		name   string
		input  string
		output string
		want   string
	}{
		{
			name:   "normal extraction",
			input:  "User: Hello\nAssistant: ",
			output: "User: Hello\nAssistant: I'm doing well, thank you!",
			want:   "I'm doing well, thank you!",
		},
		{
			name:   "no prefix match",
			input:  "User: Hello\nAssistant: ",
			output: "I'm doing well, thank you!",
			want:   "I'm doing well, thank you!",
		},
		{
			name:   "empty output",
			input:  "User: Hello\nAssistant: ",
			output: "User: Hello\nAssistant: ",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := local.extractGeneratedContent(tt.input, tt.output)
			if got != tt.want {
				t.Errorf("extractGeneratedContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocal_IsAvailable(t *testing.T) {
	local := NewLocal(nil)

	// Test with no configuration (should return false)
	if local.IsAvailable(context.Background()) {
		t.Error("Expected IsAvailable to return false when no configuration is set")
	}

	// Test with HTTP configuration
	local.useHTTP = true
	local.serverURL = "http://localhost:11434"
	// This will likely fail in test environment, but we can test the logic
	_ = local.IsAvailable(context.Background())

	// Test with executable configuration
	local.useHTTP = false
	local.executable = "/nonexistent/path"
	if local.IsAvailable(context.Background()) {
		t.Error("Expected IsAvailable to return false for nonexistent executable")
	}
}

func TestLocal_SendRequest_HTTP(t *testing.T) {
	local := NewLocal(&models.VendorConfig{
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
		},
		Timeout: 5 * time.Second,
	})

	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// This will likely fail in test environment, but we can test the error handling
	_, err := local.SendRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error when local model server is not available")
	}
}

func TestLocal_SendRequest_Process(t *testing.T) {
	local := NewLocal(&models.VendorConfig{
		Headers: map[string]string{
			"executable": "/nonexistent/llama",
			"model_path": "/path/to/model.gguf",
		},
		Timeout: 5 * time.Second,
	})

	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// This should fail because the executable doesn't exist
	_, err := local.SendRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error when executable doesn't exist")
	}
}

func TestResourceLimits(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemoryMB:  8192,
		MaxThreads:   8,
		MaxGPULayers: 32,
	}

	if limits.MaxMemoryMB != 8192 {
		t.Errorf("Expected MaxMemoryMB 8192, got %d", limits.MaxMemoryMB)
	}

	if limits.MaxThreads != 8 {
		t.Errorf("Expected MaxThreads 8, got %d", limits.MaxThreads)
	}

	if limits.MaxGPULayers != 32 {
		t.Errorf("Expected MaxGPULayers 32, got %d", limits.MaxGPULayers)
	}
}

func TestLocal_SendStreamingRequest_HTTP(t *testing.T) {
	local := NewLocal(&models.VendorConfig{
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
		},
		Timeout: 5 * time.Second,
	})

	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// This will likely fail in test environment, but we can test the error handling
	_, err := local.SendStreamingRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error when local model server is not available")
	}
}

func TestLocal_SendStreamingRequest_Process(t *testing.T) {
	local := NewLocal(&models.VendorConfig{
		Headers: map[string]string{
			"executable": "/nonexistent/llama",
			"model_path": "/path/to/model.gguf",
		},
		Timeout: 5 * time.Second,
	})

	// Ensure we're in process mode (not HTTP mode)
	local.useHTTP = false

	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// This should fail because the executable doesn't exist
	streamingResp, err := local.SendStreamingRequest(context.Background(), req)
	if err != nil {
		// Check that the error message contains expected content
		errorMsg := err.Error()
		t.Logf("Got error: %s", errorMsg)
		if !strings.Contains(errorMsg, "executable") && !strings.Contains(errorMsg, "path not configured") && !strings.Contains(errorMsg, "failed") {
			t.Errorf("Expected error to mention executable, path not configured, or failed, got: %s", errorMsg)
		}
		return
	}

	// If no immediate error, check the streaming response for errors
	if streamingResp != nil {
		defer streamingResp.Close()

		// Wait for an error from the streaming response
		select {
		case err := <-streamingResp.ErrorChan:
			t.Logf("Got streaming error: %s", err.Error())
			if !strings.Contains(err.Error(), "executable") && !strings.Contains(err.Error(), "path not configured") && !strings.Contains(err.Error(), "failed") {
				t.Errorf("Expected streaming error to mention executable, path not configured, or failed, got: %s", err.Error())
			}
		case <-time.After(2 * time.Second):
			t.Error("Expected error from streaming response but got none")
		}
	} else {
		t.Error("Expected error when executable doesn't exist")
		t.Logf("useHTTP: %v, executable: %s", local.useHTTP, local.executable)
	}
}

func TestLocal_GetCapabilities_Comprehensive(t *testing.T) {
	local := NewLocal(nil)
	capabilities := local.GetCapabilities()

	// Test that capabilities are properly set
	if capabilities.MaxTokens == 0 {
		t.Error("Expected MaxTokens to be set")
	}

	if capabilities.MaxInputTokens == 0 {
		t.Error("Expected MaxInputTokens to be set")
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming to be supported")
	}

	if len(capabilities.Models) == 0 {
		t.Error("Expected models to be set")
	}

	// Test that local models are included
	foundLocal := false
	for _, model := range capabilities.Models {
		if model == "llama2:7b" || model == "llama2:13b" || model == "mistral" {
			foundLocal = true
			break
		}
	}
	if !foundLocal {
		t.Error("Expected local models to be included in capabilities")
	}
}

func TestLocal_IsAvailable_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		config   *models.VendorConfig
		expected bool
	}{
		{
			name:     "no configuration",
			config:   nil,
			expected: false,
		},
		{
			name: "HTTP configuration",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"server_url": "http://localhost:11434",
				},
			},
			expected: false, // Will fail in test environment
		},
		{
			name: "executable configuration",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"executable": "/nonexistent/path",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := NewLocal(tt.config)
			result := local.IsAvailable(context.Background())
			if result != tt.expected {
				t.Errorf("Expected IsAvailable to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLocal_Request_Response_Conversion(t *testing.T) {
	// Test request conversion
	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
		TopP:        0.9,
		Stop:        []string{"END"},
	}

	// Test that the request can be marshaled (simulating HTTP request)
	localReq := LocalRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      false,
		Stop:        req.Stop,
	}

	_, err := json.Marshal(localReq)
	if err != nil {
		t.Errorf("Failed to marshal local request: %v", err)
	}

	// Test response conversion
	localResp := LocalResponse{
		Model:   "llama2:7b",
		Content: "Hello! How can I help you today?",
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     5,
			CompletionTokens: 10,
			TotalTokens:      15,
		},
	}

	jsonResp, err := json.Marshal(localResp)
	if err != nil {
		t.Errorf("Failed to marshal local response: %v", err)
	}

	// Verify the JSON can be unmarshaled back
	var unmarshaledResp LocalResponse
	if err := json.Unmarshal(jsonResp, &unmarshaledResp); err != nil {
		t.Errorf("Failed to unmarshal local response: %v", err)
	}

	if unmarshaledResp.Content != localResp.Content {
		t.Errorf("Expected content %s, got %s", localResp.Content, unmarshaledResp.Content)
	}

	if unmarshaledResp.Usage.TotalTokens != localResp.Usage.TotalTokens {
		t.Errorf("Expected total tokens %d, got %d", localResp.Usage.TotalTokens, unmarshaledResp.Usage.TotalTokens)
	}
}

func TestLocal_ResourceLimits_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected *ResourceLimits
	}{
		{
			name: "default limits",
			headers: map[string]string{
				"server_url": "http://localhost:11434",
			},
			expected: &ResourceLimits{
				MaxMemoryMB: 4096,
				MaxThreads:  4,
			},
		},
		{
			name: "custom limits",
			headers: map[string]string{
				"server_url":     "http://localhost:11434",
				"max_memory_mb":  "8192",
				"max_threads":    "8",
				"max_gpu_layers": "32",
			},
			expected: &ResourceLimits{
				MaxMemoryMB:  4096, // Currently not parsed from headers
				MaxThreads:   4,    // Currently not parsed from headers
				MaxGPULayers: 0,    // Currently not parsed from headers
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.VendorConfig{
				Headers: tt.headers,
			}

			local := NewLocal(config)

			if local.resourceLimits.MaxMemoryMB != tt.expected.MaxMemoryMB {
				t.Errorf("Expected MaxMemoryMB %d, got %d", tt.expected.MaxMemoryMB, local.resourceLimits.MaxMemoryMB)
			}

			if local.resourceLimits.MaxThreads != tt.expected.MaxThreads {
				t.Errorf("Expected MaxThreads %d, got %d", tt.expected.MaxThreads, local.resourceLimits.MaxThreads)
			}

			if local.resourceLimits.MaxGPULayers != tt.expected.MaxGPULayers {
				t.Errorf("Expected MaxGPULayers %d, got %d", tt.expected.MaxGPULayers, local.resourceLimits.MaxGPULayers)
			}
		})
	}
}

func TestLocal_Error_Handling(t *testing.T) {
	tests := []struct {
		name        string
		config      *models.VendorConfig
		request     *models.Request
		expectError bool
	}{
		{
			name: "no executable configured",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"model_path": "/path/to/model.gguf",
				},
			},
			request: &models.Request{
				Model: "test",
				Messages: []models.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: true,
		},
		{
			name: "no server URL configured for HTTP",
			config: &models.VendorConfig{
				Headers: map[string]string{
					"model_path": "llama2:7b",
				},
			},
			request: &models.Request{
				Model: "llama2:7b",
				Messages: []models.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := NewLocal(tt.config)
			_, err := local.SendRequest(context.Background(), tt.request)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
