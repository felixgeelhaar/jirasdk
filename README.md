# Jira Connect - Idiomatic Go Library

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A production-grade, idiomatic Go client library for Jira Cloud and Server/Data Center REST APIs.

## Features

- ✅ **Idiomatic Go** - Follows Go best practices and conventions
- ✅ **Context Support** - Full context propagation for cancellation and timeouts
- ✅ **Functional Options** - Flexible, extensible configuration pattern
- ✅ **Automatic Retries** - Exponential backoff with jitter
- ✅ **Rate Limiting** - Automatic handling of rate limits
- ✅ **Type Safe** - Strongly typed domain models
- ✅ **Middleware** - Extensible request/response pipeline
- ✅ **Multiple Auth** - OAuth 2.0, API Tokens, PAT, Basic Auth
- ✅ **Enterprise Ready** - Production-grade error handling and logging

## Installation

```bash
go get github.com/felixgeelhaar/jira-connect
```

## Quick Start

### Jira Cloud (API Token)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    jira "github.com/felixgeelhaar/jira-connect"
)

func main() {
    // Create client
    client, err := jira.NewClient(
        jira.WithBaseURL("https://your-domain.atlassian.net"),
        jira.WithAPIToken("your-email@example.com", "your-api-token"),
        jira.WithTimeout(30*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use client
    ctx := context.Background()
    issue, err := client.Issue.Get(ctx, "PROJ-123", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
}
```

### Jira Server/Data Center (PAT)

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://jira.your-company.com"),
    jira.WithPAT("your-personal-access-token"),
)
```

## Configuration Options

### Authentication

```go
// API Token (Jira Cloud - Recommended)
jira.WithAPIToken("email@example.com", "token")

// Personal Access Token (Server/Data Center - Recommended)
jira.WithPAT("token")

// Basic Auth (Legacy)
jira.WithBasicAuth("username", "password")
```

### HTTP Client Configuration

```go
// Timeout
jira.WithTimeout(60 * time.Second)

// Custom HTTP Client
jira.WithHTTPClient(&http.Client{
    Timeout: 60 * time.Second,
    Transport: customTransport,
})

// User Agent
jira.WithUserAgent("MyApp/1.0.0")
```

### Retry and Rate Limiting

```go
// Max retries
jira.WithMaxRetries(5)

// Rate limit buffer
jira.WithRateLimitBuffer(10 * time.Second)
```

### Custom Middleware

```go
// Add logging middleware
loggingMiddleware := func(next transport.RoundTripFunc) transport.RoundTripFunc {
    return func(ctx context.Context, req *http.Request) (*http.Response, error) {
        log.Printf("Request: %s %s", req.Method, req.URL)
        resp, err := next(ctx, req)
        if resp != nil {
            log.Printf("Response: %d", resp.StatusCode)
        }
        return resp, err
    }
}

client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithAPIToken("email", "token"),
    jira.WithMiddleware(loggingMiddleware),
)
```

## API Coverage

### Issues

```go
// Get issue
issue, err := client.Issue.Get(ctx, "PROJ-123", nil)

// Create issue
input := &issue.CreateInput{
    Fields: &issue.IssueFields{
        Project:   &issue.Project{Key: "PROJ"},
        Summary:   "New issue",
        IssueType: &issue.IssueType{Name: "Task"},
    },
}
created, err := client.Issue.Create(ctx, input)

// Update issue
updateInput := &issue.UpdateInput{
    Fields: map[string]interface{}{
        "summary": "Updated summary",
    },
}
err = client.Issue.Update(ctx, "PROJ-123", updateInput)

// Delete issue
err = client.Issue.Delete(ctx, "PROJ-123")

// Transition issue
transitionInput := &issue.TransitionInput{
    Transition: &issue.Transition{ID: "11"},
}
err = client.Issue.DoTransition(ctx, "PROJ-123", transitionInput)
```

### Projects (Planned)

```go
// Get project
project, err := client.Project.Get(ctx, "PROJ")

// List projects
projects, err := client.Project.List(ctx)
```

### Users (Planned)

```go
// Get user
user, err := client.User.Get(ctx, "account-id")
```

### Workflows (Planned)

```go
// Get available transitions
transitions, err := client.Workflow.GetTransitions(ctx, "PROJ-123")
```

## Architecture

This library follows **Hexagonal Architecture** (Ports and Adapters) principles:

```
jira-connect/
├── client.go              # Main client with functional options
├── auth/                  # Authentication adapters
│   ├── oauth2.go         # OAuth 2.0 (planned)
│   ├── apitoken.go       # API Token
│   └── pat.go            # Personal Access Token
├── core/                  # Business logic & domain models
│   ├── issue/            # Issue domain
│   ├── project/          # Project domain
│   ├── user/             # User domain
│   └── workflow/         # Workflow domain
├── transport/             # HTTP client abstraction
│   ├── middleware.go     # Middleware chain
│   └── backoff.go        # Retry logic
└── internal/             # Internal utilities
    └── pagination/       # Pagination helpers
```

## Design Principles

### 1. Context-First API

All operations accept `context.Context` as the first parameter for cancellation and timeout control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

issue, err := client.Issue.Get(ctx, "PROJ-123", nil)
```

### 2. Functional Options Pattern

Flexible configuration without breaking backward compatibility:

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://example.atlassian.net"),
    jira.WithAPIToken("email", "token"),
    jira.WithTimeout(30*time.Second),
    jira.WithMaxRetries(5),
)
```

### 3. Automatic Retry with Exponential Backoff

Retries are handled automatically for transient failures (5xx, 429):

- Exponential backoff: `min(100ms * 2^attempt, 30s)`
- Jitter: ±25% randomization to avoid thundering herd
- Context-aware: respects cancellation

### 4. Rate Limit Handling

Automatic detection and handling of rate limits:

- Respects `Retry-After` header
- Configurable buffer time
- Transparent retry after waiting

### 5. Middleware Pipeline

Extensible request/response processing:

```go
Request → Retry → RateLimit → UserAgent → Auth → HTTP
```

## Error Handling

All errors are wrapped with context for better debugging:

```go
issue, err := client.Issue.Get(ctx, "INVALID", nil)
if err != nil {
    // Error includes full context: authentication failed, HTTP 401, etc.
    log.Printf("Error: %v", err)
}
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...
```

## Roadmap

### Phase 1: Foundation (Current)
- [x] Core client architecture
- [x] Authentication (API Token, PAT, Basic Auth)
- [x] HTTP transport with middleware
- [x] Retry logic and rate limiting
- [ ] Comprehensive testing

### Phase 2: Core Resources
- [ ] Issue CRUD operations
- [ ] Project management
- [ ] User operations
- [ ] Workflow transitions

### Phase 3: Advanced Features
- [ ] JQL search
- [ ] Custom fields
- [ ] Attachments
- [ ] Comments and watchers

### Phase 4: Enterprise Features
- [ ] OAuth 2.0 authentication
- [ ] Webhook support
- [ ] Pagination helpers
- [ ] Observability (metrics, tracing)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with inspiration from:
- [andygrunwald/go-jira](https://github.com/andygrunwald/go-jira)
- [Official Jira REST API Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/)
- Go community best practices

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/felixgeelhaar/jira-connect)
- 🐛 [Issue Tracker](https://github.com/felixgeelhaar/jira-connect/issues)
- 💬 [Discussions](https://github.com/felixgeelhaar/jira-connect/discussions)
