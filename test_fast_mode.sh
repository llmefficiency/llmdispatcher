#!/bin/bash

echo "🧪 Testing FastMode functionality..."

# Start the server in background
echo "🚀 Starting web server..."
go run apps/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "📡 Testing FastMode request..."

# Test FastMode request
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "fast",
    "messages": [{"role": "user", "content": "Hello! Can you tell me a short joke?"}]
  }' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "📊 Testing stats endpoint..."

# Test stats endpoint
curl -X GET "http://localhost:8080/api/v1/stats?mode=fast" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "🔄 Testing mode comparison..."

# Test mode comparison
curl -X GET "http://localhost:8080/api/v1/stats/modes" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

# Stop the server
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo "✅ Test completed!" 