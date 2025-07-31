package llmdispatcher

import (
	"context"

	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
)

// NewOpenAIVendor creates a new OpenAI vendor
func NewOpenAIVendor(config *VendorConfig) Vendor {
	internalConfig := &models.VendorConfig{}

	if config != nil {
		internalConfig.APIKey = config.APIKey
		internalConfig.BaseURL = config.BaseURL
		internalConfig.Timeout = config.Timeout
		internalConfig.Headers = config.Headers
		internalConfig.RateLimit = models.RateLimit{
			RequestsPerMinute: config.RateLimit.RequestsPerMinute,
			TokensPerMinute:   config.RateLimit.TokensPerMinute,
		}
	}

	return &vendorAdapter{
		vendor: vendors.NewOpenAI(internalConfig),
	}
}

// NewAnthropicVendor creates a new Anthropic vendor
func NewAnthropicVendor(config *VendorConfig) Vendor {
	internalConfig := &models.VendorConfig{}

	if config != nil {
		internalConfig.APIKey = config.APIKey
		internalConfig.BaseURL = config.BaseURL
		internalConfig.Timeout = config.Timeout
		internalConfig.Headers = config.Headers
		internalConfig.RateLimit = models.RateLimit{
			RequestsPerMinute: config.RateLimit.RequestsPerMinute,
			TokensPerMinute:   config.RateLimit.TokensPerMinute,
		}
	}

	return &vendorAdapter{
		vendor: vendors.NewAnthropic(internalConfig),
	}
}

// NewGoogleVendor creates a new Google vendor
func NewGoogleVendor(config *VendorConfig) Vendor {
	internalConfig := &models.VendorConfig{}

	if config != nil {
		internalConfig.APIKey = config.APIKey
		internalConfig.BaseURL = config.BaseURL
		internalConfig.Timeout = config.Timeout
		internalConfig.Headers = config.Headers
		internalConfig.RateLimit = models.RateLimit{
			RequestsPerMinute: config.RateLimit.RequestsPerMinute,
			TokensPerMinute:   config.RateLimit.TokensPerMinute,
		}
	}

	return &vendorAdapter{
		vendor: vendors.NewGoogle(internalConfig),
	}
}

// NewAzureOpenAIVendor creates a new Azure OpenAI vendor
func NewAzureOpenAIVendor(config *VendorConfig) Vendor {
	internalConfig := &models.VendorConfig{}

	if config != nil {
		internalConfig.APIKey = config.APIKey
		internalConfig.BaseURL = config.BaseURL
		internalConfig.Timeout = config.Timeout
		internalConfig.Headers = config.Headers
		internalConfig.RateLimit = models.RateLimit{
			RequestsPerMinute: config.RateLimit.RequestsPerMinute,
			TokensPerMinute:   config.RateLimit.TokensPerMinute,
		}
	}

	return &vendorAdapter{
		vendor: vendors.NewAzureOpenAI(internalConfig),
	}
}

// vendorAdapter adapts the internal vendor interface to the public interface
type vendorAdapter struct {
	vendor models.LLMVendor
}

func (a *vendorAdapter) Name() string {
	return a.vendor.Name()
}

func (a *vendorAdapter) SendRequest(ctx context.Context, req *Request) (*Response, error) {
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

	internalResp, err := a.vendor.SendRequest(ctx, internalReq)
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

func (a *vendorAdapter) GetCapabilities() Capabilities {
	internalCaps := a.vendor.GetCapabilities()
	return Capabilities{
		Models:            internalCaps.Models,
		SupportsStreaming: internalCaps.SupportsStreaming,
		MaxTokens:         internalCaps.MaxTokens,
		MaxInputTokens:    internalCaps.MaxInputTokens,
	}
}

func (a *vendorAdapter) IsAvailable(ctx context.Context) bool {
	return a.vendor.IsAvailable(ctx)
}

func (a *vendorAdapter) SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error) {
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

	internalStreamingResp, err := a.vendor.SendStreamingRequest(ctx, internalReq)
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
