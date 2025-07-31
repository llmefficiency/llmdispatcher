package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/llmefficiency/llmdispatcher/internal/dispatcher"
	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
)

// LocalVendorDemo demonstrates the local vendor functionality
func LocalVendorDemo() {
	fmt.Println("🚀 Local Vendor Demo")
	fmt.Println("====================")

	// Test 1: HTTP Mode (Ollama)
	fmt.Println("\n📡 Test 1: HTTP Mode (Ollama)")
	testHTTPMode()

	// Test 2: Process Mode (llama.cpp)
	fmt.Println("\n⚙️  Test 2: Process Mode (llama.cpp)")
	testProcessMode()

	// Test 3: Cost Optimization
	fmt.Println("\n💰 Test 3: Cost Optimization")
	testCostOptimization()

	// Test 4: Streaming
	fmt.Println("\n🔄 Test 4: Streaming")
	testStreaming()

	// Test 5: Error Handling
	fmt.Println("\n⚠️  Test 5: Error Handling")
	testErrorHandling()

	fmt.Println("\n✅ Local vendor demo completed!")
}

func testHTTPMode() {
	// Create dispatcher with local configuration
	config := &models.Config{
		DefaultVendor: "local",
		Timeout:       30 * time.Second,
		EnableLogging: true,
		EnableMetrics: true,
	}

	disp := dispatcher.NewWithConfig(config)

	// Create and register local vendor (HTTP mode)
	localConfig := &models.VendorConfig{
		APIKey: "dummy",
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
			"model_path": "llama2:7b",
		},
		Timeout: 30 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register local vendor: %v", err)
		return
	}

	// Test basic request
	ctx := context.Background()
	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "What is the capital of France?"},
		},
		Temperature: 0.7,
		MaxTokens:   50,
	}

	fmt.Println("  Sending HTTP request...")
	resp, err := disp.Send(ctx, req)
	if err != nil {
		fmt.Printf("  ❌ HTTP request failed: %v\n", err)
		return
	}

	fmt.Printf("  ✅ HTTP response: %s\n", resp.Content)
}

func testProcessMode() {
	// Create dispatcher
	config := &models.Config{
		DefaultVendor: "local",
		Timeout:       60 * time.Second,
		EnableLogging: true,
	}

	disp := dispatcher.NewWithConfig(config)

	// Create and register local vendor (Process mode)
	localConfig := &models.VendorConfig{
		APIKey: "dummy",
		Headers: map[string]string{
			"executable": "/usr/local/bin/llama", // This will likely fail in demo
			"model_path": "/path/to/model.gguf",
		},
		Timeout: 60 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register local vendor: %v", err)
		return
	}

	// Test process request
	ctx := context.Background()
	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Write a short poem."},
		},
		Temperature: 0.8,
		MaxTokens:   100,
	}

	fmt.Println("  Sending process request...")
	resp, err := disp.Send(ctx, req)
	if err != nil {
		fmt.Printf("  ❌ Process request failed (expected): %v\n", err)
		return
	}

	fmt.Printf("  ✅ Process response: %s\n", resp.Content)
}

func testCostOptimization() {
	// Create dispatcher with cost optimization
	config := &models.Config{
		DefaultVendor: "local",
		Timeout:       30 * time.Second,
		EnableLogging: true,
		CostOptimization: &models.CostOptimization{
			Enabled:     true,
			PreferCheap: true,
			VendorCosts: map[string]float64{
				"local":     0.0001, // Cheapest option
				"openai":    0.0020,
				"anthropic": 0.0015,
			},
		},
	}

	disp := dispatcher.NewWithConfig(config)

	// Register local vendor
	localConfig := &models.VendorConfig{
		APIKey: "dummy",
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
			"model_path": "llama2:7b",
		},
		Timeout: 30 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register local vendor: %v", err)
		return
	}

	// Test cost optimization
	ctx := context.Background()
	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Explain quantum computing in simple terms."},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	fmt.Println("  Testing cost optimization...")
	resp, err := disp.Send(ctx, req)
	if err != nil {
		fmt.Printf("  ❌ Cost optimization test failed: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Cost optimization response from %s: %s\n", resp.Vendor, resp.Content)
}

func testStreaming() {
	// Create dispatcher
	config := &models.Config{
		DefaultVendor: "local",
		Timeout:       30 * time.Second,
		EnableLogging: true,
	}

	disp := dispatcher.NewWithConfig(config)

	// Register local vendor
	localConfig := &models.VendorConfig{
		APIKey: "dummy",
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
			"model_path": "llama2:7b",
		},
		Timeout: 30 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register local vendor: %v", err)
		return
	}

	// Test streaming
	ctx := context.Background()
	req := &models.Request{
		Model: "llama2:7b",
		Messages: []models.Message{
			{Role: "user", Content: "Write a haiku about AI."},
		},
		Temperature: 0.8,
		MaxTokens:   50,
	}

	fmt.Println("  Testing streaming...")
	streamResp, err := disp.SendStreaming(ctx, req)
	if err != nil {
		fmt.Printf("  ❌ Streaming test failed: %v\n", err)
		return
	}

	fmt.Print("  ✅ Streaming response: ")
	for {
		select {
		case content := <-streamResp.ContentChan:
			fmt.Print(content)
		case <-streamResp.DoneChan:
			fmt.Println("\n  [Stream completed]")
			return
		case err := <-streamResp.ErrorChan:
			fmt.Printf("\n  ❌ Streaming error: %v\n", err)
			return
		}
	}
}

func testErrorHandling() {
	// Create dispatcher
	config := &models.Config{
		DefaultVendor:  "local",
		FallbackVendor: "openai", // Fallback to OpenAI if local fails
		Timeout:        30 * time.Second,
		EnableLogging:  true,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:      2,
			BackoffStrategy: models.ExponentialBackoff,
			RetryableErrors: []string{
				"server unavailable",
				"timeout",
				"model not found",
			},
		},
	}

	disp := dispatcher.NewWithConfig(config)

	// Register local vendor with invalid configuration
	localConfig := &models.VendorConfig{
		APIKey: "dummy",
		Headers: map[string]string{
			"server_url": "http://localhost:9999", // Invalid port
			"model_path": "nonexistent-model",
		},
		Timeout: 5 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register local vendor: %v", err)
		return
	}

	// Test error handling
	ctx := context.Background()
	req := &models.Request{
		Model: "nonexistent-model",
		Messages: []models.Message{
			{Role: "user", Content: "This should fail."},
		},
		Temperature: 0.7,
		MaxTokens:   50,
	}

	fmt.Println("  Testing error handling...")
	resp, err := disp.Send(ctx, req)
	if err != nil {
		fmt.Printf("  ✅ Error handling test passed: %v\n", err)
		return
	}

	fmt.Printf("  ❌ Unexpected success: %s\n", resp.Content)
}

// RunLocalDemo runs the local vendor demo
func RunLocalDemo() {
	fmt.Println("🎯 Running Local Vendor Demo")
	fmt.Println("=============================")
	fmt.Println("This demo will test various aspects of the local vendor:")
	fmt.Println("1. HTTP mode (Ollama)")
	fmt.Println("2. Process mode (llama.cpp)")
	fmt.Println("3. Cost optimization")
	fmt.Println("4. Streaming")
	fmt.Println("5. Error handling")
	fmt.Println()

	LocalVendorDemo()
}
