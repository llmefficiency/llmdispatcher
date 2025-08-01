package models

// VendorModels contains the mapping of vendors to their available models
var VendorModels = map[string][]string{
	"openai": {
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	},
	"anthropic": {
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	},
	"google": {
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-pro",
		"gemini-pro-vision",
	},
	"azure": {
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
	},
	"local": {
		"llama2:7b",
		"llama2:13b",
		"llama2:70b",
		"llama3:8b",
		"llama3:70b",
		"mistral:7b",
		"codellama:7b",
		"codellama:13b",
	},
}

// GetVendorModels returns the list of models for a given vendor
func GetVendorModels(vendor string) []string {
	if models, exists := VendorModels[vendor]; exists {
		return models
	}
	return []string{}
}

// IsValidModel checks if a model is valid for a given vendor
func IsValidModel(vendor, model string) bool {
	models := GetVendorModels(vendor)
	for _, m := range models {
		if m == model {
			return true
		}
	}
	return false
}

// GetVendorForModel returns the vendor that supports a given model
func GetVendorForModel(model string) string {
	for vendor, models := range VendorModels {
		for _, m := range models {
			if m == model {
				return vendor
			}
		}
	}
	return ""
}
