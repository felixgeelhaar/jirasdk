// Package main demonstrates comment operations with the jira-connect library.
//
// This example covers:
//   - Creating comments with plain text (convenience method)
//   - Creating comments with rich ADF formatting
//   - Reading comment bodies and metadata safely
//   - Updating existing comments
//   - Deleting comments
//   - Best practices for comment handling
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
	issueKey := "PROJ-123" // Replace with your issue key

	fmt.Println("=== Jira Comment Operations Examples ===")
	fmt.Println()

	// Example 1: Add a simple plain text comment (recommended for most use cases)
	fmt.Println("1. Adding plain text comment...")
	plainComment := &issue.AddCommentInput{}
	plainComment.SetBodyText("This is a simple comment added via the SDK. Plain text is automatically converted to ADF format.")

	added, err := client.Issue.AddComment(ctx, issueKey, plainComment)
	if err != nil {
		log.Fatalf("Failed to add plain comment: %v", err)
	}
	fmt.Printf("   Added comment ID: %s\n", added.ID)
	fmt.Printf("   Comment body: %s\n", added.GetBodyText())
	fmt.Println()

	// Example 2: Add a comment with rich ADF formatting
	fmt.Println("2. Adding rich formatted comment...")
	richComment := &issue.AddCommentInput{}
	richADF := issue.NewADF().
		AddHeading("Status Update", 3).
		AddParagraph("The following items have been completed:").
		AddBulletList([]string{
			"Database schema updated",
			"API endpoints implemented",
			"Unit tests written",
		}).
		AddParagraph("Next steps:").
		AddBulletList([]string{
			"Integration testing",
			"Documentation update",
		})

	richComment.SetBody(richADF)

	addedRich, err := client.Issue.AddComment(ctx, issueKey, richComment)
	if err != nil {
		log.Fatalf("Failed to add rich comment: %v", err)
	}
	fmt.Printf("   Added rich comment ID: %s\n", addedRich.ID)
	fmt.Println()

	// Example 3: List all comments and safely read their content
	fmt.Println("3. Listing all comments on the issue...")
	comments, err := client.Issue.ListComments(ctx, issueKey)
	if err != nil {
		log.Fatalf("Failed to list comments: %v", err)
	}

	fmt.Printf("   Total comments: %d\n", len(comments))
	for i, c := range comments {
		fmt.Printf("\n   Comment #%d:\n", i+1)
		fmt.Printf("   ID: %s\n", c.ID)

		// Safe: use GetAuthorName to avoid nil pointer panic
		authorName := c.GetAuthorName()
		if authorName == "" {
			authorName = "Unknown"
		}
		fmt.Printf("   Author: %s\n", authorName)

		// Safe: use GetCreatedTime to avoid nil pointer panic
		created := c.GetCreatedTime()
		if !created.IsZero() {
			fmt.Printf("   Created: %s\n", created.Format(time.RFC3339))
		}

		// Safe: use GetUpdatedTime to avoid nil pointer panic
		updated := c.GetUpdatedTime()
		if !updated.IsZero() {
			fmt.Printf("   Updated: %s\n", updated.Format(time.RFC3339))
		}

		// Safe: use GetBodyText to extract plain text from ADF
		bodyText := c.GetBodyText()
		if len(bodyText) > 100 {
			bodyText = bodyText[:100] + "..."
		}
		fmt.Printf("   Body: %s\n", bodyText)
	}
	fmt.Println()

	// Example 4: Update an existing comment
	fmt.Println("4. Updating a comment...")
	if len(comments) > 0 {
		firstComment := comments[0]

		updateInput := &issue.UpdateCommentInput{}
		updateInput.SetBodyText("This comment has been updated via the SDK. Updates are timestamped automatically.")

		updated, err := client.Issue.UpdateComment(ctx, issueKey, firstComment.ID, updateInput)
		if err != nil {
			log.Printf("   Warning: Could not update comment: %v\n", err)
		} else {
			fmt.Printf("   Updated comment ID: %s\n", updated.ID)
			fmt.Printf("   New body: %s\n", updated.GetBodyText())

			// Safe: check update timestamp
			updateTime := updated.GetUpdatedTime()
			if !updateTime.IsZero() {
				fmt.Printf("   Updated at: %s\n", updateTime.Format(time.RFC3339))
			}
		}
	} else {
		fmt.Println("   No comments available to update")
	}
	fmt.Println()

	// Example 5: Update a comment with rich formatting
	fmt.Println("5. Updating a comment with rich formatting...")
	if len(comments) >= 2 {
		secondComment := comments[1]

		updateRich := &issue.UpdateCommentInput{}
		richUpdateADF := issue.NewADF().
			AddHeading("Updated Status", 3).
			AddParagraph("This comment has been updated with rich formatting:").
			AddBulletList([]string{
				"✓ Task completed",
				"✓ Review requested",
				"⏳ Waiting for approval",
			})

		updateRich.SetBody(richUpdateADF)

		updated, err := client.Issue.UpdateComment(ctx, issueKey, secondComment.ID, updateRich)
		if err != nil {
			log.Printf("   Warning: Could not update comment: %v\n", err)
		} else {
			fmt.Printf("   Updated comment ID: %s\n", updated.ID)
			// Extract text to show the content
			text := updated.GetBodyText()
			fmt.Printf("   New content (as text):\n   %s\n", text)
		}
	}
	fmt.Println()

	// Example 6: Working with comment metadata
	fmt.Println("6. Safely accessing comment metadata...")
	if len(comments) > 0 {
		comment := comments[0]

		// Get ADF body directly (for advanced use cases)
		adf := comment.GetBody()
		if adf != nil {
			fmt.Println("   Comment has ADF body")
			// You can inspect or manipulate the ADF structure
			if !adf.IsEmpty() {
				fmt.Println("   ADF body is not empty")
			}
		}

		// Get plain text body (recommended for most use cases)
		text := comment.GetBodyText()
		fmt.Printf("   Plain text body length: %d characters\n", len(text))

		// Get author information
		author := comment.GetAuthor()
		if author != nil {
			fmt.Printf("   Author account ID: %s\n", author.AccountID)
			fmt.Printf("   Author display name: %s\n", author.DisplayName)
		}
	}
	fmt.Println()

	// Example 7: Delete a comment (use with caution!)
	fmt.Println("7. Deleting a comment...")
	fmt.Println("   (Skipped in this example to preserve data)")
	fmt.Println("   To delete a comment, use:")
	fmt.Println("   err := client.Issue.DeleteComment(ctx, issueKey, commentID)")
	fmt.Println()

	// Example 8: Best practices summary
	fmt.Println("=== Best Practices for Comment Operations ===")
	fmt.Println()
	fmt.Println("1. Use SetBodyText() for simple plain text comments")
	fmt.Println("   input := &issue.AddCommentInput{}")
	fmt.Println("   input.SetBodyText(\"My comment\")")
	fmt.Println()
	fmt.Println("2. Use SetBody() with NewADF() for rich formatting")
	fmt.Println("   adf := issue.NewADF().AddHeading(\"Title\", 3).AddParagraph(\"Text\")")
	fmt.Println("   input.SetBody(adf)")
	fmt.Println()
	fmt.Println("3. Always use safe accessor methods when reading:")
	fmt.Println("   - GetBodyText() instead of direct Body access")
	fmt.Println("   - GetAuthorName() instead of comment.Author.DisplayName")
	fmt.Println("   - GetCreatedTime() instead of *comment.Created")
	fmt.Println()
	fmt.Println("4. Check for zero time values:")
	fmt.Println("   if !comment.GetCreatedTime().IsZero() {")
	fmt.Println("       // Use the time value")
	fmt.Println("   }")
	fmt.Println()
	fmt.Println("5. Handle errors appropriately:")
	fmt.Println("   - AddComment returns error if body is empty")
	fmt.Println("   - UpdateComment returns error if comment ID is invalid")
	fmt.Println()

	fmt.Println("=== Comment operations completed! ===")
}
