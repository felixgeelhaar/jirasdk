package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issuelinktype"
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

	// Example 1: List all issue link types
	fmt.Println("=== Listing All Issue Link Types ===")
	linkTypes, err := client.IssueLinkType.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list issue link types: %v", err)
	}

	fmt.Printf("Found %d issue link types:\n", len(linkTypes))
	for _, lt := range linkTypes {
		fmt.Printf("\n- %s (ID: %s)\n", lt.Name, lt.ID)
		fmt.Printf("  Inward: %s\n", lt.Inward)
		fmt.Printf("  Outward: %s\n", lt.Outward)
		if lt.Self != "" {
			fmt.Printf("  Self: %s\n", lt.Self)
		}
	}

	// Example 2: Get specific link type details
	if len(linkTypes) > 0 {
		linkTypeID := linkTypes[0].ID
		fmt.Printf("\n=== Getting Link Type Details (ID: %s) ===\n", linkTypeID)

		linkType, err := client.IssueLinkType.Get(ctx, linkTypeID)
		if err != nil {
			log.Printf("Failed to get link type: %v", err)
		} else {
			fmt.Printf("ID: %s\n", linkType.ID)
			fmt.Printf("Name: %s\n", linkType.Name)
			fmt.Printf("Inward Description: %s\n", linkType.Inward)
			fmt.Printf("Outward Description: %s\n", linkType.Outward)
			fmt.Printf("Self: %s\n", linkType.Self)
		}
	}

	// Example 3: Create a custom issue link type
	fmt.Println("\n=== Creating Custom Issue Link Type ===")

	newLinkType, err := client.IssueLinkType.Create(ctx, &issuelinktype.CreateInput{
		Name:    "API Dependency",
		Inward:  "depends on",
		Outward: "is depended on by",
	})
	if err != nil {
		log.Fatalf("Failed to create link type: %v", err)
	}

	fmt.Printf("Created link type: %s (ID: %s)\n", newLinkType.Name, newLinkType.ID)
	fmt.Printf("  Inward: %s\n", newLinkType.Inward)
	fmt.Printf("  Outward: %s\n", newLinkType.Outward)

	linkTypeID := newLinkType.ID

	// Example 4: Update issue link type
	fmt.Println("\n=== Updating Issue Link Type ===")

	updated, err := client.IssueLinkType.Update(ctx, linkTypeID, &issuelinktype.UpdateInput{
		Name:    "Updated API Dependency",
		Inward:  "has dependency on",
		Outward: "is dependency for",
	})
	if err != nil {
		log.Printf("Failed to update link type: %v", err)
	} else {
		fmt.Printf("Updated link type: %s\n", updated.Name)
		fmt.Printf("  New Inward: %s\n", updated.Inward)
		fmt.Printf("  New Outward: %s\n", updated.Outward)
	}

	// Example 5: Common Link Type Patterns
	fmt.Println("\n=== Common Link Type Patterns ===")

	patterns := []struct {
		name    string
		inward  string
		outward string
		usage   string
	}{
		{
			name:    "Blocks",
			inward:  "is blocked by",
			outward: "blocks",
			usage:   "When one issue prevents another from being completed",
		},
		{
			name:    "Duplicates",
			inward:  "is duplicated by",
			outward: "duplicates",
			usage:   "When issues represent the same work",
		},
		{
			name:    "Relates",
			inward:  "relates to",
			outward: "relates to",
			usage:   "General relationship between issues",
		},
		{
			name:    "Causes",
			inward:  "is caused by",
			outward: "causes",
			usage:   "When one issue is the root cause of another",
		},
		{
			name:    "Dependency",
			inward:  "depends on",
			outward: "is dependency for",
			usage:   "When one issue requires another to be completed first",
		},
		{
			name:    "Parent/Child",
			inward:  "is parent of",
			outward: "is child of",
			usage:   "Hierarchical relationship between issues",
		},
	}

	for _, pattern := range patterns {
		fmt.Printf("📌 %s\n", pattern.name)
		fmt.Printf("   Inward: \"%s\"\n", pattern.inward)
		fmt.Printf("   Outward: \"%s\"\n", pattern.outward)
		fmt.Printf("   Usage: %s\n", pattern.usage)
		fmt.Println()
	}

	// Example 6: Link Type Statistics
	fmt.Println("=== Issue Link Type Statistics ===")

	// Categorize link types
	directional := []string{}
	bidirectional := []string{}
	hierarchical := []string{}

	for _, lt := range linkTypes {
		if lt.Inward == lt.Outward {
			bidirectional = append(bidirectional, lt.Name)
		} else if contains(lt.Name, "parent") || contains(lt.Name, "child") ||
			contains(lt.Name, "epic") || contains(lt.Name, "subtask") {
			hierarchical = append(hierarchical, lt.Name)
		} else {
			directional = append(directional, lt.Name)
		}
	}

	fmt.Printf("\nTotal Link Types: %d\n", len(linkTypes))

	if len(directional) > 0 {
		fmt.Printf("\nDirectional Links (%d):\n", len(directional))
		for _, name := range directional {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(bidirectional) > 0 {
		fmt.Printf("\nBidirectional Links (%d):\n", len(bidirectional))
		for _, name := range bidirectional {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(hierarchical) > 0 {
		fmt.Printf("\nHierarchical Links (%d):\n", len(hierarchical))
		for _, name := range hierarchical {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Example 7: Understanding Link Directionality
	fmt.Println("\n=== Understanding Link Directionality ===")

	fmt.Println("When creating a link from Issue A to Issue B:")
	fmt.Println()
	fmt.Println("Inward description:")
	fmt.Println("  - Describes the relationship FROM Issue B TO Issue A")
	fmt.Println("  - Displayed on Issue B")
	fmt.Println("  - Example: 'is blocked by' means B is blocked by A")
	fmt.Println()
	fmt.Println("Outward description:")
	fmt.Println("  - Describes the relationship FROM Issue A TO Issue B")
	fmt.Println("  - Displayed on Issue A")
	fmt.Println("  - Example: 'blocks' means A blocks B")
	fmt.Println()

	// Practical example
	fmt.Println("Practical Example:")
	fmt.Println("  Link Type: 'Blocks'")
	fmt.Println("  Inward: 'is blocked by'")
	fmt.Println("  Outward: 'blocks'")
	fmt.Println()
	fmt.Println("  Creating link from PROJ-123 to PROJ-456:")
	fmt.Println("  - On PROJ-123: 'blocks PROJ-456' (outward)")
	fmt.Println("  - On PROJ-456: 'is blocked by PROJ-123' (inward)")

	// Example 8: Best Practices
	fmt.Println("\n=== Best Practices for Issue Link Types ===")

	fmt.Println("✅ DO:")
	fmt.Println("  - Use clear, descriptive names")
	fmt.Println("  - Make inward/outward descriptions grammatically correct")
	fmt.Println("  - Keep link types consistent across projects")
	fmt.Println("  - Document the purpose of custom link types")
	fmt.Println("  - Use existing link types when appropriate")
	fmt.Println()

	fmt.Println("❌ DON'T:")
	fmt.Println("  - Create duplicate link types")
	fmt.Println("  - Use vague or ambiguous descriptions")
	fmt.Println("  - Delete link types that are in use")
	fmt.Println("  - Create too many custom link types")
	fmt.Println("  - Use link types for workflow states")
	fmt.Println()

	// Example 9: Design Considerations
	fmt.Println("=== Link Type Design Considerations ===")

	considerations := []struct {
		question string
		guidance string
	}{
		{
			question: "When should I create a custom link type?",
			guidance: "When existing link types don't express the relationship you need",
		},
		{
			question: "Should my link type be directional?",
			guidance: "Most relationships are directional (blocks, depends on). Use bidirectional only for mutual relationships",
		},
		{
			question: "How do I name inward/outward descriptions?",
			guidance: "Use verb phrases that complete the sentence 'This issue [description] that issue'",
		},
		{
			question: "Can I modify system link types?",
			guidance: "No, only custom link types can be modified",
		},
	}

	for i, c := range considerations {
		fmt.Printf("%d. %s\n", i+1, c.question)
		fmt.Printf("   → %s\n\n", c.guidance)
	}

	// Clean up: Delete the custom link type
	fmt.Println("=== Cleaning Up ===")
	fmt.Println("Deleting custom link type...")
	err = client.IssueLinkType.Delete(ctx, linkTypeID)
	if err != nil {
		log.Printf("Failed to delete link type: %v", err)
	} else {
		fmt.Printf("Link type '%s' deleted successfully\n", updated.Name)
	}

	fmt.Println("\n=== Issue Link Types Example Complete ===")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
