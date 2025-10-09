# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.1] - 2025-01-09

### Security
- **Fixed log injection vulnerabilities (CWE-117)** in example code
  - Sanitized user input in OAuth2 callback handler (`examples/oauth2/main.go`)
  - Sanitized URL paths in logging middleware (`examples/advanced/main.go`)
  - Implemented `sanitizeForLog()` helper using `strings.NewReplacer` for improved readability

### Added - Phase 6: Extended API Coverage

#### New Services (7 services, 38 new methods)

- **Dashboard Service** (11 methods) - Complete dashboard management
  - `List()` - List all accessible dashboards
  - `Get()` - Get specific dashboard details
  - `Create()` - Create new dashboard with permissions
  - `Update()` - Update dashboard properties
  - `Delete()` - Delete dashboard
  - `Copy()` - Copy existing dashboard
  - `GetGadgets()` - List gadgets on dashboard
  - `AddGadget()` - Add gadget to dashboard
  - `UpdateGadget()` - Update gadget position/properties
  - `RemoveGadget()` - Remove gadget from dashboard
  - `SetItemProperty()` - Set dashboard item property

- **Group Service** (9 methods) - Comprehensive group administration
  - `Find()` - Search for groups
  - `Get()` - Get group details with expansion
  - `Create()` - Create new group
  - `Delete()` - Delete group
  - `GetMembers()` - List group members with pagination
  - `AddUser()` - Add user to group
  - `RemoveUser()` - Remove user from group
  - `BulkGet()` - Get multiple groups in bulk
  - `GetUsersFromGroup()` - Get users from specific group

- **Application Properties Service** (3 methods) - System-wide configuration
  - `GetAdvancedSettings()` - Get all advanced settings
  - `GetApplicationProperty()` - Get specific property
  - `SetApplicationProperty()` - Set property value

- **Server Info Service** (2 methods) - Instance metadata and health
  - `Get()` - Get server information (version, build, deployment type)
  - `GetConfiguration()` - Get Jira configuration (voting, time tracking, etc.)

- **Myself Service** (6 methods) - Current user preferences
  - `Get()` - Get current user details
  - `GetPreferences()` - Get all user preferences
  - `SetPreferences()` - Set multiple preferences
  - `GetPreference()` - Get specific preference
  - `SetPreference()` - Set individual preference
  - `DeletePreference()` - Delete preference

- **Jira Expressions Service** (2 methods) - Dynamic expression evaluation
  - `Evaluate()` - Evaluate Jira expression with context
  - `Analyze()` - Analyze expressions for syntax and complexity

- **Issue Link Type Service** (5 methods) - Custom relationship management
  - `List()` - List all issue link types
  - `Get()` - Get specific link type
  - `Create()` - Create custom link type
  - `Update()` - Update link type properties
  - `Delete()` - Delete link type

#### Enhanced Existing Services (18 new methods)

- **User Service Extensions** (9 new methods)
  - `SetDefaultColumns()` - Set default columns for user
  - `ResetDefaultColumns()` - Reset columns to system defaults
  - `GetUserProperty()` - Get user property value
  - `SetUserProperty()` - Set user property
  - `DeleteUserProperty()` - Delete user property
  - `GetUserGroups()` - Get groups user belongs to
  - `GetUserPermissions()` - Get user's permission details
  - `FindUsersWithAllPermissions()` - Find users with all specified permissions
  - `FindUsersWithBrowsePermission()` - Find users with browse permission

- **Workflow Service Extensions** (9 new methods)
  - `CreateWorkflowScheme()` - Create new workflow scheme
  - `UpdateWorkflowScheme()` - Update workflow scheme
  - `DeleteWorkflowScheme()` - Delete workflow scheme
  - `GetStatusCategories()` - Get all status categories
  - `GetStatusCategory()` - Get specific status category
  - `DoTransition()` - Execute issue transition
  - `GetTransitionProperties()` - Get transition properties
  - `SetWorkflowSchemeIssueType()` - Set issue type mapping
  - `DeleteWorkflowSchemeIssueType()` - Delete issue type mapping

### Summary

**Total API Coverage:**
- 27 services (7 new + 20 existing)
- 250+ methods (56 added in Phase 6)
- Comprehensive Jira REST API v3 coverage

**New Capabilities:**
- Dashboard visualization and gadget management
- Group administration and membership control
- System configuration and advanced settings
- Server health monitoring and metadata
- User preference customization
- Dynamic expression evaluation for automation
- Custom issue relationship type management
- Enhanced user and workflow operations

**Documentation:**
- Updated README.md with all new service examples
- Added Phase 6 completion to roadmap
- Comprehensive usage examples for each service
- Updated architecture documentation

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
