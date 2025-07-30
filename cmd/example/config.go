package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/llmefficiency/llmdispatcher/pkg/llmdispatcher"
)

// VendorConfig holds configuration for all vendors
type VendorConfig struct {
	OpenAI      *llmdispatcher.VendorConfig
	Anthropic   *llmdispatcher.VendorConfig
	Google      *llmdispatcher.VendorConfig
	AzureOpenAI *llmdispatcher.VendorConfig
	Cohere      *llmdispatcher.VendorConfig
}

// LoadVendorConfigs loads vendor configurations from environment variables
func LoadVendorConfigs() *VendorConfig {
	config := &VendorConfig{}

	// OpenAI Configuration
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI = &llmdispatcher.VendorConfig{
			APIKey:  apiKey,
			BaseURL: getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			Timeout: parseDuration(getEnvOrDefault("OPENAI_TIMEOUT", "30s")),
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}
	}

	// Anthropic Configuration
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.Anthropic = &llmdispatcher.VendorConfig{
			APIKey:  apiKey,
			BaseURL: getEnvOrDefault("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
			Timeout: parseDuration(getEnvOrDefault("ANTHROPIC_TIMEOUT", "30s")),
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}
	}

	// Google Configuration
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		config.Google = &llmdispatcher.VendorConfig{
			APIKey:  apiKey,
			BaseURL: getEnvOrDefault("GOOGLE_BASE_URL", "https://generativelanguage.googleapis.com"),
			Timeout: parseDuration(getEnvOrDefault("GOOGLE_TIMEOUT", "30s")),
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}
	}

	// Azure OpenAI Configuration
	if apiKey := os.Getenv("AZURE_OPENAI_API_KEY"); apiKey != "" {
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		if endpoint == "" {
			fmt.Println("⚠️  AZURE_OPENAI_ENDPOINT not set, skipping Azure OpenAI")
		} else {
			config.AzureOpenAI = &llmdispatcher.VendorConfig{
				APIKey:  apiKey,
				BaseURL: endpoint,
				Timeout: parseDuration(getEnvOrDefault("AZURE_OPENAI_TIMEOUT", "30s")),
				Headers: map[string]string{
					"User-Agent": "llmdispatcher/1.0",
				},
			}
		}
	}

	// Cohere Configuration
	if apiKey := os.Getenv("COHERE_API_KEY"); apiKey != "" {
		config.Cohere = &llmdispatcher.VendorConfig{
			APIKey:  apiKey,
			BaseURL: getEnvOrDefault("COHERE_BASE_URL", "https://api.cohere.ai"),
			Timeout: parseDuration(getEnvOrDefault("COHERE_TIMEOUT", "30s")),
			Headers: map[string]string{
				"User-Agent": "llmdispatcher/1.0",
			},
		}
	}

	return config
}

// PrintVendorStatus prints the status of all vendor configurations
func PrintVendorStatus(config *VendorConfig) {
	fmt.Println("\n🔑 Vendor Configuration Status:")

	vendors := []struct {
		name   string
		config *llmdispatcher.VendorConfig
	}{
		{"OpenAI", config.OpenAI},
		{"Anthropic", config.Anthropic},
		{"Google", config.Google},
		{"Azure OpenAI", config.AzureOpenAI},
		{"Cohere", config.Cohere},
	}

	for _, vendor := range vendors {
		if vendor.config != nil {
			fmt.Printf("✅ %s: Configured\n", vendor.name)
		} else {
			fmt.Printf("❌ %s: Not configured\n", vendor.name)
		}
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string (simplified version)
func parseDuration(duration string) time.Duration {
	// This is a simplified parser - in production you'd want more robust parsing
	if strings.HasSuffix(duration, "s") {
		seconds := strings.TrimSuffix(duration, "s")
		if seconds == "30" {
			return 30 * time.Second
		}
	}
	return 30 * time.Second // default
}

// Example of loading from a .env file (you'd need a .env parser library)
func LoadFromEnvFile(filename string) error {
	// In a real application, you might use a library like "github.com/joho/godotenv"
	// to load environment variables from a .env file

	// Example .env file content:
	// OPENAI_API_KEY=sk-your-openai-key
	// ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
	// GOOGLE_API_KEY=your-google-key
	// AZURE_OPENAI_API_KEY=your-azure-key
	// AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
	// COHERE_API_KEY=your-cohere-key

	fmt.Printf("📁 Loading environment from %s (not implemented)\n", filename)
	return nil
}

// Example of loading from a configuration file
func LoadFromConfigFile(filename string) (*VendorConfig, error) {
	// In a real application, you might use a library like "gopkg.in/yaml.v3"
	// to load configuration from a YAML or JSON file

	// Example config.yaml:
	// vendors:
	//   openai:
	//     api_key: ${OPENAI_API_KEY}
	//     base_url: https://api.openai.com/v1
	//     timeout: 30s
	//   anthropic:
	//     api_key: ${ANTHROPIC_API_KEY}
	//     base_url: https://api.anthropic.com
	//     timeout: 30s

	fmt.Printf("📄 Loading configuration from %s (not implemented)\n", filename)
	return &VendorConfig{}, nil
}
