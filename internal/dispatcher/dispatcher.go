package dispatcher

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// Dispatcher manages routing of LLM requests to different vendors
type Dispatcher struct {
	vendors    map[string]models.LLMVendor
	config     *models.Config
	stats      *models.DispatcherStats
	statsMutex sync.RWMutex
	logger     *log.Logger
}

// New creates a new dispatcher with default configuration
func New() *Dispatcher {
	return NewWithConfig(&models.Config{
		DefaultVendor: "",
		Timeout:       30 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
	})
}

// NewWithConfig creates a new dispatcher with custom configuration
func NewWithConfig(config *models.Config) *Dispatcher {
	if config == nil {
		config = &models.Config{}
	}

	dispatcher := &Dispatcher{
		vendors: make(map[string]models.LLMVendor),
		config:  config,
		stats: &models.DispatcherStats{
			VendorStats: make(map[string]models.VendorStats),
		},
		logger: log.New(log.Writer(), "[LLMDispatcher] ", log.LstdFlags),
	}

	return dispatcher
}

// RegisterVendor registers a new vendor with the dispatcher
func (d *Dispatcher) RegisterVendor(vendor models.LLMVendor) error {
	if vendor == nil {
		return fmt.Errorf("%w: vendor cannot be nil", models.ErrInvalidConfig)
	}

	name := vendor.Name()
	if name == "" {
		return fmt.Errorf("%w: vendor name cannot be empty", models.ErrInvalidConfig)
	}

	d.vendors[name] = vendor
	d.logger.Printf("Registered vendor: %s", name)
	return nil
}

// Send sends a request to the appropriate vendor based on routing rules
func (d *Dispatcher) Send(ctx context.Context, req *models.Request) (*models.Response, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: context cannot be nil", models.ErrInvalidRequest)
	}

	if req == nil {
		return nil, fmt.Errorf("%w: request cannot be nil", models.ErrInvalidRequest)
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	start := time.Now()

	// Update stats
	d.statsMutex.Lock()
	d.stats.TotalRequests++
	d.stats.LastRequestTime = time.Now()
	d.statsMutex.Unlock()

	// Apply timeout if configured
	if d.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.config.Timeout)
		defer cancel()
	}

	// Determine which vendor to use
	vendor, err := d.selectVendor(ctx, req)
	if err != nil {
		d.updateStats(false, "", time.Since(start))
		return nil, fmt.Errorf("failed to select vendor: %w", err)
	}

	// Send request with retry logic
	response, err := d.sendWithRetry(ctx, vendor, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start))
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start))
	return response, nil
}

// SendStreaming sends a streaming request to the appropriate vendor
func (d *Dispatcher) SendStreaming(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: context cannot be nil", models.ErrInvalidRequest)
	}

	if req == nil {
		return nil, fmt.Errorf("%w: request cannot be nil", models.ErrInvalidRequest)
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Set streaming flag
	req.Stream = true

	start := time.Now()

	// Update stats
	d.statsMutex.Lock()
	d.stats.TotalRequests++
	d.stats.LastRequestTime = time.Now()
	d.statsMutex.Unlock()

	// Apply timeout if configured
	if d.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.config.Timeout)
		defer cancel()
	}

	// Determine which vendor to use
	vendor, err := d.selectVendor(ctx, req)
	if err != nil {
		d.updateStats(false, "", time.Since(start))
		return nil, fmt.Errorf("failed to select vendor: %w", err)
	}

	// Check if vendor supports streaming
	if !vendor.GetCapabilities().SupportsStreaming {
		return nil, fmt.Errorf("vendor %s does not support streaming", vendor.Name())
	}

	// Send streaming request
	streamingResp, err := vendor.SendStreamingRequest(ctx, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start))
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start))
	return streamingResp, nil
}

// SendToVendor sends a request to a specific vendor
func (d *Dispatcher) SendToVendor(ctx context.Context, vendorName string, req *models.Request) (*models.Response, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: context cannot be nil", models.ErrInvalidRequest)
	}

	if req == nil {
		return nil, fmt.Errorf("%w: request cannot be nil", models.ErrInvalidRequest)
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	start := time.Now()

	// Update stats
	d.statsMutex.Lock()
	d.stats.TotalRequests++
	d.stats.LastRequestTime = time.Now()
	d.statsMutex.Unlock()

	// Get the specified vendor
	vendor, exists := d.vendors[vendorName]
	if !exists {
		return nil, fmt.Errorf("vendor %s not found", vendorName)
	}

	// Check if vendor is available
	if !vendor.IsAvailable(ctx) {
		return nil, fmt.Errorf("vendor %s is not available", vendorName)
	}

	// Send request
	response, err := d.sendWithRetry(ctx, vendor, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start))
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start))
	return response, nil
}

