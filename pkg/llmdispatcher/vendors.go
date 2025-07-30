package llmdispatcher

import (
	"context"

	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
)

// NewOpenAIVendor creates a new OpenAI vendor
func NewOpenAIVendor(config *VendorConfig) Vendor {
	internalConfig := &models.VendorConfig{
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
		Timeout: config.Timeout,
		Headers: config.Headers,
		RateLimit: models.RateLimit{
			RequestsPerMinute: config.RateLimit.RequestsPerMinute,
			TokensPerMinute:   config.RateLimit.TokensPerMinute,
		},
	}

	return &vendorAdapter{
		vendor: vendors.NewOpenAI(internalConfig),
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
