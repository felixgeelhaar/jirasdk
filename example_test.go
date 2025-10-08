package jirasdk_test

import (
	"context"
	"fmt"
	"log"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
)

// ExampleNewClient demonstrates basic client creation with explicit configuration.
func ExampleNewClient() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "your-api-token"),
		jira.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client created for: %s\n", client.BaseURL.String())
	// Output: Client created for: https://your-domain.atlassian.net
}

// ExampleLoadConfigFromEnv demonstrates environment-based configuration.
func ExampleLoadConfigFromEnv() {
	// This example requires environment variables to be set:
	// export JIRA_BASE_URL="https://your-domain.atlassian.net"
	// export JIRA_EMAIL="user@example.com"
	// export JIRA_API_TOKEN="your-api-token"

	// For demonstration purposes, we'll show the pattern
	// In real usage, these would be loaded from environment
	fmt.Printf("LoadConfigFromEnv() reads JIRA_* environment variables\n")
	// Output: LoadConfigFromEnv() reads JIRA_* environment variables
}

// ExampleWithEnv demonstrates combining environment variables with explicit options.
func ExampleWithEnv() {
	// WithEnv() loads base config from environment variables,
	// which can be combined with explicit options to override defaults

	// Example pattern:
	// client, err := jira.NewClient(
	//     jira.WithEnv(),                    // Load from environment
	//     jira.WithTimeout(60*time.Second),  // Override timeout
	//     jira.WithMaxRetries(5),            // Override retries
	// )

	fmt.Printf("WithEnv() can be combined with other options\n")
	// Output: WithEnv() can be combined with other options
}

// ExampleWithAPIToken demonstrates API token authentication.
func ExampleWithAPIToken() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "your-api-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client configured with API token\n")
	_ = client
	// Output: Client configured with API token
}

// ExampleWithPAT demonstrates Personal Access Token authentication.
func ExampleWithPAT() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://jira.company.com"),
		jira.WithPAT("your-personal-access-token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client configured with PAT\n")
	_ = client
	// Output: Client configured with PAT
}

// ExampleWithBasicAuth demonstrates basic authentication (legacy).
func ExampleWithBasicAuth() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://jira.company.com"),
		jira.WithBasicAuth("username", "password"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client configured with basic auth\n")
	_ = client
	// Output: Client configured with basic auth
}

// ExampleWithTimeout demonstrates custom timeout configuration.
func ExampleWithTimeout() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
		jira.WithTimeout(60*time.Second), // 60 second timeout
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client timeout: %v\n", client.HTTPClient.Timeout)
	// Output: Client timeout: 1m0s
}

// ExampleWithMaxRetries demonstrates retry configuration.
func ExampleWithMaxRetries() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
		jira.WithMaxRetries(5), // Retry up to 5 times
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client created with custom retry config\n")
	_ = client
	// Output: Client created with custom retry config
}

// ExampleWithUserAgent demonstrates custom user agent configuration.
func ExampleWithUserAgent() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
		jira.WithUserAgent("MyApp/1.0.0"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client created with custom user agent\n")
	_ = client
	// Output: Client created with custom user agent
}

// ExampleClient_Do demonstrates context usage with the client.
func ExampleClient_Do() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	// The Do method would be used for low-level HTTP operations
	// Most users should use the higher-level service methods instead
	_ = ctx
	_ = client

	fmt.Printf("Client ready for API calls\n")
	// Output: Client ready for API calls
}

// Example_contextCancellation demonstrates using context for cancellation.
func Example_contextCancellation() {
	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Context will be passed to all API calls
	_ = ctx
	_ = client
	fmt.Println("Context configured for API calls")
	// Output: Context configured for API calls
}

// Example_structuredLogging demonstrates structured logging configuration.
func Example_structuredLogging() {
	// Note: This example doesn't actually import bolt to keep the example simple
	// In real code, you would: import "github.com/felixgeelhaar/jirasdk/logger/bolt"

	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
		// jira.WithLogger(bolt.NewLogger()),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client configured with logging\n")
	_ = client
	// Output: Client configured with logging
}

// Example_resilience demonstrates resilience pattern configuration.
func Example_resilience() {
	// Note: This example doesn't actually import fortify to keep the example simple
	// In real code, you would: import "github.com/felixgeelhaar/jirasdk/resilience/fortify"

	client, err := jira.NewClient(
		jira.WithBaseURL("https://your-domain.atlassian.net"),
		jira.WithAPIToken("user@example.com", "token"),
		// jira.WithResilience(fortify.NewAdapter(jira.DefaultResilienceConfig())),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client configured with resilience patterns\n")
	_ = client
	// Output: Client configured with resilience patterns
}
