package dispatcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// Dispatcher manages routing of LLM requests to different vendors
type Dispatcher struct {
	vendors      map[string]models.LLMVendor
	config       *models.Config
	stats        *models.DispatcherStats
	statsMutex   sync.RWMutex
	logger       *log.Logger
	modeRegistry *models.ModeRegistry
}

// New creates a new dispatcher with default configuration
func New() *Dispatcher {
	return NewWithConfig(&models.Config{
		Mode:          models.AutoMode,
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
			ModeStats:   make(map[models.Mode]*models.ModeStats),
		},
		logger:       log.New(log.Writer(), "[LLMDispatcher] ", log.LstdFlags),
		modeRegistry: models.NewModeRegistry(),
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

// Send sends a request to the appropriate vendor based on routing strategy
func (d *Dispatcher) Send(ctx context.Context, req *models.Request) (*models.Response, error) {
	if ctx == nil {
		return nil, models.ErrInvalidRequest
	}

	if req == nil {
		return nil, models.ErrInvalidRequest
	}

	// Validate request
	fmt.Printf("DEBUG: Dispatcher validating request with Model='%s', Mode='%s'\n", req.Model, req.Mode)
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrInvalidRequest, err)
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

	// Use mode-based vendor selection with context preprocessing
	vendor, err := d.selectVendorWithMode(ctx, req)
	if err != nil {
		d.updateStats(false, "", time.Since(start), 0.0)
		return nil, fmt.Errorf("failed to select vendor: %w", err)
	}

	response, err := d.sendWithRetry(ctx, vendor, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start), 0.0)
		return nil, err
	}

	// Calculate estimated cost
	var estimatedCost float64
	if response != nil && response.Usage.TotalTokens > 0 {
		totalTokens := response.Usage.PromptTokens + response.Usage.CompletionTokens
		estimatedCost = estimateCost(totalTokens, vendor.Name())
		response.EstimatedCost = estimatedCost
	}

	d.updateStats(true, vendor.Name(), time.Since(start), estimatedCost)
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
	fmt.Printf("DEBUG: Dispatcher validating streaming request with Model='%s', Mode='%s'\n", req.Model, req.Mode)
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrInvalidRequest, err)
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

	// Use mode-based vendor selection with context preprocessing
	vendor, err := d.selectVendorWithMode(ctx, req)
	if err != nil {
		d.updateStats(false, "", time.Since(start), 0.0)
		return nil, fmt.Errorf("failed to select vendor: %w", err)
	}

	// Check if vendor supports streaming
	if !vendor.GetCapabilities().SupportsStreaming {
		return nil, fmt.Errorf("vendor %s does not support streaming", vendor.Name())
	}

	// Send streaming request
	streamingResp, err := vendor.SendStreamingRequest(ctx, req)
	if err != nil {
		d.updateStats(false, vendor.Name(), time.Since(start), 0.0)
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start), 0.0) // Cost not available for streaming
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
		d.updateStats(false, vendor.Name(), time.Since(start), 0.0)
		return nil, err
	}

	// Calculate estimated cost
	var estimatedCost float64
	if response != nil && response.Usage.TotalTokens > 0 {
		totalTokens := response.Usage.PromptTokens + response.Usage.CompletionTokens
		estimatedCost = estimateCost(totalTokens, vendor.Name())
		response.EstimatedCost = estimatedCost
	}

	d.updateStats(true, vendor.Name(), time.Since(start), estimatedCost)
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
		d.updateStats(false, vendor.Name(), time.Since(start), 0.0)
		return nil, err
	}

	d.updateStats(true, vendor.Name(), time.Since(start), 0.0) // Cost not available for streaming
	return streamingResp, nil
}

