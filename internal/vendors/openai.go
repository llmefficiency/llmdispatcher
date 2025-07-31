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

// OpenAI vendor implementation
type OpenAI struct {
	config *models.VendorConfig
	client *http.Client
}

// OpenAIRequest represents the OpenAI API request format
type OpenAIRequest struct {
	Model       string           `json:"model"`
	Messages    []models.Message `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Stop        []string         `json:"stop,omitempty"`
	User        string           `json:"user,omitempty"`
}

// OpenAIResponse represents the OpenAI API response format
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIError represents an OpenAI API error
type OpenAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

// NewOpenAI creates a new OpenAI vendor instance
func NewOpenAI(config *models.VendorConfig) *OpenAI {
	if config == nil {
		config = &models.VendorConfig{}
	}

	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Set default base URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &OpenAI{
		config: config,
		client: client,
	}
}

// Name returns the vendor name
func (o *OpenAI) Name() string {
	return "openai"
}

// SendRequest sends a request to OpenAI
func (o *OpenAI) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Convert to OpenAI format
	openaiReq := OpenAIRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
		User:        req.User,
	}

	// Marshal request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	// Add custom headers
	for key, value := range o.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var openaiErr OpenAIError
		if err := json.Unmarshal(body, &openaiErr); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", openaiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to standard format
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	response := &models.Response{
		Content:      choice.Message.Content,
		Model:        openaiResp.Model,
		Vendor:       o.Name(),
		FinishReason: choice.FinishReason,
		CreatedAt:    time.Unix(openaiResp.Created, 0),
		Usage: models.Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}

	return response, nil
}

// GetCapabilities returns OpenAI's capabilities
func (o *OpenAI) GetCapabilities() models.Capabilities {
	return models.Capabilities{
		Models: []string{
			"gpt-4",
			"gpt-4-turbo",
			"gpt-4-turbo-preview",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
		},
		SupportsStreaming: true,
		MaxTokens:         4096,
		MaxInputTokens:    128000,
	}
}

// IsAvailable checks if OpenAI is available
func (o *OpenAI) IsAvailable(ctx context.Context) bool {
	// Simple availability check - could be enhanced with actual health check
	return o.config.APIKey != ""
}

// SendStreamingRequest sends a streaming request to OpenAI
func (o *OpenAI) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	// Create streaming response
	streamingResp := models.NewStreamingResponse(req.Model, o.Name())

	// Convert to OpenAI format with streaming enabled
	openaiReq := OpenAIRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      true, // Enable streaming
		Stop:        req.Stop,
		User:        req.User,
	}

	// Marshal request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	// Add custom headers
	for key, value := range o.config.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := o.client.Do(httpReq)
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
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					streamingResp.ErrorChan <- fmt.Errorf("failed to parse stream data: %w", err)
					return
				}

				if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
					streamingResp.ContentChan <- streamResp.Choices[0].Delta.Content
				}
			}
		}
	}()

	return streamingResp, nil
}
