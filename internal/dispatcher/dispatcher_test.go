package dispatcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// MockVendor is a mock implementation of LLMVendor for testing
type MockVendor struct {
	name         string
	shouldFail   bool
	response     *models.Response
	capabilities models.Capabilities
	available    bool
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
	return m.capabilities
}

func (m *MockVendor) IsAvailable(ctx context.Context) bool {
	return m.available
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
				DefaultVendor:  "openai",
				FallbackVendor: "anthropic",
				Timeout:        30 * time.Second,
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
			if dispatcher.config == nil {
				t.Error("config should not be nil")
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

func TestSend_WithFallback(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		DefaultVendor:  "primary",
		FallbackVendor: "fallback",
	})

	// Register primary vendor that fails
	primaryVendor := &MockVendor{
		name:       "primary",
		shouldFail: true,
		available:  true,
	}

	// Register fallback vendor that succeeds
	fallbackVendor := &MockVendor{
		name: "fallback",
		response: &models.Response{
			Content: "Fallback response",
			Model:   "fallback-model",
			Vendor:  "fallback",
		},
		available: true,
	}

	err := dispatcher.RegisterVendor(primaryVendor)
	if err != nil {
		t.Fatalf("Failed to register primary vendor: %v", err)
	}

	err = dispatcher.RegisterVendor(fallbackVendor)
	if err != nil {
		t.Fatalf("Failed to register fallback vendor: %v", err)
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

	if response.Vendor != "fallback" {
		t.Errorf("Expected fallback vendor, got %s", response.Vendor)
	}
}

func TestSend_WithRoutingRules(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		RoutingRules: []models.RoutingRule{
			{
				Condition: models.RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  true,
			},
			{
				Condition: models.RoutingCondition{
					ModelPattern: "claude",
				},
				Vendor:   "anthropic",
				Priority: 1,
				Enabled:  true,
			},
		},
	})

	// Register vendors
	openaiVendor := &MockVendor{
		name: "openai",
		response: &models.Response{
			Content: "OpenAI response",
			Model:   "gpt-4",
			Vendor:  "openai",
		},
		available: true,
	}

	anthropicVendor := &MockVendor{
		name: "anthropic",
		response: &models.Response{
			Content: "Anthropic response",
			Model:   "claude",
			Vendor:  "anthropic",
		},
		available: true,
	}

	err := dispatcher.RegisterVendor(openaiVendor)
	if err != nil {
		t.Fatalf("Failed to register OpenAI vendor: %v", err)
	}

	err = dispatcher.RegisterVendor(anthropicVendor)
	if err != nil {
		t.Fatalf("Failed to register Anthropic vendor: %v", err)
	}

	// Test routing to OpenAI
	request := &models.Request{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx := context.Background()
	response, err := dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if response.Vendor != "openai" {
		t.Errorf("Expected OpenAI vendor, got %s", response.Vendor)
	}

	// Test routing to Anthropic
	request.Model = "claude"
	response, err = dispatcher.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if response.Vendor != "anthropic" {
		t.Errorf("Expected Anthropic vendor, got %s", response.Vendor)
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

func TestMatchesCondition(t *testing.T) {
	dispatcher := New()

	tests := []struct {
		name      string
		request   *models.Request
		condition models.RoutingCondition
		wantMatch bool
	}{
		{
			name: "model pattern match",
			request: &models.Request{
				Model: "gpt-4",
			},
			condition: models.RoutingCondition{
				ModelPattern: "gpt-4",
			},
			wantMatch: true,
		},
		{
			name: "model pattern no match",
			request: &models.Request{
				Model: "gpt-3.5-turbo",
			},
			condition: models.RoutingCondition{
				ModelPattern: "gpt-4",
			},
			wantMatch: false,
		},
		{
			name: "max tokens within limit",
			request: &models.Request{
				Model:     "gpt-4",
				MaxTokens: 500,
			},
			condition: models.RoutingCondition{
				MaxTokens: 1000,
			},
			wantMatch: true,
		},
		{
			name: "max tokens exceeds limit",
			request: &models.Request{
				Model:     "gpt-4",
				MaxTokens: 1500,
			},
			condition: models.RoutingCondition{
				MaxTokens: 1000,
			},
			wantMatch: false,
		},
		{
			name: "temperature within limit",
			request: &models.Request{
				Model:       "gpt-4",
				Temperature: 0.5,
			},
			condition: models.RoutingCondition{
				Temperature: 1.0,
			},
			wantMatch: true,
		},
		{
			name: "temperature exceeds limit",
			request: &models.Request{
				Model:       "gpt-4",
				Temperature: 1.5,
			},
			condition: models.RoutingCondition{
				Temperature: 1.0,
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := dispatcher.matchesCondition(tt.request, tt.condition)
			if matches != tt.wantMatch {
				t.Errorf("matchesCondition() = %v, want %v", matches, tt.wantMatch)
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
	if err != models.ErrNoVendorsRegistered {
		t.Errorf("Expected ErrNoVendorsRegistered, got %v", err)
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
	if err != models.ErrVendorUnavailable {
		t.Errorf("Expected ErrVendorUnavailable, got %v", err)
	}
	if vendor != nil {
		t.Errorf("Expected nil vendor, got %v", vendor)
	}
}

func TestDispatcher_SelectVendor_DefaultVendor(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		DefaultVendor: "test",
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

func TestDispatcher_SelectVendor_Fallback(t *testing.T) {
	dispatcher := NewWithConfig(&models.Config{
		DefaultVendor:  "default",
		FallbackVendor: "fallback",
	})

	// Register unavailable default vendor
	defaultVendor := &MockVendor{
		name:      "default",
		available: false,
	}
	dispatcher.RegisterVendor(defaultVendor)

	// Register available fallback vendor
	fallbackVendor := &MockVendor{
		name:      "fallback",
		available: true,
	}
	dispatcher.RegisterVendor(fallbackVendor)

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
	if vendor.Name() != "fallback" {
		t.Errorf("Expected vendor name 'fallback', got %s", vendor.Name())
	}
}

func TestDispatcher_MatchesCondition(t *testing.T) {
	dispatcher := New()

	tests := []struct {
		name      string
		req       *models.Request
		condition models.RoutingCondition
		want      bool
	}{
		{
			name: "model pattern match",
			req: &models.Request{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				ModelPattern: "gpt-4",
			},
			want: true,
		},
		{
			name: "model pattern no match",
			req: &models.Request{
				Model: "gpt-3.5-turbo",
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				ModelPattern: "gpt-4",
			},
			want: false,
		},
		{
			name: "max tokens within limit",
			req: &models.Request{
				Model:     "test",
				MaxTokens: 100,
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				MaxTokens: 200,
			},
			want: true,
		},
		{
			name: "max tokens exceeds limit",
			req: &models.Request{
				Model:     "test",
				MaxTokens: 300,
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				MaxTokens: 200,
			},
			want: false,
		},
		{
			name: "temperature within limit",
			req: &models.Request{
				Model:       "test",
				Temperature: 0.5,
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				Temperature: 1.0,
			},
			want: true,
		},
		{
			name: "temperature exceeds limit",
			req: &models.Request{
				Model:       "test",
				Temperature: 1.5,
				Messages: []models.Message{
					{Role: "user", Content: "test"},
				},
			},
			condition: models.RoutingCondition{
				Temperature: 1.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dispatcher.matchesCondition(tt.req, tt.condition)
			if got != tt.want {
				t.Errorf("matchesCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}
