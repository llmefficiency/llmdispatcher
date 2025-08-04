package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/llmefficiency/llmdispatcher/internal/dispatcher"
	"github.com/llmefficiency/llmdispatcher/internal/models"
	"github.com/llmefficiency/llmdispatcher/internal/vendors"
)

// WebService represents the web service
type WebService struct {
	dispatcher *dispatcher.Dispatcher
	config     *models.Config
	server     *http.Server
}

// RequestPayload represents the incoming request payload
type RequestPayload struct {
	Model       string           `json:"model"`
	Messages    []models.Message `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Stop        []string         `json:"stop,omitempty"`
	User        string           `json:"user,omitempty"`
	Vendor      string           `json:"vendor,omitempty"` // Optional vendor override
}

// ResponsePayload represents the response payload
type ResponsePayload struct {
	Success bool                    `json:"success"`
	Data    *models.Response        `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
	Stats   *models.DispatcherStats `json:"stats,omitempty"`
}

// StreamingResponsePayload represents the streaming response payload
type StreamingResponsePayload struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// VendorTestPayload represents vendor test configuration
type VendorTestPayload struct {
	Vendor      string           `json:"vendor"`
	Model       string           `json:"model"`
	Messages    []models.Message `json:"messages"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

// NewWebService creates a new web service instance
func NewWebService() *WebService {
	// Load environment variables
	if err := loadEnv(".env"); err != nil {
		log.Printf("⚠️  Could not load .env file: %v", err)
	}

	// Create dispatcher configuration
	config := &models.Config{
		DefaultVendor:  "openai",
		FallbackVendor: "anthropic",
		Timeout:        60 * time.Second,
		EnableLogging:  true,
		EnableMetrics:  true,
		RetryPolicy: &models.RetryPolicy{
			MaxRetries:      3,
			BackoffStrategy: models.ExponentialBackoff,
			RetryableErrors: []string{"rate limit exceeded", "timeout"},
		},
		RoutingRules: []models.RoutingRule{
			{
				Condition: models.RoutingCondition{
					ModelPattern: "gpt-4",
				},
				Vendor:   "openai",
				Priority: 1,
				Enabled:  true,
			},
			{
				Condition: models.RoutingCondition{
					ModelPattern: "claude",
				},
				Vendor:   "anthropic",
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	// Create dispatcher
	disp := dispatcher.NewWithConfig(config)

	// Register vendors
	registerVendors(disp)

	return &WebService{
		dispatcher: disp,
		config:     config,
	}
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

// registerVendors registers all available vendors
func registerVendors(disp *dispatcher.Dispatcher) {
	// Register OpenAI vendor
	if openaiAPIKey := os.Getenv("OPENAI_API_KEY"); openaiAPIKey != "" {
		openaiConfig := &models.VendorConfig{
			APIKey:  openaiAPIKey,
			BaseURL: "https://api.openai.com/v1",
			Timeout: 120 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		openaiVendor := vendors.NewOpenAI(openaiConfig)
		if err := disp.RegisterVendor(openaiVendor); err != nil {
			log.Printf("Failed to register OpenAI vendor: %v", err)
		} else {
			log.Println("✅ Registered OpenAI vendor")
		}
	} else {
		log.Println("⚠️  OPENAI_API_KEY not set, skipping OpenAI vendor")
	}

	// Register Anthropic vendor
	if anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicAPIKey != "" {
		anthropicConfig := &models.VendorConfig{
			APIKey:  anthropicAPIKey,
			BaseURL: "https://api.anthropic.com",
			Timeout: 120 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		anthropicVendor := vendors.NewAnthropic(anthropicConfig)
		if err := disp.RegisterVendor(anthropicVendor); err != nil {
			log.Printf("Failed to register Anthropic vendor: %v", err)
		} else {
			log.Println("✅ Registered Anthropic vendor")
		}
	} else {
		log.Println("⚠️  ANTHROPIC_API_KEY not set, skipping Anthropic vendor")
	}

	// Register Google vendor
	if googleAPIKey := os.Getenv("GOOGLE_API_KEY"); googleAPIKey != "" {
		googleConfig := &models.VendorConfig{
			APIKey:  googleAPIKey,
			BaseURL: "https://generativelanguage.googleapis.com",
			Timeout: 120 * time.Second,
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}

		googleVendor := vendors.NewGoogle(googleConfig)
		if err := disp.RegisterVendor(googleVendor); err != nil {
			log.Printf("Failed to register Google vendor: %v", err)
		} else {
			log.Println("✅ Registered Google vendor")
		}
	} else {
		log.Println("⚠️  GOOGLE_API_KEY not set, skipping Google vendor")
	}

	// Register Azure OpenAI vendor
	if azureOpenAIAPIKey := os.Getenv("AZURE_OPENAI_API_KEY"); azureOpenAIAPIKey != "" {
		azureConfig := &models.VendorConfig{
			APIKey:  azureOpenAIAPIKey,
			BaseURL: os.Getenv("AZURE_OPENAI_ENDPOINT"),
			Timeout: 120 * time.Second,
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
		log.Println("⚠️  AZURE_OPENAI_API_KEY not set, skipping Azure OpenAI vendor")
	}

	// Register Local vendor (Ollama)
	localConfig := &models.VendorConfig{
		APIKey: "dummy", // Not used for local models
		Headers: map[string]string{
			"server_url": "http://localhost:11434",
			"model_path": "llama2:7b",
		},
		Timeout: 120 * time.Second,
	}

	localVendor := vendors.NewLocal(localConfig)
	if err := disp.RegisterVendor(localVendor); err != nil {
		log.Printf("Failed to register Local vendor: %v", err)
	} else {
		log.Println("✅ Registered Local vendor")
	}
}

// setupRoutes sets up the HTTP routes
func (ws *WebService) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", ws.healthHandler).Methods("GET")

	// Chat completion (direct reply)
	api.HandleFunc("/chat/completions", ws.chatCompletionsHandler).Methods("POST")

	// Streaming chat completion
	api.HandleFunc("/chat/completions/stream", ws.streamingChatCompletionsHandler).Methods("POST")

	// Vendor test endpoint
	api.HandleFunc("/test/vendor", ws.testVendorHandler).Methods("POST")

	// Statistics endpoint
	api.HandleFunc("/stats", ws.statsHandler).Methods("GET")

	// Vendors endpoint
	api.HandleFunc("/vendors", ws.vendorsHandler).Methods("GET")

	// Models endpoint
	api.HandleFunc("/models", ws.modelsHandler).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("apps/server/static"))
	router.PathPrefix("/").Handler(fs)

	// CORS middleware
	router.Use(corsMiddleware)

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthHandler handles health check requests
func (ws *WebService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get available vendors with enabled status (same logic as vendorsHandler)
	availableVendors := ws.dispatcher.GetVendors()

	// Create vendor info with enabled status
	vendorInfo := make([]map[string]interface{}, 0)

	// Check which vendors have API keys configured
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	googleKey := os.Getenv("GOOGLE_API_KEY")
	azureKey := os.Getenv("AZURE_OPENAI_API_KEY")

	// Check if keys are valid (not placeholder values)
	openaiEnabled := openaiKey != "" && !strings.Contains(openaiKey, "your-") && (strings.HasPrefix(openaiKey, "sk-") || strings.HasPrefix(openaiKey, "sk-proj-"))
	anthropicEnabled := anthropicKey != "" && !strings.Contains(anthropicKey, "your-") && (strings.HasPrefix(anthropicKey, "sk-ant-") || strings.HasPrefix(anthropicKey, "sk-ant-api"))
	googleEnabled := googleKey != "" && !strings.Contains(googleKey, "your-") && !strings.Contains(googleKey, "here")
	azureEnabled := azureKey != "" && !strings.Contains(azureKey, "your-") && !strings.Contains(azureKey, "here")
	localEnabled := true // Local is always available

	for _, vendor := range availableVendors {
		enabled := false
		switch vendor {
		case "openai":
			enabled = openaiEnabled
		case "anthropic":
			enabled = anthropicEnabled
		case "google":
			enabled = googleEnabled
		case "azure":
			enabled = azureEnabled
		case "local":
			enabled = localEnabled
		}

		vendorInfo = append(vendorInfo, map[string]interface{}{
			"name":    vendor,
			"enabled": enabled,
		})
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"vendors":   vendorInfo,
		"timestamp": time.Now().UTC(),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// chatCompletionsHandler handles direct chat completion requests
func (ws *WebService) chatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to internal request
	req := &models.Request{
		Model:       payload.Model,
		Messages:    payload.Messages,
		Temperature: payload.Temperature,
		MaxTokens:   payload.MaxTokens,
		TopP:        payload.TopP,
		Stream:      false, // Force non-streaming for this endpoint
		Stop:        payload.Stop,
		User:        payload.User,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), ws.config.Timeout)
	defer cancel()

	// Send request
	var response *models.Response
	var err error

	if payload.Vendor != "" {
		// Use specific vendor if provided
		response, err = ws.dispatcher.SendToVendor(ctx, payload.Vendor, req)
	} else {
		// Use automatic vendor selection
		response, err = ws.dispatcher.Send(ctx, req)
	}
	if err != nil {
		responsePayload := ResponsePayload{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	responsePayload := ResponsePayload{
		Success: true,
		Data:    response,
		Stats:   ws.dispatcher.GetStats(),
	}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// streamingChatCompletionsHandler handles streaming chat completion requests
func (ws *WebService) streamingChatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	// Set proper headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		// Send error as SSE
		fmt.Fprintf(w, "data: [ERROR] Invalid request body: %v\n\n", err)
		w.(http.Flusher).Flush()
		return
	}

	// Convert to internal request
	req := &models.Request{
		Model:       payload.Model,
		Messages:    payload.Messages,
		Temperature: payload.Temperature,
		MaxTokens:   payload.MaxTokens,
		TopP:        payload.TopP,
		Stream:      true, // Force streaming for this endpoint
		Stop:        payload.Stop,
		User:        payload.User,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		// Send error as SSE
		fmt.Fprintf(w, "data: [ERROR] Invalid request: %v\n\n", err)
		w.(http.Flusher).Flush()
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), ws.config.Timeout)
	defer cancel()

	// Send streaming request
	var streamResp *models.StreamingResponse
	var err error

	if payload.Vendor != "" {
		// Use specific vendor if provided
		streamResp, err = ws.dispatcher.SendStreamingToVendor(ctx, payload.Vendor, req)
	} else {
		// Use automatic vendor selection
		streamResp, err = ws.dispatcher.SendStreaming(ctx, req)
	}
	if err != nil {
		// Send error as SSE
		fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
		w.(http.Flusher).Flush()
		return
	}

	// Stream the response
	done := false
	for !done {
		select {
		case chunk, ok := <-streamResp.ContentChan:
			if !ok {
				done = true
			} else {
				// Send chunk as Server-Sent Events
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				w.(http.Flusher).Flush()
			}
		case err := <-streamResp.ErrorChan:
			if err != nil {
				fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
				w.(http.Flusher).Flush()
			}
			done = true
		case <-streamResp.DoneChan:
			done = true
		case <-ctx.Done():
			done = true
		}
	}

	// Close the streaming response
	streamResp.Close()
}

// testVendorHandler handles vendor testing requests
func (ws *WebService) testVendorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var payload VendorTestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create request
	req := &models.Request{
		Model:       payload.Model,
		Messages:    payload.Messages,
		Temperature: payload.Temperature,
		MaxTokens:   payload.MaxTokens,
		Stream:      payload.Stream,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get vendor
	vendor, exists := ws.dispatcher.GetVendor(payload.Vendor)
	if !exists {
		http.Error(w, fmt.Sprintf("Vendor %s not found", payload.Vendor), http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), ws.config.Timeout)
	defer cancel()

	var response *models.Response
	var err error

	if payload.Stream {
		// Test streaming
		streamResp, err := vendor.SendStreamingRequest(ctx, req)
		if err != nil {
			responsePayload := ResponsePayload{
				Success: false,
				Error:   err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}

		// For testing, we'll collect the streamed content
		var content strings.Builder
		done := false
		for !done {
			select {
			case chunk, ok := <-streamResp.ContentChan:
				if !ok {
					done = true
				} else {
					content.WriteString(chunk)
				}
			case err := <-streamResp.ErrorChan:
				if err != nil {
					responsePayload := ResponsePayload{
						Success: false,
						Error:   err.Error(),
					}
					w.WriteHeader(http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
						http.Error(w, "Failed to encode response", http.StatusInternalServerError)
					}
					return
				}
				done = true
			case <-streamResp.DoneChan:
				done = true
			case <-ctx.Done():
				done = true
			}
		}

		streamResp.Close()

		// Create response from streamed content
		response = &models.Response{
			Content:   content.String(),
			Model:     req.Model,
			Vendor:    payload.Vendor,
			CreatedAt: time.Now(),
		}
	} else {
		// Test direct request
		response, err = vendor.SendRequest(ctx, req)
		if err != nil {
			responsePayload := ResponsePayload{
				Success: false,
				Error:   err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}
	}

	// Return success response
	responsePayload := ResponsePayload{
		Success: true,
		Data:    response,
		Stats:   ws.dispatcher.GetStats(),
	}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// statsHandler handles statistics requests
func (ws *WebService) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ws.dispatcher.GetStats()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// vendorsHandler handles vendors list requests
func (ws *WebService) vendorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get available vendors
	availableVendors := ws.dispatcher.GetVendors()

	// Create vendor info with enabled status
	vendorInfo := make([]map[string]interface{}, 0)

	// Check which vendors have API keys configured
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	googleKey := os.Getenv("GOOGLE_API_KEY")
	azureKey := os.Getenv("AZURE_OPENAI_API_KEY")

	// Check if keys are valid (not placeholder values)
	openaiEnabled := openaiKey != "" && !strings.Contains(openaiKey, "your-") && (strings.HasPrefix(openaiKey, "sk-") || strings.HasPrefix(openaiKey, "sk-proj-"))
	anthropicEnabled := anthropicKey != "" && !strings.Contains(anthropicKey, "your-") && (strings.HasPrefix(anthropicKey, "sk-ant-") || strings.HasPrefix(anthropicKey, "sk-ant-api"))
	googleEnabled := googleKey != "" && !strings.Contains(googleKey, "your-") && !strings.Contains(googleKey, "here")
	azureEnabled := azureKey != "" && !strings.Contains(azureKey, "your-") && !strings.Contains(azureKey, "here")
	localEnabled := true // Local is always available

	for _, vendor := range availableVendors {
		enabled := false
		switch vendor {
		case "openai":
			enabled = openaiEnabled
		case "anthropic":
			enabled = anthropicEnabled
		case "google":
			enabled = googleEnabled
		case "azure":
			enabled = azureEnabled
		case "local":
			enabled = localEnabled
		}

		vendorInfo = append(vendorInfo, map[string]interface{}{
			"name":    vendor,
			"enabled": enabled,
		})
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"vendors": vendorInfo,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// modelsHandler handles model listing requests
func (ws *WebService) modelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get vendor from query parameter
	vendor := r.URL.Query().Get("vendor")

	if vendor != "" {
		// Return models for specific vendor
		models := models.GetVendorModels(vendor)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"vendor": vendor,
			"models": models,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Return all vendor models
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models.VendorModels,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Start starts the web service
func (ws *WebService) Start(addr string) error {
	router := ws.setupRoutes()

	ws.server = &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
	}

	log.Printf("🚀 Starting web service on %s", addr)
	log.Printf("📊 Available endpoints:")
	log.Printf("   GET  /api/v1/health")
	log.Printf("   POST /api/v1/chat/completions")
	log.Printf("   POST /api/v1/chat/completions/stream")
	log.Printf("   POST /api/v1/test/vendor")
	log.Printf("   GET  /api/v1/stats")
	log.Printf("   GET  /api/v1/vendors")

	return ws.server.ListenAndServe()
}

// Shutdown gracefully shuts down the web service
func (ws *WebService) Shutdown(ctx context.Context) error {
	if ws.server != nil {
		return ws.server.Shutdown(ctx)
	}
	return nil
}

func main() {
	// Create web service
	ws := NewWebService()

	// Start the service
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Fatal(ws.Start(addr))
}
