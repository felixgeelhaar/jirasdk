// Package main demonstrates custom field usage with the jira-connect library.
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
	// Create client with PAT authentication
	client, err := jira.NewClient(
		jira.WithBaseURL(os.Getenv("JIRA_BASE_URL")),
		jira.WithPAT(os.Getenv("JIRA_PAT")),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Create an issue with custom fields
	fmt.Println("=== Creating Issue with Custom Fields ===")

	customFields := issue.NewCustomFields().
		SetString("customfield_10001", "Sprint 23").
		SetNumber("customfield_10002", 8.5).                             // Story points
		SetDate("customfield_10003", time.Now().AddDate(0, 0, 14)).      // Due date
		SetSelect("customfield_10004", "High").                          // Priority
		SetMultiSelect("customfield_10005", []string{"Backend", "API"}). // Components
		SetLabels("customfield_10006", []string{"feature", "customer-request"})

	created, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project: &issue.Project{
				Key: "PROJ",
			},
			Summary: "Implement user authentication",
			IssueType: &issue.IssueType{
				Name: "Story",
			},
			Description: issue.ADFFromText("Add OAuth 2.0 authentication to the application"),
			Custom:      customFields,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}

	fmt.Printf("Created issue: %s\n", created.Key)

	// Example 2: Retrieve an issue and read custom fields
	fmt.Println("\n=== Reading Issue with Custom Fields ===")

	retrieved, err := client.Issue.Get(ctx, created.Key, nil)
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	fmt.Printf("Issue: %s - %s\n", retrieved.Key, retrieved.Fields.Summary)

	// Read custom fields using type-safe methods
	if sprint, ok := retrieved.Fields.Custom.GetString("customfield_10001"); ok {
		fmt.Printf("Sprint: %s\n", sprint)
	}

	if storyPoints, ok := retrieved.Fields.Custom.GetNumber("customfield_10002"); ok {
		fmt.Printf("Story Points: %.1f\n", storyPoints)
	}

	if dueDate, ok := retrieved.Fields.Custom.GetDate("customfield_10003"); ok {
		fmt.Printf("Due Date: %s\n", dueDate.Format("2006-01-02"))
	}

	if priority, ok := retrieved.Fields.Custom.GetSelect("customfield_10004"); ok {
		fmt.Printf("Priority: %s\n", priority)
	}

	if components, ok := retrieved.Fields.Custom.GetMultiSelect("customfield_10005"); ok {
		fmt.Printf("Components: %v\n", components)
	}

	if labels, ok := retrieved.Fields.Custom.GetLabels("customfield_10006"); ok {
		fmt.Printf("Labels: %v\n", labels)
	}

	// Example 3: Update custom fields
	fmt.Println("\n=== Updating Custom Fields ===")

	updatedFields := issue.NewCustomFields().
		SetString("customfield_10001", "Sprint 24"). // Move to next sprint
		SetNumber("customfield_10002", 13)           // Update story points

	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: updatedFields.ToMap(),
	})
	if err != nil {
		log.Fatalf("Failed to update issue: %v", err)
	}

	fmt.Printf("Updated custom fields for issue: %s\n", created.Key)

	// Example 4: Working with complex custom fields
	fmt.Println("\n=== Complex Custom Fields ===")

	// User picker field
	complexFields := issue.NewCustomFields().
		SetUser("customfield_10007", "5b10ac8d82e05b22cc7d4ef5") // Account ID

	// Raw field for complex structures
	complexFields.SetRaw("customfield_10008", map[string]interface{}{
		"value": "Custom Value",
		"child": map[string]interface{}{
			"value": "Child Value",
		},
	})

	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: complexFields.ToMap(),
	})
	if err != nil {
		log.Fatalf("Failed to update complex fields: %v", err)
	}

	fmt.Printf("Updated complex custom fields for issue: %s\n", created.Key)

	// Example 5: Merge custom fields
	fmt.Println("\n=== Merging Custom Fields ===")

	baseFields := issue.NewCustomFields().
		SetString("customfield_10001", "Sprint 25").
		SetNumber("customfield_10002", 5)

	additionalFields := issue.NewCustomFields().
		SetLabels("customfield_10006", []string{"urgent", "bug-fix"})

	// Merge additional fields into base fields
	baseFields.Merge(additionalFields)

	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: baseFields.ToMap(),
	})
	if err != nil {
		log.Fatalf("Failed to merge and update fields: %v", err)
	}

	fmt.Printf("Merged and updated custom fields for issue: %s\n", created.Key)

	fmt.Println("\n=== Custom Fields Example Complete ===")
}
