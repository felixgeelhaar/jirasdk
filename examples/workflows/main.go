package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/workflow"
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

	// Example 1: List all workflows
	fmt.Println("=== Listing All Workflows ===")
	workflows, err := client.Workflow.List(ctx, &workflow.ListOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Fatalf("Failed to list workflows: %v", err)
	}

	fmt.Printf("Found %d workflows:\n", len(workflows))
	for _, w := range workflows {
		fmt.Printf("- %s (ID: %s)\n", w.Name, w.ID)
		if w.Description != "" {
			fmt.Printf("  Description: %s\n", w.Description)
		}
		if w.IsDefault {
			fmt.Printf("  (Default workflow)\n")
		}
	}

	// Example 2: Get specific workflow
	if len(workflows) > 0 {
		fmt.Printf("\n=== Getting Workflow Details: %s ===\n", workflows[0].Name)
		workflowDetails, err := client.Workflow.Get(ctx, workflows[0].ID)
		if err != nil {
			log.Printf("Failed to get workflow: %v", err)
		} else {
			fmt.Printf("ID: %s\n", workflowDetails.ID)
			fmt.Printf("Name: %s\n", workflowDetails.Name)
			fmt.Printf("Description: %s\n", workflowDetails.Description)

			if len(workflowDetails.Statuses) > 0 {
				fmt.Println("\nStatuses:")
				for _, status := range workflowDetails.Statuses {
					fmt.Printf("- %s (ID: %s)\n", status.Name, status.ID)
					if status.StatusCategory != nil {
						fmt.Printf("  Category: %s (%s)\n", status.StatusCategory.Name, status.StatusCategory.Key)
					}
				}
			}

			if len(workflowDetails.Transitions) > 0 {
				fmt.Println("\nTransitions:")
				for _, transition := range workflowDetails.Transitions {
					fmt.Printf("- %s (ID: %s)\n", transition.Name, transition.ID)
					if transition.To != nil {
						fmt.Printf("  To: %s\n", transition.To.Name)
					}
				}
			}
		}
	}

	// Example 3: Get available transitions for an issue
	issueKey := "PROJ-123" // Replace with your issue key
	fmt.Printf("\n=== Getting Available Transitions for %s ===\n", issueKey)
	transitions, err := client.Workflow.GetTransitions(ctx, issueKey, &workflow.GetTransitionsOptions{
		Expand: []string{"transitions.fields"},
	})
	if err != nil {
		log.Printf("Failed to get transitions: %v", err)
	} else {
		fmt.Printf("Found %d available transitions:\n", len(transitions))
		for _, t := range transitions {
			fmt.Printf("\n- %s (ID: %s)\n", t.Name, t.ID)
			if t.To != nil {
				fmt.Printf("  Moves to: %s\n", t.To.Name)
			}
			if t.HasScreen {
				fmt.Printf("  Has screen: Yes\n")
			}
			if t.IsGlobal {
				fmt.Printf("  Global transition: Yes\n")
			}
			if t.IsConditional {
				fmt.Printf("  Conditional: Yes\n")
			}

			// Show required fields for this transition
			if len(t.Fields) > 0 {
				fmt.Println("  Required fields:")
				for fieldKey, field := range t.Fields {
					if field.Required {
						fmt.Printf("    - %s (%s)\n", field.Name, fieldKey)
					}
				}
			}
		}
	}

	// Example 4: Get all statuses
	fmt.Println("\n=== Getting All Statuses ===")
	statuses, err := client.Workflow.GetAllStatuses(ctx)
	if err != nil {
		log.Printf("Failed to get statuses: %v", err)
	} else {
		fmt.Printf("Found %d statuses:\n", len(statuses))

		// Group statuses by category
		byCategory := make(map[string][]*workflow.Status)
		for _, status := range statuses {
			categoryKey := "Unknown"
			if status.StatusCategory != nil {
				categoryKey = status.StatusCategory.Name
			}
			byCategory[categoryKey] = append(byCategory[categoryKey], status)
		}

		for category, categoryStatuses := range byCategory {
			fmt.Printf("\n%s:\n", category)
			for _, status := range categoryStatuses {
				fmt.Printf("  - %s (ID: %s)\n", status.Name, status.ID)
				if status.Description != "" {
					fmt.Printf("    %s\n", status.Description)
				}
			}
		}
	}

	// Example 5: Get specific status
	if len(statuses) > 0 {
		statusID := statuses[0].ID
		fmt.Printf("\n=== Getting Status Details (ID: %s) ===\n", statusID)
		status, err := client.Workflow.GetStatus(ctx, statusID)
		if err != nil {
			log.Printf("Failed to get status: %v", err)
		} else {
			fmt.Printf("Name: %s\n", status.Name)
			fmt.Printf("Description: %s\n", status.Description)
			if status.StatusCategory != nil {
				fmt.Printf("Category: %s (%s)\n", status.StatusCategory.Name, status.StatusCategory.Key)
				fmt.Printf("Color: %s\n", status.StatusCategory.ColorName)
			}
		}
	}

	// Example 6: List workflow schemes
	fmt.Println("\n=== Listing Workflow Schemes ===")
	schemes, err := client.Workflow.ListWorkflowSchemes(ctx, &workflow.ListWorkflowSchemesOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to list workflow schemes: %v", err)
	} else {
		fmt.Printf("Found %d workflow schemes:\n", len(schemes))
		for _, scheme := range schemes {
			fmt.Printf("\n- %s (ID: %d)\n", scheme.Name, scheme.ID)
			if scheme.Description != "" {
				fmt.Printf("  Description: %s\n", scheme.Description)
			}
			if scheme.DefaultWorkflow != "" {
				fmt.Printf("  Default Workflow: %s\n", scheme.DefaultWorkflow)
			}
			if scheme.Draft {
				fmt.Printf("  Status: Draft\n")
			}

			if len(scheme.IssueTypeMappings) > 0 {
				fmt.Println("  Issue Type Mappings:")
				for issueType, workflowName := range scheme.IssueTypeMappings {
					fmt.Printf("    %s -> %s\n", issueType, workflowName)
				}
			}
		}
	}

	// Example 7: Get specific workflow scheme
	if len(schemes) > 0 {
		schemeID := schemes[0].ID
		fmt.Printf("\n=== Getting Workflow Scheme Details (ID: %d) ===\n", schemeID)
		scheme, err := client.Workflow.GetWorkflowScheme(ctx, schemeID)
		if err != nil {
			log.Printf("Failed to get workflow scheme: %v", err)
		} else {
			fmt.Printf("Name: %s\n", scheme.Name)
			fmt.Printf("Description: %s\n", scheme.Description)
			fmt.Printf("Default Workflow: %s\n", scheme.DefaultWorkflow)
			fmt.Printf("Draft: %t\n", scheme.Draft)

			if scheme.LastModifiedUser != nil {
				fmt.Printf("Last Modified By: %s\n", scheme.LastModifiedUser.DisplayName)
			}
			if scheme.LastModified != "" {
				fmt.Printf("Last Modified: %s\n", scheme.LastModified)
			}
		}
	}

	// Example 8: Search for specific workflow
	fmt.Println("\n=== Searching for Specific Workflow ===")
	searchResults, err := client.Workflow.List(ctx, &workflow.ListOptions{
		WorkflowName: "Software",
		MaxResults:   10,
	})
	if err != nil {
		log.Printf("Failed to search workflows: %v", err)
	} else {
		fmt.Printf("Found %d workflows matching 'Software':\n", len(searchResults))
		for _, w := range searchResults {
			fmt.Printf("- %s (ID: %s)\n", w.Name, w.ID)
		}
	}

	// Example 9: Analyze workflow complexity
	fmt.Println("\n=== Workflow Complexity Analysis ===")
	for _, w := range workflows {
		if len(w.Statuses) > 0 || len(w.Transitions) > 0 {
			fmt.Printf("\n%s:\n", w.Name)
			fmt.Printf("  Statuses: %d\n", len(w.Statuses))
			fmt.Printf("  Transitions: %d\n", len(w.Transitions))

			// Calculate complexity score (simple heuristic)
			complexity := len(w.Statuses) + (len(w.Transitions) * 2)
			complexityLevel := "Simple"
			if complexity > 20 {
				complexityLevel = "Complex"
			} else if complexity > 10 {
				complexityLevel = "Moderate"
			}
			fmt.Printf("  Complexity: %s (score: %d)\n", complexityLevel, complexity)
		}
	}

	fmt.Println("\n=== Workflows Example Complete ===")
}
