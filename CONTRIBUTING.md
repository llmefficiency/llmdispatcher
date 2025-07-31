# Contributing to LLM Dispatcher

Thank you for your interest in contributing! 🚀

## How to Contribute

### Reporting Bugs
- Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include steps to reproduce
- Provide environment information

### Suggesting Features
- Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md)
- Describe the problem you're solving
- Include examples if possible

### Code Contributions

#### Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/llmdispatcher.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Install dependencies: `go mod download`

#### Development
- Follow Go conventions
- Add tests for new functionality
- Update documentation
- Run tests: `go test ./...`

#### Pull Request Process
1. Ensure tests pass: `go test ./...`
2. Update documentation if needed
3. Create a pull request with a clear description
4. Address any review comments

## Project Structure

```
llmdispatcher/
├── cmd/example/          # Example application
├── internal/             # Private implementation
│   ├── dispatcher/       # Core dispatcher logic
│   ├── models/          # Data models
│   └── vendors/         # Vendor implementations
├── pkg/llmdispatcher/   # Public API
└── docs/                # Documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Getting Help

- [Documentation](docs/INDEX.md)
- [API Reference](docs/API_REFERENCE.md)
- [Examples](docs/EXAMPLES.md)
- GitHub Issues for bugs and features

---

Thank you for contributing! 🌟 