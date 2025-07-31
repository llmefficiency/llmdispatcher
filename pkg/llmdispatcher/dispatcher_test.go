package llmdispatcher

import (
	"context"
	"testing"
	"time"
)

// MockVendor is a mock implementation of Vendor for testing
type MockVendor struct {
	name         string
	shouldFail   bool
	response     *Response
	capabilities Capabilities
	available    bool
}

func (m *MockVendor) Name() string {
	return m.name
}

func (m *MockVendor) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	if m.shouldFail {
		return nil, &MockError{message: "mock error"}
	}
	return m.response, nil
}

func (m *MockVendor) GetCapabilities() Capabilities {
	return m.capabilities
}

func (m *MockVendor) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *MockVendor) SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error) {
	if m.shouldFail {
		return nil, &MockError{message: "mock streaming error"}
	}

	streamingResp := NewStreamingResponse(req.Model, m.name)

	// Simulate streaming response
	go func() {
		defer streamingResp.Close()
		streamingResp.ContentChan <- "Mock streaming response"
		streamingResp.DoneChan <- true
	}()

	return streamingResp, nil
}

type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

func TestNew(t *testing.T) {
	dispatcher := New()
	if dispatcher == nil {
		t.Fatal("New() returned nil")
	}
	if dispatcher.dispatcher == nil {
		t.Error("internal dispatcher should not be nil")
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "with config",
			config: &Config{
				DefaultVendor:  "openai",
				FallbackVendor: "anthropic",
				Timeout:        30 * time.Second,
				EnableLogging:  true,
				EnableMetrics:  true,
			},
		},
		{
			name:   "nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := NewWithConfig(tt.config)
			if dispatcher == nil {
				t.Fatal("NewWithConfig() returned nil")
			}
			if dispatcher.dispatcher == nil {
				t.Error("internal dispatcher should not be nil")
			}
		})
	}
}

