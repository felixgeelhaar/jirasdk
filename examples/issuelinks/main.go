package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/felixgeelhaar/jira-connect"
	"github.com/felixgeelhaar/jira-connect/auth"
	"github.com/felixgeelhaar/jira-connect/core/issue"
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
	client := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAuth(auth.NewBasicAuth(email, apiToken)),
	)

	ctx := context.Background()

	// Example 1: List available issue link types
	fmt.Println("=== Available Issue Link Types ===")
	linkTypes, err := client.Issue.ListIssueLinkTypes(ctx)
	if err != nil {
		log.Fatalf("Failed to list issue link types: %v", err)
	}

	for _, linkType := range linkTypes {
		fmt.Printf("- %s (ID: %s)\n", linkType.Name, linkType.ID)
		fmt.Printf("  Inward: %s\n", linkType.Inward)
		fmt.Printf("  Outward: %s\n", linkType.Outward)
		fmt.Println()
	}

	// Example 2: Create a "blocks" relationship
	// PROJ-123 blocks PROJ-456
	fmt.Println("=== Creating Issue Link: PROJ-123 blocks PROJ-456 ===")
	err = client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
		Type:         issue.BlocksLinkType(),
		InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
		OutwardIssue: &issue.IssueRef{Key: "PROJ-456"},
		Comment: &issue.LinkComment{
			Body: "These issues are related through a blocking dependency",
		},
	})
	if err != nil {
		log.Printf("Failed to create issue link: %v", err)
	} else {
		fmt.Println("✓ Issue link created successfully")
	}

	// Example 3: Create a "duplicates" relationship
	fmt.Println("\n=== Creating Issue Link: PROJ-789 duplicates PROJ-123 ===")
	err = client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
		Type:         issue.DuplicatesLinkType(),
		InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
		OutwardIssue: &issue.IssueRef{Key: "PROJ-789"},
	})
	if err != nil {
		log.Printf("Failed to create duplicate link: %v", err)
	} else {
		fmt.Println("✓ Duplicate link created successfully")
	}

	// Example 4: Create a "relates to" relationship
	fmt.Println("\n=== Creating Issue Link: PROJ-111 relates to PROJ-222 ===")
	err = client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
		Type:         issue.RelatesToLinkType(),
		InwardIssue:  &issue.IssueRef{Key: "PROJ-111"},
		OutwardIssue: &issue.IssueRef{Key: "PROJ-222"},
	})
	if err != nil {
		log.Printf("Failed to create relates link: %v", err)
	} else {
		fmt.Println("✓ Relates link created successfully")
	}

	// Example 5: Get specific link type details
	if len(linkTypes) > 0 {
		fmt.Printf("\n=== Getting Link Type Details (ID: %s) ===\n", linkTypes[0].ID)
		linkType, err := client.Issue.GetIssueLinkType(ctx, linkTypes[0].ID)
		if err != nil {
			log.Printf("Failed to get link type: %v", err)
		} else {
			fmt.Printf("Name: %s\n", linkType.Name)
			fmt.Printf("Inward: %s\n", linkType.Inward)
			fmt.Printf("Outward: %s\n", linkType.Outward)
		}
	}

	// Example 6: Get links for an issue
	fmt.Println("\n=== Getting Issue Links for PROJ-123 ===")
	links, err := client.Issue.GetIssueLinks(ctx, "PROJ-123")
	if err != nil {
		log.Printf("Failed to get issue links: %v", err)
	} else {
		fmt.Printf("Found %d links\n", len(links))
		for _, link := range links {
			if link.InwardIssue != nil {
				fmt.Printf("- %s %s (ID: %s)\n",
					link.Type.Inward,
					link.InwardIssue.Key,
					link.ID,
				)
			}
			if link.OutwardIssue != nil {
				fmt.Printf("- %s %s (ID: %s)\n",
					link.Type.Outward,
					link.OutwardIssue.Key,
					link.ID,
				)
			}
		}
	}

	// Example 7: Delete an issue link
	// Note: Replace "10000" with an actual link ID
	fmt.Println("\n=== Deleting Issue Link (Example) ===")
	err = client.Issue.DeleteIssueLink(ctx, "10000")
	if err != nil {
		log.Printf("Failed to delete issue link: %v", err)
	} else {
		fmt.Println("✓ Issue link deleted successfully")
	}

	// Example 8: Using helper functions for different link types
	fmt.Println("\n=== Available Helper Functions ===")
	helpers := map[string]*issue.IssueLinkType{
		"Blocks":     issue.BlocksLinkType(),
		"Duplicates": issue.DuplicatesLinkType(),
		"Relates":    issue.RelatesToLinkType(),
		"Causes":     issue.CausesLinkType(),
		"Clones":     issue.ClonesLinkType(),
	}

	for name, linkType := range helpers {
		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  Function: %sLinkType()\n", name)
		fmt.Printf("  Name: %s\n", linkType.Name)
		fmt.Printf("  Inward: %s\n", linkType.Inward)
		fmt.Printf("  Outward: %s\n", linkType.Outward)
	}

	// Example 9: Create link with visibility restrictions
	fmt.Println("\n=== Creating Link with Comment Visibility ===")
	err = client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
		Type:         issue.BlocksLinkType(),
		InwardIssue:  &issue.IssueRef{Key: "PROJ-100"},
		OutwardIssue: &issue.IssueRef{Key: "PROJ-200"},
		Comment: &issue.LinkComment{
			Body: "This link is only visible to developers",
			Visibility: &issue.CommentVisibility{
				Type:  "role",
				Value: "Developers",
			},
		},
	})
	if err != nil {
		log.Printf("Failed to create link with visibility: %v", err)
	} else {
		fmt.Println("✓ Link with restricted comment created successfully")
	}

	fmt.Println("\n=== Issue Links Example Complete ===")
}
