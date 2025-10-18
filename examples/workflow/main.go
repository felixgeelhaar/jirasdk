// Package main demonstrates workflow and comment operations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/core/search"
	"github.com/felixgeelhaar/jirasdk/core/workflow"
)

func main() {
	// Get configuration from environment
	baseURL := os.Getenv("JIRA_BASE_URL")
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if baseURL == "" || email == "" || apiToken == "" {
		log.Fatal("Please set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables")
	}

	// Create client
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Jira Workflow and Comment Operations ===")
	fmt.Println()

	// 1. Complex JQL Search with QueryBuilder
	fmt.Println("1. Building and executing complex JQL query...")
	query := search.NewQueryBuilder().
		Project("PROJ").
		And().
		Status("In Progress").
		And().
		Assignee("currentUser()").
		OrderBy("created", "DESC")

	results, err := client.Search.Search(ctx, &search.SearchOptions{
		JQL:        query.Build(),
		MaxResults: 10,
		Fields:     []string{"summary", "status", "priority", "assignee", "created"},
	})
	if err != nil {
		log.Printf("   Warning: Search failed: %v\n", err)
	} else {
		fmt.Printf("   Found %d issues assigned to you\n", len(results.Issues))
		for i, iss := range results.Issues {
			// Safe: use GetSummary, GetStatusName, and GetPriorityName to avoid nil pointer panics
			priority := iss.GetPriorityName()
			if priority == "" {
				priority = "None"
			}
			fmt.Printf("   %d. [%s] %s - %s (Priority: %s)\n",
				i+1, iss.Key, iss.GetSummary(), iss.GetStatusName(), priority)
		}
		fmt.Println()
	}

	// 2. Workflow Transitions
	issueKey := "PROJ-123" // Replace with your issue key
	fmt.Printf("2. Managing Workflow Transitions for %s...\n", issueKey)

	// Get available transitions
	transitions, err := client.Workflow.GetTransitions(ctx, issueKey, &workflow.GetTransitionsOptions{
		Expand: []string{"transitions.fields"},
	})
	if err != nil {
		log.Printf("   Warning: Could not get transitions: %v\n", err)
	} else {
		fmt.Printf("   Available transitions:\n")
		for i, transition := range transitions {
			fmt.Printf("   %d. %s (ID: %s)\n", i+1, transition.Name, transition.ID)
			if transition.To != nil {
				fmt.Printf("      -> Transitions to: %s\n", transition.To.Name)
			}
		}
		fmt.Println()
	}

	// 3. Comment Operations
	fmt.Println("3. Working with Comments...")

	// Add a comment
	newComment := &issue.AddCommentInput{
		Body: "This is an automated comment from the jira-connect library.",
	}

	addedComment, err := client.Issue.AddComment(ctx, issueKey, newComment)
	if err != nil {
		log.Printf("   Warning: Could not add comment: %v\n", err)
	} else {
		fmt.Printf("   Added comment: %s\n", addedComment.ID)
	}

	// List all comments
	comments, err := client.Issue.ListComments(ctx, issueKey)
	if err != nil {
		log.Printf("   Warning: Could not list comments: %v\n", err)
	} else {
		fmt.Printf("   Total comments on issue: %d\n", len(comments))
		for i, c := range comments {
			// Safe: check for nil before accessing Author and Created
			authorName := "Unknown"
			if c.Author != nil {
				authorName = c.Author.DisplayName
			}
			createdTime := "Unknown"
			if c.Created != nil {
				createdTime = c.Created.Format(time.RFC3339)
			}
			fmt.Printf("   %d. By %s at %s\n", i+1, authorName, createdTime)
		}
		fmt.Println()
	}

	// 4. Issue Watchers and Voters
	fmt.Println("4. Managing Watchers and Voters...")

	// Get current watchers
	watchers, err := client.Issue.GetWatchers(ctx, issueKey)
	if err != nil {
		log.Printf("   Warning: Could not get watchers: %v\n", err)
	} else {
		fmt.Printf("   Current watchers: %d\n", watchers.WatchCount)
		fmt.Printf("   Is watching: %v\n", watchers.IsWatching)
	}

	// Get current voters
	votes, err := client.Issue.GetVotes(ctx, issueKey)
	if err != nil {
		log.Printf("   Warning: Could not get votes: %v\n", err)
	} else {
		fmt.Printf("   Current votes: %d\n", votes.Votes)
		fmt.Printf("   Has voted: %v\n", votes.HasVoted)
		fmt.Println()
	}

	// 5. Issue Creation
	fmt.Println("5. Creating Issue with Labels...")

	newIssue := &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project: &issue.Project{
				Key: "PROJ",
			},
			Summary: "Workflow example issue created via jira-connect",
			IssueType: &issue.IssueType{
				Name: "Task",
			},
			Priority: &issue.Priority{
				Name: "Medium",
			},
			Labels: []string{"automated", "jira-connect", "workflow-example"},
		},
	}

	createdIssue, err := client.Issue.Create(ctx, newIssue)
	if err != nil {
		log.Printf("   Warning: Could not create issue: %v\n", err)
	} else {
		fmt.Printf("   Created issue: %s\n", createdIssue.Key)
		fmt.Printf("   Issue URL: %s\n", createdIssue.Self)
		fmt.Println()
	}

	// 6. Working with Statuses
	fmt.Println("6. Listing All Workflow Statuses...")

	statuses, err := client.Workflow.GetAllStatuses(ctx)
	if err != nil {
		log.Printf("   Warning: Could not get statuses: %v\n", err)
	} else {
		fmt.Printf("   Available statuses: %d\n", len(statuses))

		// Group by category
		categories := make(map[string][]*workflow.Status)
		for _, status := range statuses {
			category := "Unknown"
			if status.StatusCategory != nil {
				category = status.StatusCategory.Name
			}
			categories[category] = append(categories[category], status)
		}

		for category, categoryStatuses := range categories {
			fmt.Printf("\n   %s:\n", category)
			for _, status := range categoryStatuses {
				fmt.Printf("     - %s\n", status.Name)
			}
		}
		fmt.Println()
	}

	fmt.Println("=== Workflow operations completed successfully! ===")
}