// SendStreamingToVendor sends a streaming request to a specific vendor
func (d *Dispatcher) SendStreamingToVendor(ctx context.Context, vendorName string, req *models.Request) (*models.StreamingResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: context cannot be nil", models.ErrInvalidRequest)
	}

	if req == nil {
		return nil, fmt.Errorf("%w: request cannot be nil", models.ErrInvalidRequest)
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Set streaming flag
	req.Stream = true

	start := time.Now()

	// Update stats
	d.statsMutex.Lock()
	d.stats.TotalRequests++
	d.stats.LastRequestTime = time.Now()
	d.statsMutex.Unlock()

	// Get the specified vendor
	vendor, exists := d.vendors[vendorName]
	if !exists {
		return nil, fmt.Errorf("vendor %s not found", vendorName)
	}

	// Check if vendor is available
	if !vendor.IsAvailable(ctx) {
		return nil, fmt.Errorf("vendor %s is not available", vendorName)
	}

	// Check if vendor supports streaming
	if !vendor.GetCapabilities().SupportsStreaming {
		return nil, fmt.Errorf("vendor %s does not support streaming", vendorName)
	}

	// Send streaming request
	streamingResp, err := vendor.SendStreamingRequest(ctx, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start))
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start))
	return streamingResp, nil
}

// selectVendor determines which vendor should handle the request
func (d *Dispatcher) selectVendor(ctx context.Context, req *models.Request) (models.LLMVendor, error) {
	if len(d.vendors) == 0 {
		return nil, models.ErrNoVendorsRegistered
	}

	// Check routing rules first
	if len(d.config.RoutingRules) > 0 {
		if vendor := d.applyRoutingRules(req); vendor != nil {
			if vendor.IsAvailable(ctx) {
				return vendor, nil
			}
			d.logger.Printf("Vendor %s from routing rule is not available", vendor.Name())
		}
	}

	// Use default vendor if specified
	if d.config.DefaultVendor != "" {
		if vendor, exists := d.vendors[d.config.DefaultVendor]; exists {
			if vendor.IsAvailable(ctx) {
				return vendor, nil
			}
			d.logger.Printf("Default vendor %s is not available", d.config.DefaultVendor)
		} else {
			d.logger.Printf("Default vendor %s not found", d.config.DefaultVendor)
		}
	}

	// Try fallback vendor
	if d.config.FallbackVendor != "" && d.config.FallbackVendor != d.config.DefaultVendor {
		if vendor, exists := d.vendors[d.config.FallbackVendor]; exists {
			if vendor.IsAvailable(ctx) {
				return vendor, nil
			}
			d.logger.Printf("Fallback vendor %s is not available", d.config.FallbackVendor)
		}
	}

	// Fallback to first available vendor
	for name, vendor := range d.vendors {
		if vendor.IsAvailable(ctx) {
			d.logger.Printf("Using fallback vendor: %s", name)
			return vendor, nil
		}
	}

	return nil, models.ErrVendorUnavailable
}

// applyRoutingRules applies routing rules to determine the appropriate vendor
func (d *Dispatcher) applyRoutingRules(req *models.Request) models.LLMVendor {
	// Sort rules by priority (higher priority first)
	rules := make([]models.RoutingRule, len(d.config.RoutingRules))
	copy(rules, d.config.RoutingRules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if d.matchesCondition(req, rule.Condition) {
			if vendor, exists := d.vendors[rule.Vendor]; exists {
				return vendor
			}
		}
	}

	return nil
}

// matchesCondition checks if a request matches a routing condition
func (d *Dispatcher) matchesCondition(req *models.Request, condition models.RoutingCondition) bool {
	// Check model pattern
	if condition.ModelPattern != "" {
		// Simple pattern matching - could be enhanced with regex
		if req.Model != condition.ModelPattern {
			return false
		}
	}

	// Check max tokens
	if condition.MaxTokens > 0 && req.MaxTokens > condition.MaxTokens {
		return false
	}

	// Check temperature
	if condition.Temperature > 0 && req.Temperature > condition.Temperature {
		return false
	}

	return true
}

