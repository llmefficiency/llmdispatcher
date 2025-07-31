# Architecture Overview

## Project Structure

```
llmdispatcher/
├── cmd/                    # Command-line applications
│   └── example/           # Example application demonstrating usage
├── internal/              # Private implementation details
│   ├── dispatcher/        # Core dispatcher logic
│   ├── models/           # Data models and types
│   └── vendors/          # Vendor-specific implementations
├── pkg/                   # Public API package
│   └── llmdispatcher/    # Main library package
├── scripts/              # Build and test scripts
└── docs/                 # Documentation
```

## Design Principles

### 1. Separation of Concerns
- **Public API** (`pkg/llmdispatcher/`): Clean, stable interface for consumers
- **Internal Implementation** (`internal/`): Private implementation details
- **Vendor Abstraction**: Common interface for all LLM vendors

### 2. Interface-Based Design
The dispatcher uses interfaces to abstract vendor implementations:

```go
type Vendor interface {
    Name() string
    SendRequest(ctx context.Context, req *Request) (*Response, error)
    SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error)
    GetCapabilities() Capabilities
    IsAvailable(ctx context.Context) bool
}
```

### 3. Configuration-Driven
- Centralized configuration management
- Environment variable support
- File-based configuration (YAML/JSON)
- Runtime configuration updates

## Core Components

### Dispatcher
The central orchestrator that:
- Manages vendor registration and lifecycle
- Implements routing logic
- Handles retry policies and fallbacks
- Tracks metrics and statistics
- Provides streaming support

### Vendor Implementations
Each vendor implements the common `Vendor` interface:
- **OpenAI**: GPT models with streaming support
- **Anthropic**: Claude models with large context windows
- **Google**: Gemini models with massive context
- **Azure OpenAI**: Enterprise OpenAI deployment

### Models and Types
- **Request/Response**: Standardized message formats
- **Configuration**: Flexible configuration structures
- **Statistics**: Comprehensive metrics tracking
- **Errors**: Vendor-specific error handling

## Data Flow

```
User Request → Dispatcher → Routing Logic → Vendor Selection → Vendor API → Response Processing → User Response
```

### Request Flow
1. **Request Creation**: User creates a `Request` with model and messages
2. **Routing Decision**: Dispatcher applies routing rules and vendor selection
3. **Vendor Execution**: Selected vendor processes the request
4. **Response Processing**: Dispatcher processes and formats the response
5. **Statistics Update**: Metrics are updated for monitoring

### Streaming Flow
1. **Stream Request**: User requests streaming response
2. **Channel Creation**: Dispatcher creates communication channels
3. **Vendor Streaming**: Vendor streams chunks to channels
4. **Real-time Processing**: User receives chunks as they arrive
5. **Completion**: Channels are closed when streaming completes

## Error Handling Strategy

### Multi-Level Error Handling
1. **Vendor-Level**: Vendor-specific error handling and retries
2. **Dispatcher-Level**: Global retry policies and fallbacks
3. **Application-Level**: User-defined error handling

### Error Types
- **Retryable Errors**: Rate limits, timeouts, temporary failures
- **Non-Retryable Errors**: Authentication, invalid requests
- **Fallback Triggers**: Vendor unavailability, quota exceeded

## Performance Considerations

### Concurrency
- Thread-safe dispatcher design
- Concurrent vendor operations
- Channel-based streaming communication
- Connection pooling for HTTP clients

### Caching
- Vendor capability caching
- Configuration caching
- Response caching (future enhancement)

### Resource Management
- Automatic channel cleanup
- Connection timeout handling
- Memory-efficient streaming

## Security Model

### API Key Management
- Environment variable loading
- Secure key storage
- Key rotation support
- Permission-based access

### Request Validation
- Input sanitization
- Model validation
- Token limit enforcement
- Rate limiting

## Testing Strategy

### Test Coverage
- **Unit Tests**: Individual component testing
- **Integration Tests**: Vendor API testing
- **End-to-End Tests**: Full workflow testing
- **Performance Tests**: Load and stress testing

### Test Structure
- **Mock Vendors**: Simulated vendor responses
- **Test Utilities**: Common testing helpers
- **Test Data**: Comprehensive test scenarios
- **Coverage Reporting**: Detailed coverage metrics

## Future Enhancements

### Planned Features
- **Plugin System**: Dynamic vendor loading
- **Advanced Caching**: Response and capability caching
- **Load Balancing**: Intelligent load distribution
- **Metrics Export**: Prometheus/OpenTelemetry integration
- **Web UI**: Management dashboard

### Scalability Improvements
- **Horizontal Scaling**: Multi-instance support
- **Database Integration**: Persistent statistics
- **Message Queues**: Asynchronous processing
- **Microservices**: Service decomposition 