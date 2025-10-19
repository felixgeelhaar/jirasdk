// Package main demonstrates date and time handling with the jira-connect library.
//
// This example covers:
//   - Reading standard date fields (Created, Updated, DueDate)
//   - Setting DueDate when creating and updating issues
//   - Working with custom date and datetime fields
//   - Safe date handling patterns to avoid nil pointer panics
//   - Date formatting and parsing best practices
//   - Automatic flexible date/time parsing for Jira's various formats
//
// Note: The SDK automatically handles Jira's various date/time formats:
//   - Date only: "2025-10-30"
//   - DateTime with timezone: "2024-01-01T10:30:00.000+0000"
//   - RFC3339: "2024-01-01T10:30:00.000Z"
//   - Time only: "15:30:00"
//
// This works for both standard fields AND custom date/datetime fields!
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
	// Create client with API token authentication
	client, err := jira.NewClient(
		jira.WithBaseURL(os.Getenv("JIRA_BASE_URL")),
		jira.WithAPIToken(os.Getenv("JIRA_EMAIL"), os.Getenv("JIRA_API_TOKEN")),
		jira.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Date and Time Handling Examples ===")
	fmt.Println()

	// Example 1: Create an issue with DueDate
	fmt.Println("1. Creating issue with DueDate...")
	dueDate := time.Now().AddDate(0, 0, 14) // 14 days from now
	createInput := &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project: &issue.Project{
				Key: "PROJ", // Replace with your project key
			},
			Summary: "Task with due date",
			IssueType: &issue.IssueType{
				Name: "Task",
			},
			Description: "This task has a due date set",
			DueDate:     &dueDate,
		},
	}

	created, err := client.Issue.Create(ctx, createInput)
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("   Created issue: %s with due date: %s\n", created.Key, dueDate.Format("2006-01-02"))
	fmt.Println()

	// Example 2: Read standard date fields safely
	fmt.Println("2. Reading date fields safely...")
	retrieved, err := client.Issue.Get(ctx, created.Key, &issue.GetOptions{
		Fields: []string{"summary", "created", "updated", "duedate"},
	})
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	fmt.Printf("   Issue: %s\n", retrieved.Key)

	// ✅ SAFE: Use GetCreatedTime() - returns zero time if nil
	if created := retrieved.GetCreatedTime(); !created.IsZero() {
		fmt.Printf("   Created: %s\n", created.Format(time.RFC3339))
	} else {
		fmt.Println("   Created: not set")
	}

	// ✅ SAFE: Use GetUpdatedTime() - returns zero time if nil
	if updated := retrieved.GetUpdatedTime(); !updated.IsZero() {
		fmt.Printf("   Updated: %s\n", updated.Format(time.RFC3339))
	} else {
		fmt.Println("   Updated: not set")
	}

	// ✅ SAFE: Use GetDueDateValue() - returns zero time if nil
	if dueDate := retrieved.GetDueDateValue(); !dueDate.IsZero() {
		fmt.Printf("   Due Date: %s\n", dueDate.Format("2006-01-02"))

		// Check if overdue
		if time.Now().After(dueDate) {
			fmt.Println("   ⚠️  This task is OVERDUE!")
		} else {
			daysUntilDue := int(time.Until(dueDate).Hours() / 24)
			fmt.Printf("   ℹ️  %d days until due\n", daysUntilDue)
		}
	} else {
		fmt.Println("   Due Date: not set")
	}
	fmt.Println()

	// Example 3: Alternative safe pattern using GetDueDate() for pointer access
	fmt.Println("3. Alternative safe pattern with pointer access...")
	if dueDatePtr := retrieved.GetDueDate(); dueDatePtr != nil {
		fmt.Printf("   Due Date (via pointer): %s\n", dueDatePtr.Format("2006-01-02"))
	} else {
		fmt.Println("   Due Date: nil (not set)")
	}
	fmt.Println()

	// Example 4: Update DueDate
	fmt.Println("4. Updating due date...")
	newDueDate := time.Now().AddDate(0, 0, 30) // 30 days from now
	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: map[string]interface{}{
			"duedate": newDueDate.Format("2006-01-02"), // Jira expects YYYY-MM-DD format
		},
	})
	if err != nil {
		log.Fatalf("Failed to update due date: %v", err)
	}
	fmt.Printf("   Updated due date to: %s\n", newDueDate.Format("2006-01-02"))
	fmt.Println()

	// Example 5: Working with custom date fields
	fmt.Println("5. Working with custom date fields...")

	// Set custom date field (YYYY-MM-DD format)
	customDate := time.Now().AddDate(0, 1, 0) // 1 month from now
	customFields := issue.NewCustomFields().
		SetDate("customfield_10001", customDate)

	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: customFields.ToMap(),
	})
	if err != nil {
		log.Printf("   Warning: Failed to set custom date field: %v\n", err)
	} else {
		fmt.Printf("   Set custom date field to: %s\n", customDate.Format("2006-01-02"))
	}

	// Set custom datetime field (RFC3339 format with timezone)
	customDateTime := time.Now().AddDate(0, 0, 7) // 1 week from now
	customFields = issue.NewCustomFields().
		SetDateTime("customfield_10002", customDateTime)

	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: customFields.ToMap(),
	})
	if err != nil {
		log.Printf("   Warning: Failed to set custom datetime field: %v\n", err)
	} else {
		fmt.Printf("   Set custom datetime field to: %s\n", customDateTime.Format(time.RFC3339))
	}
	fmt.Println()

	// Example 6: Read custom date fields
	fmt.Println("6. Reading custom date fields...")
	retrieved, err = client.Issue.Get(ctx, created.Key, nil)
	if err != nil {
		log.Fatalf("Failed to get issue: %v", err)
	}

	if customDate, ok := retrieved.Fields.Custom.GetDate("customfield_10001"); ok {
		fmt.Printf("   Custom Date Field: %s\n", customDate.Format("2006-01-02"))
	} else {
		fmt.Println("   Custom Date Field: not set or not found")
	}

	if customDateTime, ok := retrieved.Fields.Custom.GetDateTime("customfield_10002"); ok {
		fmt.Printf("   Custom DateTime Field: %s\n", customDateTime.Format(time.RFC3339))
	} else {
		fmt.Println("   Custom DateTime Field: not set or not found")
	}
	fmt.Println()

	// Example 7: Clear a due date
	fmt.Println("7. Clearing due date...")
	err = client.Issue.Update(ctx, created.Key, &issue.UpdateInput{
		Fields: map[string]interface{}{
			"duedate": nil, // Set to nil to clear
		},
	})
	if err != nil {
		log.Fatalf("Failed to clear due date: %v", err)
	}
	fmt.Println("   Due date cleared")
	fmt.Println()

	// Example 8: Common date formats for display
	fmt.Println("8. Common date formatting patterns...")
	now := time.Now()
	fmt.Printf("   ISO 8601 Date: %s\n", now.Format("2006-01-02"))
	fmt.Printf("   RFC3339 (with timezone): %s\n", now.Format(time.RFC3339))
	fmt.Printf("   US Format: %s\n", now.Format("01/02/2006"))
	fmt.Printf("   EU Format: %s\n", now.Format("02/01/2006"))
	fmt.Printf("   Human Readable: %s\n", now.Format("January 2, 2006"))
	fmt.Printf("   With Time: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Example 9: DANGEROUS - What NOT to do
	fmt.Println("9. ⚠️  DANGEROUS patterns to AVOID:")
	fmt.Println(`
   ❌ NEVER do this (can cause nil pointer panic):

   dueDate := issue.Fields.DueDate
   fmt.Println(dueDate.Format("2006-01-02"))  // PANIC if DueDate is nil!

   ✅ ALWAYS use safe accessors instead:

   if dueDate := issue.GetDueDateValue(); !dueDate.IsZero() {
       fmt.Println(dueDate.Format("2006-01-02"))  // Safe!
   }

   Or use the pointer accessor:

   if dueDate := issue.GetDueDate(); dueDate != nil {
       fmt.Println(dueDate.Format("2006-01-02"))  // Safe!
   }
`)

	fmt.Println("=== Date handling examples completed! ===")
}