// sendWithRetry sends a request with retry logic
func (d *Dispatcher) sendWithRetry(ctx context.Context, vendor models.LLMVendor, req *models.Request) (*models.Response, error) {
	var lastErr error
	maxAttempts := 1

	if d.config.RetryPolicy != nil {
		maxAttempts = d.config.RetryPolicy.MaxRetries + 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		response, err := vendor.SendRequest(ctx, req)
		if err == nil {
			return response, nil
		}

		lastErr = err
		d.logger.Printf("Attempt %d failed for vendor %s: %v", attempt, vendor.Name(), err)

		// Check if we should retry
		if attempt < maxAttempts && d.shouldRetry(err) {
			backoff := d.calculateBackoff(attempt)
			d.logger.Printf("Retrying in %v", backoff)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		break
	}

	// If we have a fallback vendor, try it
	if d.config.FallbackVendor != "" && vendor.Name() != d.config.FallbackVendor {
		if fallbackVendor, exists := d.vendors[d.config.FallbackVendor]; exists {
			d.logger.Printf("Trying fallback vendor: %s", d.config.FallbackVendor)
			fallbackResponse, fallbackErr := fallbackVendor.SendRequest(ctx, req)
			if fallbackErr == nil {
				return fallbackResponse, nil
			}
			d.logger.Printf("Fallback vendor also failed: %v", fallbackErr)
		}
	}

	return nil, fmt.Errorf("all attempts failed: %w", lastErr)
}

// shouldRetry determines if an error should trigger a retry
func (d *Dispatcher) shouldRetry(err error) bool {
	if d.config.RetryPolicy == nil {
		return false
	}

	// Handle nil error
	if err == nil {
		return false
	}

	// Check if error is in retryable errors list
	errStr := err.Error()
	for _, retryableErr := range d.config.RetryPolicy.RetryableErrors {
		if errStr == retryableErr {
			return true
		}
	}

	// Default retryable errors
	defaultRetryableErrors := []string{
		"rate limit exceeded",
		"timeout",
		"connection refused",
		"network error",
	}

	for _, retryableErr := range defaultRetryableErrors {
		if errStr == retryableErr {
			return true
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for retries
func (d *Dispatcher) calculateBackoff(attempt int) time.Duration {
	if d.config.RetryPolicy == nil {
		return time.Second
	}

	baseDelay := time.Second
	switch d.config.RetryPolicy.BackoffStrategy {
	case models.ExponentialBackoff:
		// Use int64 to avoid integer overflow, cap at reasonable maximum
		backoff := int64(1 << (attempt - 1))
		if backoff > 60 { // Cap at 60 seconds
			backoff = 60
		}
		return baseDelay * time.Duration(backoff)
	case models.LinearBackoff:
		return baseDelay * time.Duration(attempt)
	case models.FixedBackoff:
		return baseDelay
	default:
		return baseDelay
	}
}

// updateStats updates the dispatcher statistics
func (d *Dispatcher) updateStats(success bool, vendorName string, latency time.Duration) {
	d.statsMutex.Lock()
	defer d.statsMutex.Unlock()

	if success {
		d.stats.SuccessfulRequests++
	} else {
		d.stats.FailedRequests++
	}

	// Update vendor-specific stats
	if vendorName != "" {
		stats := d.stats.VendorStats[vendorName]
		stats.Requests++
		if success {
			stats.Successes++
		} else {
			stats.Failures++
		}
		stats.LastUsed = time.Now()

		// Update average latency
		if stats.AverageLatency == 0 {
			stats.AverageLatency = latency
		} else {
			stats.AverageLatency = (stats.AverageLatency + latency) / 2
		}

		d.stats.VendorStats[vendorName] = stats
	}

	// Update global average latency
	if d.stats.AverageLatency == 0 {
		d.stats.AverageLatency = latency
	} else {
		d.stats.AverageLatency = (d.stats.AverageLatency + latency) / 2
	}
}

// GetStats returns the current dispatcher statistics
func (d *Dispatcher) GetStats() *models.DispatcherStats {
	d.statsMutex.RLock()
	defer d.statsMutex.RUnlock()

	// Return a copy to avoid race conditions
	stats := *d.stats
	stats.VendorStats = make(map[string]models.VendorStats)
	for k, v := range d.stats.VendorStats {
		stats.VendorStats[k] = v
	}

	return &stats
}

// GetVendors returns a list of registered vendor names
func (d *Dispatcher) GetVendors() []string {
	vendors := make([]string, 0, len(d.vendors))
	for name := range d.vendors {
		vendors = append(vendors, name)
	}
	return vendors
}

// GetVendor returns a specific vendor by name
func (d *Dispatcher) GetVendor(name string) (models.LLMVendor, bool) {
	vendor, exists := d.vendors[name]
	return vendor, exists
}
