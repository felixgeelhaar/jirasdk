// Package jirasdk provides an idiomatic Go client for Jira Cloud and Server/Data Center REST APIs.
//
// # Overview
//
// jirasdk is an enterprise-grade Go library for interacting with Atlassian Jira's REST APIs.
// It provides a clean, type-safe interface with support for modern Go patterns including
// context propagation, functional options, and comprehensive error handling.
//
// # Features
//
//   - Functional options pattern for flexible configuration
//   - Context propagation for cancellation and timeouts
//   - Automatic retry with exponential backoff
//   - Rate limiting and quota management
//   - Circuit breaker pattern for fault tolerance
//   - Structured logging with bolt integration
//   - Type-safe domain models
//   - Environment variable configuration
//   - Multiple authentication methods (API Token, PAT, Basic Auth, OAuth 2.0)
//
// # Installation
//
// Install using go get:
//
//	go get github.com/felixgeelhaar/jirasdk
//
// # Quick Start
//
// Create a client with explicit configuration:
//
//	import jira "github.com/felixgeelhaar/jirasdk"
//
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://your-domain.atlassian.net"),
//	    jira.WithAPIToken("your-email@example.com", "your-api-token"),
//	    jira.WithTimeout(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get an issue
//	issue, err := client.Issue.Get(ctx, "PROJ-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
//
// Or use environment variables for configuration:
//
//	export JIRA_BASE_URL="https://your-domain.atlassian.net"
//	export JIRA_EMAIL="your-email@example.com"
//	export JIRA_API_TOKEN="your-api-token"
//
//	client, err := jira.LoadConfigFromEnv()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Authentication
//
// The library supports multiple authentication methods:
//
// API Token (Recommended for Jira Cloud):
//
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://your-domain.atlassian.net"),
//	    jira.WithAPIToken("user@example.com", "your-api-token"),
//	)
//
// Personal Access Token (Jira Server/Data Center):
//
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://jira.company.com"),
//	    jira.WithPAT("your-personal-access-token"),
//	)
//
// Basic Authentication (Legacy):
//
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://jira.company.com"),
//	    jira.WithBasicAuth("username", "password"),
//	)
//
// OAuth 2.0:
//
//	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURL:  "http://localhost:8080/callback",
//	    Scopes:       []string{"read:jira-work", "write:jira-work"},
//	})
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://your-domain.atlassian.net"),
//	    jira.WithOAuth2(oauth),
//	)
//
// # Environment Variable Configuration
//
// The library follows AWS SDK and Azure SDK patterns for environment-based configuration:
//
// Required variables:
//   - JIRA_BASE_URL: Your Jira instance URL
//
// Authentication (choose one):
//   - JIRA_EMAIL + JIRA_API_TOKEN: API token authentication (Jira Cloud)
//   - JIRA_PAT: Personal Access Token (Jira Server/Data Center)
//   - JIRA_USERNAME + JIRA_PASSWORD: Basic authentication (legacy)
//   - JIRA_OAUTH_CLIENT_ID + JIRA_OAUTH_CLIENT_SECRET + JIRA_OAUTH_REDIRECT_URL: OAuth 2.0
//
// Optional configuration:
//   - JIRA_TIMEOUT: HTTP timeout in seconds (default: 30)
//   - JIRA_MAX_RETRIES: Maximum retry attempts (default: 3)
//   - JIRA_RATE_LIMIT_BUFFER: Rate limit buffer in seconds (default: 5)
//   - JIRA_USER_AGENT: Custom user agent string
//
// # Domain Services
//
// The client provides access to various Jira domains through service objects:
//
//	// Issue operations
//	issue, err := client.Issue.Get(ctx, "PROJ-123")
//	newIssue, err := client.Issue.Create(ctx, &CreateIssueRequest{...})
//	err = client.Issue.Update(ctx, "PROJ-123", &UpdateIssueRequest{...})
//
//	// Project operations
//	project, err := client.Project.Get(ctx, "PROJ")
//	projects, err := client.Project.List(ctx)
//
//	// Search operations
//	results, err := client.Search.Search(ctx, "project = PROJ AND status = Open", nil)
//
//	// User operations
//	user, err := client.User.GetMyself(ctx)
//
//	// Agile operations (Scrum/Kanban boards)
//	boards, err := client.Agile.GetBoards(ctx, nil)
//	sprints, err := client.Agile.GetBoardSprints(ctx, boardID, nil)
//
//	// Workflow operations
//	workflows, err := client.Workflow.List(ctx)
//	transitions, err := client.Workflow.GetTransitions(ctx, "PROJ-123")
//
// # Advanced Configuration
//
// Configure resilience patterns with fortify:
//
//	import "github.com/felixgeelhaar/jirasdk/resilience/fortify"
//
//	resilience := fortify.NewAdapter(jira.DefaultResilienceConfig())
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://your-domain.atlassian.net"),
//	    jira.WithAPIToken("email", "token"),
//	    jira.WithResilience(resilience),
//	)
//
// Configure structured logging with bolt:
//
//	import "github.com/felixgeelhaar/jirasdk/logger/bolt"
//
//	logger := bolt.NewLogger()
//	client, err := jira.NewClient(
//	    jira.WithBaseURL("https://your-domain.atlassian.net"),
//	    jira.WithAPIToken("email", "token"),
//	    jira.WithLogger(logger),
//	)
//
// # Error Handling
//
// The library provides structured error types for better error handling:
//
//	issue, err := client.Issue.Get(ctx, "INVALID-123")
//	if err != nil {
//	    // Handle different error types
//	    switch {
//	    case errors.Is(err, context.DeadlineExceeded):
//	        log.Println("Request timed out")
//	    case errors.Is(err, context.Canceled):
//	        log.Println("Request was canceled")
//	    default:
//	        log.Printf("API error: %v", err)
//	    }
//	    return err
//	}
//
// # Pagination
//
// The library handles pagination automatically for list operations:
//
//	// Search with pagination
//	opts := &SearchOptions{
//	    StartAt:    0,
//	    MaxResults: 50,
//	}
//	results, err := client.Search.Search(ctx, "project = PROJ", opts)
//
// # Context and Cancellation
//
// All API methods accept context.Context for cancellation and timeout control:
//
//	// With timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123")
//
//	// With cancellation
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    time.Sleep(5 * time.Second)
//	    cancel()
//	}()
//
//	results, err := client.Search.Search(ctx, "project = PROJ", nil)
//
// # Best Practices
//
//   - Always use context.Context for cancellation and timeout control
//   - Reuse Client instances across requests (they are safe for concurrent use)
//   - Use environment variables for production configuration
//   - Enable resilience patterns for production deployments
//   - Configure appropriate timeouts based on your use case
//   - Use structured logging to monitor API usage and errors
//   - Handle rate limiting gracefully (built-in by default)
//
// # Thread Safety
//
// The Client is safe for concurrent use by multiple goroutines. All service
// methods can be called concurrently without additional synchronization.
//
// # Examples
//
// See the examples directory for comprehensive usage examples:
//   - examples/basic: Basic client usage
//   - examples/environment: Environment variable configuration
//   - examples/advanced: Advanced features and pagination
//   - examples/resilience: Resilience patterns with fortify
//   - examples/observability: Structured logging with bolt
//   - examples/oauth2: OAuth 2.0 authentication flow
//
// For more information, visit: https://github.com/felixgeelhaar/jirasdk
package jirasdk
