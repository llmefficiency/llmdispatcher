# Contributing to LLM Dispatcher

Thank you for your interest in contributing to LLM Dispatcher! 🚀

## 🤝 How to Contribute

### 🐛 Reporting Bugs
- Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include detailed steps to reproduce
- Provide environment information
- Add screenshots if applicable

### 💡 Suggesting Features
- Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md)
- Describe the problem you're solving
- Include mockups or examples
- Assess impact and complexity

### 🔧 Code Contributions

#### Prerequisites
- Go 1.21+ installed
- Git configured
- Basic understanding of Go

#### Development Setup
1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/llmdispatcher.git
   cd llmdispatcher
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

#### Coding Standards

##### Code Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golint` for style checking
- Keep functions small and focused
- Add comprehensive comments

##### Testing
- Write tests for new functionality
- Maintain 90%+ test coverage
- Include both unit and integration tests
- Test error scenarios
- Add benchmarks for performance-critical code

##### Documentation
- Update README.md for user-facing changes
- Add API documentation for new functions
- Include usage examples
- Update CHANGELOG.md

#### Pull Request Process

1. **Ensure your code is ready**
   - All tests pass
   - Code is formatted
   - Documentation is updated
   - No linting errors

2. **Create a pull request**
   - Use the PR template
   - Describe your changes clearly
   - Link related issues
   - Add screenshots if applicable

3. **Review process**
   - Address review comments
   - Make requested changes
   - Ensure CI checks pass

## 🏗️ Project Structure

```
llmdispatcher/
├── cmd/example/          # Example application
├── internal/             # Private implementation
│   ├── dispatcher/       # Core dispatcher logic
│   ├── models/          # Data models
│   └── vendors/         # Vendor implementations
├── pkg/llmdispatcher/   # Public API
├── docs/                # Documentation
└── scripts/             # Build and test scripts
```

## 🧪 Testing Guidelines

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...

# Run specific test
go test ./internal/dispatcher/
```

### Test Coverage Requirements
- **Core logic**: 90%+ coverage
- **Vendor implementations**: 80%+ coverage
- **Public API**: 100% coverage
- **Error handling**: Must be tested

## 📚 Documentation Standards

### Code Documentation
- Document all exported functions
- Include usage examples
- Explain complex algorithms
- Add package-level documentation

### User Documentation
- Keep README.md up-to-date
- Add examples for new features
- Update API reference
- Include troubleshooting guides

## 🔒 Security Guidelines

### API Key Security
- Never commit API keys
- Use environment variables
- Add keys to .gitignore
- Use secure key management

### Code Security
- Validate all inputs
- Sanitize user data
- Use secure HTTP clients
- Implement proper error handling

## 🚀 Release Process

### Version Management
- Follow [Semantic Versioning](https://semver.org/)
- Update CHANGELOG.md
- Tag releases with git
- Create GitHub releases

### Pre-release Checklist
- [ ] All tests pass
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version is bumped
- [ ] Release notes are prepared

## 🎯 Contribution Areas

### High Priority
- **Bug fixes**: Critical issues affecting users
- **Security patches**: Vulnerabilities and security improvements
- **Documentation**: Improving clarity and completeness
- **Test coverage**: Adding missing tests

### Medium Priority
- **New vendors**: Adding support for additional LLM providers
- **Performance improvements**: Optimizing existing code
- **Feature enhancements**: Improving existing functionality
- **CI/CD improvements**: Better automation

### Low Priority
- **Code refactoring**: Improving code structure
- **Style improvements**: Better formatting and organization
- **Additional examples**: More usage examples
- **Tooling**: Development tool improvements

## 🤝 Community Guidelines

### Code of Conduct
- Be respectful and inclusive
- Help others learn and grow
- Provide constructive feedback
- Follow project conventions

### Communication
- Use clear, descriptive language
- Ask questions when unsure
- Provide context for suggestions
- Be patient with newcomers

## 📞 Getting Help

### Resources
- [Documentation](docs/INDEX.md)
- [API Reference](docs/API_REFERENCE.md)
- [Examples](docs/EXAMPLES.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)

### Support Channels
- GitHub Issues for bugs and features
- GitHub Discussions for questions
- Pull requests for code contributions

## 🏆 Recognition

### Contributors
- All contributors are listed in [CONTRIBUTORS.md](CONTRIBUTORS.md)
- Significant contributions are highlighted
- Contributors receive recognition in releases

### Hall of Fame
- Top contributors are featured
- Special recognition for major features
- Community awards for outstanding work

---

Thank you for contributing to LLM Dispatcher! Your contributions help make this project better for everyone. 🌟 