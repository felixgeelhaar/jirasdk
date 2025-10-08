// Package main demonstrates basic usage of the jira-connect library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/core/project"
	"github.com/felixgeelhaar/jirasdk/core/search"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Basic Jira Operations ===")
	fmt.Println()

	// 1. Get authenticated user
	fmt.Println("1. Getting current user...")
	user, err := client.User.GetMyself(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	fmt.Printf("   Logged in as: %s (%s)\n\n", user.DisplayName, user.EmailAddress)

	// 2. Get an issue
	issueKey := "PROJ-123" // Replace with your issue key
	fmt.Printf("2. Getting issue %s...\n", issueKey)
	iss, err := client.Issue.Get(ctx, issueKey, &issue.GetOptions{
		Fields: []string{"summary", "status", "assignee", "created"},
	})
	if err != nil {
		log.Printf("   Warning: Could not get issue %s: %v\n", issueKey, err)
	} else {
		fmt.Printf("   Issue: %s\n", iss.Key)
		fmt.Printf("   Summary: %s\n", iss.Fields.Summary)
		fmt.Printf("   Status: %s\n", iss.Fields.Status.Name)
		if iss.Fields.Assignee != nil {
			fmt.Printf("   Assignee: %s\n", iss.Fields.Assignee.DisplayName)
		}
		fmt.Printf("   Created: %s\n\n", iss.Fields.Created.Format(time.RFC3339))
	}

	// 3. Search for issues using JQL
	fmt.Println("3. Searching for open issues...")
	searchResult, err := client.Search.Search(ctx, &search.SearchOptions{
		JQL:        "status = Open ORDER BY created DESC",
		MaxResults: 5,
	})
	if err != nil {
		log.Printf("   Warning: Could not search issues: %v\n", err)
	} else {
		fmt.Printf("   Found %d issues:\n", len(searchResult.Issues))
		for i, issue := range searchResult.Issues {
			fmt.Printf("   %d. %s - %s\n", i+1, issue.Key, issue.Fields.Summary)
		}
		fmt.Println()
	}

	// 4. List projects
	fmt.Println("4. Listing projects...")
	projects, err := client.Project.List(ctx, &project.ListOptions{
		Recent: 5,
	})
	if err != nil {
		log.Printf("   Warning: Could not list projects: %v\n", err)
	} else {
		fmt.Printf("   Found %d projects:\n", len(projects))
		for i, proj := range projects {
			fmt.Printf("   %d. [%s] %s\n", i+1, proj.Key, proj.Name)
		}
		fmt.Println()
	}

	fmt.Println("=== Basic operations completed successfully! ===")
}
