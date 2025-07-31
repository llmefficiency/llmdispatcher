package vendors

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// AnthropicVendor implements the LLMVendor interface for Anthropic's Claude models
type AnthropicVendor struct {
	config *models.VendorConfig
	client *http.Client
}

// NewAnthropic creates a new Anthropic vendor
func NewAnthropic(config *models.VendorConfig) *AnthropicVendor {
	if config == nil {
		config = &models.VendorConfig{
			BaseURL: "https://api.anthropic.com",
			Timeout: 30 * time.Second,
		}
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &AnthropicVendor{
		config: config,
		client: client,
	}
}

// Name returns the vendor name
func (a *AnthropicVendor) Name() string {
	return "anthropic"
}

// SendRequest sends a request to Anthropic's API
func (a *AnthropicVendor) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Anthropic format
	anthropicReq := a.convertRequest(req)

	// Create HTTP request
	jsonData, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.BaseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("User-Agent", "llmdispatcher/1.0")

	// Add custom headers
	for key, value := range a.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := a.client.Do(httpReq)
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
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to standard response
	response := a.convertResponse(&anthropicResp, req.Model)
	return response, nil
}

// GetCapabilities returns the vendor's capabilities
func (a *AnthropicVendor) GetCapabilities() models.Capabilities {
	return models.Capabilities{
		Models: []string{
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
		},
		SupportsStreaming: true,
		MaxTokens:         4096,
		MaxInputTokens:    200000,
	}
}

// IsAvailable checks if Anthropic is available
func (a *AnthropicVendor) IsAvailable(ctx context.Context) bool {
	return a.config.APIKey != ""
}

// SendStreamingRequest sends a streaming request to Anthropic
func (a *AnthropicVendor) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	// Create streaming response
	streamingResp := models.NewStreamingResponse(req.Model, a.Name())

	// Convert to Anthropic format
	anthropicReq := a.convertRequest(req)
	anthropicReq.Stream = true // Enable streaming

	// Marshal request
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.BaseURL+"/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Add custom headers
	for key, value := range a.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := a.client.Do(httpReq)
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Handle streaming response in goroutine
	go func() {
		defer resp.Body.Close()
		defer streamingResp.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					streamingResp.DoneChan <- true
					return
				}
				streamingResp.ErrorChan <- fmt.Errorf("failed to read stream: %w", err)
				return
			}

			// Skip empty lines
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Remove "data: " prefix
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					streamingResp.DoneChan <- true
					return
				}

				// Parse the JSON data
				var streamResp struct {
					Type  string `json:"type"`
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}

				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					streamingResp.ErrorChan <- fmt.Errorf("failed to parse stream data: %w", err)
					return
				}

				if streamResp.Type == "content_block_delta" && streamResp.Delta.Type == "text_delta" {
					streamingResp.ContentChan <- streamResp.Delta.Text
				}
			}
		}
	}()

	return streamingResp, nil
}

// convertRequest converts our standard request to Anthropic format
func (a *AnthropicVendor) convertRequest(req *models.Request) *anthropicRequest {
	// Convert messages to Anthropic format
	messages := make([]anthropicMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = anthropicMessage{
			Role:    msg.Role,
			Content: []anthropicContent{{Type: "text", Text: msg.Content}},
		}
	}

	anthropicReq := &anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	return anthropicReq
}

// convertResponse converts Anthropic response to our standard format
func (a *AnthropicVendor) convertResponse(anthropicResp *anthropicResponse, model string) *models.Response {
	// Extract content from response
	var content string
	if len(anthropicResp.Content) > 0 {
		content = anthropicResp.Content[0].Text
	}

	// Calculate token usage
	usage := models.Usage{
		PromptTokens:     anthropicResp.Usage.InputTokens,
		CompletionTokens: anthropicResp.Usage.OutputTokens,
		TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
	}

	return &models.Response{
		Content:   content,
		Model:     model,
		Vendor:    a.Name(),
		Usage:     usage,
		CreatedAt: time.Now(),
	}
}

// Anthropic API request/response structures
type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"` // Added for streaming
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
	Usage   anthropicUsage     `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
