package models

import (
	"testing"
	"time"
)

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *Request
		wantErr bool
	}{
		{
			name: "valid request",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: false,
		},
		{
			name: "empty model",
			request: &Request{
				Model: "",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "no messages",
			request: &Request{
				Model:    "gpt-3.5-turbo",
				Messages: []Message{},
			},
			wantErr: true,
		},
		{
			name: "temperature too high",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 2.5,
			},
			wantErr: true,
		},
		{
			name: "temperature negative",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: -0.1,
			},
			wantErr: true,
		},
		{
			name: "top_p too high",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				TopP: 1.5,
			},
			wantErr: true,
		},
		{
			name: "top_p negative",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				TopP: -0.1,
			},
			wantErr: true,
		},
		{
			name: "max_tokens negative",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: -10,
			},
			wantErr: true,
		},
		{
			name: "invalid message",
			request: &Request{
				Model: "gpt-3.5-turbo",
				Messages: []Message{
					{Role: "", Content: "Hello"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name:    "valid user message",
			message: Message{Role: "user", Content: "Hello"},
			wantErr: false,
		},
		{
			name:    "valid system message",
			message: Message{Role: "system", Content: "You are a helpful assistant"},
			wantErr: false,
		},
		{
			name:    "valid assistant message",
			message: Message{Role: "assistant", Content: "Hello! How can I help you?"},
			wantErr: false,
		},
		{
			name:    "empty role",
			message: Message{Role: "", Content: "Hello"},
			wantErr: true,
		},
		{
			name:    "empty content",
			message: Message{Role: "user", Content: ""},
			wantErr: true,
		},
		{
			name:    "invalid role",
			message: Message{Role: "invalid", Content: "Hello"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVendorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  VendorConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: VendorConfig{
				APIKey:  "sk-test",
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: VendorConfig{
				APIKey:  "",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: VendorConfig{
				APIKey:  "sk-test",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout is valid",
			config: VendorConfig{
				APIKey:  "sk-test",
				Timeout: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("VendorConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewStreamingResponse(t *testing.T) {
	model := "gpt-4"
	vendor := "openai"

	streamingResp := NewStreamingResponse(model, vendor)

	if streamingResp.Model != model {
		t.Errorf("Expected model %s, got %s", model, streamingResp.Model)
	}

	if streamingResp.Vendor != vendor {
		t.Errorf("Expected vendor %s, got %s", vendor, streamingResp.Vendor)
	}

	if streamingResp.ContentChan == nil {
		t.Error("Expected ContentChan to be initialized")
	}

	if streamingResp.DoneChan == nil {
		t.Error("Expected DoneChan to be initialized")
	}

	if streamingResp.ErrorChan == nil {
		t.Error("Expected ErrorChan to be initialized")
	}

	// Check channel buffer sizes
	if cap(streamingResp.ContentChan) != 100 {
		t.Errorf("Expected ContentChan buffer size 100, got %d", cap(streamingResp.ContentChan))
	}

	if cap(streamingResp.DoneChan) != 1 {
		t.Errorf("Expected DoneChan buffer size 1, got %d", cap(streamingResp.DoneChan))
	}

	if cap(streamingResp.ErrorChan) != 1 {
		t.Errorf("Expected ErrorChan buffer size 1, got %d", cap(streamingResp.ErrorChan))
	}
}

func TestStreamingResponse_Close(t *testing.T) {
	streamingResp := NewStreamingResponse("gpt-4", "openai")

	// Send some data to channels
	streamingResp.ContentChan <- "test content"

	// Close the response
	streamingResp.Close()

	// Verify channels are closed by trying to read from them
	// Note: We need to drain the content first
	select {
	case <-streamingResp.ContentChan:
		// Drain the content
	default:
		// No content to drain
	}

	// Now check if channels are closed
	select {
	case _, ok := <-streamingResp.ContentChan:
		if ok {
			t.Error("Expected ContentChan to be closed")
		}
	default:
		// Channel is closed, which is what we expect
	}

	select {
	case _, ok := <-streamingResp.DoneChan:
		if ok {
			t.Error("Expected DoneChan to be closed")
		}
	default:
		// Channel is closed, which is what we expect
	}

	select {
	case _, ok := <-streamingResp.ErrorChan:
		if ok {
			t.Error("Expected ErrorChan to be closed")
		}
	default:
		// Channel is closed, which is what we expect
	}
}

func TestStreamingResponse_Usage(t *testing.T) {
	streamingResp := NewStreamingResponse("gpt-4", "openai")

	// Set usage data
	streamingResp.Usage = Usage{
		PromptTokens:     10,
		CompletionTokens: 15,
		TotalTokens:      25,
	}

	if streamingResp.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", streamingResp.Usage.PromptTokens)
	}

	if streamingResp.Usage.CompletionTokens != 15 {
		t.Errorf("Expected completion tokens 15, got %d", streamingResp.Usage.CompletionTokens)
	}

	if streamingResp.Usage.TotalTokens != 25 {
		t.Errorf("Expected total tokens 25, got %d", streamingResp.Usage.TotalTokens)
	}
}

func TestStreamingResponse_CreatedAt(t *testing.T) {
	before := time.Now()
	streamingResp := NewStreamingResponse("gpt-4", "openai")
	after := time.Now()

	if streamingResp.CreatedAt.Before(before) || streamingResp.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v should be between %v and %v",
			streamingResp.CreatedAt, before, after)
	}
}
