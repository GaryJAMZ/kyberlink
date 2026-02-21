# Contributing to KyberLink

Thank you for your interest in contributing to KyberLink! We welcome contributions from developers of all experience levels. This document provides guidelines for getting involved.

## Reporting Bugs

Found a bug? Please report it using [GitHub Issues](https://github.com/GaryJAMZ/kyberlink/issues).

When reporting a bug, please include:
- A clear, descriptive title
- A detailed description of the issue
- Steps to reproduce the problem
- Expected behavior vs. actual behavior
- Your environment (OS, Go version, Node.js version, etc.)
- Any relevant error messages or logs

## Reporting Security Vulnerabilities

**Do NOT open a public GitHub issue for security vulnerabilities.** This could expose the vulnerability before a fix is available.

Instead, please report security issues directly to **antoniogaribay6@gmail.com** with details about the vulnerability. Please see our [SECURITY.md](SECURITY.md) policy for more information.

## Submitting Pull Requests

We'd love to have your contributions! Here's how to get started:

1. **Fork the repository** on GitHub
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following the code style guidelines below
4. **Commit with clear messages** describing your changes
5. **Push to your fork** and **open a Pull Request** against the main repository
6. Respond to any review feedback

## Code Style Guidelines

### Go
- Format all Go code with `gofmt`
- Use standard Go conventions and idioms
- Write clear comments for exported functions and complex logic
- Aim for clear, maintainable code

### TypeScript
- Use strict mode (`"strict": true` in tsconfig.json)
- Follow standard TypeScript conventions
- Use meaningful variable and function names
- Write clear comments for complex logic

## Development Setup

Before you start, ensure you have the following prerequisites installed:

- **Go 1.23 or higher** - https://golang.org/dl/
- **Node.js 18 or higher** - https://nodejs.org/

### Getting Started

1. Clone your fork of the repository
2. Install dependencies:
   ```bash
   # For Go dependencies
   go mod download

   # For Node.js dependencies (if applicable)
   npm install
   ```
3. Build and test your changes
4. Ensure tests pass before submitting your PR

## License

By contributing to KyberLink, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE). All contributions must be compatible with this license.

---

Thank you for helping make KyberLink better!
