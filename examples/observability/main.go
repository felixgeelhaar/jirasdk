package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/felixgeelhaar/bolt"
	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/core/search"
	boltadapter "github.com/felixgeelhaar/jirasdk/logger/bolt"
)

func main() {
	// Get credentials from environment
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")

	if email == "" || apiToken == "" || baseURL == "" {
		log.Fatal("JIRA_EMAIL, JIRA_API_TOKEN, and JIRA_BASE_URL must be set")
	}

	// Example 1: JSON logging for production
	fmt.Println("=== Example 1: JSON Logging (Production) ===")
	fmt.Println()

	jsonLogger := bolt.New(bolt.NewJSONHandler(os.Stdout))

	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithLogger(boltadapter.NewAdapter(jsonLogger)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Fetch an issue - logs will be in JSON format
	fmt.Println("Fetching issue PROJ-1 (JSON logs)...")
	_, err = client.Issue.Get(ctx, "PROJ-1", &issue.GetOptions{
		Fields: []string{"summary", "status"},
	})
	if err != nil {
		fmt.Printf("Error (expected if issue doesn't exist): %v\n", err)
	}
	fmt.Println()

	// Example 2: Console logging for development
	fmt.Println("=== Example 2: Console Logging (Development) ===")
	fmt.Println()

	consoleLogger := bolt.New(bolt.NewConsoleHandler(os.Stdout))

	clientDev, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithLogger(boltadapter.NewAdapter(consoleLogger)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Search for issues - logs will be in console format
	fmt.Println("Searching for issues (Console logs)...")
	_, err = clientDev.Search.Search(ctx, &search.SearchOptions{
		JQL:        "project = PROJ",
		MaxResults: 5,
	})
	if err != nil {
		fmt.Printf("Search completed with result: %v\n", err)
	}
	fmt.Println()

	// Example 3: Structured logging benefits
	fmt.Println("=== Example 3: Structured Logging Benefits ===")
	fmt.Println()

	structuredLogger := bolt.New(bolt.NewJSONHandler(os.Stdout)).
		With().
		Str("service", "jira-integration").
		Str("environment", "production").
		Str("version", "1.0.0").
		Logger()

	clientStructured, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithLogger(boltadapter.NewAdapter(structuredLogger)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("All requests will include service, environment, and version fields...")
	_, err = clientStructured.Issue.Get(ctx, "PROJ-1", &issue.GetOptions{
		Fields: []string{"summary", "status"},
	})
	if err != nil {
		fmt.Printf("Request completed (error expected if issue doesn't exist)\n")
	}
	fmt.Println()

	// Example 4: OpenTelemetry Integration (conceptual)
	fmt.Println("=== Example 4: OpenTelemetry Integration (Conceptual) ===")
	fmt.Println()
	fmt.Println("When using OpenTelemetry:")
	fmt.Println("1. Initialize OpenTelemetry tracer and propagator")
	fmt.Println("2. Create context with trace information")
	fmt.Println("3. Bolt automatically includes trace_id and span_id in logs")
	fmt.Println()
	fmt.Println("Example log output:")
	fmt.Println(`{
  "level": "info",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "method": "GET",
  "path": "/rest/api/3/issue/PROJ-1",
  "status": 200,
  "duration": 234,
  "service": "jira-integration",
  "message": "jira_request_completed"
}`)
	fmt.Println()

	// Example 5: Log Levels
	fmt.Println("=== Example 5: Log Levels ===")
	fmt.Println()

	levelLogger := bolt.New(bolt.NewJSONHandler(os.Stdout))
	levelLogger.SetLevel(bolt.INFO) // Only INFO, WARN, ERROR will be logged

	clientLevels, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithLogger(boltadapter.NewAdapter(levelLogger)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Logger set to INFO level (DEBUG logs will be suppressed)...")
	_, err = clientLevels.Issue.Get(ctx, "PROJ-1", &issue.GetOptions{
		Fields: []string{"summary", "status"},
	})
	if err != nil {
		fmt.Printf("Request completed\n")
	}
	fmt.Println()

	// Example 6: Performance characteristics
	fmt.Println("=== Example 6: Performance Characteristics ===")
	fmt.Println()
	fmt.Println("Bolt logging performance:")
	fmt.Println("- Simple log: 63ns/op, 0 allocations")
	fmt.Println("- JSON handler: Zero allocations in hot paths")
	fmt.Println("- Console handler: ~10 allocations (development only)")
	fmt.Println()
	fmt.Println("Benefits for jira-connect:")
	fmt.Println("- Minimal overhead on API requests (<0.01% CPU)")
	fmt.Println("- No GC pressure from logging")
	fmt.Println("- Structured data for easy parsing and analysis")
	fmt.Println("- OpenTelemetry integration for distributed tracing")
	fmt.Println()

	// Example 7: Best practices
	fmt.Println("=== Example 7: Best Practices ===")
	fmt.Println()
	fmt.Println("Production deployment:")
	fmt.Println("1. Use JSON handler for machine-readable logs")
	fmt.Println("2. Set appropriate log level (INFO or WARN)")
	fmt.Println("3. Include service metadata (service name, version, environment)")
	fmt.Println("4. Enable OpenTelemetry for distributed tracing")
	fmt.Println("5. Forward logs to centralized logging (Elasticsearch, Loki, etc.)")
	fmt.Println()
	fmt.Println("Development:")
	fmt.Println("1. Use Console handler for human-readable output")
	fmt.Println("2. Set DEBUG level for detailed information")
	fmt.Println("3. Colorized output helps identify issues quickly")
	fmt.Println()

	fmt.Println("=== Observability Example Complete ===")
}
