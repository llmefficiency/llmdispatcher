package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bufio"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// AzureOpenAIVendor implements the LLMVendor interface for Azure OpenAI
type AzureOpenAIVendor struct {
	config *models.VendorConfig
	client *http.Client
}

// NewAzureOpenAI creates a new Azure OpenAI vendor
func NewAzureOpenAI(config *models.VendorConfig) *AzureOpenAIVendor {
	if config == nil {
		config = &models.VendorConfig{
			Timeout: 30 * time.Second,
		}
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &AzureOpenAIVendor{
		config: config,
		client: client,
	}
}

// Name returns the vendor name
func (a *AzureOpenAIVendor) Name() string {
	return "azure-openai"
}

// SendRequest sends a request to Azure OpenAI API
func (a *AzureOpenAIVendor) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Azure OpenAI format
	azureReq := a.convertRequest(req)

	// Create HTTP request
	jsonData, err := json.Marshal(azureReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview",
		a.config.BaseURL, req.Model)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", a.config.APIKey)
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
	var azureResp azureResponse
	if err := json.Unmarshal(body, &azureResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to standard response
	response := a.convertResponse(&azureResp, req.Model)
	return response, nil
}

// GetCapabilities returns the vendor's capabilities
func (a *AzureOpenAIVendor) GetCapabilities() models.Capabilities {
	return models.Capabilities{
		Models: []string{
			"gpt-4",
			"gpt-4-turbo",
			"gpt-4o",
			"gpt-35-turbo",
			"gpt-35-turbo-16k",
		},
		SupportsStreaming: true,
		MaxTokens:         4096,
		MaxInputTokens:    128000,
	}
}

// IsAvailable checks if Azure OpenAI is available
func (a *AzureOpenAIVendor) IsAvailable(ctx context.Context) bool {
	return a.config.APIKey != "" && a.config.BaseURL != ""
}

// SendStreamingRequest sends a streaming request to Azure OpenAI
func (a *AzureOpenAIVendor) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	// Create streaming response
	streamingResp := models.NewStreamingResponse(req.Model, a.Name())

	// Convert to Azure OpenAI format
	azureReq := a.convertRequest(req)
	azureReq.Stream = true // Enable streaming

	// Marshal request
	reqBody, err := json.Marshal(azureReq)
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.BaseURL+"/openai/deployments/"+req.Model+"/chat/completions?api-version=2024-02-15-preview", bytes.NewBuffer(reqBody))
	if err != nil {
		streamingResp.Close()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", a.config.APIKey)

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

// convertRequest converts our standard request to Azure OpenAI format
func (a *AzureOpenAIVendor) convertRequest(req *models.Request) *azureRequest {
	// Convert messages to Azure OpenAI format
	messages := make([]azureMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = azureMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	azureReq := &azureRequest{
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	return azureReq
}

// convertResponse converts Azure OpenAI response to our standard format
func (a *AzureOpenAIVendor) convertResponse(azureResp *azureResponse, model string) *models.Response {
	// Extract content from response
	var content string
	if len(azureResp.Choices) > 0 {
		content = azureResp.Choices[0].Message.Content
	}

	// Calculate token usage
	usage := models.Usage{
		PromptTokens:     azureResp.Usage.PromptTokens,
		CompletionTokens: azureResp.Usage.CompletionTokens,
		TotalTokens:      azureResp.Usage.TotalTokens,
	}

	return &models.Response{
		Content:   content,
		Model:     model,
		Vendor:    a.Name(),
		Usage:     usage,
		CreatedAt: time.Now(),
	}
}

// Azure OpenAI API request/response structures
type azureRequest struct {
	Messages    []azureMessage `json:"messages"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
}

type azureMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type azureResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []azureChoice `json:"choices"`
	Usage   azureUsage    `json:"usage"`
}

type azureChoice struct {
	Index   int          `json:"index"`
	Message azureMessage `json:"message"`
	Delta   azureMessage `json:"delta,omitempty"`
}

type azureUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
