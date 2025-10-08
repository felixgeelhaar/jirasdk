package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
)

func main() {
	// Get credentials from environment
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")

	if email == "" || apiToken == "" || baseURL == "" {
		log.Fatal("JIRA_EMAIL, JIRA_API_TOKEN, and JIRA_BASE_URL must be set")
	}

	// Create authenticated client
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	issueKey := "PROJ-123" // Replace with your issue key

	// Example 1: Add worklog with time spent string
	fmt.Println("=== Adding Worklog with Time String ===")
	now := time.Now()
	worklog1, err := client.Issue.AddWorklog(ctx, issueKey, &issue.AddWorklogInput{
		TimeSpent: "3h 20m",
		Started:   &now,
		Comment:   "Implemented new feature for user authentication",
	})
	if err != nil {
		log.Printf("Failed to add worklog: %v", err)
	} else {
		fmt.Printf("✓ Worklog added successfully (ID: %s)\n", worklog1.ID)
		fmt.Printf("  Time spent: %s (%d seconds)\n", worklog1.TimeSpent, worklog1.TimeSpentSeconds)
		fmt.Printf("  Comment: %s\n", worklog1.Comment)
	}

	// Example 2: Add worklog with seconds
	fmt.Println("\n=== Adding Worklog with Seconds ===")
	worklog2, err := client.Issue.AddWorklog(ctx, issueKey, &issue.AddWorklogInput{
		TimeSpentSeconds: 7200, // 2 hours
		Started:          &now,
		Comment:          "Code review and testing",
	})
	if err != nil {
		log.Printf("Failed to add worklog: %v", err)
	} else {
		fmt.Printf("✓ Worklog added successfully (ID: %s)\n", worklog2.ID)
		fmt.Printf("  Time spent: %s (%d seconds)\n", worklog2.TimeSpent, worklog2.TimeSpentSeconds)
	}

	// Example 3: Add worklog with visibility restrictions
	fmt.Println("\n=== Adding Worklog with Visibility Restrictions ===")
	worklog3, err := client.Issue.AddWorklog(ctx, issueKey, &issue.AddWorklogInput{
		TimeSpent: "1h 30m",
		Started:   &now,
		Comment:   "Internal team discussion - developers only",
		Visibility: &issue.WorklogVisibility{
			Type:  "role",
			Value: "Developers",
		},
	})
	if err != nil {
		log.Printf("Failed to add restricted worklog: %v", err)
	} else {
		fmt.Printf("✓ Restricted worklog added successfully (ID: %s)\n", worklog3.ID)
	}

	// Example 4: List all worklogs for an issue
	fmt.Println("\n=== Listing All Worklogs ===")
	worklogs, err := client.Issue.ListWorklogs(ctx, issueKey, nil)
	if err != nil {
		log.Printf("Failed to list worklogs: %v", err)
	} else {
		fmt.Printf("Found %d worklogs:\n", len(worklogs))
		for i, w := range worklogs {
			authorName := "Unknown"
			if w.Author != nil {
				authorName = w.Author.DisplayName
			}
			fmt.Printf("%d. ID: %s | Author: %s | Time: %s | Comment: %s\n",
				i+1, w.ID, authorName, w.TimeSpent, w.Comment)
		}
	}

	// Example 5: List worklogs with date filters
	fmt.Println("\n=== Listing Worklogs with Date Filter ===")
	yesterday := time.Now().AddDate(0, 0, -1)
	filteredWorklogs, err := client.Issue.ListWorklogs(ctx, issueKey, &issue.ListWorklogsOptions{
		StartedAfter: &yesterday,
		MaxResults:   10,
	})
	if err != nil {
		log.Printf("Failed to list filtered worklogs: %v", err)
	} else {
		fmt.Printf("Found %d worklogs from yesterday onwards\n", len(filteredWorklogs))
	}

	// Example 6: Get specific worklog
	if len(worklogs) > 0 {
		worklogID := worklogs[0].ID
		fmt.Printf("\n=== Getting Worklog (ID: %s) ===\n", worklogID)
		w, err := client.Issue.GetWorklog(ctx, issueKey, worklogID)
		if err != nil {
			log.Printf("Failed to get worklog: %v", err)
		} else {
			fmt.Printf("Worklog Details:\n")
			fmt.Printf("  ID: %s\n", w.ID)
			fmt.Printf("  Time Spent: %s (%d seconds)\n", w.TimeSpent, w.TimeSpentSeconds)
			fmt.Printf("  Comment: %s\n", w.Comment)
			if w.Author != nil {
				fmt.Printf("  Author: %s (%s)\n", w.Author.DisplayName, w.Author.EmailAddress)
			}
			if w.Started != nil {
				fmt.Printf("  Started: %s\n", w.Started.Format(time.RFC3339))
			}
			if w.Created != nil {
				fmt.Printf("  Created: %s\n", w.Created.Format(time.RFC3339))
			}
		}
	}

	// Example 7: Update worklog
	if len(worklogs) > 0 {
		worklogID := worklogs[0].ID
		fmt.Printf("\n=== Updating Worklog (ID: %s) ===\n", worklogID)
		updatedWorklog, err := client.Issue.UpdateWorklog(ctx, issueKey, worklogID, &issue.UpdateWorklogInput{
			TimeSpent: "4h",
			Comment:   "Updated time estimate after review",
		})
		if err != nil {
			log.Printf("Failed to update worklog: %v", err)
		} else {
			fmt.Printf("✓ Worklog updated successfully\n")
			fmt.Printf("  New time spent: %s\n", updatedWorklog.TimeSpent)
			fmt.Printf("  New comment: %s\n", updatedWorklog.Comment)
		}
	}

	// Example 8: Format duration examples
	fmt.Println("\n=== Duration Formatting Examples ===")
	durations := []int64{
		60,      // 1 minute
		3600,    // 1 hour
		7200,    // 2 hours
		12000,   // 3 hours 20 minutes
		86400,   // 1 day
		604800,  // 1 week
		694800,  // 1 week 1 day 1 hour
	}

	for _, seconds := range durations {
		formatted := issue.FormatDuration(seconds)
		fmt.Printf("%6d seconds = %s\n", seconds, formatted)
	}

	// Example 9: Add worklog with estimate adjustment
	fmt.Println("\n=== Adding Worklog with Estimate Adjustment ===")
	worklog4, err := client.Issue.AddWorklog(ctx, issueKey, &issue.AddWorklogInput{
		TimeSpent: "2h",
		Started:   &now,
		Comment:   "Bug fix implementation",
		AdjustEstimate: &issue.AdjustEstimate{
			Type:        "new",
			NewEstimate: "4h",
		},
	})
	if err != nil {
		log.Printf("Failed to add worklog with estimate: %v", err)
	} else {
		fmt.Printf("✓ Worklog added with estimate adjustment (ID: %s)\n", worklog4.ID)
	}

	// Example 10: Calculate total time logged
	fmt.Println("\n=== Calculating Total Time Logged ===")
	totalSeconds := int64(0)
	for _, w := range worklogs {
		totalSeconds += w.TimeSpentSeconds
	}
	fmt.Printf("Total time logged: %s (%d seconds)\n",
		issue.FormatDuration(totalSeconds),
		totalSeconds)

	// Example 11: Delete worklog (commented out for safety)
	fmt.Println("\n=== Deleting Worklog (Example) ===")
	// Uncomment to actually delete
	// if len(worklogs) > 0 {
	// 	worklogID := worklogs[len(worklogs)-1].ID
	// 	err := client.Issue.DeleteWorklog(ctx, issueKey, worklogID)
	// 	if err != nil {
	// 		log.Printf("Failed to delete worklog: %v", err)
	// 	} else {
	// 		fmt.Printf("✓ Worklog %s deleted successfully\n", worklogID)
	// 	}
	// }
	fmt.Println("(Delete example commented out for safety)")

	// Example 12: Group worklogs by author
	fmt.Println("\n=== Grouping Worklogs by Author ===")
	authorTotals := make(map[string]int64)
	for _, w := range worklogs {
		authorName := "Unknown"
		if w.Author != nil {
			authorName = w.Author.DisplayName
		}
		authorTotals[authorName] += w.TimeSpentSeconds
	}

	for author, seconds := range authorTotals {
		fmt.Printf("%s: %s\n", author, issue.FormatDuration(seconds))
	}

	fmt.Println("\n=== Worklogs Example Complete ===")
}
