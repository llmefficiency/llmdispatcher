package llmdispatcher

import (
	"context"
	"fmt"

	"github.com/llmefficiency/llmdispatcher/internal/dispatcher"
	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// Dispatcher is the main public interface for the LLM dispatcher
type Dispatcher struct {
	dispatcher *dispatcher.Dispatcher
}

// New creates a new dispatcher with default configuration
func New() *Dispatcher {
	return &Dispatcher{
		dispatcher: dispatcher.New(),
	}
}

// NewWithConfig creates a new dispatcher with custom configuration
func NewWithConfig(config *Config) *Dispatcher {
	internalConfig := &models.Config{}

	if config != nil {
		internalConfig.DefaultVendor = config.DefaultVendor
		internalConfig.FallbackVendor = config.FallbackVendor
		internalConfig.Timeout = config.Timeout
		internalConfig.EnableLogging = config.EnableLogging
		internalConfig.EnableMetrics = config.EnableMetrics
	}

	if config != nil && config.RetryPolicy != nil {
		internalConfig.RetryPolicy = &models.RetryPolicy{
			MaxRetries:      config.RetryPolicy.MaxRetries,
			BackoffStrategy: models.BackoffStrategy(config.RetryPolicy.BackoffStrategy),
			RetryableErrors: config.RetryPolicy.RetryableErrors,
		}
	}

	if config != nil && len(config.RoutingRules) > 0 {
		internalConfig.RoutingRules = make([]models.RoutingRule, len(config.RoutingRules))
		for i, rule := range config.RoutingRules {
			internalConfig.RoutingRules[i] = models.RoutingRule{
				Condition: models.RoutingCondition{
					ModelPattern:     rule.Condition.ModelPattern,
					MaxTokens:        rule.Condition.MaxTokens,
					Temperature:      rule.Condition.Temperature,
					CostThreshold:    rule.Condition.CostThreshold,
					LatencyThreshold: rule.Condition.LatencyThreshold,
				},
				Vendor:   rule.Vendor,
				Priority: rule.Priority,
				Enabled:  rule.Enabled,
			}
		}
	}

	return &Dispatcher{
		dispatcher: dispatcher.NewWithConfig(internalConfig),
	}
}

// Send sends a request to the appropriate vendor
func (d *Dispatcher) Send(ctx context.Context, req *Request) (*Response, error) {
	internalReq := &models.Request{
		Model:       req.Model,
		Messages:    make([]models.Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		internalReq.Messages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	internalResp, err := d.dispatcher.Send(ctx, internalReq)
	if err != nil {
		return nil, err
	}

	return &Response{
		Content:      internalResp.Content,
		Model:        internalResp.Model,
		Vendor:       internalResp.Vendor,
		FinishReason: internalResp.FinishReason,
		CreatedAt:    internalResp.CreatedAt,
		Usage: Usage{
			PromptTokens:     internalResp.Usage.PromptTokens,
			CompletionTokens: internalResp.Usage.CompletionTokens,
			TotalTokens:      internalResp.Usage.TotalTokens,
		},
	}, nil
}

// SendStreaming sends a streaming request to the appropriate vendor
func (d *Dispatcher) SendStreaming(ctx context.Context, req *Request) (*StreamingResponse, error) {
	internalReq := &models.Request{
		Model:       req.Model,
		Messages:    make([]models.Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		internalReq.Messages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	internalStreamingResp, err := d.dispatcher.SendStreaming(ctx, internalReq)
	if err != nil {
		return nil, err
	}

	// Create public streaming response
	publicStreamingResp := NewStreamingResponse(internalStreamingResp.Model, internalStreamingResp.Vendor)

	// Copy the channels and data
	go func() {
		defer publicStreamingResp.Close()
		defer internalStreamingResp.Close()

		for {
			select {
			case content := <-internalStreamingResp.ContentChan:
				publicStreamingResp.ContentChan <- content
			case done := <-internalStreamingResp.DoneChan:
				publicStreamingResp.DoneChan <- done
				return
			case err := <-internalStreamingResp.ErrorChan:
				publicStreamingResp.ErrorChan <- err
				return
			}
		}
	}()

	return publicStreamingResp, nil
}

// RegisterVendor registers a vendor with the dispatcher
func (d *Dispatcher) RegisterVendor(vendor Vendor) error {
	// Create an adapter to convert between public and internal interfaces
	adapter := &internalVendorAdapter{vendor: vendor}
	return d.dispatcher.RegisterVendor(adapter)
}

// GetStats returns the current dispatcher statistics
func (d *Dispatcher) GetStats() *Stats {
	internalStats := d.dispatcher.GetStats()

	stats := &Stats{
		TotalRequests:      internalStats.TotalRequests,
		SuccessfulRequests: internalStats.SuccessfulRequests,
		FailedRequests:     internalStats.FailedRequests,
		AverageLatency:     internalStats.AverageLatency,
		LastRequestTime:    internalStats.LastRequestTime,
		VendorStats:        make(map[string]VendorStats),
	}

	for name, vendorStats := range internalStats.VendorStats {
		stats.VendorStats[name] = VendorStats{
			Requests:       vendorStats.Requests,
			Successes:      vendorStats.Successes,
			Failures:       vendorStats.Failures,
			AverageLatency: vendorStats.AverageLatency,
			LastUsed:       vendorStats.LastUsed,
		}
	}

	return stats
}

// GetVendors returns a list of registered vendor names
func (d *Dispatcher) GetVendors() []string {
	return d.dispatcher.GetVendors()
}

// GetVendor returns a specific vendor by name
func (d *Dispatcher) GetVendor(name string) (Vendor, bool) {
	vendor, exists := d.dispatcher.GetVendor(name)
	if !exists {
		return nil, false
	}
	return &vendorWrapper{vendor: vendor}, true
}

// internalVendorAdapter adapts the public vendor interface to the internal interface
type internalVendorAdapter struct {
	vendor Vendor
}

func (a *internalVendorAdapter) Name() string {
	if a.vendor == nil {
		return ""
	}
	return a.vendor.Name()
}

func (a *internalVendorAdapter) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	if a.vendor == nil {
		return nil, fmt.Errorf("vendor is nil")
	}
	// Convert internal request to public request
	publicReq := &Request{
		Model:       req.Model,
		Messages:    make([]Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		publicReq.Messages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call the public vendor
	publicResp, err := a.vendor.SendRequest(ctx, publicReq)
	if err != nil {
		return nil, err
	}

	// Convert public response to internal response
	return &models.Response{
		Content:      publicResp.Content,
		Model:        publicResp.Model,
		Vendor:       publicResp.Vendor,
		FinishReason: publicResp.FinishReason,
		CreatedAt:    publicResp.CreatedAt,
		Usage: models.Usage{
			PromptTokens:     publicResp.Usage.PromptTokens,
			CompletionTokens: publicResp.Usage.CompletionTokens,
			TotalTokens:      publicResp.Usage.TotalTokens,
		},
	}, nil
}

func (a *internalVendorAdapter) GetCapabilities() models.Capabilities {
	if a.vendor == nil {
		return models.Capabilities{}
	}
	publicCaps := a.vendor.GetCapabilities()
	return models.Capabilities{
		Models:            publicCaps.Models,
		SupportsStreaming: publicCaps.SupportsStreaming,
		MaxTokens:         publicCaps.MaxTokens,
		MaxInputTokens:    publicCaps.MaxInputTokens,
	}
}

func (a *internalVendorAdapter) IsAvailable(ctx context.Context) bool {
	if a.vendor == nil {
		return false
	}
	return a.vendor.IsAvailable(ctx)
}

func (a *internalVendorAdapter) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	if a.vendor == nil {
		return nil, fmt.Errorf("vendor is nil")
	}
	// Convert internal request to public request
	publicReq := &Request{
		Model:       req.Model,
		Messages:    make([]Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		publicReq.Messages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call the public vendor's streaming method
	publicStreamingResp, err := a.vendor.SendStreamingRequest(ctx, publicReq)
	if err != nil {
		return nil, err
	}

	// Convert public streaming response to internal streaming response
	internalStreamingResp := models.NewStreamingResponse(req.Model, a.vendor.Name())

	// Copy the channels and data
	go func() {
		defer internalStreamingResp.Close()
		defer publicStreamingResp.Close()

		for {
			select {
			case content := <-publicStreamingResp.ContentChan:
				internalStreamingResp.ContentChan <- content
			case done := <-publicStreamingResp.DoneChan:
				internalStreamingResp.DoneChan <- done
				return
			case err := <-publicStreamingResp.ErrorChan:
				internalStreamingResp.ErrorChan <- err
				return
			}
		}
	}()

	return internalStreamingResp, nil
}

// vendorWrapper wraps the internal vendor interface
type vendorWrapper struct {
	vendor models.LLMVendor
}

func (w *vendorWrapper) Name() string {
	return w.vendor.Name()
}

func (w *vendorWrapper) SendRequest(ctx context.Context, req *Request) (*Response, error) {
	internalReq := &models.Request{
		Model:       req.Model,
		Messages:    make([]models.Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		internalReq.Messages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	internalResp, err := w.vendor.SendRequest(ctx, internalReq)
	if err != nil {
		return nil, err
	}

	return &Response{
		Content:      internalResp.Content,
		Model:        internalResp.Model,
		Vendor:       internalResp.Vendor,
		FinishReason: internalResp.FinishReason,
		CreatedAt:    internalResp.CreatedAt,
		Usage: Usage{
			PromptTokens:     internalResp.Usage.PromptTokens,
			CompletionTokens: internalResp.Usage.CompletionTokens,
			TotalTokens:      internalResp.Usage.TotalTokens,
		},
	}, nil
}

func (w *vendorWrapper) GetCapabilities() Capabilities {
	internalCaps := w.vendor.GetCapabilities()
	return Capabilities{
		Models:            internalCaps.Models,
		SupportsStreaming: internalCaps.SupportsStreaming,
		MaxTokens:         internalCaps.MaxTokens,
		MaxInputTokens:    internalCaps.MaxInputTokens,
	}
}

func (w *vendorWrapper) IsAvailable(ctx context.Context) bool {
	return w.vendor.IsAvailable(ctx)
}

func (w *vendorWrapper) SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error) {
	internalReq := &models.Request{
		Model:       req.Model,
		Messages:    make([]models.Message, len(req.Messages)),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	for i, msg := range req.Messages {
		internalReq.Messages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	internalStreamingResp, err := w.vendor.SendStreamingRequest(ctx, internalReq)
	if err != nil {
		return nil, err
	}

	// Create public streaming response
	publicStreamingResp := NewStreamingResponse(internalStreamingResp.Model, internalStreamingResp.Vendor)

	// Copy the channels and data
	go func() {
		defer publicStreamingResp.Close()
		defer internalStreamingResp.Close()

		for {
			select {
			case content := <-internalStreamingResp.ContentChan:
				publicStreamingResp.ContentChan <- content
			case done := <-internalStreamingResp.DoneChan:
				publicStreamingResp.DoneChan <- done
				return
			case err := <-internalStreamingResp.ErrorChan:
				publicStreamingResp.ErrorChan <- err
				return
			}
		}
	}()

	return publicStreamingResp, nil
}
