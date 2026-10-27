# Contributing to LabKit

Thank you for your interest in contributing to LabKit! This document provides guidelines for contributing to the project.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [maintainers].

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- A clear description of the problem
- Steps to reproduce the issue
- Expected behavior vs. actual behavior
- Environment details (OS, Docker version, labctl version)
- Relevant logs or error messages

### Suggesting Enhancements

Enhancement suggestions are welcome! Please:

- Use a clear and descriptive title
- Provide a detailed description of the proposed enhancement
- Explain why the enhancement would be useful
- List any potential implementation considerations

### Contributing Labs

The most valuable contribution is a new lab template. Follow these steps:

1. **Use the scaffold**: Start from `templates/_scaffold/` to ensure your lab follows the template contract
2. **Validate locally**: Run `labctl validate ./templates/your-lab` to check compliance
3. **Test thoroughly**: Ensure smoke tests pass and the lab can be started, accessed, and cleaned up
4. **Document clearly**: Include learning objectives, concepts, and remediation guidance
5. **Submit a PR**: Open a pull request with your lab template

See [docs/template-authoring.md](docs/template-authoring.md) for detailed guidance.

### Contributing to the CLI

When contributing to the `labctl` CLI:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Ensure all tests pass: `go test ./...`
5. Run linters: `go fmt ./...` and `go vet ./...`
6. Commit with a clear message
7. Push and create a pull request

## Development Workflow

### Setting Up

```bash
# Clone the repository
git clone https://github.com/your-org/vulnerable-lab-platform.git
cd vulnerable-lab-platform

# Build the CLI
go build -o labctl ./cmd/labctl

# Add to PATH (optional)
export PATH=$PATH:$(pwd)
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/manifest
```

### Validating Templates

```bash
# Validate a specific template
./labctl validate ./templates/web/intro-lab

# Validate all templates
./scripts/validate-all.sh
```

## Pull Request Process

1. Update the CHANGELOG.md with a description of your changes
2. Ensure your code follows the project's style guidelines
3. Update documentation if needed
4. Ensure all tests pass
5. Your PR will be reviewed by maintainers

## Coding Standards

- Follow Go conventions for CLI code
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Write tests for new functionality

## Lab Template Standards

- All ports must bind to `127.0.0.1` by default
- Images should be pinned to specific versions or digests
- Include comprehensive documentation
- Provide smoke tests for validation
- Declare all endpoints in the manifest
- Avoid privileged containers and host mounts

## CI/CD

All pull requests must pass CI checks including:

- Manifest schema validation
- Compose configuration validation
- Policy compliance checks
- Smoke tests on disposable infrastructure
- License checks

## Questions?

Feel free to open an issue for questions or clarification.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
