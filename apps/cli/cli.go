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
)

// printResponse prints a formatted response
func printResponse(vendor, model, content string, usage models.Usage) {
	fmt.Printf("\n📝 Response from %s:\n", vendor)
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Content: %s\n", content)
	fmt.Printf("Usage: %d prompt tokens, %d completion tokens, %d total tokens\n",
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
}

// printDetailedStats prints detailed statistics with vendor breakdown
func printDetailedStats(stats *models.DispatcherStats) {
	fmt.Printf("\n📊 Detailed Statistics:\n")
	fmt.Printf("┌─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│                    OVERALL STATS                          │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ Total Requests: %-8d │ Successful: %-8d │ Failed: %-8d │\n",
		stats.TotalRequests, stats.SuccessfulRequests, stats.FailedRequests)
	fmt.Printf("│ Average Latency: %-35s │\n", stats.AverageLatency.String())
	if stats.TotalCost > 0 {
		fmt.Printf("│ Total Cost: $%-8.4f │ Average Cost: $%-8.4f │\n",
			stats.TotalCost, stats.AverageCost)
	}
	fmt.Printf("│ Last Request: %-35s │\n", stats.LastRequestTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("└─────────────────────────────────────────────────────────────┘\n")

	if len(stats.VendorStats) > 0 {
		fmt.Printf("\n🔍 VENDOR BREAKDOWN:\n")
		fmt.Printf("┌─────────────────────────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Vendor      │ Requests │ Successes │ Failures │ Avg Latency │ Last Used      │\n")
		fmt.Printf("├─────────────────────────────────────────────────────────────────────────────────┤\n")

		for vendorName, vendorStats := range stats.VendorStats {
			lastUsed := vendorStats.LastUsed.Format("01-02 15:04")
			fmt.Printf("│ %-11s │ %-8d │ %-9d │ %-8d │ %-11s │ %-14s │\n",
				vendorName,
				vendorStats.Requests,
				vendorStats.Successes,
				vendorStats.Failures,
				vendorStats.AverageLatency.String(),
				lastUsed)
		}
		fmt.Printf("└─────────────────────────────────────────────────────────────────────────────────┘\n")
	}
}

// printStrategyComparison prints a clean comparison of key metrics across strategies
func printStrategyComparison(strategyStats map[models.Strategy]*models.DispatcherStats) {
	fmt.Printf("\n🎯 STRATEGY COMPARISON:\n")
	fmt.Printf("┌────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Strategy   │ Latency     │ Cost/Request │ Success Rate │ Vendor │\n")
	fmt.Printf("├────────────────────────────────────────────────────────────────────┤\n")

	for strategy, stats := range strategyStats {
		successRate := 0.0
		if stats.TotalRequests > 0 {
			successRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
		}

		// Get the primary vendor used
		primaryVendor := "none"
		if len(stats.VendorStats) > 0 {
			for vendor := range stats.VendorStats {
				primaryVendor = vendor
				break // Get first (and likely only) vendor
			}
		}

		costPerRequest := 0.0
		if stats.TotalRequests > 0 && stats.TotalCost > 0 {
			costPerRequest = stats.TotalCost / float64(stats.TotalRequests)
		}

		fmt.Printf("│ %-10s │ %-11s │ $%-10.4f │ %-11.0f%% │ %-6s │\n",
			string(strategy),
			stats.AverageLatency.String(),
			costPerRequest,
			successRate,
			primaryVendor)
	}
	fmt.Printf("└────────────────────────────────────────────────────────────────────┘\n")
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

// createVendorConfig creates a vendor configuration with standard settings
func createVendorConfig(apiKey, baseURL string, timeout time.Duration) *models.VendorConfig {
	return &models.VendorConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Timeout: timeout,
		Headers: map[string]string{
			"User-Agent": "llmdispatcher/1.0",
		},
	}
}

// registerVendors registers all available vendors to the dispatcher
func registerVendors(disp *dispatcher.Dispatcher) {
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	googleAPIKey := os.Getenv("GOOGLE_API_KEY")

	if openaiAPIKey != "" {
		config := createVendorConfig(openaiAPIKey, "https://api.openai.com/v1", 30*time.Second)
		vendor := vendors.NewOpenAI(config)
		if err := disp.RegisterVendor(vendor); err != nil {
			log.Printf("⚠️  Failed to register OpenAI vendor: %v", err)
		}
	}

	if anthropicAPIKey != "" {
		config := createVendorConfig(anthropicAPIKey, "https://api.anthropic.com", 30*time.Second)
		vendor := vendors.NewAnthropic(config)
		if err := disp.RegisterVendor(vendor); err != nil {
			log.Printf("⚠️  Failed to register Anthropic vendor: %v", err)
		}
	}

	if googleAPIKey != "" {
		config := createVendorConfig(googleAPIKey, "https://generativelanguage.googleapis.com", 30*time.Second)
		vendor := vendors.NewGoogle(config)
		if err := disp.RegisterVendor(vendor); err != nil {
			log.Printf("⚠️  Failed to register Google vendor: %v", err)
		}
	}
}

// runStrategyTest runs a test with a specific strategy and returns the stats
func runStrategyTest(strategy models.Strategy, baseMessages []models.Message) *models.DispatcherStats {
	// Create dispatcher with strategy-specific configuration
	config := &models.Config{
		Strategy:      strategy,
		Timeout:       30 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: models.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
	}

	disp := dispatcher.NewWithConfig(config)
	registerVendors(disp)

	// We need to determine which vendor would be selected for this strategy
	// For now, we'll try different models and see which one works
	modelNames := []string{"gpt-3.5-turbo", "claude-3-haiku-20240307", "gemini-pro"}

	var lastErr error
	for _, modelName := range modelNames {
		testRequest := &models.Request{
			Model:       modelName,
			Messages:    baseMessages,
			Temperature: 0.7,
			MaxTokens:   100,
		}

		ctx := context.Background()
		_, err := disp.Send(ctx, testRequest)
		if err != nil {
			lastErr = err
			continue // Try next model
		} else {
			// Success! Break out of loop
			break
		}
	}

	if lastErr != nil {
		log.Printf("⚠️  Strategy %s test failed with all models: %v", strategy, lastErr)
	}

	return disp.GetStats()
}

// runStrategyComparison runs tests across all strategies and shows comparison
func runStrategyComparison() {
	fmt.Printf("\n🚀 Running Strategy Comparison Test\n")
	fmt.Printf("Testing all strategies with vendor-appropriate models...\n")

	// We'll use a generic request that gets modified per vendor
	baseMessages := []models.Message{
		{
			Role:    "user",
			Content: "Hello! Can you tell me a short joke?",
		},
	}

	strategies := []models.Strategy{
		models.BalancedStrategy,
		models.SpeedStrategy,
		models.QualityStrategy,
		models.BudgetStrategy,
	}

	strategyStats := make(map[models.Strategy]*models.DispatcherStats)

	for _, strategy := range strategies {
		fmt.Printf("Testing %s strategy...\n", strategy)
		stats := runStrategyTest(strategy, baseMessages)
		strategyStats[strategy] = stats
	}

	// Print the comparison
	printStrategyComparison(strategyStats)
}

func main() {
	// Load environment variables from .env file
	if err := loadEnv(".env"); err != nil {
		log.Printf("⚠️  Could not load .env file: %v", err)
	}

	// Parse command line flags
	var localMode = flag.Bool("local", false, "Run in local strategy with Ollama")
	var vendorMode = flag.Bool("vendor", false, "Run in vendor strategy")
	var vendorOverride = flag.String("vendor-override", "", "Override vendor to use (anthropic, openai). If not specified, uses default vendor")
	var modelPath = flag.String("model", "llama2:7b", "Model to use in local strategy")
	var serverURL = flag.String("server", "http://localhost:11434", "Ollama server URL")
	var compareModes = flag.Bool("compare", false, "Run comparison test across all strategies")
	flag.Parse()

	// Check if running strategy comparison
	if *compareModes {
		runStrategyComparison()
		return
	}

	// Check if running in local strategy
	if *localMode {
		runLocalMode(*modelPath, *serverURL)
		return
	}

	// Check if running in vendor strategy
	if *vendorMode {
		runVendorMode(*vendorOverride, *modelPath, *serverURL)
		return
	}

	// Get API keys from environment variables
	azureOpenAIAPIKey := os.Getenv("AZURE_OPENAI_API_KEY")

	// Create dispatcher with configuration
	config := &models.Config{
		Strategy:      models.BalancedStrategy,
		Timeout:       30 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: models.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
	}

	disp := dispatcher.NewWithConfig(config)
	registerVendors(disp)

	// Log registered vendors
	if os.Getenv("OPENAI_API_KEY") != "" {
		log.Println("✅ Registered OpenAI vendor")
	} else {
		log.Println("⚠️  OPENAI_API_KEY not set, skipping OpenAI vendor")
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		log.Println("✅ Registered Anthropic vendor")
	} else {
		log.Println("⚠️  ANTHROPIC_API_KEY not set")
	}

	if os.Getenv("GOOGLE_API_KEY") != "" {
		log.Println("✅ Registered Google vendor")
	} else {
		log.Println("⚠️  GOOGLE_API_KEY not set")
	}

	// Register Azure OpenAI vendor (when implemented)
	if azureOpenAIAPIKey != "" {
		azureConfig := &models.VendorConfig{
			APIKey:  azureOpenAIAPIKey,
			BaseURL: os.Getenv("AZURE_OPENAI_ENDPOINT"),
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		azureVendor := vendors.NewAzureOpenAI(azureConfig)
		if err := disp.RegisterVendor(azureVendor); err != nil {
			log.Printf("Failed to register Azure OpenAI vendor: %v", err)
		} else {
			log.Println("✅ Registered Azure OpenAI vendor")
		}
	} else {
		log.Println("⚠️  AZURE_OPENAI_API_KEY not set")
	}

	// Check if we have any vendors registered
	vendors := disp.GetVendors()
	if len(vendors) == 0 {
		log.Fatal("No vendors registered. Please set at least one API key.")
	}

	log.Printf("✅ Registered vendors: %v", vendors)

	// Create a request
	request := &models.Request{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
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
	response, err := disp.Send(ctx, request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Print the response
	printResponse(response.Vendor, response.Model, response.Content, response.Usage)

	// Print detailed statistics
	stats := disp.GetStats()
	printDetailedStats(stats)
}

// runLocalMode runs the dispatcher in local strategy using Ollama
func runLocalMode(modelPath, serverURL string) {
	log.Printf("🚀 Starting local strategy with model: %s", modelPath)
	log.Printf("📡 Connecting to Ollama server: %s", serverURL)

	// Create dispatcher with local configuration
	config := &models.Config{
		Strategy:      models.BudgetStrategy, // Use budget strategy for local
		Timeout:       60 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
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

	// Print detailed statistics
	stats := disp.GetStats()
	printDetailedStats(stats)

	log.Println("🎉 Local strategy test completed successfully!")
}

// runVendorMode runs the dispatcher in vendor strategy to test specific vendors
func runVendorMode(vendorOverride, modelPath, serverURL string) {
	log.Printf("🚀 Starting vendor strategy")

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
		Strategy:      models.BalancedStrategy, // Use balanced strategy for vendor testing
		Timeout:       60 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
	}

	disp := dispatcher.NewWithConfig(config)

	// Register vendor based on target
	switch targetVendor {
	case "anthropic":
		anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
		if anthropicAPIKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable is required for Anthropic vendor")
		}
		config := createVendorConfig(anthropicAPIKey, "https://api.anthropic.com", 120*time.Second)
		vendor := vendors.NewAnthropic(config)
		if err := disp.RegisterVendor(vendor); err != nil {
			log.Fatalf("Failed to register Anthropic vendor: %v", err)
		}
		log.Println("✅ Anthropic vendor registered successfully")

	case "openai":
		openaiAPIKey := os.Getenv("OPENAI_API_KEY")
		if openaiAPIKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable is required for OpenAI vendor")
		}
		config := createVendorConfig(openaiAPIKey, "https://api.openai.com/v1", 120*time.Second)
		vendor := vendors.NewOpenAI(config)
		if err := disp.RegisterVendor(vendor); err != nil {
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

	// Print detailed statistics
	stats := disp.GetStats()
	printDetailedStats(stats)

	log.Printf("🎉 Vendor strategy test completed successfully for %s!", targetVendor)
}
