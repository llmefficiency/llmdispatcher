# Development Guide

## Getting Started

### Prerequisites
- Go 1.21+ (latest stable version recommended)
- Git
- Make (optional, for build automation)

### Environment Setup

1. **Clone the repository**:
```bash
git clone https://github.com/llmefficiency/llmdispatcher.git
cd llmdispatcher
```

2. **Install dependencies**:
```bash
go mod download
```

3. **Set up environment variables**:
```bash
cp cmd/example/env.example .env
# Edit .env with your API keys
```

## Project Structure for Developers

### Key Directories

#### `pkg/llmdispatcher/`
Public API package - this is what users import:
- `dispatcher.go`: Main dispatcher interface
- `types.go`: Public types and interfaces
- `vendors.go`: Vendor factory functions

#### `internal/`
Private implementation details:
- `dispatcher/`: Core dispatcher logic
- `models/`: Data models and configuration
- `vendors/`: Vendor-specific implementations

#### `cmd/example/`
Example application demonstrating usage:
- `cli.go`: Complete usage example with CLI interface
- `config.go`: Configuration loading examples
- `setup.sh`: Environment setup script

### CLI Testing

The example application provides a command-line interface for testing:

```bash
# Test vendor mode with default vendor
go run cmd/example/cli.go --vendor

# Test vendor mode with specific vendor
go run cmd/example/cli.go --vendor --vendor-override anthropic

# Test local mode with Ollama
go run cmd/example/cli.go --local

# Test with custom model
go run cmd/example/cli.go --local --model llama2:13b
```

## Development Workflow

### 1. Making Changes

#### Adding a New Vendor
1. Create vendor implementation in `internal/vendors/`
2. Implement the `Vendor` interface
3. Add factory function in `pkg/llmdispatcher/vendors.go`
4. Add tests in `internal/vendors/`
5. Update documentation

#### Modifying the Dispatcher
1. Update `internal/dispatcher/dispatcher.go`
2. Update public interface in `pkg/llmdispatcher/dispatcher.go`
3. Add/update tests
4. Update documentation

#### Adding New Features
1. Design the feature in `internal/`
2. Expose public API in `pkg/llmdispatcher/`
3. Add comprehensive tests
4. Update documentation and examples

### 2. Testing

#### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test files
go test ./internal/dispatcher/
go test ./pkg/llmdispatcher/
```

#### Test Structure
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test vendor interactions
- **Mock Tests**: Test with simulated responses
- **Benchmark Tests**: Performance testing

#### Test Utilities
```go
// Mock vendor for testing
mockVendor := &MockVendor{
    NameFunc: func() string { return "mock" },
    SendRequestFunc: func(ctx context.Context, req *Request) (*Response, error) {
        return &Response{Content: "mock response"}, nil
    },
}

// Test dispatcher with mock vendor
dispatcher := llmdispatcher.New()
dispatcher.RegisterVendor(mockVendor)
```

### 3. Code Quality

#### Code Style
- Follow Go standard formatting (`gofmt`)
- Use `golint` for style checking
- Follow Go naming conventions
- Add comprehensive comments for public APIs

#### Pre-commit Checks
```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test ./...

# Check coverage
go test -cover ./...
```

### 4. Documentation

#### Code Documentation
- Document all public functions and types
- Include usage examples in comments
- Keep documentation up-to-date with code changes

#### API Documentation
- Update README.md for user-facing changes
- Update ARCHITECTURE.md for architectural changes
- Add examples for new features

## Adding New Vendors

### Step-by-Step Guide

1. **Create Vendor File**
```bash
touch internal/vendors/newvendor.go
touch internal/vendors/newvendor_test.go
```

2. **Implement Vendor Interface**
```go
type NewVendor struct {
    config *VendorConfig
    client *http.Client
}

func (v *NewVendor) Name() string {
    return "newvendor"
}

func (v *NewVendor) SendRequest(ctx context.Context, req *Request) (*Response, error) {
    // Implementation
}

func (v *NewVendor) SendStreamingRequest(ctx context.Context, req *Request) (*StreamingResponse, error) {
    // Implementation
}

