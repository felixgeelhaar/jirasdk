package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jira-connect"
	"github.com/felixgeelhaar/jira-connect/auth"
	"github.com/felixgeelhaar/jira-connect/resilience/fortify"
)

func main() {
	// Get credentials from environment
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")

	if email == "" || apiToken == "" || baseURL == "" {
		log.Fatal("JIRA_EMAIL, JIRA_API_TOKEN, and JIRA_BASE_URL must be set")
	}

	fmt.Println("=== Jira Connect - Resilience Patterns Example ===")
	fmt.Println()

	// Example 1: Default resilience configuration
	fmt.Println("=== Example 1: Default Resilience Configuration ===")
	fmt.Println()

	defaultConfig := jira.DefaultResilienceConfig()
	fmt.Printf("Default configuration:\n")
	fmt.Printf("  Circuit Breaker: %v (threshold: %d failures, interval: %v)\n",
		defaultConfig.CircuitBreakerEnabled, defaultConfig.CircuitBreakerThreshold, defaultConfig.CircuitBreakerInterval)
	fmt.Printf("  Retry: %v (max attempts: %d, initial delay: %v)\n",
		defaultConfig.RetryEnabled, defaultConfig.RetryMaxAttempts, defaultConfig.RetryInitialDelay)
	fmt.Printf("  Rate Limiting: %v (%d req/%v, burst: %d)\n",
		defaultConfig.RateLimitEnabled, defaultConfig.RateLimitRate, defaultConfig.RateLimitWindow, defaultConfig.RateLimitBurst)
	fmt.Printf("  Timeout: %v (%v)\n",
		defaultConfig.TimeoutEnabled, defaultConfig.TimeoutDuration)
	fmt.Printf("  Bulkhead: %v (max concurrent: %d, max queue: %d)\n",
		defaultConfig.BulkheadEnabled, defaultConfig.BulkheadMaxConcurrent, defaultConfig.BulkheadMaxQueue)
	fmt.Println()

	// Create client with default resilience
	resilience := fortify.NewAdapter(defaultConfig)
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAuth(auth.NewBasicAuth(email, apiToken)),
		jira.WithResilience(resilience),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("Making API request with full resilience protection...")
	_, err = client.Issue.Get(ctx, "PROJ-1")
	if err != nil {
		fmt.Printf("Request completed (error expected if issue doesn't exist): %v\n", err)
	} else {
		fmt.Println("Request successful!")
	}
	fmt.Println()

	// Example 2: Custom resilience configuration
	fmt.Println("=== Example 2: Custom Resilience Configuration ===")
	fmt.Println()

	customConfig := jira.ResilienceConfig{
		// Aggressive circuit breaker for critical operations
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 3,  // Open after 3 consecutive failures
		CircuitBreakerInterval:  30,  // seconds
		CircuitBreakerTimeout:   60,  // seconds

		// More retries with exponential backoff
		RetryEnabled:      true,
		RetryMaxAttempts:  5,
		RetryInitialDelay: 50,   // milliseconds
		RetryMaxDelay:     5000, // milliseconds
		RetryMultiplier:   2.0,
		RetryJitter:       true,

		// Conservative rate limiting
		RateLimitEnabled: true,
		RateLimitRate:    50,  // 50 req/min
		RateLimitBurst:   5,
		RateLimitWindow:  60, // seconds

		// Shorter timeout for critical paths
		TimeoutEnabled:  true,
		TimeoutDuration: 10, // seconds

		// Smaller bulkhead for resource protection
		BulkheadEnabled:      true,
		BulkheadMaxConcurrent: 5,
		BulkheadMaxQueue:      10,
		BulkheadQueueTimeout:  3, // seconds
	}

	customResilience := fortify.NewAdapter(customConfig)
	fmt.Println("Created client with custom aggressive resilience settings")
	fmt.Println()

	// Example 3: Resilience patterns explained
	fmt.Println("=== Example 3: Resilience Patterns Explained ===")
	fmt.Println()

	fmt.Println("1. Circuit Breaker")
	fmt.Println("   Prevents cascading failures by temporarily blocking requests")
	fmt.Println("   to failing services. States: Closed → Open → Half-Open → Closed")
	fmt.Println("   Current state:", customResilience.GetCircuitBreakerState())
	fmt.Println()

	fmt.Println("2. Retry with Exponential Backoff")
	fmt.Println("   Automatically retries failed operations with increasing delays")
	fmt.Println("   and jitter to prevent thundering herd problems")
	fmt.Println()

	fmt.Println("3. Rate Limiting (Token Bucket)")
	fmt.Println("   Controls request rate to comply with API limits and")
	fmt.Println("   prevent overwhelming the service")
	fmt.Println()

	fmt.Println("4. Timeout")
	fmt.Println("   Enforces time limits on operations to prevent resource leaks")
	fmt.Println("   and ensure SLA compliance")
	fmt.Println()

	fmt.Println("5. Bulkhead")
	fmt.Println("   Limits concurrent operations to prevent resource exhaustion")
	fmt.Println("   and isolate failures")
	fmt.Println()

	// Example 4: Pattern composition
	fmt.Println("=== Example 4: How Patterns Work Together ===")
	fmt.Println()
	fmt.Println("Request flow with all patterns enabled:")
	fmt.Println("  1. Bulkhead checks concurrency limit")
	fmt.Println("  2. Rate limiter checks quota (blocks if needed)")
	fmt.Println("  3. Timeout wraps the request")
	fmt.Println("  4. Circuit breaker checks service health")
	fmt.Println("  5. Retry handles transient failures")
	fmt.Println("  6. Finally, HTTP request is executed")
	fmt.Println()

	// Example 5: Performance impact
	fmt.Println("=== Example 5: Performance Characteristics ===")
	fmt.Println()
	fmt.Println("Fortify resilience patterns:")
	fmt.Println("  - Circuit Breaker: ~30ns overhead, 0 allocations")
	fmt.Println("  - Retry: ~25ns overhead, 0 allocations")
	fmt.Println("  - Rate Limiter: ~45ns overhead, 0 allocations")
	fmt.Println("  - Timeout: ~50ns overhead, 0 allocations")
	fmt.Println("  - Bulkhead: ~39ns overhead, 0 allocations")
	fmt.Println()
	fmt.Println("Total overhead: <200ns per request (<1µs)")
	fmt.Println("Impact: Negligible for most use cases")
	fmt.Println()

	// Example 6: Use cases
	fmt.Println("=== Example 6: When to Use Each Pattern ===")
	fmt.Println()
	fmt.Println("Circuit Breaker:")
	fmt.Println("  ✓ External dependencies that can fail")
	fmt.Println("  ✓ Preventing cascading failures")
	fmt.Println("  ✓ Fast failure for unhealthy services")
	fmt.Println()

	fmt.Println("Retry:")
	fmt.Println("  ✓ Transient network failures")
	fmt.Println("  ✓ Rate-limited APIs")
	fmt.Println("  ✓ Temporary service unavailability")
	fmt.Println()

	fmt.Println("Rate Limiting:")
	fmt.Println("  ✓ Complying with API quotas")
	fmt.Println("  ✓ Protecting your own services")
	fmt.Println("  ✓ Fair resource usage")
	fmt.Println()

	fmt.Println("Timeout:")
	fmt.Println("  ✓ Enforcing SLAs")
	fmt.Println("  ✓ Preventing resource leaks")
	fmt.Println("  ✓ Setting operation deadlines")
	fmt.Println()

	fmt.Println("Bulkhead:")
	fmt.Println("  ✓ Preventing resource exhaustion")
	fmt.Println("  ✓ Isolating critical operations")
	fmt.Println("  ✓ Managing concurrent access")
	fmt.Println()

	fmt.Println("=== Resilience Example Complete ===")
}
