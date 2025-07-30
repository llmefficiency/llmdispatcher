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
