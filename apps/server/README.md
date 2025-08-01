# LLM Dispatcher Web Service

A web service that refactors the CLI logic into a REST API with support for both direct reply and streaming modes, and vendor testing capabilities.

## Features

- **Direct Reply Mode**: Get complete responses from LLM vendors
- **Streaming Mode**: Real-time streaming responses using Server-Sent Events
- **Vendor Testing**: Test specific vendors with custom configurations
- **Statistics**: Monitor dispatcher performance and vendor usage
- **Web Interface**: Beautiful HTML client for easy testing
- **CORS Support**: Cross-origin requests enabled
- **Health Checks**: Service health monitoring

## Quick Start

1. **Set up environment variables** (create a `.env` file):
```bash
OPENAI_API_KEY=your_openai_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key
GOOGLE_API_KEY=your_google_api_key
AZURE_OPENAI_API_KEY=your_azure_api_key
AZURE_OPENAI_ENDPOINT=your_azure_endpoint
```

2. **Run the web service**:
```bash
cd cmd/webservice
go run main.go
```

3. **Access the web interface**:
   - Open your browser and go to `http://localhost:8080`
   - Use the interactive web interface to test the service

## API Endpoints

### Health Check
```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "vendors": ["openai", "anthropic", "local"],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Direct Chat Completion
```http
POST /api/v1/chat/completions
Content-Type: application/json

{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "Hello! Can you tell me a short joke?"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 100
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "content": "Why don't scientists trust atoms? Because they make up everything!",
    "usage": {
      "prompt_tokens": 10,
      "completion_tokens": 15,
      "total_tokens": 25
    },
    "model": "gpt-3.5-turbo",
    "vendor": "openai",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "stats": {
    "total_requests": 1,
    "successful_requests": 1,
    "failed_requests": 0,
    "average_latency": "1.2s"
  }
}
```

### Streaming Chat Completion
```http
POST /api/v1/chat/completions/stream
Content-Type: application/json

{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "Write a short story about a robot learning to paint."
    }
  ],
  "temperature": 0.8,
  "max_tokens": 200
}
```

**Response:** Server-Sent Events stream
```
data: Once upon a time, in a world where robots and humans coexisted...
data: there was a curious robot named Pixel who lived in a small studio...
data: Unlike other robots who focused on efficiency and precision...
```

### Vendor Testing
```http
POST /api/v1/test/vendor
Content-Type: application/json

{
  "vendor": "anthropic",
  "model": "claude-3-haiku-20240307",
  "messages": [
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 100,
  "stream": false
}
```

### Get Statistics
```http
GET /api/v1/stats
```

**Response:**
```json
{
  "total_requests": 10,
  "successful_requests": 9,
  "failed_requests": 1,
  "average_latency": "1.5s",
  "vendor_stats": {
    "openai": {
      "requests": 5,
      "successes": 5,
      "failures": 0,
      "average_latency": "1.2s"
    },
    "anthropic": {
      "requests": 3,
      "successes": 3,
      "failures": 0,
      "average_latency": "2.1s"
    }
  }
}
```

### Get Available Vendors
```http
GET /api/v1/vendors
```

**Response:**
```json
{
  "vendors": ["openai", "anthropic", "google", "azure", "local"]
}
```

## Web Interface

The web service includes a beautiful HTML interface accessible at `http://localhost:8080` that provides:

- **Health Check**: Monitor service status and available vendors
- **Direct Chat**: Test direct completion requests
- **Streaming Chat**: Test real-time streaming responses
- **Vendor Testing**: Test specific vendors with custom configurations
- **Statistics**: View dispatcher performance metrics

## Configuration

The web service automatically registers vendors based on available environment variables:

- **OpenAI**: Requires `OPENAI_API_KEY`
- **Anthropic**: Requires `ANTHROPIC_API_KEY`
- **Google**: Requires `GOOGLE_API_KEY`
- **Azure OpenAI**: Requires `AZURE_OPENAI_API_KEY` and `AZURE_OPENAI_ENDPOINT`
- **Local (Ollama)**: Automatically registered (requires Ollama running on `http://localhost:11434`)

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `OPENAI_API_KEY` | OpenAI API key | No |
| `ANTHROPIC_API_KEY` | Anthropic API key | No |
| `GOOGLE_API_KEY` | Google API key | No |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI API key | No |
| `AZURE_OPENAI_ENDPOINT` | Azure OpenAI endpoint | No |
| `PORT` | Server port (default: 8080) | No |

## Testing with curl

### Direct Request
```bash
curl -X POST http://localhost:8080/api/v1/chat/completions \
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
  }'
```

### Streaming Request
```bash
curl -X POST http://localhost:8080/api/v1/chat/completions/stream \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "Write a short story about a robot learning to paint."
      }
    ],
    "temperature": 0.8,
    "max_tokens": 200
  }'
```

### Vendor Test
```bash
curl -X POST http://localhost:8080/api/v1/test/vendor \
  -H "Content-Type: application/json" \
  -d '{
    "vendor": "anthropic",
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 100,
    "stream": false
  }'
```

## Architecture

The web service refactors the CLI logic into a REST API with the following components:

- **WebService**: Main service struct managing the dispatcher and HTTP server
- **Request/Response Payloads**: JSON structures for API communication
- **Handlers**: HTTP handlers for different endpoints
- **Static Files**: HTML client for easy testing
- **CORS Middleware**: Cross-origin request support

## Error Handling

The service provides comprehensive error handling:

- **400 Bad Request**: Invalid request format or validation errors
- **500 Internal Server Error**: Dispatcher or vendor errors
- **JSON Error Responses**: Detailed error messages in JSON format

## Performance

- **Timeout Configuration**: Configurable request timeouts
- **Retry Policy**: Automatic retry with exponential backoff
- **Statistics Tracking**: Real-time performance monitoring
- **Vendor Routing**: Intelligent routing based on model patterns

## Security

- **CORS Headers**: Proper cross-origin request handling
- **Input Validation**: Request validation and sanitization
- **Error Sanitization**: Safe error message handling
- **Environment Variables**: Secure API key management 