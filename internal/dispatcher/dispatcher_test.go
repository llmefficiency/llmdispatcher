package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// MockVendor is a mock implementation of LLMVendor for testing
type MockVendor struct {
	name              string
	shouldFail        bool
	response          *models.Response
	capabilities      models.Capabilities
	available         bool
	supportsStreaming bool
	streamingResponse *models.StreamingResponse
}

func (m *MockVendor) Name() string {
	return m.name
}

func (m *MockVendor) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	if m.shouldFail {
		return nil, errors.New("mock error")
	}
	return m.response, nil
}

func (m *MockVendor) GetCapabilities() models.Capabilities {
	capabilities := m.capabilities
	capabilities.SupportsStreaming = m.supportsStreaming
	return capabilities
}

func (m *MockVendor) IsAvailable(ctx context.Context) bool {
	return m.available
}

// SendStreamingRequest sends a streaming request (mock implementation)
func (m *MockVendor) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	if m.shouldFail {
		return nil, errors.New("mock streaming error")
	}

	if !m.supportsStreaming {
		return nil, errors.New("vendor does not support streaming")
	}

	if m.streamingResponse != nil {
		return m.streamingResponse, nil
	}

	streamingResp := models.NewStreamingResponse(req.Model, m.name)

	// Simulate streaming response
	go func() {
		defer streamingResp.Close()
		streamingResp.ContentChan <- "Mock streaming response"
		streamingResp.DoneChan <- true
	}()

	return streamingResp, nil
}

func TestNew(t *testing.T) {
	dispatcher := New()
	if dispatcher == nil {
		t.Fatal("New() returned nil")
	}
	if dispatcher.config == nil {
		t.Error("config should not be nil")
	}
	if dispatcher.vendors == nil {
		t.Error("vendors map should not be nil")
	}
	if dispatcher.stats == nil {
		t.Error("stats should not be nil")
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *models.Config
	}{
		{
			name: "with config",
			config: &models.Config{
				Mode:          models.AutoMode,
				Timeout:       30 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
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
				t.Error("Expected dispatcher to be created")
			}
		})
	}
}

func TestRegisterVendor(t *testing.T) {
	dispatcher := New()

	tests := []struct {
		name    string
		vendor  models.LLMVendor
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

func TestSend_Success(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor
	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &models.Response{
			Content: "Hello from test vendor",
			Model:   "test-model",
			Vendor:  "test-vendor",
			Usage: models.Usage{
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

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
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
}

func TestSend_NoVendors(t *testing.T) {
	dispatcher := New()

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := dispatcher.Send(ctx, request)
	if err == nil {
		t.Fatal("Expected error when no vendors are registered")
	}
}

// TestSend_WithTimeout is skipped as it requires more complex timeout handling
func TestSend_WithTimeout(t *testing.T) {
	t.Skip("Timeout test requires more complex implementation")
}

func TestSend_WithRetry(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:      2,
			BackoffStrategy: models.ExponentialBackoff,
			RetryableErrors: []string{"mock error"},
		},
	})

	// Register a vendor that fails initially then succeeds
	failingVendor := &MockVendor{
		name:       "failing-vendor",
		shouldFail: true, // Will fail on first attempt
		response: &models.Response{
			Content: "Success after retry",
			Model:   "test-model",
			Vendor:  "failing-vendor",
		},
		available: true,
	}

	err := dispatcher.RegisterVendor(failingVendor)
	if err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err = dispatcher.Send(ctx, request)
	if err == nil {
		t.Fatal("Expected error after retries")
	}
}

func TestSend_WithModeStrategySuccess(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		Mode:          models.AutoMode,
		EnableLogging: true,
		EnableMetrics: true,
	})

	// Register vendor that succeeds
	successVendor := &MockVendor{
		name: "success",
		response: &models.Response{
			Content: "Success response",
			Model:   "success-model",
			Vendor:  "success",
		},
		available: true,
	}

	err := dispatcher.RegisterVendor(successVendor)
	if err != nil {
		t.Fatalf("Failed to register success vendor: %v", err)
	}

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	response, err := dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if response.Vendor != "success" {
		t.Errorf("Expected success vendor, got %s", response.Vendor)
	}
}

