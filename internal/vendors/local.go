package vendors

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/models"
)

// Local vendor implementation for local model inference
type Local struct {
	config *models.VendorConfig
	client *http.Client
	// Local-specific configuration
	modelPath      string
	serverURL      string
	executable     string
	useHTTP        bool
	resourceLimits *ResourceLimits
}

// ResourceLimits defines resource constraints for local models
type ResourceLimits struct {
	MaxMemoryMB  int `json:"max_memory_mb"`
	MaxThreads   int `json:"max_threads"`
	MaxGPULayers int `json:"max_gpu_layers"`
}

// LocalRequest represents the local model request format
type LocalRequest struct {
	Model       string           `json:"model"`
	Messages    []models.Message `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Stop        []string         `json:"stop,omitempty"`
}

// LocalResponse represents the local model response format
type LocalResponse struct {
	Model   string `json:"model"`
	Content string `json:"content"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewLocal creates a new Local vendor instance
func NewLocal(config *models.VendorConfig) *Local {
	if config == nil {
		config = &models.VendorConfig{}
	}

	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second // Longer timeout for local models
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	local := &Local{
		config: config,
		client: client,
		resourceLimits: &ResourceLimits{
			MaxMemoryMB: 4096, // 4GB default
			MaxThreads:  4,    // 4 threads default
		},
	}

	// Parse local-specific configuration from headers
	if config.Headers != nil {
		if modelPath, ok := config.Headers["model_path"]; ok {
			local.modelPath = modelPath
		}
		if serverURL, ok := config.Headers["server_url"]; ok {
			local.serverURL = serverURL
			local.useHTTP = true
		}
		if executable, ok := config.Headers["executable"]; ok {
			local.executable = executable
		}
	}

	// Set default server URL for Ollama if not provided
	if local.useHTTP && local.serverURL == "" {
		local.serverURL = "http://localhost:11434"
	}

	return local
}

// Name returns the vendor name
func (l *Local) Name() string {
	return "local"
}

// SendRequest sends a request to the local model
func (l *Local) SendRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	if l.useHTTP {
		return l.sendHTTPRequest(ctx, req)
	}
	return l.sendProcessRequest(ctx, req)
}

// sendHTTPRequest sends a request via HTTP (Ollama)
func (l *Local) sendHTTPRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	localReq := LocalRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      false,
		Stop:        req.Stop,
	}

	jsonData, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", l.serverURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local model error: %s - %s", resp.Status, string(body))
	}

	var localResp LocalResponse
	if err := json.NewDecoder(resp.Body).Decode(&localResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &models.Response{
		Content: localResp.Content,
		Usage: models.Usage{
			PromptTokens:     localResp.Usage.PromptTokens,
			CompletionTokens: localResp.Usage.CompletionTokens,
			TotalTokens:      localResp.Usage.TotalTokens,
		},
		Model:     localResp.Model,
		Vendor:    l.Name(),
		CreatedAt: time.Now(),
	}, nil
}

// sendProcessRequest sends a request via direct process execution (llama.cpp)
func (l *Local) sendProcessRequest(ctx context.Context, req *models.Request) (*models.Response, error) {
	if l.executable == "" {
		return nil, fmt.Errorf("executable path not configured for local model")
	}

	// Prepare input for the model
	input := l.prepareInput(req.Messages)

	// Build command arguments
	args := []string{
		"-m", l.modelPath,
		"--temp", fmt.Sprintf("%.2f", req.Temperature),
		"--top-p", fmt.Sprintf("%.2f", req.TopP),
	}

	if req.MaxTokens > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", req.MaxTokens))
	}

	// Add resource limits
	if l.resourceLimits.MaxThreads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", l.resourceLimits.MaxThreads))
	}
	if l.resourceLimits.MaxMemoryMB > 0 {
		args = append(args, "--ctx-size", fmt.Sprintf("%d", l.resourceLimits.MaxMemoryMB))
	}

	// Create command
	// nolint:gosec // executable path is controlled by configuration, not user input
	cmd := exec.CommandContext(ctx, l.executable, args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("local model execution failed: %w, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	// Extract the generated content (remove the input prompt)
	content := l.extractGeneratedContent(input, output)

	// Estimate token usage (rough approximation)
	totalTokens := len(strings.Fields(input)) + len(strings.Fields(content))

	return &models.Response{
		Content: content,
		Usage: models.Usage{
			PromptTokens:     len(strings.Fields(input)),
			CompletionTokens: len(strings.Fields(content)),
			TotalTokens:      totalTokens,
		},
		Model:     req.Model,
		Vendor:    l.Name(),
		CreatedAt: time.Now(),
	}, nil
}

// prepareInput formats messages for local model input
func (l *Local) prepareInput(messages []models.Message) string {
	var input strings.Builder

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			input.WriteString(fmt.Sprintf("System: %s\n", msg.Content))
		case "user":
			input.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
		case "assistant":
			input.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
		}
	}

	input.WriteString("Assistant: ")
	return input.String()
}

