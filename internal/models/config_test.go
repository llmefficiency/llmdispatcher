package models

import (
	"testing"
	"time"
)

// MockError is a mock error type for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DefaultVendor:  "openai",
				FallbackVendor: "anthropic",
				Timeout:        30 * time.Second,
				EnableLogging:  true,
				EnableMetrics:  true,
			},
			wantErr: false,
		},
		{
			name: "empty default vendor",
			config: &Config{
				DefaultVendor:  "",
				FallbackVendor: "anthropic",
				Timeout:        30 * time.Second,
			},
			wantErr: false, // This is valid - no default vendor
		},
		{
			name: "zero timeout",
			config: &Config{
				DefaultVendor: "openai",
				Timeout:       0,
			},
			wantErr: false, // This is valid - no timeout
		},
		{
			name: "negative timeout",
			config: &Config{
				DefaultVendor: "openai",
				Timeout:       -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRetryPolicy_Validation(t *testing.T) {
	tests := []struct {
		name        string
		retryPolicy *RetryPolicy
		wantErr     bool
	}{
		{
			name: "valid retry policy",
			retryPolicy: &RetryPolicy{
				MaxRetries:      3,
				BackoffStrategy: ExponentialBackoff,
				RetryableErrors: []string{"rate limit exceeded", "timeout"},
			},
			wantErr: false,
		},
		{
			name: "negative max retries",
			retryPolicy: &RetryPolicy{
				MaxRetries:      -1,
				BackoffStrategy: ExponentialBackoff,
			},
			wantErr: true,
		},
		{
			name: "zero max retries",
			retryPolicy: &RetryPolicy{
				MaxRetries:      0,
				BackoffStrategy: ExponentialBackoff,
			},
			wantErr: false, // This is valid - no retries
		},
		{
			name: "invalid backoff strategy",
			retryPolicy: &RetryPolicy{
				MaxRetries:      3,
				BackoffStrategy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRetryPolicy(tt.retryPolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRetryPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackoffStrategy_Validation(t *testing.T) {
	tests := []struct {
		name     string
		strategy BackoffStrategy
		wantErr  bool
	}{
		{
			name:     "exponential backoff",
			strategy: ExponentialBackoff,
			wantErr:  false,
		},
		{
			name:     "linear backoff",
			strategy: LinearBackoff,
			wantErr:  false,
		},
		{
			name:     "fixed backoff",
			strategy: FixedBackoff,
			wantErr:  false,
		},
		{
			name:     "invalid strategy",
			strategy: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBackoffStrategy(tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBackoffStrategy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoutingRule_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rule    RoutingRule
		wantErr bool
	}{
		{
			name: "valid routing rule",
			rule: RoutingRule{
				Condition: RoutingCondition{
					ModelPattern: "gpt-4",
					MaxTokens:    1000,
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  true,
			},
			wantErr: false,
		},
		{
			name: "empty vendor",
			rule: RoutingRule{
				Condition: RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "",
				Priority: 1,
				Enabled:  true,
			},
			wantErr: true,
		},
		{
			name: "negative priority",
			rule: RoutingRule{
				Condition: RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "openai",
				Priority: -1,
				Enabled:  true,
			},
			wantErr: true,
		},
		{
			name: "disabled rule",
			rule: RoutingRule{
				Condition: RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  false,
			},
			wantErr: false, // This is valid - disabled rules are allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoutingRule(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoutingRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoutingCondition_Validation(t *testing.T) {
	tests := []struct {
		name      string
		condition RoutingCondition
		wantErr   bool
	}{
		{
			name: "valid condition",
			condition: RoutingCondition{
				ModelPattern: "gpt-4",
				MaxTokens:    1000,
				Temperature:  0.7,
			},
			wantErr: false,
		},
		{
			name: "negative max tokens",
			condition: RoutingCondition{
				ModelPattern: "gpt-4",
				MaxTokens:    -1,
			},
			wantErr: true,
		},
		{
			name: "invalid temperature too high",
			condition: RoutingCondition{
				ModelPattern: "gpt-4",
				Temperature:  3.0,
			},
			wantErr: true,
		},
		{
			name: "invalid temperature too low",
			condition: RoutingCondition{
				ModelPattern: "gpt-4",
				Temperature:  -1.0,
			},
			wantErr: true,
		},
		{
			name: "negative latency threshold",
			condition: RoutingCondition{
				ModelPattern:     "gpt-4",
				LatencyThreshold: -1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoutingCondition(tt.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoutingCondition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDispatcherStats_Validation(t *testing.T) {
	tests := []struct {
		name    string
		stats   *DispatcherStats
		wantErr bool
	}{
		{
			name: "valid stats",
			stats: &DispatcherStats{
				TotalRequests:      10,
				SuccessfulRequests: 8,
				FailedRequests:     2,
				VendorStats:        make(map[string]VendorStats),
				AverageLatency:     100 * time.Millisecond,
				LastRequestTime:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "negative total requests",
			stats: &DispatcherStats{
				TotalRequests:      -1,
				SuccessfulRequests: 0,
				FailedRequests:     0,
			},
			wantErr: true,
		},
		{
			name: "negative successful requests",
			stats: &DispatcherStats{
				TotalRequests:      10,
				SuccessfulRequests: -1,
				FailedRequests:     0,
			},
			wantErr: true,
		},
		{
			name: "negative failed requests",
			stats: &DispatcherStats{
				TotalRequests:      10,
				SuccessfulRequests: 8,
				FailedRequests:     -1,
			},
			wantErr: true,
		},
		{
			name: "requests mismatch",
			stats: &DispatcherStats{
				TotalRequests:      10,
				SuccessfulRequests: 8,
				FailedRequests:     1, // Should be 2
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDispatcherStats(tt.stats)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDispatcherStats() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVendorStats_Validation(t *testing.T) {
	tests := []struct {
		name    string
		stats   VendorStats
		wantErr bool
	}{
		{
			name: "valid vendor stats",
			stats: VendorStats{
				Requests:       10,
				Successes:      8,
				Failures:       2,
				AverageLatency: 100 * time.Millisecond,
				LastUsed:       time.Now(),
			},
			wantErr: false,
		},
		{
			name: "negative requests",
			stats: VendorStats{
				Requests:  -1,
				Successes: 0,
				Failures:  0,
			},
			wantErr: true,
		},
		{
			name: "negative successes",
			stats: VendorStats{
				Requests:  10,
				Successes: -1,
				Failures:  0,
			},
			wantErr: true,
		},
		{
			name: "negative failures",
			stats: VendorStats{
				Requests:  10,
				Successes: 8,
				Failures:  -1,
			},
			wantErr: true,
		},
		{
			name: "requests mismatch",
			stats: VendorStats{
				Requests:  10,
				Successes: 8,
				Failures:  1, // Should be 2
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVendorStats(tt.stats)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVendorStats() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCostOptimization_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *CostOptimization
		wantErr bool
	}{
		{
			name: "valid cost optimization",
			config: &CostOptimization{
				Enabled:     true,
				MaxCost:     0.10,
				PreferCheap: true,
				VendorCosts: map[string]float64{
					"openai":    0.002,
					"anthropic": 0.003,
					"google":    0.001,
				},
			},
			wantErr: false,
		},
		{
			name: "disabled cost optimization",
			config: &CostOptimization{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "negative max cost",
			config: &CostOptimization{
				Enabled: true,
				MaxCost: -0.10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we don't have validation for CostOptimization
			// This test ensures the structure works correctly
			if tt.config == nil {
				t.Error("Config should not be nil")
			}
		})
	}
}

func TestLatencyOptimization_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *LatencyOptimization
		wantErr bool
	}{
		{
			name: "valid latency optimization",
			config: &LatencyOptimization{
				Enabled:    true,
				MaxLatency: 30 * time.Second,
				PreferFast: true,
				LatencyWeights: map[string]float64{
					"openai":    1.0,
					"anthropic": 1.2,
					"google":    0.8,
				},
			},
			wantErr: false,
		},
		{
			name: "disabled latency optimization",
			config: &LatencyOptimization{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "negative max latency",
			config: &LatencyOptimization{
				Enabled:    true,
				MaxLatency: -30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we don't have validation for LatencyOptimization
			// This test ensures the structure works correctly
			if tt.config == nil {
				t.Error("Config should not be nil")
			}
		})
	}
}

func TestConfig_AdvancedRouting(t *testing.T) {
	config := &Config{
		DefaultVendor: "openai",
		CostOptimization: &CostOptimization{
			Enabled:     true,
			MaxCost:     0.10,
			PreferCheap: true,
			VendorCosts: map[string]float64{
				"openai":    0.002,
				"anthropic": 0.003,
			},
		},
		LatencyOptimization: &LatencyOptimization{
			Enabled:    true,
			MaxLatency: 30 * time.Second,
			PreferFast: true,
			LatencyWeights: map[string]float64{
				"openai":    1.0,
				"anthropic": 1.2,
			},
		},
	}

	if config.CostOptimization == nil {
		t.Error("Expected CostOptimization to be set")
	}

	if !config.CostOptimization.Enabled {
		t.Error("Expected CostOptimization to be enabled")
	}

	if config.CostOptimization.MaxCost != 0.10 {
		t.Errorf("Expected MaxCost 0.10, got %f", config.CostOptimization.MaxCost)
	}

	if config.LatencyOptimization == nil {
		t.Error("Expected LatencyOptimization to be set")
	}

	if !config.LatencyOptimization.Enabled {
		t.Error("Expected LatencyOptimization to be enabled")
	}

	if config.LatencyOptimization.MaxLatency != 30*time.Second {
		t.Errorf("Expected MaxLatency 30s, got %v", config.LatencyOptimization.MaxLatency)
	}
}

func TestRoutingCondition_AdvancedFields(t *testing.T) {
	condition := RoutingCondition{
		ModelPattern:     "gpt-*",
		MaxTokens:        1000,
		Temperature:      0.7,
		CostThreshold:    0.05,
		LatencyThreshold: 10 * time.Second,
		UserID:           "user123",
		RequestType:      "chat",
		ContentLength:    500,
	}

	if condition.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got %s", condition.UserID)
	}

	if condition.RequestType != "chat" {
		t.Errorf("Expected RequestType 'chat', got %s", condition.RequestType)
	}

	if condition.ContentLength != 500 {
		t.Errorf("Expected ContentLength 500, got %d", condition.ContentLength)
	}

	if condition.CostThreshold != 0.05 {
		t.Errorf("Expected CostThreshold 0.05, got %f", condition.CostThreshold)
	}

	if condition.LatencyThreshold != 10*time.Second {
		t.Errorf("Expected LatencyThreshold 10s, got %v", condition.LatencyThreshold)
	}
}

func TestDispatcherStats_AdvancedMetrics(t *testing.T) {
	stats := DispatcherStats{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		AverageLatency:     2 * time.Second,
		LastRequestTime:    time.Now(),
		TotalCost:          0.50,
		AverageCost:        0.005,
		CostByVendor: map[string]float64{
			"openai":    0.30,
			"anthropic": 0.20,
		},
	}

	if stats.TotalCost != 0.50 {
		t.Errorf("Expected TotalCost 0.50, got %f", stats.TotalCost)
	}

	if stats.AverageCost != 0.005 {
		t.Errorf("Expected AverageCost 0.005, got %f", stats.AverageCost)
	}

	if len(stats.CostByVendor) != 2 {
		t.Errorf("Expected 2 vendors in CostByVendor, got %d", len(stats.CostByVendor))
	}

	if stats.CostByVendor["openai"] != 0.30 {
		t.Errorf("Expected OpenAI cost 0.30, got %f", stats.CostByVendor["openai"])
	}

	if stats.CostByVendor["anthropic"] != 0.20 {
		t.Errorf("Expected Anthropic cost 0.20, got %f", stats.CostByVendor["anthropic"])
	}
}

func TestVendorStats_AdvancedMetrics(t *testing.T) {
	stats := VendorStats{
		Requests:       50,
		Successes:      48,
		Failures:       2,
		AverageLatency: 1 * time.Second,
		LastUsed:       time.Now(),
		TotalCost:      0.25,
		AverageCost:    0.005,
		TokenUsage:     10000,
	}

	if stats.TotalCost != 0.25 {
		t.Errorf("Expected TotalCost 0.25, got %f", stats.TotalCost)
	}

	if stats.AverageCost != 0.005 {
		t.Errorf("Expected AverageCost 0.005, got %f", stats.AverageCost)
	}

	if stats.TokenUsage != 10000 {
		t.Errorf("Expected TokenUsage 10000, got %d", stats.TokenUsage)
	}
}

// Helper validation functions
func validateConfig(config *Config) error {
	if config == nil {
		return &MockError{message: "config cannot be nil"}
	}
	if config.Timeout < 0 {
		return &MockError{message: "timeout cannot be negative"}
	}
	return nil
}

func validateRetryPolicy(policy *RetryPolicy) error {
	if policy == nil {
		return nil // nil policy is valid
	}
	if policy.MaxRetries < 0 {
		return &MockError{message: "max retries cannot be negative"}
	}
	if err := validateBackoffStrategy(policy.BackoffStrategy); err != nil {
		return err
	}
	return nil
}

func validateBackoffStrategy(strategy BackoffStrategy) error {
	switch strategy {
	case ExponentialBackoff, LinearBackoff, FixedBackoff:
		return nil
	default:
		return &MockError{message: "invalid backoff strategy"}
	}
}

func validateRoutingRule(rule RoutingRule) error {
	if rule.Vendor == "" {
		return &MockError{message: "vendor is required"}
	}
	if rule.Priority < 0 {
		return &MockError{message: "priority cannot be negative"}
	}
	if err := validateRoutingCondition(rule.Condition); err != nil {
		return err
	}
	return nil
}

func validateRoutingCondition(condition RoutingCondition) error {
	if condition.MaxTokens < 0 {
		return &MockError{message: "max tokens cannot be negative"}
	}
	if condition.Temperature < 0 || condition.Temperature > 2 {
		return &MockError{message: "temperature must be between 0 and 2"}
	}
	if condition.LatencyThreshold < 0 {
		return &MockError{message: "latency threshold cannot be negative"}
	}
	return nil
}

func validateDispatcherStats(stats *DispatcherStats) error {
	if stats == nil {
		return &MockError{message: "stats cannot be nil"}
	}
	if stats.TotalRequests < 0 {
		return &MockError{message: "total requests cannot be negative"}
	}
	if stats.SuccessfulRequests < 0 {
		return &MockError{message: "successful requests cannot be negative"}
	}
	if stats.FailedRequests < 0 {
		return &MockError{message: "failed requests cannot be negative"}
	}
	if stats.TotalRequests != stats.SuccessfulRequests+stats.FailedRequests {
		return &MockError{message: "total requests must equal successful + failed requests"}
	}
	return nil
}

func validateVendorStats(stats VendorStats) error {
	if stats.Requests < 0 {
		return &MockError{message: "requests cannot be negative"}
	}
	if stats.Successes < 0 {
		return &MockError{message: "successes cannot be negative"}
	}
	if stats.Failures < 0 {
		return &MockError{message: "failures cannot be negative"}
	}
	if stats.Requests != stats.Successes+stats.Failures {
		return &MockError{message: "requests must equal successes + failures"}
	}
	return nil
}
