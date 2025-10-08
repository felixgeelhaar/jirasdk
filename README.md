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

### OAuth 2.0 Authentication

```go
// Create OAuth 2.0 authenticator
oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURL:  "http://localhost:8080/callback",
    Scopes:       []string{"read:jira-work", "write:jira-work"},
})

// Get authorization URL
authURL := oauth.GetAuthURL("state-string")
fmt.Println("Visit:", authURL)

// Exchange authorization code for token
token, err := oauth.Exchange(ctx, authorizationCode)

// Create client with OAuth 2.0
client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithOAuth2(oauth),
)

// Token is automatically refreshed when expired
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
// Get issue with specific fields
issue, err := client.Issue.Get(ctx, "PROJ-123", &issue.GetOptions{
    Fields: []string{"summary", "status", "assignee", "priority"},
})

// Create issue
input := &issue.CreateInput{
    Fields: &issue.IssueFields{
        Project:   &issue.Project{Key: "PROJ"},
        Summary:   "New issue",
        IssueType: &issue.IssueType{Name: "Task"},
        Priority:  &issue.Priority{Name: "High"},
        Labels:    []string{"bug", "urgent"},
    },
}
created, err := client.Issue.Create(ctx, input)

// Update issue
updateInput := &issue.UpdateInput{
    Fields: map[string]interface{}{
        "summary": "Updated summary",
        "priority": map[string]string{"name": "Medium"},
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

// Assign issue
err = client.Issue.Assign(ctx, "PROJ-123", "accountId")

// Comments
comment, err := client.Issue.AddComment(ctx, "PROJ-123", &issue.AddCommentInput{
    Body: "This is a comment",
})
comments, err := client.Issue.ListComments(ctx, "PROJ-123")

// Watchers and Voters
watchers, err := client.Issue.GetWatchers(ctx, "PROJ-123")
err = client.Issue.AddWatcher(ctx, "PROJ-123", "accountId")
votes, err := client.Issue.GetVotes(ctx, "PROJ-123")
err = client.Issue.AddVote(ctx, "PROJ-123")

// Attachments
file, _ := os.Open("document.pdf")
defer file.Close()
attachments, err := client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
    Filename: "document.pdf",
    Content:  file,
})

// Upload from string/bytes
report := strings.NewReader("Report content here")
attachments, err = client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
    Filename: "report.txt",
    Content:  report,
})

// Get attachment metadata
metadata, err := client.Issue.GetAttachment(ctx, "10000")
fmt.Printf("File: %s, Size: %d bytes\n", metadata.Filename, metadata.Size)

// Download attachment
content, err := client.Issue.DownloadAttachment(ctx, "10000")
defer content.Close()
data, _ := io.ReadAll(content)

// Delete attachment
err = client.Issue.DeleteAttachment(ctx, "10000")

// Custom Fields
customFields := issue.NewCustomFields().
    SetString("customfield_10001", "Sprint 23").
    SetNumber("customfield_10002", 8.5).
    SetDate("customfield_10003", time.Now()).
    SetSelect("customfield_10004", "High").
    SetMultiSelect("customfield_10005", []string{"Backend", "API"}).
    SetLabels("customfield_10006", []string{"feature", "urgent"}).
    SetUser("customfield_10007", "accountId123")

// Create issue with custom fields
created, err := client.Issue.Create(ctx, &issue.CreateInput{
    Fields: &issue.IssueFields{
        Project:   &issue.Project{Key: "PROJ"},
        Summary:   "New issue",
        IssueType: &issue.IssueType{Name: "Task"},
        Custom:    customFields,
    },
})

// Read custom fields from an issue
retrieved, err := client.Issue.Get(ctx, "PROJ-123", nil)
if sprint, ok := retrieved.Fields.Custom.GetString("customfield_10001"); ok {
    fmt.Printf("Sprint: %s\n", sprint)
}
if storyPoints, ok := retrieved.Fields.Custom.GetNumber("customfield_10002"); ok {
    fmt.Printf("Story Points: %.1f\n", storyPoints)
}

// Update custom fields
updates := issue.NewCustomFields().
    SetString("customfield_10001", "Sprint 24")
err = client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
    Fields: updates.ToMap(),
})
```

