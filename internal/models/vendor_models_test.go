package models

import (
	"testing"
)

func TestGetVendorModels(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		expected []string
	}{
		{
			name:     "openai vendor",
			vendor:   "openai",
			expected: []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo", "gpt-3.5-turbo-16k"},
		},
		{
			name:     "anthropic vendor",
			vendor:   "anthropic",
			expected: []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"},
		},
		{
			name:     "google vendor",
			vendor:   "google",
			expected: []string{"gemini-1.5-pro", "gemini-1.5-flash", "gemini-pro", "gemini-pro-vision"},
		},
		{
			name:     "azure vendor",
			vendor:   "azure",
			expected: []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"},
		},
		{
			name:     "local vendor",
			vendor:   "local",
			expected: []string{"llama2:7b", "llama2:13b", "llama2:70b", "llama3:8b", "llama3:70b", "mistral:7b", "codellama:7b", "codellama:13b"},
		},
		{
			name:     "non-existent vendor",
			vendor:   "nonexistent",
			expected: []string{},
		},
		{
			name:     "empty vendor",
			vendor:   "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVendorModels(tt.vendor)
			if len(result) != len(tt.expected) {
				t.Errorf("GetVendorModels(%s) = %v, expected %v", tt.vendor, result, tt.expected)
			}
			for i, model := range result {
				if model != tt.expected[i] {
					t.Errorf("GetVendorModels(%s)[%d] = %s, expected %s", tt.vendor, i, model, tt.expected[i])
				}
			}
		})
	}
}

func TestIsValidModel(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		expected bool
	}{
		{
			name:     "valid openai model",
			vendor:   "openai",
			model:    "gpt-4o",
			expected: true,
		},
		{
			name:     "valid anthropic model",
			vendor:   "anthropic",
			model:    "claude-3-5-sonnet-20241022",
			expected: true,
		},
		{
			name:     "valid google model",
			vendor:   "google",
			model:    "gemini-1.5-pro",
			expected: true,
		},
		{
			name:     "valid azure model",
			vendor:   "azure",
			model:    "gpt-4",
			expected: true,
		},
		{
			name:     "valid local model",
			vendor:   "local",
			model:    "llama2:7b",
			expected: true,
		},
		{
			name:     "invalid model for vendor",
			vendor:   "openai",
			model:    "invalid-model",
			expected: false,
		},
		{
			name:     "non-existent vendor",
			vendor:   "nonexistent",
			model:    "any-model",
			expected: false,
		},
		{
			name:     "empty vendor",
			vendor:   "",
			model:    "any-model",
			expected: false,
		},
		{
			name:     "empty model",
			vendor:   "openai",
			model:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidModel(tt.vendor, tt.model)
			if result != tt.expected {
				t.Errorf("IsValidModel(%s, %s) = %v, expected %v", tt.vendor, tt.model, result, tt.expected)
			}
		})
	}
}

func TestGetVendorForModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected []string // Multiple possible vendors for shared models
	}{
		{
			name:     "openai model",
			model:    "gpt-4o",
			expected: []string{"openai", "azure"}, // gpt-4o is in both
		},
		{
			name:     "anthropic model",
			model:    "claude-3-5-sonnet-20241022",
			expected: []string{"anthropic"},
		},
		{
			name:     "google model",
			model:    "gemini-1.5-pro",
			expected: []string{"google"},
		},
		{
			name:     "azure model",
			model:    "gpt-4",
			expected: []string{"openai", "azure"}, // gpt-4 is in both
		},
		{
			name:     "local model",
			model:    "llama2:7b",
			expected: []string{"local"},
		},
		{
			name:     "non-existent model",
			model:    "invalid-model",
			expected: []string{},
		},
		{
			name:     "empty model",
			model:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVendorForModel(tt.model)
			if len(tt.expected) == 0 {
				if result != "" {
					t.Errorf("GetVendorForModel(%s) = %s, expected empty string", tt.model, result)
				}
			} else {
				found := false
				for _, expected := range tt.expected {
					if result == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetVendorForModel(%s) = %s, expected one of %v", tt.model, result, tt.expected)
				}
			}
		})
	}
}

func TestVendorModelsCompleteness(t *testing.T) {
	// Test that all vendors have at least one model
	for vendor, models := range VendorModels {
		if len(models) == 0 {
			t.Errorf("Vendor %s has no models defined", vendor)
		}

		// Test that each model is unique within the vendor
		modelSet := make(map[string]bool)
		for _, model := range models {
			if modelSet[model] {
				t.Errorf("Vendor %s has duplicate model: %s", vendor, model)
			}
			modelSet[model] = true
		}
	}
}

func TestModelVendorMapping(t *testing.T) {
	// Test that each model maps back to a valid vendor (some models may be in multiple vendors)
	for _, models := range VendorModels {
		for _, model := range models {
			mappedVendor := GetVendorForModel(model)
			if mappedVendor == "" {
				t.Errorf("Model %s does not map to any vendor", model)
			}
			// Check if the mapped vendor is valid
			if _, exists := VendorModels[mappedVendor]; !exists {
				t.Errorf("Model %s maps to invalid vendor %s", model, mappedVendor)
			}
		}
	}
}
