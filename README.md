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

// Issue Links
// Create a "blocks" relationship
err = client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
    Type:         issue.BlocksLinkType(),
    InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
    OutwardIssue: &issue.IssueRef{Key: "PROJ-456"},
    Comment: &issue.LinkComment{
        Body: "These issues are related",
    },
})

// List available link types
linkTypes, err := client.Issue.ListIssueLinkTypes(ctx)

// Get links for an issue
links, err := client.Issue.GetIssueLinks(ctx, "PROJ-123")

// Delete issue link
err = client.Issue.DeleteIssueLink(ctx, "10000")

// Available helper functions:
// - issue.BlocksLinkType() - "blocks" / "is blocked by"
// - issue.DuplicatesLinkType() - "duplicates" / "is duplicated by"
// - issue.RelatesToLinkType() - "relates to" / "relates to"
// - issue.CausesLinkType() - "causes" / "is caused by"
// - issue.ClonesLinkType() - "clones" / "is cloned by"

// Time Tracking / Worklogs
now := time.Now()

// Add worklog with time string
worklog, err := client.Issue.AddWorklog(ctx, "PROJ-123", &issue.AddWorklogInput{
    TimeSpent: "3h 20m",
    Started:   &now,
    Comment:   "Implemented feature",
})

// Add worklog with seconds
worklog, err = client.Issue.AddWorklog(ctx, "PROJ-123", &issue.AddWorklogInput{
    TimeSpentSeconds: 7200, // 2 hours
    Started:          &now,
    Comment:          "Code review",
})

// List worklogs
worklogs, err := client.Issue.ListWorklogs(ctx, "PROJ-123", nil)

// List with date filters
yesterday := time.Now().AddDate(0, 0, -1)
worklogs, err = client.Issue.ListWorklogs(ctx, "PROJ-123", &issue.ListWorklogsOptions{
    StartedAfter: &yesterday,
    MaxResults:   10,
})

// Get specific worklog
worklog, err = client.Issue.GetWorklog(ctx, "PROJ-123", "10000")

// Update worklog
worklog, err = client.Issue.UpdateWorklog(ctx, "PROJ-123", "10000", &issue.UpdateWorklogInput{
    TimeSpent: "4h",
    Comment:   "Updated estimate",
})

// Delete worklog
err = client.Issue.DeleteWorklog(ctx, "PROJ-123", "10000")

// Format duration helper
formatted := issue.FormatDuration(12000) // Returns "3h 20m"

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

// Workflows
// List all workflows
workflows, err := client.Workflow.List(ctx, &workflow.ListOptions{
    MaxResults: 50,
})

// Get specific workflow
workflow, err := client.Workflow.Get(ctx, "classic-default-workflow")

// Get available transitions for an issue
transitions, err := client.Workflow.GetTransitions(ctx, "PROJ-123", &workflow.GetTransitionsOptions{
    Expand: []string{"transitions.fields"},
})

// Get all statuses
statuses, err := client.Workflow.GetAllStatuses(ctx)

// Get specific status
status, err := client.Workflow.GetStatus(ctx, "10000")

// Workflow Schemes
// List all workflow schemes
schemes, err := client.Workflow.ListWorkflowSchemes(ctx, nil)

// Get specific workflow scheme
scheme, err := client.Workflow.GetWorkflowScheme(ctx, 10000)

// Check required fields for transition
for _, transition := range transitions {
    for fieldKey, field := range transition.Fields {
        if field.Required {
            fmt.Printf("Field %s is required\n", field.Name)
        }
    }
}
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
proj, err := client.Project.Get(ctx, "PROJ")

// List projects
projects, err := client.Project.List(ctx, &project.ListOptions{
    Recent: 10,
})

// Create project
newProject, err := client.Project.Create(ctx, &project.CreateInput{
    Key:            "DEMO",
    Name:           "Demo Project",
    ProjectTypeKey: "software",
    LeadAccountID:  "accountId123",
})

// Update project
_, err = client.Project.Update(ctx, "PROJ", &project.UpdateInput{
    Name:        "Updated Name",
    Description: "Updated description",
})

// Archive and restore
err = client.Project.Archive(ctx, "PROJ")
err = client.Project.Restore(ctx, "PROJ")

// Delete project
err = client.Project.Delete(ctx, "PROJ")

// Component Management
// List components
components, err := client.Project.ListProjectComponents(ctx, "PROJ")