func TestSend_Success(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor
	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &Response{
			Content: "Hello from test vendor",
			Model:   "test-model",
			Vendor:  "test-vendor",
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
			CreatedAt: time.Now(),
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

	if response.Content != "Hello from test vendor" {
		t.Errorf("Expected content 'Hello from test vendor', got '%s'", response.Content)
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

func TestSend_NoVendors(t *testing.T) {
	dispatcher := New()

	request := &Request{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := dispatcher.Send(ctx, request)
	if err == nil {
		t.Fatal("Expected error when no vendors are registered")
	}
}

func TestRegisterVendor(t *testing.T) {
	dispatcher := New()

	tests := []struct {
		name    string
		vendor  Vendor
		wantErr bool
	}{
		{
			name: "valid vendor",
			vendor: &MockVendor{
				name: "test-vendor",
			},
			wantErr: false,
		},
		{
			name:    "nil vendor",
			vendor:  nil,
			wantErr: true,
		},
		{
			name: "empty vendor name",
			vendor: &MockVendor{
				name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dispatcher.RegisterVendor(tt.vendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterVendor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	dispatcher := New()

	// Send a request to generate some stats
	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &Response{
			Content: "Test response",
			Model:   "test-model",
			Vendor:  "test-vendor",
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
	}

	ctx := context.Background()
	_, err = dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	stats := dispatcher.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", stats.SuccessfulRequests)
	}

	if len(stats.VendorStats) != 1 {
		t.Errorf("Expected 1 vendor stat, got %d", len(stats.VendorStats))
	}
}

func TestGetVendors(t *testing.T) {
	dispatcher := New()

	// Register some vendors
	vendor1 := &MockVendor{name: "vendor1"}
	vendor2 := &MockVendor{name: "vendor2"}

	err := dispatcher.RegisterVendor(vendor1)
	if err != nil {
		t.Fatalf("Failed to register vendor1: %v", err)
	}

	err = dispatcher.RegisterVendor(vendor2)
	if err != nil {
		t.Fatalf("Failed to register vendor2: %v", err)
	}

	vendors := dispatcher.GetVendors()
	if len(vendors) != 2 {
		t.Errorf("Expected 2 vendors, got %d", len(vendors))
	}

	// Check that both vendors are in the list
	vendorMap := make(map[string]bool)
	for _, vendor := range vendors {
		vendorMap[vendor] = true
	}

	if !vendorMap["vendor1"] {
		t.Error("vendor1 not found in vendor list")
	}

	if !vendorMap["vendor2"] {
		t.Error("vendor2 not found in vendor list")
	}
}

func TestGetVendor(t *testing.T) {
	dispatcher := New()

	// Register a vendor
	mockVendor := &MockVendor{name: "test-vendor"}
	err := dispatcher.RegisterVendor(mockVendor)
	if err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	// Test getting existing vendor
	vendor, exists := dispatcher.GetVendor("test-vendor")
	if !exists {
		t.Error("Expected vendor to exist")
	}
	if vendor == nil {
		t.Error("Expected vendor to not be nil")
	}

	// Test getting non-existent vendor
	vendor, exists = dispatcher.GetVendor("non-existent")
	if exists {
		t.Error("Expected vendor to not exist")
	}
	if vendor != nil {
		t.Error("Expected vendor to be nil")
	}
}

func TestVendorCapabilities(t *testing.T) {
	// Test vendor capabilities through the public API
	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &Response{
			Content: "Test response",
			Model:   "test-model",
			Vendor:  "test-vendor",
		},
		capabilities: Capabilities{
			Models:            []string{"test-model"},
			SupportsStreaming: true,
			MaxTokens:         1000,
			MaxInputTokens:    10000,
		},
		available: true,
	}

	dispatcher := New()
	err := dispatcher.RegisterVendor(mockVendor)
	if err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	// Test getting vendor through public API
	vendor, exists := dispatcher.GetVendor("test-vendor")
	if !exists {
		t.Fatal("Expected vendor to exist")
	}

	// Test vendor capabilities
	capabilities := vendor.GetCapabilities()
	if len(capabilities.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(capabilities.Models))
	}

	if !capabilities.SupportsStreaming {
		t.Error("Expected streaming to be supported")
	}

	if capabilities.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", capabilities.MaxTokens)
	}

	// Test vendor availability
	available := vendor.IsAvailable(context.Background())
	if !available {
		t.Error("Expected vendor to be available")
	}
}

func TestNewWithConfig_Complex(t *testing.T) {
	config := &Config{
		DefaultVendor:  "openai",
		FallbackVendor: "anthropic",
		Timeout:        30 * time.Second,
		EnableLogging:  true,
		EnableMetrics:  true,
		RetryPolicy: &RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
		RoutingRules: []RoutingRule{
			{
				Condition: RoutingCondition{
					ModelPattern: "gpt-4",
					MaxTokens:    1000,
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	dispatcher := NewWithConfig(config)
	if dispatcher == nil {
		t.Fatal("NewWithConfig() returned nil")
	}

	// Test that the dispatcher was created successfully
	if dispatcher.dispatcher == nil {
		t.Error("internal dispatcher should not be nil")
	}
}

func TestSend_WithComplexRequest(t *testing.T) {
	dispatcher := New()

	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &Response{
			Content:      "Complex response",
			Model:        "test-model",
			Vendor:       "test-vendor",
			FinishReason: "stop",
			CreatedAt:    time.Now(),
			Usage: Usage{
				PromptTokens:     20,
				CompletionTokens: 10,
				TotalTokens:      30,
			},
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
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
		Temperature: 0.8,
		MaxTokens:   200,
		TopP:        0.9,
		Stream:      false,
		Stop:        []string{"\n", "END"},
		User:        "test-user",
	}

	ctx := context.Background()
	response, err := dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Content != "Complex response" {
		t.Errorf("Expected content 'Complex response', got '%s'", response.Content)
	}

	if response.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got '%s'", response.FinishReason)
	}

	if response.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", response.Usage.TotalTokens)
	}
}