// extractGeneratedContent extracts the generated content from the model output
func (l *Local) extractGeneratedContent(input, output string) string {
	// Simple extraction - remove the input prompt from the output
	if strings.HasPrefix(output, input) {
		return strings.TrimSpace(output[len(input):])
	}
	return strings.TrimSpace(output)
}

// SendStreamingRequest sends a streaming request to the local model
func (l *Local) SendStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	if l.useHTTP {
		return l.sendHTTPStreamingRequest(ctx, req)
	}
	return l.sendProcessStreamingRequest(ctx, req)
}

// sendHTTPStreamingRequest sends a streaming request via HTTP
func (l *Local) sendHTTPStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	localReq := LocalRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      true,
		Stop:        req.Stop,
	}

	jsonData, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", l.serverURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("local model error: %s - %s", resp.Status, string(body))
	}

	streamingResp := models.NewStreamingResponse(req.Model, l.Name())

	go func() {
		defer resp.Body.Close()
		defer streamingResp.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Parse streaming response (Ollama format)
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					streamingResp.DoneChan <- true
					return
				}

				var streamData map[string]interface{}
				if err := json.Unmarshal([]byte(data), &streamData); err != nil {
					streamingResp.ErrorChan <- fmt.Errorf("failed to parse stream data: %w", err)
					return
				}

				if content, ok := streamData["content"].(string); ok {
					streamingResp.ContentChan <- content
				}
			}
		}

		if err := scanner.Err(); err != nil {
			streamingResp.ErrorChan <- fmt.Errorf("stream reading error: %w", err)
		}
	}()

	return streamingResp, nil
}

// sendProcessStreamingRequest sends a streaming request via process execution
func (l *Local) sendProcessStreamingRequest(ctx context.Context, req *models.Request) (*models.StreamingResponse, error) {
	if l.executable == "" {
		return nil, fmt.Errorf("executable path not configured for local model")
	}

	input := l.prepareInput(req.Messages)

	args := []string{
		"-m", l.modelPath,
		"--temp", fmt.Sprintf("%.2f", req.Temperature),
		"--top-p", fmt.Sprintf("%.2f", req.TopP),
		"--stream",
	}

	if req.MaxTokens > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", req.MaxTokens))
	}

	if l.resourceLimits.MaxThreads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", l.resourceLimits.MaxThreads))
	}

	// nolint:gosec // executable path is controlled by configuration, not user input
	cmd := exec.CommandContext(ctx, l.executable, args...)
	cmd.Stdin = strings.NewReader(input)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	streamingResp := models.NewStreamingResponse(req.Model, l.Name())

	go func() {
		defer streamingResp.Close()

		if err := cmd.Start(); err != nil {
			streamingResp.ErrorChan <- fmt.Errorf("failed to start process: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				streamingResp.ContentChan <- line
			}
		}

		if err := cmd.Wait(); err != nil {
			streamingResp.ErrorChan <- fmt.Errorf("process execution failed: %w", err)
			return
		}

		streamingResp.DoneChan <- true
	}()

	return streamingResp, nil
}

// GetCapabilities returns the local model capabilities
func (l *Local) GetCapabilities() models.Capabilities {
	return models.Capabilities{
		Models:            models.GetVendorModels("local"),
		SupportsStreaming: true,
		MaxTokens:         4096,
		MaxInputTokens:    4096,
	}
}

// IsAvailable checks if the local model is available
func (l *Local) IsAvailable(ctx context.Context) bool {
	if l.useHTTP {
		return l.checkHTTPServer(ctx)
	}
	return l.checkExecutable()
}

// checkHTTPServer checks if the HTTP server is available
func (l *Local) checkHTTPServer(ctx context.Context) bool {
	url := fmt.Sprintf("%s/api/tags", l.serverURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// checkExecutable checks if the executable is available
func (l *Local) checkExecutable() bool {
	if l.executable == "" {
		return false
	}

	// nolint:gosec // executable path is controlled by configuration, not user input
	cmd := exec.Command(l.executable, "--help")
	return cmd.Run() == nil
}
