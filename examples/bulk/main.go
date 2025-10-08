package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/bulk"
)

func main() {
	// Get credentials from environment
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")
	projectKey := os.Getenv("JIRA_PROJECT_KEY")

	if email == "" || apiToken == "" || baseURL == "" {
		log.Fatal("JIRA_EMAIL, JIRA_API_TOKEN, and JIRA_BASE_URL must be set")
	}

	if projectKey == "" {
		projectKey = "PROJ" // Default project key
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

	// Example 1: Bulk create issues
	fmt.Println("=== Bulk Creating Issues ===")

	createInput := &bulk.CreateIssuesInput{
		IssueUpdates: []bulk.IssueUpdate{
			{
				Fields: map[string]interface{}{
					"project": map[string]string{
						"key": projectKey,
					},
					"summary": "Bulk Created Issue 1",
					"description": "This issue was created via bulk API",
					"issuetype": map[string]string{
						"name": "Task",
					},
					"priority": map[string]string{
						"name": "Medium",
					},
				},
			},
			{
				Fields: map[string]interface{}{
					"project": map[string]string{
						"key": projectKey,
					},
					"summary": "Bulk Created Issue 2",
					"description": "This is the second bulk created issue",
					"issuetype": map[string]string{
						"name": "Task",
					},
					"priority": map[string]string{
						"name": "Low",
					},
					"labels": []string{"bulk-created", "demo"},
				},
			},
			{
				Fields: map[string]interface{}{
					"project": map[string]string{
						"key": projectKey,
					},
					"summary": "Bulk Created Issue 3",
					"description": "Third issue created in bulk",
					"issuetype": map[string]string{
						"name": "Bug",
					},
					"priority": map[string]string{
						"name": "High",
					},
					"labels": []string{"bulk-created", "demo", "urgent"},
				},
			},
		},
	}

	result, err := client.Bulk.CreateIssues(ctx, createInput)
	if err != nil {
		log.Printf("Failed to create issues: %v", err)
	} else {
		fmt.Printf("Successfully created %d issues:\n", len(result.Issues))
		var createdKeys []string
		for _, issue := range result.Issues {
			fmt.Printf("- %s (ID: %s)\n", issue.Key, issue.ID)
			createdKeys = append(createdKeys, issue.Key)
		}

		if len(result.Errors) > 0 {
			fmt.Printf("\nEncountered %d errors during creation:\n", len(result.Errors))
			for i, err := range result.Errors {
				fmt.Printf("Error %d (element %d):\n", i+1, err.FailedElementNumber)
				if err.ElementErrors != nil {
					for _, msg := range err.ElementErrors.ErrorMessages {
						fmt.Printf("  - %s\n", msg)
					}
				}
			}
		}

		// Example 2: Bulk delete the created issues
		// Note: Uncomment to actually delete the demo issues
		/*
		if len(createdKeys) > 0 {
			fmt.Println("\n=== Bulk Deleting Issues ===")
			fmt.Printf("Deleting %d issues: %v\n", len(createdKeys), createdKeys)

			deleteInput := &bulk.DeleteIssuesInput{
				IssueIDs: createdKeys,
			}

			err = client.Bulk.DeleteIssues(ctx, deleteInput)
			if err != nil {
				log.Printf("Failed to delete issues: %v", err)
			} else {
				fmt.Printf("Successfully deleted %d issues\n", len(createdKeys))
			}
		}
		*/
	}

	// Example 3: Demonstrating bulk operation progress tracking
	// (This is a conceptual example - actual long-running operations would return a task ID)
	fmt.Println("\n=== Bulk Operation Progress Tracking ===")
	fmt.Println("For long-running bulk operations, you can track progress:")
	fmt.Println()
	fmt.Println("1. Submit a bulk operation that returns a task ID")
	fmt.Println("2. Poll for progress using GetProgress()")
	fmt.Println("3. Or use WaitForCompletion() to block until done")
	fmt.Println()
	fmt.Println("Example code:")
	fmt.Println(`
    // Option 1: Poll manually
    taskID := "returned-task-id"
    progress, err := client.Bulk.GetProgress(ctx, taskID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Operation is %d%% complete\n", progress.ProgressPercent)

    // Option 2: Wait for completion (blocking)
    progress, err := client.Bulk.WaitForCompletion(ctx, taskID, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }

    if progress.Status == bulk.BulkOperationStatusComplete {
        fmt.Printf("Success! Processed %d items\n", progress.Result.SuccessCount)
    }
	`)

	// Example 4: Demonstrating best practices
	fmt.Println("\n=== Best Practices for Bulk Operations ===")
	fmt.Println("1. Maximum 1000 issues per request")
	fmt.Println("   - For larger batches, split into multiple requests")
	fmt.Println()
	fmt.Println("2. Handle partial failures gracefully")
	fmt.Println("   - Check result.Errors for failures")
	fmt.Println("   - Retry failed items if needed")
	fmt.Println()
	fmt.Println("3. Use appropriate field operations")
	fmt.Println("   - Use 'add' for multi-value fields (labels, components)")
	fmt.Println("   - Use 'set' to replace field values")
	fmt.Println("   - Use 'remove' to delete specific values")
	fmt.Println()
	fmt.Println("4. Consider performance")
	fmt.Println("   - Bulk operations are async for large batches")
	fmt.Println("   - Use progress tracking for operations > 100 items")
	fmt.Println("   - Implement retry logic with exponential backoff")
	fmt.Println()

	// Example 5: Splitting large batches
	fmt.Println("=== Handling Large Batches ===")

	// Simulate a large batch of issues to create
	largeBatch := make([]bulk.IssueUpdate, 2500)
	for i := range largeBatch {
		largeBatch[i] = bulk.IssueUpdate{
			Fields: map[string]interface{}{
				"project": map[string]string{"key": projectKey},
				"summary": fmt.Sprintf("Issue %d", i+1),
				"issuetype": map[string]string{"name": "Task"},
			},
		}
	}

	fmt.Printf("Have %d issues to create (exceeds limit of %d)\n",
		len(largeBatch), bulk.MaxBulkIssues)

	// Split into batches
	batches := splitIntoBatches(largeBatch, bulk.MaxBulkIssues)
	fmt.Printf("Split into %d batches\n", len(batches))
	fmt.Println()
	fmt.Println("Processing batches:")

	for i, batch := range batches {
		fmt.Printf("  Batch %d: %d issues\n", i+1, len(batch))

		// Note: Uncomment to actually create these issues
		/*
		batchInput := &bulk.CreateIssuesInput{
			IssueUpdates: batch,
		}

		result, err := client.Bulk.CreateIssues(ctx, batchInput)
		if err != nil {
			log.Printf("Batch %d failed: %v", i+1, err)
			continue
		}

		fmt.Printf("  Batch %d: Created %d issues\n", i+1, len(result.Issues))

		// Small delay between batches to avoid rate limiting
		time.Sleep(time.Second)
		*/
	}

	fmt.Println("\n=== Bulk Operations Example Complete ===")
	fmt.Println("\nNote: The actual bulk delete is commented out to avoid")
	fmt.Println("accidentally deleting demo issues. Uncomment if needed.")
}

// splitIntoBatches splits a slice of issue updates into batches of specified size
func splitIntoBatches(issues []bulk.IssueUpdate, batchSize int) [][]bulk.IssueUpdate {
	var batches [][]bulk.IssueUpdate

	for i := 0; i < len(issues); i += batchSize {
		end := i + batchSize
		if end > len(issues) {
			end = len(issues)
		}
		batches = append(batches, issues[i:end])
	}

	return batches
}
