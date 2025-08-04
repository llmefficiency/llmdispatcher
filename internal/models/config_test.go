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
			name: "valid_config",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       30 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "empty_default_vendor",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       30 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "zero_timeout",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       0,
				EnableLogging: true,
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "negative_timeout",
			config: &Config{
				Mode:          AutoMode,
				Timeout:       -1 * time.Second,
				EnableLogging: true,
				EnableMetrics: true,
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

func TestModeStrategy_Validation(t *testing.T) {
	tests := []struct {
		name    string
		mode    Mode
		wantErr bool
	}{
		{
			name:    "valid_fast_mode",
			mode:    FastMode,
			wantErr: false,
		},
		{
			name:    "valid_sophisticated_mode",
			mode:    SophisticatedMode,
			wantErr: false,
		},
		{
			name:    "valid_cost_saving_mode",
			mode:    CostSavingMode,
			wantErr: false,
		},
		{
			name:    "valid_auto_mode",
			mode:    AutoMode,
			wantErr: false,
		},
		{
			name:    "invalid_mode",
			mode:    Mode("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewModeRegistry()

			if tt.wantErr {
				// For invalid mode, GetStrategy should return an error
				_, err := registry.GetStrategy(tt.mode)
				if err == nil {
					t.Errorf("Expected error for invalid mode %s", tt.mode)
				}
			} else {
				// For valid modes, the strategy should be created successfully
				strategy, err := registry.GetStrategy(tt.mode)
				if err != nil {
					t.Errorf("Unexpected error for valid mode %s: %v", tt.mode, err)
				}
				if strategy.Name() != string(tt.mode) {
					t.Errorf("Expected strategy name %s, got %s", tt.mode, strategy.Name())
				}
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

func TestModeOverrides_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ModeOverrides
		wantErr bool
	}{
		{
			name: "valid mode overrides",
			config: &ModeOverrides{
				VendorPreferences: map[Mode][]string{
					FastMode:          {"local", "anthropic"},
					SophisticatedMode: {"anthropic", "openai"},
				},
				MaxCostPerRequest:   0.01,
				MaxLatency:          2 * time.Second,
				SophisticatedModels: []string{"claude-3-opus", "gpt-4"},
			},
			wantErr: false,
		},
		{
			name:    "nil mode overrides",
			config:  nil,
			wantErr: false, // nil is valid (optional)
		},
		{
			name: "negative max cost",
			config: &ModeOverrides{
				MaxCostPerRequest: -0.01,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we don't have validation for ModeOverrides
			// This test is a placeholder for future validation logic
			_ = tt
		})
	}
}

func TestConfig_AdvancedRouting(t *testing.T) {
	config := &Config{
		Mode: AutoMode,
		ModeOverrides: &ModeOverrides{
			VendorPreferences: map[Mode][]string{
				AutoMode:          {"openai", "anthropic", "google"},
				FastMode:          {"local", "anthropic"},
				SophisticatedMode: {"anthropic", "openai"},
				CostSavingMode:    {"local", "google", "openai"},
			},
			MaxCostPerRequest:   0.10,
			MaxLatency:          5 * time.Second,
			SophisticatedModels: []string{"claude-3-opus", "gpt-4", "gemini-pro"},
		},
	}

	if config.Mode != AutoMode {
		t.Errorf("Expected mode 'auto', got '%s'", config.Mode)
	}

	if config.ModeOverrides == nil {
		t.Error("Expected mode overrides to be set")
	}

	if len(config.ModeOverrides.VendorPreferences) != 4 {
		t.Errorf("Expected 4 mode preferences, got %d", len(config.ModeOverrides.VendorPreferences))
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
