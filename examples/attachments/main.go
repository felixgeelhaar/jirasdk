// Package main demonstrates attachment upload/download with the jira-connect library.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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

	// Example 1: Upload a single attachment
	fmt.Println("=== Uploading Single Attachment ===")

	file, err := os.Open("document.pdf")
	if err != nil {
		log.Printf("Skipping file upload (file not found): %v", err)
	} else {
		defer file.Close()

		attachments, err := client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
			Filename: "document.pdf",
			Content:  file,
		})
		if err != nil {
			log.Fatalf("Failed to upload attachment: %v", err)
		}

		for _, att := range attachments {
			fmt.Printf("Uploaded: %s (ID: %s, Size: %d bytes)\n", att.Filename, att.ID, att.Size)
		}
	}

	// Example 2: Upload attachment from string/bytes
	fmt.Println("\n=== Uploading Attachment from String ===")

	reportContent := `# Daily Report
Date: 2024-01-15

## Summary
- Tasks completed: 5
- Issues resolved: 3
- Code reviews: 2

## Notes
All sprint goals achieved.
`

	attachments, err := client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
		Filename: "daily-report.md",
		Content:  strings.NewReader(reportContent),
	})
	if err != nil {
		log.Fatalf("Failed to upload report: %v", err)
	}

	fmt.Printf("Uploaded report: %s (ID: %s)\n", attachments[0].Filename, attachments[0].ID)
	attachmentID := attachments[0].ID

	// Example 3: Get attachment metadata
	fmt.Println("\n=== Getting Attachment Metadata ===")

	metadata, err := client.Issue.GetAttachment(ctx, attachmentID)
	if err != nil {
		log.Fatalf("Failed to get attachment: %v", err)
	}

	fmt.Printf("Attachment Details:\n")
	fmt.Printf("  Filename: %s\n", metadata.Filename)
	fmt.Printf("  Size: %d bytes\n", metadata.Size)
	fmt.Printf("  MIME Type: %s\n", metadata.MimeType)
	if metadata.Created != nil {
		fmt.Printf("  Created: %s\n", metadata.Created.Format("2006-01-02 15:04:05"))
	}
	if metadata.Author != nil {
		fmt.Printf("  Author: %s\n", metadata.Author.DisplayName)
	}

	// Example 4: Download attachment
	fmt.Println("\n=== Downloading Attachment ===")

	content, err := client.Issue.DownloadAttachment(ctx, attachmentID)
	if err != nil {
		log.Fatalf("Failed to download attachment: %v", err)
	}
	defer content.Close()

	// Read content
	data, err := io.ReadAll(content)
	if err != nil {
		log.Fatalf("Failed to read content: %v", err)
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	fmt.Println("Content preview:")
	fmt.Println(string(data))

	// Example 5: Save downloaded attachment to file
	fmt.Println("\n=== Saving Attachment to File ===")

	downloaded, err := client.Issue.DownloadAttachment(ctx, attachmentID)
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	defer downloaded.Close()

	outFile, err := os.Create("downloaded-report.md")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, downloaded)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}

	fmt.Printf("Saved %d bytes to downloaded-report.md\n", written)

	// Example 6: Delete attachment
	fmt.Println("\n=== Deleting Attachment ===")

	err = client.Issue.DeleteAttachment(ctx, attachmentID)
	if err != nil {
		log.Fatalf("Failed to delete attachment: %v", err)
	}

	fmt.Printf("Deleted attachment: %s\n", attachmentID)

	// Example 7: Upload multiple files at once
	fmt.Println("\n=== Uploading Multiple Attachments ===")

	// Create sample files in memory
	files := []struct {
		filename string
		content  string
	}{
		{"notes.txt", "Meeting notes from daily standup"},
		{"config.json", `{"environment": "production", "debug": false}`},
		{"readme.md", "# Project Documentation\n\nThis is a sample readme."},
	}

	for _, f := range files {
		attachments, err := client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
			Filename: f.filename,
			Content:  strings.NewReader(f.content),
		})
		if err != nil {
			log.Printf("Failed to upload %s: %v", f.filename, err)
			continue
		}

		if len(attachments) > 0 {
			fmt.Printf("Uploaded: %s (ID: %s, %d bytes)\n",
				attachments[0].Filename,
				attachments[0].ID,
				attachments[0].Size)
		}
	}

	// Example 8: Error handling
	fmt.Println("\n=== Error Handling ===")

	// Try to download non-existent attachment
	_, err = client.Issue.DownloadAttachment(ctx, "99999")
	if err != nil {
		fmt.Printf("Expected error for non-existent attachment: %v\n", err)
	}

	// Try to upload with empty filename
	_, err = client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
		Filename: "",
		Content:  strings.NewReader("content"),
	})
	if err != nil {
		fmt.Printf("Expected error for empty filename: %v\n", err)
	}

	fmt.Println("\n=== Attachments Example Complete ===")
}
