## Contributing to Jira Connect

Thank you for considering contributing to the jira-connect library! This document outlines the development workflow and standards.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/felixgeelhaar/jira-connect.git
cd jira-connect

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover
```

## Code Standards

### Architecture Principles

This library follows **Hexagonal Architecture** (Ports and Adapters):

- **Core Domain** (`core/`): Business logic and domain models
- **Adapters** (`auth/`, `transport/`): External integrations
- **Infrastructure** (`internal/`): Shared utilities

### Go Best Practices

1. **Context-First APIs**: All operations accept `context.Context` as the first parameter
2. **Functional Options**: Use the functional options pattern for flexible configuration
3. **Error Handling**: Wrap errors with context using `fmt.Errorf` and `%w`
4. **Table-Driven Tests**: Use table-driven tests with testify for comprehensive coverage

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Maintain >80% test coverage for new code

## Testing

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -cover

# With race detector
go test ./... -race

# Specific package
go test ./core/issue/...

# Verbose output
go test ./... -v
```

### Writing Tests

Use table-driven tests:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)

            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

## Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write** tests for your changes
4. **Ensure** all tests pass
5. **Commit** your changes using conventional commits
6. **Push** to your fork
7. **Open** a Pull Request

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat(issue): add support for custom fields
fix(auth): handle token refresh errors
docs(readme): update installation instructions
test(transport): add retry logic tests
```

## Code Review

All submissions require review. We aim to review PRs within 48 hours.

Review criteria:
- ✅ Tests pass and coverage is maintained
- ✅ Code follows style guidelines
- ✅ Documentation is updated
- ✅ No breaking changes (or clearly marked)
- ✅ Commit messages follow conventions

## Continuous Integration

### GitHub Actions Workflow

All pull requests and commits to `main` or `develop` branches trigger automated CI/CD pipelines:

#### Test Job
- Runs on Ubuntu Latest
- Tests against Go versions 1.21, 1.22, and 1.23
- Executes full test suite with race detection
- Generates coverage reports
- Uploads coverage to Codecov (on Go 1.23)

#### Lint Job
- Runs `golangci-lint` with all linters enabled
- 5-minute timeout for comprehensive analysis
- Enforces code style and best practices

#### Security Job
- Runs Gosec security scanner
- Generates SARIF report for GitHub Security tab
- Scans for common security vulnerabilities

#### Coverage Job
- Verifies all packages maintain >80% test coverage
- Fails if coverage drops below threshold

#### Build Job
- Verifies all packages build successfully
- Builds all example programs

### Running CI Locally

Before pushing, run these checks locally:

```bash
# Run tests
go test -race ./...

# Run linter
golangci-lint run --timeout=5m

# Run security scanner
gosec ./...

# Check coverage
go test -cover ./...

# Build all
go build ./...
```

### CI/CD Best Practices

1. **All tests must pass** before merging
2. **Maintain >80% coverage** for all packages
3. **Fix linting issues** before committing
4. **Address security findings** promptly
5. **Ensure builds succeed** on all platforms

## Release Process

Releases follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. Update CHANGELOG.md with release notes
2. Update version in go.mod if needed
3. Create and push git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. GitHub Actions will create the release automatically

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues and PRs before creating new ones

Thank you for contributing! 🎉
