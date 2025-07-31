package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/dispatcher"
	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
	"github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

func main() {
	// Parse command line flags
	var localMode = flag.Bool("local", false, "Run in local mode with Ollama")
	var demoMode = flag.Bool("demo", false, "Run local vendor demo")
	var modelPath = flag.String("model", "llama2:7b", "Model to use in local mode")
	var serverURL = flag.String("server", "http://localhost:11434", "Ollama server URL")
	flag.Parse()

	// Check if running in demo mode
	if *demoMode {
		RunLocalDemo()
		return
	}

	// Check if running in local mode
	if *localMode {
		runLocalMode(*modelPath, *serverURL)
		return
	}

	// Get API keys from environment variables
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	azureOpenAIAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")
	cohereAPIKey := os.Getenv("COHERE_API_KEY")

	// Create dispatcher with configuration
	config := &llmdispatcher.Config{
		DefaultVendor:  "openai",
		FallbackVendor: "anthropic", // Will be used when implemented
		Timeout:        30 * time.Second,
		EnableLogging:  true,
		EnableMetrics:  true,
		RetryPolicy: &llmdispatcher.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: llmdispatcher.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
		RoutingRules: []llmdispatcher.RoutingRule{
			{
				Condition: llmdispatcher.RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  true,
			},
			{
				Condition: llmdispatcher.RoutingCondition{
					ModelPattern: "claude",
				},
				Vendor:   "anthropic",
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	dispatcher := llmdispatcher.NewWithConfig(config)

	// Register OpenAI vendor (if API key is available)
	if openaiAPIKey != "" {
		openaiConfig := &llmdispatcher.VendorConfig{
			APIKey:  openaiAPIKey,
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		openaiVendor := llmdispatcher.NewOpenAIVendor(openaiConfig)
		if err := dispatcher.RegisterVendor(openaiVendor); err != nil {
			log.Printf("Failed to register OpenAI vendor: %v", err)
		} else {
			log.Println("✅ Registered OpenAI vendor")
		}
	} else {
		log.Println("⚠️  OPENAI_API_KEY not set, skipping OpenAI vendor")
	}

	// Register Anthropic vendor (when implemented)
	if anthropicAPIKey != "" {
		anthropicConfig := &llmdispatcher.VendorConfig{
			APIKey:  anthropicAPIKey,
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		anthropicVendor := llmdispatcher.NewAnthropicVendor(anthropicConfig)
		if err := dispatcher.RegisterVendor(anthropicVendor); err != nil {
			log.Printf("Failed to register Anthropic vendor: %v", err)
		} else {
			log.Println("✅ Registered Anthropic vendor")
		}
	} else {
		log.Println("⚠️  ANTHROPIC_API_KEY not set")
	}

	// Register Google vendor (when implemented)
	if googleAPIKey != "" {
		googleConfig := &llmdispatcher.VendorConfig{
			APIKey:  googleAPIKey,
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		googleVendor := llmdispatcher.NewGoogleVendor(googleConfig)
		if err := dispatcher.RegisterVendor(googleVendor); err != nil {
			log.Printf("Failed to register Google vendor: %v", err)
		} else {
			log.Println("✅ Registered Google vendor")
		}
	} else {
		log.Println("⚠️  GOOGLE_API_KEY not set")
	}

	// Register Azure OpenAI vendor (when implemented)
	if azureOpenAIAPIKey != "" {
		azureConfig := &llmdispatcher.VendorConfig{
			APIKey:  azureOpenAIAPIKey,
			BaseURL: os.Getenv("AZURE_OPENAI_ENDPOINT"),
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		azureVendor := llmdispatcher.NewAzureOpenAIVendor(azureConfig)
		if err := dispatcher.RegisterVendor(azureVendor); err != nil {
			log.Printf("Failed to register Azure OpenAI vendor: %v", err)
		} else {
			log.Println("✅ Registered Azure OpenAI vendor")
		}
	} else {
		log.Println("⚠️  AZURE_OPENAI_API_KEY not set")
	}

	// Register Cohere vendor (when implemented)
	if cohereAPIKey != "" {
		log.Println("ℹ️  Cohere vendor not yet implemented")
		// cohereVendor := llmdispatcher.NewCohereVendor(&llmdispatcher.VendorConfig{
		//     APIKey: cohereAPIKey,
		//     Timeout: 30 * time.Second,
		// })
		// dispatcher.RegisterVendor(cohereVendor)
	} else {
		log.Println("⚠️  COHERE_API_KEY not set")
	}

	// Check if we have any vendors registered
	vendors := dispatcher.GetVendors()
	if len(vendors) == 0 {
		log.Fatal("No vendors registered. Please set at least one API key.")
	}

	log.Printf("✅ Registered vendors: %v", vendors)

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
	fmt.Printf("\n📝 Response from %s:\n", response.Vendor)
	fmt.Printf("Model: %s\n", response.Model)
	fmt.Printf("Content: %s\n", response.Content)
	fmt.Printf("Usage: %d prompt tokens, %d completion tokens, %d total tokens\n",
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		response.Usage.TotalTokens)

	// Print statistics
	stats := dispatcher.GetStats()
	fmt.Printf("\n📊 Dispatcher Statistics:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

	// Print vendor statistics
	for vendorName, vendorStats := range stats.VendorStats {
		fmt.Printf("\n🔍 %s Vendor Statistics:\n", vendorName)
		fmt.Printf("  Requests: %d\n", vendorStats.Requests)
		fmt.Printf("  Successes: %d\n", vendorStats.Successes)
		fmt.Printf("  Failures: %d\n", vendorStats.Failures)
		fmt.Printf("  Average Latency: %v\n", vendorStats.AverageLatency)
		fmt.Printf("  Last Used: %v\n", vendorStats.LastUsed)
	}
}

// runLocalMode runs the dispatcher in local mode using Ollama
func runLocalMode(modelPath, serverURL string) {
	log.Printf("🚀 Starting local mode with model: %s", modelPath)
	log.Printf("📡 Connecting to Ollama server: %s", serverURL)

	// Create dispatcher with local configuration
	config := &models.Config{
		DefaultVendor: "local",
		Timeout:       60 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
		CostOptimization: &models.CostOptimization{
			Enabled:     true,
			PreferCheap: true,
			VendorCosts: map[string]float64{
				"local": 0.0001, // Cheapest option
			},
		},
	}

	disp := dispatcher.NewWithConfig(config)

	// Create and register local vendor
	localConfig := &models.VendorConfig{
		APIKey: "dummy", // Not used for local models
		Headers: map[string]string{
			"server_url": serverURL,
			"model_path": modelPath,
		},
		Timeout: 60 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Fatalf("Failed to register local vendor: %v", err)
	}

	log.Println("✅ Local vendor registered successfully")

	// Test basic functionality
	ctx := context.Background()
	req := &models.Request{
		Model: modelPath,
		Messages: []models.Message{
			{Role: "user", Content: "What is the capital of France?"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	log.Println("📤 Sending test request...")
	resp, err := disp.Send(ctx, req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	log.Println("✅ Request successful!")
	fmt.Printf("\n📝 Response from %s:\n", resp.Vendor)
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Content: %s\n", resp.Content)
	fmt.Printf("Usage: %d prompt tokens, %d completion tokens, %d total tokens\n",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens)

	// Test streaming
	log.Println("\n🔄 Testing streaming...")
	streamReq := &models.Request{
		Model: modelPath,
		Messages: []models.Message{
			{Role: "user", Content: "Write a short poem about AI."},
		},
		Temperature: 0.8,
		MaxTokens:   200,
	}

	streamResp, err := disp.SendStreaming(ctx, streamReq)
	if err != nil {
		log.Printf("⚠️  Streaming failed: %v", err)
	} else {
		log.Println("✅ Streaming successful!")
		fmt.Printf("\n📝 Streaming response from %s:\n", streamResp.Vendor)
		fmt.Printf("Model: %s\n", streamResp.Model)
		fmt.Printf("Created at: %s\n", streamResp.CreatedAt.Format(time.RFC3339))
	}

	// Print statistics
	stats := disp.GetStats()
	fmt.Printf("\n📊 Local Mode Statistics:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)

	log.Println("🎉 Local mode test completed successfully!")
}
