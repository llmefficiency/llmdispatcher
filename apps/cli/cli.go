package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/dispatcher"
	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
	"github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

// printResponse prints a formatted response
func printResponse(vendor, model, content string, usage llmdispatcher.Usage) {
	fmt.Printf("\n📝 Response from %s:\n", vendor)
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Content: %s\n", content)
	fmt.Printf("Usage: %d prompt tokens, %d completion tokens, %d total tokens\n",
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
}

// printStats prints dispatcher statistics
func printStats(stats *llmdispatcher.Stats) {
	fmt.Printf("\n📊 Dispatcher Statistics:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
}

// printInternalStats prints internal dispatcher statistics
func printInternalStats(stats *models.DispatcherStats) {
	fmt.Printf("\n📊 Dispatcher Statistics:\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
}

// loadEnv loads environment variables from .env file
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				os.Setenv(key, value)
			}
		}
	}
	return scanner.Err()
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv(".env"); err != nil {
		log.Printf("⚠️  Could not load .env file: %v", err)
	}

	// Parse command line flags
	var localMode = flag.Bool("local", false, "Run in local mode with Ollama")
	var vendorMode = flag.Bool("vendor", false, "Run in vendor mode")
	var vendorOverride = flag.String("vendor-override", "", "Override vendor to use (anthropic, openai). If not specified, uses default vendor")
	var modelPath = flag.String("model", "llama2:7b", "Model to use in local mode")
	var serverURL = flag.String("server", "http://localhost:11434", "Ollama server URL")
	flag.Parse()

	// Check if running in local mode
	if *localMode {
		runLocalMode(*modelPath, *serverURL)
		return
	}

	// Check if running in vendor mode
	if *vendorMode {
		runVendorMode(*vendorOverride, *modelPath, *serverURL)
		return
	}

	// Get API keys from environment variables
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")
	azureOpenAIAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")

	// Create dispatcher with configuration
	config := &llmdispatcher.Config{
		DefaultVendor:  "openai",
		FallbackVendor: "anthropic",
		Timeout:        30 * time.Second,
		EnableLogging:  true,
		EnableMetrics:  true,
		RetryPolicy: &llmdispatcher.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: llmdispatcher.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
		// Use cascading failure strategy
		RoutingStrategy: llmdispatcher.NewCascadingFailureStrategy([]string{"openai", "anthropic", "google"}),
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
	printResponse(response.Vendor, response.Model, response.Content, response.Usage)

	// Print statistics
	stats := dispatcher.GetStats()
	printStats(stats)

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

	// Create a background context for streaming (no timeout)
	streamCtx := context.Background()

	// Run streaming in a separate goroutine to avoid context cancellation
	streamDone := make(chan bool)
	go func() {
		defer close(streamDone)

		streamResp, err := disp.SendStreaming(streamCtx, streamReq)
		if err != nil {
			log.Printf("⚠️  Streaming failed: %v", err)
			return
		}

		log.Println("✅ Streaming successful!")
		fmt.Printf("\n📝 Streaming response from %s:\n", streamResp.Vendor)
		fmt.Printf("Model: %s\n", streamResp.Model)
		fmt.Printf("Created at: %s\n", streamResp.CreatedAt.Format(time.RFC3339))

		// Print streaming content
		fmt.Printf("\n📄 Streaming content:\n")
		fmt.Printf("Content: ")

		// Read from the streaming response channel with proper error handling
		done := false
		contentReceived := false

		for !done {
			select {
			case chunk, ok := <-streamResp.ContentChan:
				if !ok {
					// Channel closed
					done = true
					if contentReceived {
						fmt.Println() // New line after content
					}
				} else {
					// Print chunk immediately
					fmt.Print(chunk)
					contentReceived = true
				}
			case err := <-streamResp.ErrorChan:
				if err != nil {
					// Check if it's a context cancellation after receiving content
					if strings.Contains(err.Error(), "context canceled") && contentReceived {
						// This is expected when we receive content and then context is canceled
						fmt.Println() // New line after content
					} else {
						fmt.Printf("\n❌ Streaming error: %v\n", err)
					}
				}
				done = true
			case <-streamResp.DoneChan:
				// Streaming completed successfully
				done = true
				if contentReceived {
					fmt.Println() // New line after content
				}
			case <-time.After(30 * time.Second): // Simple timeout
				fmt.Printf("\n⏰ Streaming timeout after 30 seconds\n")
				done = true
			}
		}

		// Close the streaming response
		streamResp.Close()
	}()

	// Wait for streaming to complete
	<-streamDone

	// Print statistics
	stats := disp.GetStats()
	printInternalStats(stats)

	log.Println("🎉 Local mode test completed successfully!")
}