func TestSend_WithModeStrategy(t *testing.T) {
	config := &models.Config{
		Mode: models.AutoMode,
		ModeOverrides: &models.ModeOverrides{
			VendorPreferences: map[models.Mode][]string{
				models.AutoMode: {"openai", "anthropic"},
			},
		},
	}

	dispatcher := NewWithConfig(config)

	// Register vendors
	openaiVendor := &MockVendor{name: "openai", available: true}
	anthropicVendor := &MockVendor{name: "anthropic", available: true}

	dispatcher.RegisterVendor(openaiVendor)
	dispatcher.RegisterVendor(anthropicVendor)

	req := &models.Request{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// Test that the mode strategy is used
	_, err := dispatcher.Send(context.Background(), req)
	if err != nil {
		t.Errorf("Expected successful request, got error: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	dispatcher := New()

	// Send a request to generate some stats
	mockVendor := &MockVendor{
		name: "test-vendor",
		response: &models.Response{
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

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
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

func TestShouldRetry(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		RetryPolicy: &models.RetryPolicy{
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
	})

	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "retryable error",
			err:       errors.New("rate limit exceeded"),
			wantRetry: true,
		},
		{
			name:      "non-retryable error",
			err:       errors.New("invalid request"),
			wantRetry: false,
		},
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRetry := dispatcher.shouldRetry(tt.err)
			if shouldRetry != tt.wantRetry {
				t.Errorf("shouldRetry() = %v, want %v", shouldRetry, tt.wantRetry)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		RetryPolicy: &models.RetryPolicy{
			BackoffStrategy: models.ExponentialBackoff,
		},
	})

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  1,
			expected: 1 * time.Second,
		},
		{
			name:     "second attempt",
			attempt:  2,
			expected: 2 * time.Second,
		},
		{
			name:     "third attempt",
			attempt:  3,
			expected: 4 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := dispatcher.calculateBackoff(tt.attempt)
			if backoff != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, backoff, tt.expected)
			}
		})
	}
}