// Create component
component, err := client.Project.CreateComponent(ctx, &project.CreateComponentInput{
    Name:         "Backend Services",
    Description:  "All backend microservices",
    Project:      "PROJ",
    AssigneeType: "PROJECT_DEFAULT",
})

// Update component
component, err = client.Project.UpdateComponent(ctx, "10000", &project.UpdateComponentInput{
    Description: "Updated description",
})

// Get component
component, err = client.Project.GetComponent(ctx, "10000")

// Delete component
err = client.Project.DeleteComponent(ctx, "10000")

// Version Management
// List versions
versions, err := client.Project.ListProjectVersions(ctx, "PROJ")

// Create version
version, err := client.Project.CreateVersion(ctx, &project.CreateVersionInput{
    Name:        "v1.0.0",
    Description: "First release",
    Project:     "PROJ",
    StartDate:   "2024-01-01",
    ReleaseDate: "2024-06-30",
    Released:    false,
})

// Update version (mark as released)
released := true
version, err = client.Project.UpdateVersion(ctx, "10000", &project.UpdateVersionInput{
    Released: &released,
})

// Get version
version, err = client.Project.GetVersion(ctx, "10000")

// Delete version
err = client.Project.DeleteVersion(ctx, "10000")
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

### Agile/Scrum

```go
// Boards
// List all boards
boards, err := client.Agile.GetBoards(ctx, &agile.BoardsOptions{
    Type:       "scrum",  // or "kanban"
    MaxResults: 50,
})

// Get specific board
board, err := client.Agile.GetBoard(ctx, 123)

// Create board
newBoard, err := client.Agile.CreateBoard(ctx, &agile.CreateBoardInput{
    Name:     "Sprint Board",
    Type:     "scrum",
    FilterID: 10000,
})

// Delete board
err = client.Agile.DeleteBoard(ctx, 123)

// Sprints
// List board sprints
sprints, err := client.Agile.GetBoardSprints(ctx, 123, &agile.SprintsOptions{
    State:      "active,future",
    MaxResults: 50,
})

// Get specific sprint
sprint, err := client.Agile.GetSprint(ctx, 456)

// Create sprint
newSprint, err := client.Agile.CreateSprint(ctx, &agile.CreateSprintInput{
    Name:          "Sprint 25",
    OriginBoardID: 123,
    StartDate:     "2024-06-01T09:00:00.000Z",
    EndDate:       "2024-06-14T17:00:00.000Z",
    Goal:          "Complete user authentication",
})

// Update sprint
sprint, err = client.Agile.UpdateSprint(ctx, 456, &agile.UpdateSprintInput{
    State: "active",
    Goal:  "Updated goal",
})

// Delete sprint
err = client.Agile.DeleteSprint(ctx, 456)

// Move issues to sprint
err = client.Agile.MoveIssuesToSprint(ctx, 456, &agile.MoveIssuesToSprintInput{
    Issues: []string{"PROJ-123", "PROJ-124"},
})

// Epics
// List board epics
epics, err := client.Agile.GetBoardEpics(ctx, 123, &agile.EpicsOptions{
    MaxResults: 50,
})

// Get specific epic
epic, err := client.Agile.GetEpic(ctx, 789)

// Backlog
// Get backlog issues
backlog, err := client.Agile.GetBacklog(ctx, 123, &agile.BoardsOptions{
    MaxResults: 50,
})
```

### Permissions