// selectVendorWithMode uses the new mode system to select vendors with context preprocessing
func (d *Dispatcher) selectVendorWithMode(ctx context.Context, req *models.Request) (models.LLMVendor, error) {
	// Determine the mode to use
	mode := d.config.Mode
	if req.Mode != "" {
		// TODO: Validate that the request mode is valid
		// For now, we'll use the request mode if specified
		mode = models.Mode(req.Mode)
	}

	// Get the mode strategy
	strategy, err := d.modeRegistry.GetStrategy(mode)
	if err != nil {
		d.logger.Printf("Failed to get mode strategy for %s: %v", mode, err)
		// Fallback to any available vendor
		for name, vendor := range d.vendors {
			if vendor.IsAvailable(ctx) {
				d.logger.Printf("Using fallback vendor: %s", name)
				return vendor, nil
			}
		}
		return nil, fmt.Errorf("no available vendors")
	}

	// Create mode context
	modeContext := &models.ModeContext{
		Mode:             mode,
		Request:          req,
		AvailableVendors: d.vendors,
		Config:           d.config,
		Stats:            d.getModeStats(mode),
		Context:          ctx,
	}

	// Validate context
	if err := strategy.ValidateContext(modeContext); err != nil {
		d.logger.Printf("Mode context validation failed: %v", err)
		return nil, fmt.Errorf("mode context validation failed: %w", err)
	}

	// Preprocess context based on mode
	if err := strategy.PreprocessContext(modeContext); err != nil {
		d.logger.Printf("Context preprocessing failed: %v", err)
		// Continue without preprocessing rather than failing
	}

	// Optimize request for the mode
	if err := strategy.OptimizeRequest(modeContext); err != nil {
		d.logger.Printf("Request optimization failed: %v", err)
		// Continue without optimization rather than failing
	}

	// Select vendor using the mode strategy
	vendor, err := strategy.SelectVendor(modeContext)
	if err != nil {
		d.logger.Printf("Mode-based vendor selection failed: %v", err)
		// Fallback to any available vendor
		for name, vendor := range d.vendors {
			if vendor.IsAvailable(ctx) {
				d.logger.Printf("Using fallback vendor: %s", name)
				return vendor, nil
			}
		}
		return nil, fmt.Errorf("no available vendors")
	}

	d.logger.Printf("Selected vendor %s using mode %s", vendor.Name(), mode)
	return vendor, nil
}

// getModeStats returns the stats for a specific mode, creating if necessary
func (d *Dispatcher) getModeStats(mode models.Mode) *models.ModeStats {
	d.statsMutex.Lock()
	defer d.statsMutex.Unlock()

	if d.stats.ModeStats == nil {
		d.stats.ModeStats = make(map[models.Mode]*models.ModeStats)
	}

	stats, exists := d.stats.ModeStats[mode]
	if !exists {
		stats = &models.ModeStats{}
		d.stats.ModeStats[mode] = stats
	}

	return stats
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
func (d *Dispatcher) updateStats(success bool, vendorName string, latency time.Duration, cost float64) {
	d.statsMutex.Lock()
	defer d.statsMutex.Unlock()

	if success {
		d.stats.SuccessfulRequests++
	} else {
		d.stats.FailedRequests++
	}

	// Update cost statistics
	d.stats.TotalCost += cost
	if d.stats.TotalRequests > 0 {
		d.stats.AverageCost = d.stats.TotalCost / float64(d.stats.TotalRequests)
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

		// Update cost statistics for vendor
		stats.TotalCost += cost
		if stats.Requests > 0 {
			stats.AverageCost = stats.TotalCost / float64(stats.Requests)
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

	// Copy mode stats
	stats.ModeStats = make(map[models.Mode]*models.ModeStats)
	for k, v := range d.stats.ModeStats {
		stats.ModeStats[k] = v
	}

	return &stats
}

// estimateCost estimates the cost of a request based on token usage and vendor
func estimateCost(totalTokens int, vendor string) float64 {
	// Cost per 1K tokens for different vendors (approximate rates)
	costPer1KTokens := map[string]float64{
		"openai":    0.03, // GPT-3.5-turbo rate
		"anthropic": 0.15, // Claude-3-Sonnet rate
		"google":    0.05, // Gemini-Pro rate
		"azure":     0.03, // Azure OpenAI rate
		"local":     0.0,  // Local models are free
	}

	cost, exists := costPer1KTokens[vendor]
	if !exists {
		cost = 0.05 // Default cost
	}

	// Calculate cost for total tokens
	return (float64(totalTokens) / 1000.0) * cost
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

// GetModeRegistry returns the mode registry for external access
func (d *Dispatcher) GetModeRegistry() *models.ModeRegistry {
	return d.modeRegistry
}

// RegisterModeStrategy registers a custom mode strategy
func (d *Dispatcher) RegisterModeStrategy(mode models.Mode, strategy models.ModeStrategy) {
	d.modeRegistry.RegisterStrategy(mode, strategy)
}
