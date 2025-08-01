# LLM Dispatcher v0.3.0 Release Notes

## 🎉 What's New in v0.3.0

This release brings significant improvements to the LLM Dispatcher, including enhanced local vendor support, improved CLI interface, and better streaming functionality across all vendors.

### ✨ Key Features

#### 🆕 Enhanced Local Vendor Support
- **HTTP Mode**: Connect to local LLM services via HTTP endpoints
- **Executable Mode**: Run local LLM executables directly
- **Flexible Configuration**: Support for various local LLM setups (Ollama, custom APIs, etc.)

#### 🆕 Improved CLI Interface
- **Better User Experience**: Cleaner command-line interface with intuitive flags
- **Vendor Override**: Force testing with specific vendors using `--vendor-override`
- **Local Mode**: Dedicated local mode for testing with local LLM services
- **Enhanced Help**: Comprehensive help documentation and usage examples

#### 🆕 Web Service Implementation
- **REST API**: Full REST API implementation for programmatic access
- **Streaming Support**: Real-time streaming responses via HTTP
- **Vendor Management**: Dynamic vendor selection and configuration
- **Static UI**: Built-in web interface for easy testing

#### 🔧 Technical Improvements
- **Standardized Streaming**: Consistent streaming implementation across all vendors
- **Better Error Handling**: Improved error messages and validation
- **Code Refactoring**: Cleaner, more maintainable codebase
- **Enhanced Testing**: Comprehensive test coverage with updated vendor models

### 🐛 Bug Fixes
- Fixed vendor model test expectations to match actual capabilities
- Resolved streaming issues across all vendor implementations
- Improved error handling for malformed responses
- Fixed golangci-lint issues and ensured consistent code quality

### 📚 Documentation
- Comprehensive API reference documentation
- Architecture documentation with detailed component descriptions
- Development guide with setup instructions
- Troubleshooting guide for common issues
- Enhanced README with usage examples

## 🚀 Getting Started

### Quick Start
```bash
# Clone the repository
git clone https://github.com/your-org/llmdispatcher.git
cd llmdispatcher

# Build the application
make build

# Run with local mode (Ollama)
./bin/llmdispatcher --local

# Run with vendor mode
./bin/llmdispatcher --vendor --vendor-override anthropic
```

### Web Service
```bash
# Start the web service
make webservice

# Access the API at http://localhost:8080
# Access the UI at http://localhost:8080/static/
```

## 📦 Downloads

### Binary Downloads
- **Linux AMD64**: `llmdispatcher-linux-amd64`
- **Linux ARM64**: `llmdispatcher-linux-arm64`
- **macOS AMD64**: `llmdispatcher-darwin-amd64`
- **macOS ARM64**: `llmdispatcher-darwin-arm64`
- **Windows AMD64**: `llmdispatcher-windows-amd64.exe`
- **Windows ARM64**: `llmdispatcher-windows-arm64.exe`

### Installation
1. Download the appropriate binary for your platform
2. Make it executable: `chmod +x llmdispatcher-*`
3. Move to your PATH: `sudo mv llmdispatcher-* /usr/local/bin/llmdispatcher`

## 🔄 Migration from v0.2.0

This release is backward compatible with v0.2.0. No breaking changes were introduced.

### Notable Changes
- CLI interface has been improved but maintains the same core functionality
- New local mode provides better integration with local LLM services
- Web service is now available for programmatic access

## 🧪 Testing

All tests pass with comprehensive coverage:
```bash
make test
make lint
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Release Date**: July 31, 2025  
**Git Tag**: v0.3.0  
**Commit**: 684ac06 