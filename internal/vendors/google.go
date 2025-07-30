package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// GoogleVendor implements the LLMVendor interface for Google's Gemini models
type GoogleVendor struct {
	config *models.VendorConfig
	client *http.Client
}

// NewGoogle creates a new Google vendor
func NewGoogle(config *models.VendorConfig) *GoogleVendor {
	if config == nil {
		config = &models.VendorConfig{
			BaseURL: "https://generativelanguage.googleapis.com",
			Timeout: 30 * time.Second,
		}
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &GoogleVendor{
		config: config,
		client: client,
	}
}

// Name returns the vendor name
func (g *GoogleVendor) Name() string {
	return "google"
}

// SendRequest sends a request to Google's Gemini API
func (g *GoogleVendor) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Google format
	googleReq := g.convertRequest(req)

	// Create HTTP request
	jsonData, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with API key
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", 
		g.config.BaseURL, req.Model, g.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "llmdispatcher/1.0")

	// Add custom headers
	for key, value := range g.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var googleResp googleResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to standard response
	response := g.convertResponse(&googleResp, req.Model)
	return response, nil
}

// GetCapabilities returns the vendor's capabilities
func (g *GoogleVendor) GetCapabilities() models.Capabilities {
	return models.Capabilities{
		Models: []string{
			"gemini-1.5-pro",
			"gemini-1.5-flash",
			"gemini-1.0-pro",
			"gemini-pro",
			"gemini-pro-vision",
		},
		SupportsStreaming: true,
		MaxTokens:         8192,
		MaxInputTokens:    1000000,
	}
}

// IsAvailable checks if the vendor is currently available
func (g *GoogleVendor) IsAvailable(ctx context.Context) bool {
	return g.config.APIKey != ""
}

// convertRequest converts our standard request to Google format
func (g *GoogleVendor) convertRequest(req *models.Request) *googleRequest {
	// Convert messages to Google format
	contents := make([]googleContent, 0, len(req.Messages))
	for _, msg := range req.Messages {
		contents = append(contents, googleContent{
			Parts: []googlePart{{Text: msg.Content}},
		})
	}

	googleReq := &googleRequest{
		Contents: contents,
		GenerationConfig: googleGenerationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
			TopP:           req.TopP,
		},
	}

	return googleReq
}

// convertResponse converts Google response to our standard format
func (g *GoogleVendor) convertResponse(googleResp *googleResponse, model string) *models.Response {
	// Extract content from response
	var content string
	if len(googleResp.Candidates) > 0 && len(googleResp.Candidates[0].Content.Parts) > 0 {
		content = googleResp.Candidates[0].Content.Parts[0].Text
	}

	// Calculate token usage
	usage := models.Usage{
		PromptTokens:     googleResp.UsageMetadata.PromptTokenCount,
		CompletionTokens: googleResp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      googleResp.UsageMetadata.PromptTokenCount + googleResp.UsageMetadata.CandidatesTokenCount,
	}

	return &models.Response{
		Content:   content,
		Model:     model,
		Vendor:    g.Name(),
		Usage:     usage,
		CreatedAt: time.Now(),
	}
}

// Google API request/response structures
type googleRequest struct {
	Contents         []googleContent       `json:"contents"`
	GenerationConfig googleGenerationConfig `json:"generationConfig"`
}

type googleContent struct {
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	TopP           float64 `json:"topP,omitempty"`
}

type googleResponse struct {
	Candidates      []googleCandidate    `json:"candidates"`
	UsageMetadata   googleUsageMetadata  `json:"usageMetadata"`
}

type googleCandidate struct {
	Content googleContent `json:"content"`
}

type googleUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
} 