// runVendorMode runs the dispatcher in vendor mode to test specific vendors
func runVendorMode(vendorOverride, modelPath, serverURL string) {
	log.Printf("🚀 Starting vendor mode")

	// Determine which vendor to use
	var targetVendor string
	if vendorOverride == "" {
		// Use default vendor (openai)
		targetVendor = "openai"
		log.Printf("Using default vendor: %s", targetVendor)
	} else {
		// Use specified vendor
		targetVendor = vendorOverride
		log.Printf("Using specified vendor: %s", targetVendor)
	}

	// Create dispatcher with vendor configuration
	config := &models.Config{
		DefaultVendor: targetVendor,
		Timeout:       60 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
	}

	disp := dispatcher.NewWithConfig(config)

	// Register vendor based on target
	switch targetVendor {
	case "anthropic":
		// Register Anthropic vendor
		anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
		if anthropicAPIKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable is required for Anthropic vendor")
		}

		anthropicConfig := &models.VendorConfig{
			APIKey:  anthropicAPIKey,
			BaseURL: "https://api.anthropic.com",
			Timeout: 120 * time.Second, // Longer timeout for streaming
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		anthropicVendor := vendors.NewAnthropic(anthropicConfig)
		if err := disp.RegisterVendor(anthropicVendor); err != nil {
			log.Fatalf("Failed to register Anthropic vendor: %v", err)
		}
		log.Println("✅ Anthropic vendor registered successfully")

	case "openai":
		// Register OpenAI vendor
		openaiAPIKey := os.Getenv("OPENAI_API_KEY")
		if openaiAPIKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable is required for OpenAI vendor")
		}

		openaiConfig := &models.VendorConfig{
			APIKey:  openaiAPIKey,
			BaseURL: "https://api.openai.com/v1",
			Timeout: 120 * time.Second, // Longer timeout for streaming
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		openaiVendor := vendors.NewOpenAI(openaiConfig)
		if err := disp.RegisterVendor(openaiVendor); err != nil {
			log.Fatalf("Failed to register OpenAI vendor: %v", err)
		}
		log.Println("✅ OpenAI vendor registered successfully")

	default:
		log.Fatalf("Unsupported vendor: %s. Supported vendors: anthropic, openai", targetVendor)
	}

	// Test basic functionality
	ctx := context.Background()

	// Set model based on vendor
	var requestModel string
	switch targetVendor {
	case "anthropic":
		requestModel = "claude-3-haiku-20240307"
	case "openai":
		requestModel = "gpt-3.5-turbo"
	default:
		requestModel = "gpt-3.5-turbo" // Default to OpenAI model
	}

	req := &models.Request{
		Model: requestModel,
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
		Model: requestModel,
		Messages: []models.Message{
			{Role: "user", Content: "Tell me a short story about a robot."},
		},
		Temperature: 0.8,
		MaxTokens:   150,
	}

	// Create a background context for streaming (no timeout)
	streamCtx := context.Background()

	streamResp, err := disp.SendStreaming(streamCtx, streamReq)
	if err != nil {
		log.Printf("⚠️  Streaming failed: %v", err)
	} else {
		log.Println("✅ Streaming successful!")
		fmt.Printf("\n📝 Streaming response from %s:\n", streamResp.Vendor)
		fmt.Printf("Model: %s\n", streamResp.Model)
		fmt.Printf("Created at: %s\n", streamResp.CreatedAt.Format(time.RFC3339))

		// Print streaming content
		fmt.Printf("\n📄 Streaming content:\n")
		fmt.Printf("Content: ")

		// Read from the streaming response channel with proper error handling
		done := false
		contentReceived := false

		for !done {
			select {
			case chunk, ok := <-streamResp.ContentChan:
				if !ok {
					// Channel closed
					done = true
					if contentReceived {
						fmt.Println() // New line after content
					}
				} else {
					// Print chunk immediately
					fmt.Print(chunk)
					contentReceived = true
				}
			case err := <-streamResp.ErrorChan:
				if err != nil {
					// Check if it's a context cancellation after receiving content
					if strings.Contains(err.Error(), "context canceled") && contentReceived {
						// This is expected when we receive content and then context is canceled
						fmt.Println() // New line after content
					} else {
						fmt.Printf("\n❌ Streaming error: %v\n", err)
					}
				}
				done = true
			case <-streamResp.DoneChan:
				// Streaming completed successfully
				done = true
				if contentReceived {
					fmt.Println() // New line after content
				}
			case <-streamCtx.Done():
				// Context timeout or cancellation
				if streamCtx.Err() == context.DeadlineExceeded {
					fmt.Printf("\n⏰ Streaming timeout after 60 seconds\n")
				} else {
					fmt.Printf("\n⏰ Streaming canceled\n")
				}
				done = true
			}
		}

		// Close the streaming response
		streamResp.Close()
	}

	// Print statistics
	stats := disp.GetStats()
	printInternalStats(stats)

	log.Printf("🎉 Vendor mode test completed successfully for %s!", targetVendor)
}