func (v *NewVendor) GetCapabilities() Capabilities {
    return Capabilities{
        Models: []string{"model1", "model2"},
        MaxTokens: 4096,
        SupportsStreaming: true,
    }
}

func (v *NewVendor) IsAvailable(ctx context.Context) bool {
    // Health check implementation
}
```

3. **Add Factory Function**
```go
// In pkg/llmdispatcher/vendors.go
func NewNewVendor(config *VendorConfig) Vendor {
    return &NewVendor{
        config: config,
        client: &http.Client{
            Timeout: config.Timeout,
        },
    }
}
```

4. **Add Tests**
```go
// In internal/vendors/newvendor_test.go
func TestNewVendor_SendRequest(t *testing.T) {
    vendor := &NewVendor{
        config: &VendorConfig{},
    }
    
    req := &Request{
        Model: "model1",
        Messages: []Message{
            {Role: "user", Content: "Hello"},
        },
    }
    
    resp, err := vendor.SendRequest(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

5. **Update Documentation**
- Add vendor to README.md
- Update supported models list
- Add configuration examples

## Configuration Management

### Environment Variables
```bash
# Vendor-specific configuration
NEWVENDOR_API_KEY=your-api-key
NEWVENDOR_BASE_URL=https://api.newvendor.com
NEWVENDOR_TIMEOUT=30s
```

### Configuration Files
```yaml
# config.yaml
vendors:
  newvendor:
    api_key: ${NEWVENDOR_API_KEY}
    base_url: ${NEWVENDOR_BASE_URL}
    timeout: 30s
```

### Runtime Configuration
```go
config := &Config{
    DefaultVendor: "newvendor",
    Vendors: map[string]*VendorConfig{
        "newvendor": {
            APIKey: os.Getenv("NEWVENDOR_API_KEY"),
            BaseURL: os.Getenv("NEWVENDOR_BASE_URL"),
            Timeout: 30 * time.Second,
        },
    },
}
```

## Error Handling

### Vendor-Specific Errors
```go
type NewVendorError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *NewVendorError) Error() string {
    return fmt.Sprintf("newvendor error %s: %s", e.Code, e.Message)
}
```

### Error Mapping
```go
func mapNewVendorError(err error) error {
    if strings.Contains(err.Error(), "rate limit") {
        return &RateLimitError{Vendor: "newvendor", Err: err}
    }
    if strings.Contains(err.Error(), "authentication") {
        return &AuthenticationError{Vendor: "newvendor", Err: err}
    }
    return err
}
```

## Performance Optimization

### Connection Pooling
```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

### Caching
```go
type VendorCache struct {
    capabilities *Capabilities
    lastUpdated  time.Time
    mu           sync.RWMutex
}

func (c *VendorCache) GetCapabilities() *Capabilities {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.capabilities
}
```

## Debugging

### Logging
```go
import "log"

func (v *NewVendor) SendRequest(ctx context.Context, req *Request) (*Response, error) {
    log.Printf("NewVendor: Sending request to model %s", req.Model)
    // Implementation
    log.Printf("NewVendor: Request completed successfully")
    return response, nil
}
```

### Metrics
```go
type Metrics struct {
    RequestCount    int64
    SuccessCount    int64
    ErrorCount      int64
    AverageLatency  time.Duration
}

func (v *NewVendor) recordMetrics(duration time.Duration, err error) {
    atomic.AddInt64(&v.metrics.RequestCount, 1)
    if err != nil {
        atomic.AddInt64(&v.metrics.ErrorCount, 1)
    } else {
        atomic.AddInt64(&v.metrics.SuccessCount, 1)
    }
}
```

## Release Process

### Version Management
1. Update version in `go.mod`
2. Update CHANGELOG.md
3. Create git tag
4. Build and test release

### Release Checklist
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Examples working
- [ ] Performance benchmarks acceptable
- [ ] Security review completed
- [ ] Release notes prepared

## Contributing Guidelines

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Update documentation
6. Run all tests and checks
7. Submit pull request

### Code Review Checklist
- [ ] Code follows Go conventions
- [ ] Tests are comprehensive
- [ ] Documentation is updated
- [ ] No breaking changes (unless major version)
- [ ] Performance impact considered
- [ ] Security implications reviewed 