# Contributing to g0s

Thank you for contributing to g0s! We're building a fast, intuitive terminal-based server management tool and we'd love to have your help. Here's some guidance to get you started:

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Setup](#development-setup)
3. [Issues](#issues)
4. [Pull Requests](#pull-requests)
5. [Code Style](#code-style)

## Getting Started

To ensure a positive and inclusive environment, please read our code of conduct before contributing. For detailed setup instructions, please follow our [DEVELOPERS.md](./DEVELOPERS.md) guide.

## Development Setup

1. Fork and clone the repository
2. Install the required dependencies (Go 1.24+, Git, Make)
3. Set up your development environment following [DEVELOPERS.md](./DEVELOPERS.md)

## Issues

Before creating a new issue:

- Search existing issues to avoid duplicates
- Use the issue templates when available
- Include clear steps to reproduce the problem
- Provide:
  - Your OS and Go version
  - Relevant logs and error messages
  - Screenshots if applicable
  - Steps to reproduce the issue

## Pull Requests

We welcome your pull requests! To ensure smooth collaboration:

1. **Link Issues**: Always link your PR to related issues
2. **Branch Naming**: Use descriptive branch names (e.g., `feature/new-metric`, `fix/agent-crash`)
3. **Small Changes**: Keep PRs focused and reasonably sized
4. **Documentation**: Update relevant documentation
5. **Tests**: Add tests for new features

### Components

The g0s project consists of three main components:

1. **Agent**: Runs on managed servers to collect metrics and execute commands
2. **Server**: Handles central coordination and API endpoints
3. **TUI**: Terminal User Interface for user interaction

Please indicate which component(s) your changes affect in your PR description.

## Code Style

Before submitting your PR:

```sh
# Format your code
go fmt ./...

# Run tests
make test
# or
go test ./...

# Build the project
make build-agent  # For agent
go build ./cmd/... # For all components
```

### Code Guidelines

1. Follow Go's official style guide
2. Write clear commit messages
3. Add comments for complex logic
4. Use meaningful variable and function names
5. Keep functions focused and reasonably sized

Need help? Feel free to ask questions in issues or pull requests!