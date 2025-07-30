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
