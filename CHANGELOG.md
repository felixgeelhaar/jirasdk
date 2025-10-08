# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2025-01-08

### Added

#### Core Features
- **Enterprise-grade Go client** for Jira Cloud and Server/Data Center REST APIs
- **Environment variable configuration** following AWS SDK and Azure SDK patterns
  - Support for `JIRA_*` environment variables
  - `LoadConfigFromEnv()` convenience function
  - `WithEnv()` option for automatic credential loading
- **Multiple authentication methods**:
  - API Token authentication (Jira Cloud - recommended)
  - Personal Access Token (Jira Server/Data Center)
  - Basic authentication (legacy)
  - OAuth 2.0 support
- **Functional options pattern** for flexible, extensible configuration
- **Context propagation** for cancellation and timeout control

#### Domain Services
- **Issue Service**: Complete issue lifecycle management
  - Create, read, update, delete operations
  - Attachment management
  - Comment operations
  - Custom field support with type-safe API
  - Issue linking
  - Watchers management
  - Worklog tracking
- **Project Service**: Project management operations
  - Component management
  - Version management
  - Project listing and details
- **User Service**: User and account operations
- **Search Service**: JQL-based search with pagination
- **Agile Service**: Scrum/Kanban board operations
  - Board management
  - Sprint operations
  - Epic management
  - Backlog management
- **Workflow Service**: Workflow and transition management
- **Permission Service**: Permission and role management
- **Bulk Service**: Efficient batch operations

#### Resilience & Reliability
- **Production-grade resilience patterns** via fortify integration:
  - Circuit breaker pattern for fault tolerance
  - Automatic retry with exponential backoff and jitter
  - Rate limiting with configurable quotas
  - Request timeout management
  - Bulkhead pattern for resource isolation
- **Configurable retry behavior** with `WithMaxRetries()`
- **Rate limit handling** with automatic backoff

#### Observability
- **Zero-allocation structured logging** via bolt integration
- **Logger interface** for custom logging implementations
- **Request/response logging middleware**
- **Performance metrics support**

#### Developer Experience
- **Comprehensive documentation**:
  - 244 lines of package-level godoc
  - 13 testable examples
  - Complete API reference
  - Security best practices guide
  - Contributing guidelines
- **Type-safe domain models** with proper error handling
- **Middleware support** for extensible request/response pipeline
- **Thread-safe client** for concurrent use

#### Testing & Quality
- **Extensive test coverage** across all packages
- **Race condition testing** with `-race` flag
- **Example programs** demonstrating all major features
- **CI/CD pipeline** with GitHub Actions
- **Automated dependency updates** via Dependabot

#### Repository & Tooling
- **GitHub Actions workflows**:
  - Automated release workflow with multi-platform builds
  - Tag creation workflow with validation
  - Continuous integration on multiple Go versions (1.21, 1.22, 1.23)
  - Security scanning with Gosec
  - Code quality checks with golangci-lint
- **GitHub issue templates** for bugs and feature requests
- **Pull request template** with comprehensive checklist
- **Security policy** (SECURITY.md) with vulnerability reporting process
- **Release automation** with semantic versioning support

### Changed
- **Package renamed** from `jira-connect` to `jirasdk` for better Go idioms
- **Module path**: `github.com/felixgeelhaar/jirasdk`
- **Import alias**: `jira` for cleaner code

### Technical Details

#### Architecture
- **Hexagonal architecture** with clean separation of concerns
- **Transport layer** with middleware support
- **Authentication abstraction** for pluggable auth methods
- **Pagination support** for large result sets
- **Custom field handling** with type-safe API

#### Performance
- **Zero-allocation logging** with bolt
- **Connection pooling** via standard http.Client
- **Efficient JSON marshaling/unmarshaling**
- **Configurable timeouts** and retries

#### Security
- **HTTPS enforcement** for all API calls
- **Secure credential handling** via environment variables
- **No credentials in logs** or error messages
- **Security scanning** in CI/CD pipeline
- **Dependency vulnerability checks**

### Installation

```bash
go get github.com/felixgeelhaar/jirasdk@v1.0.0
```

### Quick Start

```go
import jira "github.com/felixgeelhaar/jirasdk"

client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithAPIToken("user@example.com", "your-api-token"),
)
```

Or use environment variables:

```bash
export JIRA_BASE_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="user@example.com"
export JIRA_API_TOKEN="your-api-token"
```

```go
client, err := jira.LoadConfigFromEnv()
```

### Documentation

- **pkg.go.dev**: https://pkg.go.dev/github.com/felixgeelhaar/jirasdk@v1.0.0
- **GitHub**: https://github.com/felixgeelhaar/jirasdk
- **Examples**: https://github.com/felixgeelhaar/jirasdk/tree/main/examples

### Breaking Changes

This is the initial v1.0.0 release. Future breaking changes will increment the major version.

### Upgrade Path

For users of the previous `jira-connect` package:

1. Update import path:
   ```diff
   -import jira "github.com/felixgeelhaar/jira-connect"
   +import jira "github.com/felixgeelhaar/jirasdk"
   ```

2. Update go.mod:
   ```bash
   go get github.com/felixgeelhaar/jirasdk@v1.0.0
   ```

### Contributors

- Felix Geelhaar (@felixgeelhaar)

### License

MIT License - see LICENSE file for details

---

[v1.0.0]: https://github.com/felixgeelhaar/jirasdk/releases/tag/v1.0.0
