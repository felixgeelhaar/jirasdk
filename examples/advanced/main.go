// Package main demonstrates advanced usage with custom middleware.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/transport"
)

// loggingMiddleware logs all requests and responses.
func loggingMiddleware(next transport.RoundTripFunc) transport.RoundTripFunc {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		start := time.Now()

		// Log request - sanitize URL path to prevent log injection
		sanitizedPath := sanitizeForLog(req.URL.Path)
		log.Printf("[Request] %s %s", req.Method, sanitizedPath)

		// Execute request
		resp, err := next(ctx, req)

		// Log response
		duration := time.Since(start)
		if err != nil {
			log.Printf("[Response] Error after %v: %v", duration, err)
		} else {
			log.Printf("[Response] %d in %v", resp.StatusCode, duration)
		}

		return resp, err
	}
}

// metricsMiddleware tracks request metrics.
func metricsMiddleware(next transport.RoundTripFunc) transport.RoundTripFunc {
	var totalRequests int64
	var failedRequests int64

	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		totalRequests++

		resp, err := next(ctx, req)

		if err != nil || (resp != nil && resp.StatusCode >= 400) {
			failedRequests++
		}

		// In a real application, you would export these to a metrics system
		if totalRequests%10 == 0 {
			log.Printf("[Metrics] Total: %d, Failed: %d, Success Rate: %.2f%%",
				totalRequests, failedRequests,
				float64(totalRequests-failedRequests)/float64(totalRequests)*100)
		}

		return resp, err
	}
}

func main() {
	// Get configuration from environment
	baseURL := os.Getenv("JIRA_BASE_URL")
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if baseURL == "" || email == "" || apiToken == "" {
		log.Fatal("Please set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables")
	}

	// Create client with custom middleware
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithTimeout(60*time.Second),
		jira.WithMaxRetries(5),
		jira.WithRateLimitBuffer(10*time.Second),
		jira.WithUserAgent("MyApp/1.0.0"),
		jira.WithMiddleware(loggingMiddleware),
		jira.WithMiddleware(metricsMiddleware),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example: Get multiple issues
	issueKeys := []string{"PROJ-123", "PROJ-124", "PROJ-125"}

	for _, key := range issueKeys {
		issue, err := client.Issue.Get(ctx, key, nil)
		if err != nil {
			log.Printf("Failed to get issue %s: %v", key, err)
			continue
		}

		fmt.Printf("\nIssue: %s\n", issue.Key)
		fmt.Printf("  Summary: %s\n", issue.Fields.Summary)
		fmt.Printf("  Status: %s\n", issue.Fields.Status.Name)
		if issue.Fields.Assignee != nil {
			fmt.Printf("  Assignee: %s\n", issue.Fields.Assignee.DisplayName)
		}
	}
}

// sanitizeForLog removes newline characters from strings to prevent log injection attacks.
func sanitizeForLog(s string) string {
	return strings.NewReplacer("\n", "", "\r", "").Replace(s)
}