```go
// Get all available permissions
allPermissions, err := client.Permission.GetAllPermissions(ctx)

// Check current user's permissions
myPerms, err := client.Permission.GetMyPermissions(ctx, nil)

// Check permissions for a specific project
projectPerms, err := client.Permission.GetMyPermissions(ctx, &permission.MyPermissionsOptions{
    ProjectKey:  "PROJ",
    Permissions: "BROWSE_PROJECTS,CREATE_ISSUES,EDIT_ISSUES",
})

// Permission Schemes
// List all permission schemes
schemes, err := client.Permission.ListPermissionSchemes(ctx, nil)

// Get detailed scheme with expanded information
scheme, err := client.Permission.GetPermissionScheme(ctx, 10000, &permission.GetPermissionSchemeOptions{
    Expand: []string{"permissions", "user", "group", "projectRole"},
})

// Create new permission scheme
newScheme, err := client.Permission.CreatePermissionScheme(ctx, &permission.CreatePermissionSchemeInput{
    Name:        "Custom Scheme",
    Description: "Custom permission scheme for development teams",
})

// Update permission scheme
updatedScheme, err := client.Permission.UpdatePermissionScheme(ctx, 10000, &permission.UpdatePermissionSchemeInput{
    Name:        "Updated Scheme Name",
    Description: "Updated description",
})

// Delete permission scheme
err = client.Permission.DeletePermissionScheme(ctx, 10000)

// Project Roles
// Get all roles for a project
roles, err := client.Permission.GetProjectRoles(ctx, "PROJ")

// Get specific role details
roleDetails, err := client.Permission.GetProjectRole(ctx, "PROJ", 10002)

// Add users to a project role
updatedRole, err := client.Permission.AddActorsToProjectRole(ctx, "PROJ", 10002, &permission.AddActorInput{
    User: []string{"accountId1", "accountId2"},
})

// Add groups to a project role
updatedRole, err = client.Permission.AddActorsToProjectRole(ctx, "PROJ", 10002, &permission.AddActorInput{
    Group: []string{"developers", "testers"},
})

// Remove actor from project role
err = client.Permission.RemoveActorFromProjectRole(ctx, "PROJ", 10002, "user", "accountId123")
err = client.Permission.RemoveActorFromProjectRole(ctx, "PROJ", 10002, "group", "developers")
```

### Bulk Operations

```go
// Bulk create issues (max 1000 per request)
result, err := client.Bulk.CreateIssues(ctx, &bulk.CreateIssuesInput{
    IssueUpdates: []bulk.IssueUpdate{
        {
            Fields: map[string]interface{}{
                "project":   map[string]string{"key": "PROJ"},
                "summary":   "Bulk created issue 1",
                "issuetype": map[string]string{"name": "Task"},
            },
        },
        {
            Fields: map[string]interface{}{
                "project":   map[string]string{"key": "PROJ"},
                "summary":   "Bulk created issue 2",
                "issuetype": map[string]string{"name": "Bug"},
                "labels":    []string{"bulk", "urgent"},
            },
        },
    },
})

// Check for errors
if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        fmt.Printf("Error on element %d\n", err.FailedElementNumber)
    }
}

// Bulk delete issues (max 1000 per request)
err = client.Bulk.DeleteIssues(ctx, &bulk.DeleteIssuesInput{
    IssueIDs: []string{"PROJ-123", "PROJ-124", "PROJ-125"},
})

// Track bulk operation progress
progress, err := client.Bulk.GetProgress(ctx, taskID)
fmt.Printf("Operation is %d%% complete\n", progress.ProgressPercent)

// Wait for bulk operation to complete (blocking)
progress, err := client.Bulk.WaitForCompletion(ctx, taskID, 5*time.Second)
if progress.Status == bulk.BulkOperationStatusComplete {
    fmt.Printf("Success: %d items processed\n", progress.Result.SuccessCount)
}
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
- **[examples/customfields](examples/customfields/main.go)** - Working with custom fields
- **[examples/attachments](examples/attachments/main.go)** - Upload, download, and manage attachments
- **[examples/oauth2](examples/oauth2/main.go)** - OAuth 2.0 authentication flow
- **[examples/issuelinks](examples/issuelinks/main.go)** - Create and manage issue relationships
- **[examples/worklogs](examples/worklogs/main.go)** - Time tracking and worklog management
- **[examples/workflows](examples/workflows/main.go)** - Workflow configuration, transitions, statuses, and schemes
- **[examples/projects](examples/projects/main.go)** - Project CRUD, component management, and version management
- **[examples/agile](examples/agile/main.go)** - Agile boards, sprints, epics, and backlog management
- **[examples/permissions](examples/permissions/main.go)** - Permission checking, schemes, and project role management
- **[examples/bulk](examples/bulk/main.go)** - Bulk operations for creating and deleting multiple issues efficiently

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

### Phase 3: Advanced Features ✅ **Complete**
- [x] Custom fields support with type-safe API
- [x] Attachments upload/download
- [x] Issue linking (blocks, duplicates, relates, causes, clones)
- [x] Time tracking and worklogs
- [x] OAuth 2.0 authentication

### Phase 4: Enterprise Features ✅ **Complete**
- [x] Enhanced workflow operations (transitions, statuses, schemes)
- [x] Enhanced project configuration (component and version management)
- [x] Agile/Scrum features (boards, sprints, epics, backlog)
- [x] Permissions API (schemes, project roles, permission checking)
- [x] Bulk operations (create, delete, progress tracking)
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