### Search

```go
// Simple JQL search
results, err := client.Search.Search(ctx, &search.SearchOptions{
    JQL:        "project = PROJ AND status = Open",
    MaxResults: 50,
})

// Query Builder
query := search.NewQueryBuilder().
    Project("PROJ").
    And().
    Status("In Progress").
    And().
    Assignee("currentUser()").
    OrderBy("created", "DESC")

results, err := client.Search.Search(ctx, &search.SearchOptions{
    JQL:        query.Build(),
    MaxResults: 100,
    Fields:     []string{"summary", "status", "priority"},
})

// Pagination
for i := 0; i < results.Total; i += 50 {
    page, err := client.Search.Search(ctx, &search.SearchOptions{
        JQL:        "project = PROJ",
        MaxResults: 50,
        StartAt:    i,
    })
}
```

### Projects

```go
// Get project
project, err := client.Project.Get(ctx, "PROJ", &project.GetOptions{
    Expand: []string{"lead", "description"},
})

// List projects
projects, err := client.Project.List(ctx, &project.ListOptions{
    Recent: 10,
})

// Get project components
components, err := client.Project.GetComponents(ctx, "PROJ")

// Get project versions
versions, err := client.Project.GetVersions(ctx, "PROJ", &project.VersionOptions{
    OrderBy: "releaseDate",
})

// Create version
version, err := client.Project.CreateVersion(ctx, &project.CreateVersionInput{
    Name:       "v1.0.0",
    ProjectID:  "10000",
    Released:   false,
})
```

### Users

```go
// Get current user
user, err := client.User.GetMyself(ctx)

// Get user by account ID
user, err := client.User.Get(ctx, "accountId", &user.GetOptions{
    Expand: []string{"groups", "applicationRoles"},
})

// Search users
users, err := client.User.Search(ctx, &user.SearchOptions{
    Query:      "john",
    MaxResults: 50,
})

// Find assignable users for project
users, err := client.User.FindAssignableUsers(ctx, &user.FindAssignableOptions{
    Project: "PROJ",
    Query:   "smith",
})

// Bulk get users
users, err := client.User.BulkGet(ctx, &user.BulkGetOptions{
    AccountIDs: []string{"id1", "id2", "id3"},
})
```

### Workflows

```go
// Get available transitions for an issue
transitions, err := client.Workflow.GetTransitions(ctx, "PROJ-123", &workflow.GetTransitionsOptions{
    Expand: []string{"transitions.fields"},
})

// List all workflows
workflows, err := client.Workflow.List(ctx, &workflow.ListOptions{
    WorkflowName: "Classic",
})

// Get workflow by ID
workflow, err := client.Workflow.Get(ctx, "classic-default-workflow")

// Get all statuses
statuses, err := client.Workflow.GetAllStatuses(ctx)

// Get specific status
status, err := client.Workflow.GetStatus(ctx, "10000")
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

## Examples

See the [examples](examples/) directory for complete, runnable examples:

- **[examples/basic](examples/basic/main.go)** - Basic usage patterns (get user, issues, search, projects)
- **[examples/advanced](examples/advanced/main.go)** - Custom middleware and advanced configuration
- **[examples/workflow](examples/workflow/main.go)** - Workflow operations, comments, watchers, voters

## Roadmap

### Phase 1: Foundation ✅ **Complete**
- [x] Core client architecture
- [x] Authentication (API Token, PAT, Basic Auth)
- [x] HTTP transport with middleware
- [x] Retry logic and rate limiting
- [x] Comprehensive testing (80%+ coverage)

### Phase 2: Core Resources ✅ **Complete**
- [x] Issue CRUD operations
- [x] Project management
- [x] User operations
- [x] Workflow transitions
- [x] JQL search with QueryBuilder
- [x] Comments and watchers/voters
- [x] Pagination support

### Phase 3: Advanced Features (In Progress)
- [ ] Custom fields support
- [ ] Attachments upload/download
- [ ] Advanced issue linking
- [ ] Bulk operations optimization

### Phase 4: Enterprise Features (Planned)
- [ ] OAuth 2.0 authentication
- [ ] Webhook support
- [ ] Observability (metrics, tracing)
- [ ] Connection pooling optimization

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
