#!/bin/bash

# LLM Dispatcher Setup Script
# This script helps you set up your API keys for the LLM dispatcher

echo "🔧 LLM Dispatcher Setup"
echo "========================"
echo ""

# Check if .env file exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists. Backing up to .env.backup"
    cp .env .env.backup
fi

# Copy example environment file
if [ -f "cmd/example/env.example" ]; then
    cp cmd/example/env.example .env
    echo "✅ Created .env file from template"
else
    echo "❌ Could not find cmd/example/env.example"
    exit 1
fi

echo ""
echo "📝 Please edit .env and add your API keys:"
echo ""

# Show the current .env file with instructions
cat .env | while IFS= read -r line; do
    if [[ $line =~ ^[[:space:]]*# ]]; then
        # Comment line
        echo "$line"
    elif [[ $line =~ _API_KEY= ]]; then
        # API key line
        echo "$line"
        echo "   ↑ Add your actual API key here"
    else
        # Other configuration line
        echo "$line"
    fi
done

echo ""
echo "🔑 Required API Keys:"
echo "  - OPENAI_API_KEY: Get from https://platform.openai.com/api-keys"
echo "  - ANTHROPIC_API_KEY: Get from https://console.anthropic.com/"
echo "  - GOOGLE_API_KEY: Get from https://makersuite.google.com/app/apikey"
echo "  - AZURE_OPENAI_API_KEY: Get from Azure OpenAI service"
echo "  - COHERE_API_KEY: Get from https://dashboard.cohere.ai/api-keys"
echo ""
echo "💡 Optional: You can start with just OPENAI_API_KEY for basic functionality"
echo ""
echo "🚀 After editing .env, run: go run cmd/example/main.go" 