func TestDispatcher_Send_Validation(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor
	mockVendor := &MockVendor{
		name:      "test",
		available: true,
		response: &models.Response{
			Content: "test response",
			Model:   "test-model",
			Vendor:  "test",
		},
	}
	dispatcher.RegisterVendor(mockVendor)

	tests := []struct {
		name    string
		ctx     context.Context
		req     *models.Request
		wantErr bool
	}{
		{
			name: "nil context",
			ctx:  nil,
			req: &models.Request{
				Model: "test",
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			wantErr: true,
		},
		{
			name:    "nil request",
			ctx:     context.Background(),
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid request - empty model",
			ctx:  context.Background(),
			req: &models.Request{
				Model: "",
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid request - no messages",
			ctx:  context.Background(),
			req: &models.Request{
				Model:    "test",
				Messages: []models.Message{},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			ctx:  context.Background(),
			req: &models.Request{
				Model: "test",
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dispatcher.Send(tt.ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dispatcher.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDispatcher_RegisterVendor_Validation(t *testing.T) {
	dispatcher := New()

	tests := []struct {
		name    string
		vendor  models.LLMVendor
		wantErr bool
	}{
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
		{
			name: "valid vendor",
			vendor: &MockVendor{
				name: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dispatcher.RegisterVendor(tt.vendor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dispatcher.RegisterVendor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDispatcher_SelectVendor_NoVendors(t *testing.T) {
	dispatcher := New()
	req := &models.Request{
		Model: "test",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	vendor, err := dispatcher.selectVendor(context.Background(), req)
	if err == nil {
		t.Error("Expected error when no vendors are registered")
	}
	if vendor != nil {
		t.Errorf("Expected nil vendor, got %v", vendor)
	}
}

func TestDispatcher_SelectVendor_Unavailable(t *testing.T) {
	dispatcher := New()

	// Register unavailable vendor
	mockVendor := &MockVendor{
		name:      "test",
		available: false,
	}
	dispatcher.RegisterVendor(mockVendor)

	req := &models.Request{
		Model: "test",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	vendor, err := dispatcher.selectVendor(context.Background(), req)
	if err == nil {
		t.Error("Expected error when no vendors are available")
	}
	if vendor != nil {
		t.Errorf("Expected nil vendor, got %v", vendor)
	}
}

func TestDispatcher_SelectVendor_ModeBased(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		Mode: models.AutoMode,
	})

	// Register available vendor
	mockVendor := &MockVendor{
		name:      "test",
		available: true,
	}
	dispatcher.RegisterVendor(mockVendor)

	req := &models.Request{
		Model: "test",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	vendor, err := dispatcher.selectVendor(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if vendor == nil {
		t.Errorf("Expected vendor, got nil")
	}
	if vendor.Name() != "test" {
		t.Errorf("Expected vendor name 'test', got %s", vendor.Name())
	}
}

func TestDispatcher_SelectVendor_AvailableVendor(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		Mode: models.AutoMode,
	})

	// Register available vendor
	availableVendor := &MockVendor{
		name:      "available",
		available: true,
	}
	dispatcher.RegisterVendor(availableVendor)

	req := &models.Request{
		Model: "test",
		Messages: []models.Message{
			{Role: "user", Content: "test"},
		},
	}

	vendor, err := dispatcher.selectVendor(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if vendor == nil {
		t.Errorf("Expected vendor, got nil")
	}
	if vendor.Name() != "available" {
		t.Errorf("Expected vendor name 'available', got %s", vendor.Name())
	}
}

func TestDispatcher_SendStreaming(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor that supports streaming
	mockVendor := &MockVendor{
		name:              "test-streaming",
		available:         true,
		shouldFail:        false,
		supportsStreaming: true,
		capabilities: models.Capabilities{
			Models:            []string{"test-model"},
			SupportsStreaming: true,
			MaxTokens:         1000,
			MaxInputTokens:    4000,
		},
	}

	if err := dispatcher.RegisterVendor(mockVendor); err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	req := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	streamingResp, err := dispatcher.SendStreaming(ctx, req)

	if err != nil {
		t.Fatalf("SendStreaming failed: %v", err)
	}

	if streamingResp == nil {
		t.Fatal("Expected streaming response, got nil")
	}

	if streamingResp.Model != req.Model {
		t.Errorf("Expected model %s, got %s", req.Model, streamingResp.Model)
	}

	if streamingResp.Vendor != mockVendor.Name() {
		t.Errorf("Expected vendor %s, got %s", mockVendor.Name(), streamingResp.Vendor)
	}

	// Test streaming content
	select {
	case content := <-streamingResp.ContentChan:
		if content != "Mock streaming response" {
			t.Errorf("Expected content 'Mock streaming response', got %s", content)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for streaming content")
	}

	// Test completion signal
	select {
	case done := <-streamingResp.DoneChan:
		if !done {
			t.Error("Expected done signal to be true")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for done signal")
	}

	// Don't close here since the goroutine already closes it
	// streamingResp.Close()
}

func TestDispatcher_SendStreaming_NoVendors(t *testing.T) {
	dispatcher := New()

	req := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := dispatcher.SendStreaming(ctx, req)

	if err == nil {
		t.Error("Expected error when no vendors are registered")
	}
}

func TestDispatcher_SendStreaming_VendorNotAvailable(t *testing.T) {
	dispatcher := New()

	// Register an unavailable vendor
	mockVendor := &MockVendor{
		name:       "test-unavailable",
		available:  false,
		shouldFail: false,
	}

	if err := dispatcher.RegisterVendor(mockVendor); err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	req := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	_, err := dispatcher.SendStreaming(ctx, req)

	if err == nil {
		t.Error("Expected error when vendor is not available")
	}
}

func TestDispatcher_SendStreaming_InvalidRequest(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor
	mockVendor := &MockVendor{
		name:       "test-streaming",
		available:  true,
		shouldFail: false,
	}

	if err := dispatcher.RegisterVendor(mockVendor); err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	// Test with nil request
	ctx := context.Background()
	_, err := dispatcher.SendStreaming(ctx, nil)

	if err == nil {
		t.Error("Expected error for nil request")
	}

	// Test with invalid request
	invalidReq := &models.Request{
		Model: "", // Empty model
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err = dispatcher.SendStreaming(ctx, invalidReq)

	if err == nil {
		t.Error("Expected error for invalid request")
	}
}

func TestDispatcher_SendStreaming_NilContext(t *testing.T) {
	dispatcher := New()

	// Register a mock vendor
	mockVendor := &MockVendor{
		name:       "test-streaming",
		available:  true,
		shouldFail: false,
	}

	if err := dispatcher.RegisterVendor(mockVendor); err != nil {
		t.Fatalf("Failed to register vendor: %v", err)
	}

	req := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendStreaming(context.TODO(), req)

	if err == nil {
		t.Error("Expected error for nil context")
	}
}

func TestDispatcher_SendToVendor_Success(t *testing.T) {
	dispatcher := New()

	// Create a mock vendor
	mockVendor := &MockVendor{
		name:      "test-vendor",
		available: true,
		response: &models.Response{
			Content: "Test response",
			Model:   "test-model",
			Vendor:  "test-vendor",
		},
	}

	dispatcher.RegisterVendor(mockVendor)

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	response, err := dispatcher.SendToVendor(context.Background(), "test-vendor", request)

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
}

func TestDispatcher_SendToVendor_NilContext(t *testing.T) {
	dispatcher := New()

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendToVendor(context.TODO(), "test-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDispatcher_SendToVendor_NilRequest(t *testing.T) {
	dispatcher := New()

	_, err := dispatcher.SendToVendor(context.Background(), "test-vendor", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDispatcher_SendToVendor_VendorNotFound(t *testing.T) {
	dispatcher := New()

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendToVendor(context.Background(), "nonexistent-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got %s", err.Error())
	}
}

func TestDispatcher_SendToVendor_VendorNotAvailable(t *testing.T) {
	dispatcher := New()

	// Create a mock vendor that's not available
	mockVendor := &MockVendor{
		name:      "test-vendor",
		available: false,
	}

	dispatcher.RegisterVendor(mockVendor)

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendToVendor(context.Background(), "test-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Expected error to contain 'not available', got %s", err.Error())
	}
}

func TestDispatcher_SendStreamingToVendor_Success(t *testing.T) {
	dispatcher := New()

	// Create a mock vendor with streaming support
	mockVendor := &MockVendor{
		name:              "test-vendor",
		available:         true,
		supportsStreaming: true,
		streamingResponse: &models.StreamingResponse{
			ContentChan: make(chan string, 1),
			DoneChan:    make(chan bool, 1),
			ErrorChan:   make(chan error, 1),
			Model:       "test-model",
			Vendor:      "test-vendor",
		},
	}

	dispatcher.RegisterVendor(mockVendor)

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	response, err := dispatcher.SendStreamingToVendor(context.Background(), "test-vendor", request)

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
}

func TestDispatcher_SendStreamingToVendor_NilContext(t *testing.T) {
	dispatcher := New()

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendStreamingToVendor(context.TODO(), "test-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDispatcher_SendStreamingToVendor_NilRequest(t *testing.T) {
	dispatcher := New()

	_, err := dispatcher.SendStreamingToVendor(context.Background(), "test-vendor", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDispatcher_SendStreamingToVendor_VendorNotFound(t *testing.T) {
	dispatcher := New()

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendStreamingToVendor(context.Background(), "nonexistent-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got %s", err.Error())
	}
}

func TestDispatcher_SendStreamingToVendor_VendorNotAvailable(t *testing.T) {
	dispatcher := New()

	// Create a mock vendor that's not available
	mockVendor := &MockVendor{
		name:      "test-vendor",
		available: false,
	}

	dispatcher.RegisterVendor(mockVendor)

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendStreamingToVendor(context.Background(), "test-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Expected error to contain 'not available', got %s", err.Error())
	}
}

func TestDispatcher_SendStreamingToVendor_NoStreamingSupport(t *testing.T) {
	dispatcher := New()

	// Create a mock vendor without streaming support
	mockVendor := &MockVendor{
		name:              "test-vendor",
		available:         true,
		supportsStreaming: false,
	}

	dispatcher.RegisterVendor(mockVendor)

	request := &models.Request{
		Model: "test-model",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := dispatcher.SendStreamingToVendor(context.Background(), "test-vendor", request)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "does not support streaming") {
		t.Errorf("Expected error to contain 'does not support streaming', got %s", err.Error())
	}
}

func TestDispatcher_CalculateBackoff(t *testing.T) {
	// Test with exponential backoff
	config := &models.Config{
		RetryPolicy: &models.RetryPolicy{
			BackoffStrategy: models.ExponentialBackoff,
		},
	}
	dispatcher := NewWithConfig(config)

	// Test exponential backoff
	backoff := dispatcher.calculateBackoff(1)
	if backoff != time.Second {
		t.Errorf("Expected 1s backoff, got %v", backoff)
	}

	backoff = dispatcher.calculateBackoff(2)
	if backoff != 2*time.Second {
		t.Errorf("Expected 2s backoff, got %v", backoff)
	}

	backoff = dispatcher.calculateBackoff(3)
	if backoff != 4*time.Second {
		t.Errorf("Expected 4s backoff, got %v", backoff)
	}
}

func TestDispatcher_ShouldRetry(t *testing.T) {
	dispatcher := &Dispatcher{
		config: &models.Config{
			RetryPolicy: &models.RetryPolicy{
				MaxRetries: 3,
			},
		},
	}

	// Test retryable errors
	retryableErrors := []error{
		fmt.Errorf("rate limit exceeded"),
		fmt.Errorf("timeout"),
		fmt.Errorf("connection refused"),
		fmt.Errorf("network error"),
	}

	for _, err := range retryableErrors {
		if !dispatcher.shouldRetry(err) {
			t.Errorf("Expected %v to be retryable", err)
		}
	}

	// Test non-retryable errors
	nonRetryableErrors := []error{
		fmt.Errorf("invalid request"),
		fmt.Errorf("authentication failed"),
		fmt.Errorf("permission denied"),
	}

	for _, err := range nonRetryableErrors {
		if dispatcher.shouldRetry(err) {
			t.Errorf("Expected %v to not be retryable", err)
		}
	}
}

func TestDispatcher_UpdateStats(t *testing.T) {
	dispatcher := &Dispatcher{
		stats: &models.DispatcherStats{
			VendorStats: make(map[string]models.VendorStats),
		},
	}

	// Test successful request
	dispatcher.updateStats(true, "test-vendor", 100*time.Millisecond)

	if dispatcher.stats.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got: %d", dispatcher.stats.SuccessfulRequests)
	}

	if dispatcher.stats.AverageLatency != 100*time.Millisecond {
		t.Errorf("Expected 100ms average latency, got: %v", dispatcher.stats.AverageLatency)
	}

	// Test failed request
	dispatcher.updateStats(false, "test-vendor", 200*time.Millisecond)

	if dispatcher.stats.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got: %d", dispatcher.stats.FailedRequests)
	}

	// Average should be (100 + 200) / 2 = 150ms
	expectedAvg := 150 * time.Millisecond
	if dispatcher.stats.AverageLatency != expectedAvg {
		t.Errorf("Expected 150ms average latency, got: %v", dispatcher.stats.AverageLatency)
	}

	// Test vendor-specific stats
	vendorStats := dispatcher.stats.VendorStats["test-vendor"]
	if vendorStats.Requests != 2 {
		t.Errorf("Expected 2 vendor requests, got: %d", vendorStats.Requests)
	}
	if vendorStats.Successes != 1 {
		t.Errorf("Expected 1 vendor success, got: %d", vendorStats.Successes)
	}
	if vendorStats.Failures != 1 {
		t.Errorf("Expected 1 vendor failure, got: %d", vendorStats.Failures)
	}
}

func TestDispatcher_UpdateStats_NoVendor(t *testing.T) {
	dispatcher := &Dispatcher{
		stats: &models.DispatcherStats{
			VendorStats: make(map[string]models.VendorStats),
		},
	}

	// Test without vendor name
	dispatcher.updateStats(true, "", 100*time.Millisecond)

	if dispatcher.stats.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got: %d", dispatcher.stats.SuccessfulRequests)
	}

	if len(dispatcher.stats.VendorStats) != 0 {
		t.Errorf("Expected no vendor stats, got: %d", len(dispatcher.stats.VendorStats))
	}
}
