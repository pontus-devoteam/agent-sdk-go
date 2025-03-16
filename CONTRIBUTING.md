# Contributing to Agent SDK Go

Thank you for your interest in contributing to Agent SDK Go! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

- Use the bug report template when creating an issue
- Provide detailed steps to reproduce the bug
- Include information about your environment (OS, Go version, etc.)
- If possible, provide a minimal code sample that demonstrates the issue

### Suggesting Features

- Use the feature request template when creating an issue
- Clearly describe the problem the feature would solve
- Suggest an implementation approach if you have one in mind

### Pull Requests

1. Fork the repository
2. Create a new branch for your changes
3. Implement your changes
4. Add or update tests as needed
5. Make sure all tests pass
6. Submit a pull request using the provided template

## Development Setup

1. Clone the repository
2. Install dependencies with `go mod download`
3. Build the project with `go build ./...`
4. Run tests with `go test ./...`

## Code Style

- Follow standard Go code style and conventions
- Use `gofmt` to format your code
- Write descriptive comments for public functions, types, and packages
- Add unit tests for new functionality

## Versioning

This project follows [Semantic Versioning](https://semver.org/). When proposing changes, consider whether they are:

- MAJOR version change: incompatible API changes
- MINOR version change: backwards-compatible functionality additions
- PATCH version change: backwards-compatible bug fixes

## Release Process

Releases are automated through GitHub Actions:

1. A new tag is created in the format `vX.Y.Z`
2. The release workflow is triggered
3. GoReleaser creates the release artifacts
4. Release notes are automatically generated

## License

By contributing to Agent SDK Go, you agree that your contributions will be licensed under the project's [MIT License](LICENSE). 