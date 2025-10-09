// Package main demonstrates Field resource management with the jira-connect library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/field"
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

	// Example 1: List all fields
	fmt.Println("=== Listing All Fields ===")
	fields, err := client.Field.List(ctx)
	if err != nil {
		log.Printf("Failed to list fields: %v", err)
	} else {
		fmt.Printf("Found %d fields:\n", len(fields))

		// Show system fields
		systemCount := 0
		for _, f := range fields {
			if !f.Custom {
				systemCount++
			}
		}
		fmt.Printf("  System fields: %d\n", systemCount)

		// Show custom fields
		customCount := 0
		fmt.Println("\nCustom fields:")
		for _, f := range fields {
			if f.Custom {
				customCount++
				fmt.Printf("  %s: %s\n", f.ID, f.Name)
			}
		}
		fmt.Printf("Total custom fields: %d\n", customCount)
	}

	// Example 2: Get a specific field
	if len(fields) > 0 {
		var customFieldID string
		for _, f := range fields {
			if f.Custom {
				customFieldID = f.ID
				break
			}
		}

		if customFieldID != "" {
			fmt.Printf("\n=== Getting Field %s ===\n", customFieldID)
			specificField, err := client.Field.Get(ctx, customFieldID)
			if err != nil {
				log.Printf("Failed to get field: %v", err)
			} else {
				fmt.Printf("Field: %s\n", specificField.Name)
				fmt.Printf("  ID: %s\n", specificField.ID)
				fmt.Printf("  Custom: %t\n", specificField.Custom)
				if specificField.Schema != nil {
					fmt.Printf("  Type: %s\n", specificField.Schema.Type)
				}
				fmt.Printf("  Searchable: %t\n", specificField.Searchable)
			}
		}
	}

	// Example 3: Create a custom field
	fmt.Println("\n=== Creating Custom Field ===")
	newField, err := client.Field.Create(ctx, &field.CreateFieldInput{
		Name:        "Team Velocity",
		Description: "Team velocity measurement",
		Type:        "com.atlassian.jira.plugin.system.customfieldtypes:float",
		SearcherKey: "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber",
	})
	if err != nil {
		log.Printf("Failed to create field: %v", err)
	} else {
		fmt.Printf("Created field: %s\n", newField.Name)
		fmt.Printf("  ID: %s\n", newField.ID)
		fmt.Printf("  Type: %s\n", newField.Schema.Type)
	}

	// Example 4: Update a custom field
	if newField != nil {
		fmt.Printf("\n=== Updating Field %s ===\n", newField.ID)
		updatedField, err := client.Field.Update(ctx, newField.ID, &field.UpdateFieldInput{
			Name:        "Team Velocity Points",
			Description: "Updated team velocity measurement in story points",
		})
		if err != nil {
			log.Printf("Failed to update field: %v", err)
		} else {
			fmt.Printf("Updated field: %s\n", updatedField.Name)
			fmt.Printf("  Description: %s\n", updatedField.Description)
		}
	}

	// Example 5: List field contexts
	if newField != nil {
		fmt.Printf("\n=== Listing Contexts for Field %s ===\n", newField.ID)
		contexts, err := client.Field.ListContexts(ctx, newField.ID)
		if err != nil {
			log.Printf("Failed to list contexts: %v", err)
		} else {
			fmt.Printf("Found %d contexts:\n", len(contexts))
			for i, c := range contexts {
				fmt.Printf("%d. %s (ID: %s)\n", i+1, c.Name, c.ID)
				fmt.Printf("   Global: %t, Any Issue Type: %t\n", c.IsGlobalContext, c.IsAnyIssueType)
			}
		}
	}

	// Example 6: Create a field context
	if newField != nil {
		fmt.Printf("\n=== Creating Context for Field %s ===\n", newField.ID)
		newContext, err := client.Field.CreateContext(ctx, newField.ID, &field.CreateContextInput{
			Name:        "Software Projects",
			Description: "Context for software development projects",
		})
		if err != nil {
			log.Printf("Failed to create context: %v", err)
		} else {
			fmt.Printf("Created context: %s (ID: %s)\n", newContext.Name, newContext.ID)
		}

		// Example 7: Update context
		if newContext != nil {
			fmt.Printf("\n=== Updating Context %s ===\n", newContext.ID)
			updatedContext, err := client.Field.UpdateContext(ctx, newField.ID, newContext.ID, &field.UpdateContextInput{
				Name:        "Agile Software Projects",
				Description: "Context for agile software development teams",
			})
			if err != nil {
				log.Printf("Failed to update context: %v", err)
			} else {
				fmt.Printf("Updated context: %s\n", updatedContext.Name)
			}
		}

		// Example 8: Delete context
		if newContext != nil {
			fmt.Printf("\n=== Deleting Context %s ===\n", newContext.ID)
			err := client.Field.DeleteContext(ctx, newField.ID, newContext.ID)
			if err != nil {
				log.Printf("Failed to delete context: %v", err)
			} else {
				fmt.Printf("Successfully deleted context %s\n", newContext.ID)
			}
		}
	}

	// Example 9: Delete field (cleanup)
	if newField != nil {
		fmt.Printf("\n=== Deleting Field %s ===\n", newField.ID)
		err := client.Field.Delete(ctx, newField.ID)
		if err != nil {
			log.Printf("Failed to delete field: %v", err)
		} else {
			fmt.Printf("Successfully deleted field '%s'\n", newField.Name)
		}
	}

	fmt.Println("\n=== Field Management Example Complete ===")
}
