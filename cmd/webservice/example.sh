#!/bin/bash

# LLM Dispatcher Web Service Examples
# This script demonstrates how to use the web service API endpoints

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 LLM Dispatcher Web Service Examples"
echo "======================================"

# Check if the service is running
echo ""
echo "1. Checking service health..."
curl -s "$BASE_URL/health" | jq '.'

# Get available vendors
echo ""
echo "2. Getting available vendors..."
curl -s "$BASE_URL/vendors" | jq '.'

# Test direct chat completion
echo ""
echo "3. Testing direct chat completion..."
curl -s -X POST "$BASE_URL/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "Hello! Can you tell me a short joke?"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 100
  }' | jq '.'

# Test vendor testing (if you have API keys configured)
echo ""
echo "4. Testing vendor testing endpoint..."
curl -s -X POST "$BASE_URL/test/vendor" \
  -H "Content-Type: application/json" \
  -d '{
    "vendor": "local",
    "model": "llama2:7b",
    "messages": [
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 100,
    "stream": false
  }' | jq '.'

# Get statistics
echo ""
echo "5. Getting dispatcher statistics..."
curl -s "$BASE_URL/stats" | jq '.'

echo ""
echo "✅ Examples completed!"
echo ""
echo "To test streaming, you can use:"
echo "curl -X POST $BASE_URL/chat/completions/stream \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"model\":\"gpt-3.5-turbo\",\"messages\":[{\"role\":\"user\",\"content\":\"Write a story\"}],\"temperature\":0.8,\"max_tokens\":200}'"
echo ""
echo "Or visit http://localhost:8080 for the web interface!" 