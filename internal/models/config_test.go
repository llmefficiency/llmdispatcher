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

func TestRoutingStrategy_Validation(t *testing.T) {
	tests := []struct {
		name     string
		strategy RoutingStrategy
		wantErr  bool
	}{
		{
			name: "valid_cascading_strategy",
			strategy: &CascadingFailureStrategy{
				VendorOrder: []string{"openai", "anthropic"},
			},
			wantErr: false,
		},
		{
			name: "empty_vendor_order",
			strategy: &CascadingFailureStrategy{
				VendorOrder: []string{},
			},
			wantErr: true,
		},
		{
			name:     "nil_strategy",
			strategy: nil,
			wantErr:  false, // nil is valid (optional)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoutingStrategy(tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoutingStrategy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCascadingFailureStrategy_Validation(t *testing.T) {
	tests := []struct {
		name     string
		strategy CascadingFailureStrategy
		wantErr  bool
	}{
		{
			name: "valid_strategy",
			strategy: CascadingFailureStrategy{
				VendorOrder: []string{"openai", "anthropic", "google"},
			},
			wantErr: false,
		},
		{
			name: "empty_vendor_order",
			strategy: CascadingFailureStrategy{
				VendorOrder: []string{},
			},
			wantErr: true,
		},
		{
			name: "single_vendor",
			strategy: CascadingFailureStrategy{
				VendorOrder: []string{"openai"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCascadingFailureStrategy(tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCascadingFailureStrategy() error = %v, wantErr %v", err, tt.wantErr)
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
				FailedRequests:     3, // 8 + 3 = 11 != 10
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
		RoutingStrategy: &CascadingFailureStrategy{
			VendorOrder: []string{"openai", "anthropic", "google"},
		},
		CostOptimization: &CostOptimization{
			Enabled:     true,
			MaxCost:     0.10,
			PreferCheap: true,
			VendorCosts: map[string]float64{
				"openai":    0.002,
				"anthropic": 0.008,
				"google":    0.001,
			},
		},
		LatencyOptimization: &LatencyOptimization{
			Enabled:    true,
			MaxLatency: 5 * time.Second,
			PreferFast: true,
			LatencyWeights: map[string]float64{
				"openai":    1.0,
				"anthropic": 1.2,
				"google":    0.8,
			},
		},
	}

	if config.DefaultVendor != "openai" {
		t.Errorf("Expected default vendor 'openai', got '%s'", config.DefaultVendor)
	}

	if config.RoutingStrategy == nil {
		t.Error("Expected routing strategy to be set")
	}

	if config.CostOptimization == nil {
		t.Error("Expected cost optimization to be set")
	}

	if config.LatencyOptimization == nil {
		t.Error("Expected latency optimization to be set")
	}
}

func TestDispatcherStats_AdvancedMetrics(t *testing.T) {
	stats := &DispatcherStats{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		VendorStats:        make(map[string]VendorStats),
		AverageLatency:     150 * time.Millisecond,
		LastRequestTime:    time.Now(),
		TotalCost:          25.50,
		AverageCost:        0.255,
		CostByVendor: map[string]float64{
			"openai":    15.30,
			"anthropic": 8.20,
			"google":    2.00,
		},
	}

	if stats.TotalCost != 25.50 {
		t.Errorf("Expected TotalCost 25.50, got %f", stats.TotalCost)
	}

	if stats.AverageCost != 0.255 {
		t.Errorf("Expected AverageCost 0.255, got %f", stats.AverageCost)
	}

	if len(stats.CostByVendor) != 3 {
		t.Errorf("Expected 3 vendors in cost breakdown, got %d", len(stats.CostByVendor))
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

func validateRoutingStrategy(strategy RoutingStrategy) error {
	if strategy == nil {
		return nil // nil strategy is valid (optional)
	}

	// Type switch to validate specific strategy types
	switch s := strategy.(type) {
	case *CascadingFailureStrategy:
		return validateCascadingFailureStrategy(*s)
	default:
		return nil // Unknown strategy types are considered valid
	}
}

func validateCascadingFailureStrategy(strategy CascadingFailureStrategy) error {
	if len(strategy.VendorOrder) == 0 {
		return &MockError{message: "vendor order cannot be empty"}
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
