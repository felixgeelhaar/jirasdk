// Package main demonstrates basic usage of the jira-connect library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jira-connect"
)

func main() {
	// Get configuration from environment
	baseURL := os.Getenv("JIRA_BASE_URL")
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if baseURL == "" || email == "" || apiToken == "" {
		log.Fatal("Please set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables")
	}

	// Create client with API token authentication
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithTimeout(30*time.Second),
		jira.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Example: Get an issue
	issueKey := "PROJ-123" // Replace with your issue key
	issue, err := client.Issue.Get(ctx, issueKey, nil)
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	fmt.Printf("Issue: %s\n", issue.Key)
	fmt.Printf("Summary: %s\n", issue.Fields.Summary)
	fmt.Printf("Status: %s\n", issue.Fields.Status.Name)
	fmt.Printf("Created: %s\n", issue.Fields.Created)
}
