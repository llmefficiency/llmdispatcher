package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create dispatcher with configuration
	config := &llmdispatcher.Config{
		DefaultVendor: "openai",
		Timeout:       30 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
		RetryPolicy: &llmdispatcher.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: llmdispatcher.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
	}

	dispatcher := llmdispatcher.NewWithConfig(config)

	// Create and register OpenAI vendor
	openaiConfig := &llmdispatcher.VendorConfig{
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
	if err := dispatcher.RegisterVendor(openaiVendor); err != nil {
		log.Fatalf("Failed to register OpenAI vendor: %v", err)
	}

	// Create a request
	request := &llmdispatcher.Request{
		Model: "gpt-3.5-turbo",
		Messages: []llmdispatcher.Message{
			{
				Role:    "user",
				Content: "Hello! Can you tell me a short joke?",
			},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// Send the request
	ctx := context.Background()
	response, err := dispatcher.Send(ctx, request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Print the response
	fmt.Printf("Response from %s:\n", response.Vendor)
	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Content: %s\n", response.Content)
	fmt.Printf("Usage: %d prompt tokens, %d completion tokens, %d total tokens\n",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)

	// Print statistics
	stats := dispatcher.GetStats()
	fmt.Printf("\nDispatcher Statistics:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

	// Print vendor statistics
	for vendorName, vendorStats := range stats.VendorStats {
		fmt.Printf("\n%s Vendor Statistics:\n", vendorName)
		fmt.Printf("  Requests: %d\n", vendorStats.Requests)
		fmt.Printf("  Successes: %d\n", vendorStats.Successes)
		fmt.Printf("  Failures: %d\n", vendorStats.Failures)
		fmt.Printf("  Average Latency: %v\n", vendorStats.AverageLatency)
		fmt.Printf("  Last Used: %v\n", vendorStats.LastUsed)
	}
}
