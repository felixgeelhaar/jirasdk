package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/dashboard"
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

	// Example 1: List all accessible dashboards
	fmt.Println("=== Listing All Dashboards ===")
	dashboards, err := client.Dashboard.List(ctx, &dashboard.ListOptions{
		MaxResults: 50,
	})
	if err != nil {
		log.Fatalf("Failed to list dashboards: %v", err)
	}

	fmt.Printf("Found %d dashboards:\n", len(dashboards))
	for _, dash := range dashboards {
		fmt.Printf("- %s (ID: %s)\n", dash.Name, dash.ID)
		if dash.View != "" {
			fmt.Printf("  View: %s\n", dash.View)
		}
		if dash.Owner != nil {
			fmt.Printf("  Owner: %s\n", dash.Owner.DisplayName)
		}
		if len(dash.SharePermissions) > 0 {
			fmt.Printf("  Shared with: %d permissions\n", len(dash.SharePermissions))
		}
	}

	// Example 2: Get specific dashboard details
	if len(dashboards) > 0 {
		dashboardID := dashboards[0].ID
		fmt.Printf("\n=== Getting Dashboard Details (ID: %s) ===\n", dashboardID)

		dashDetails, err := client.Dashboard.Get(ctx, dashboardID)
		if err != nil {
			log.Printf("Failed to get dashboard: %v", err)
		} else {
			fmt.Printf("ID: %s\n", dashDetails.ID)
			fmt.Printf("Name: %s\n", dashDetails.Name)
			if dashDetails.Description != "" {
				fmt.Printf("Description: %s\n", dashDetails.Description)
			}
			fmt.Printf("View: %s\n", dashDetails.View)
			fmt.Printf("Favorite: %t\n", dashDetails.IsFavourite)

			if len(dashDetails.SharePermissions) > 0 {
				fmt.Println("\nShare Permissions:")
				for _, perm := range dashDetails.SharePermissions {
					fmt.Printf("  - Type: %s\n", perm.Type)
					if perm.Group != nil {
						fmt.Printf("    Group: %s\n", perm.Group.Name)
					}
					if perm.Project != nil {
						fmt.Printf("    Project: %s\n", perm.Project.Key)
					}
					if perm.Role != nil {
						fmt.Printf("    Role: %s\n", perm.Role.Name)
					}
				}
			}
		}
	}

	// Example 3: Create a new dashboard
	fmt.Println("\n=== Creating New Dashboard ===")
	newDash, err := client.Dashboard.Create(ctx, &dashboard.CreateDashboardInput{
		Name:        "API Demo Dashboard",
		Description: "Dashboard created via jirasdk example",
		SharePermissions: []*dashboard.SharePermission{
			{Type: "global"}, // Share with everyone
		},
	})
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	fmt.Printf("Created dashboard: %s (ID: %s)\n", newDash.Name, newDash.ID)
	dashboardID := newDash.ID

	// Example 4: Update dashboard
	fmt.Println("\n=== Updating Dashboard ===")
	updated, err := client.Dashboard.Update(ctx, dashboardID, &dashboard.UpdateDashboardInput{
		Name:        "Updated API Demo Dashboard",
		Description: "Dashboard updated via jirasdk example",
	})
	if err != nil {
		log.Printf("Failed to update dashboard: %v", err)
	} else {
		fmt.Printf("Updated dashboard: %s\n", updated.Name)
		fmt.Printf("New description: %s\n", updated.Description)
	}

	// Example 5: Dashboard Gadget Management
	fmt.Println("\n=== Dashboard Gadget Management ===")

	// List existing gadgets
	fmt.Println("\nListing gadgets...")
	gadgets, err := client.Dashboard.GetGadgets(ctx, dashboardID)
	if err != nil {
		log.Printf("Failed to list gadgets: %v", err)
	} else {
		fmt.Printf("Found %d gadgets on dashboard\n", len(gadgets))
		for _, g := range gadgets {
			fmt.Printf("- ID: %d, Title: %s\n", g.ID, g.Title)
			fmt.Printf("  Module: %s\n", g.ModuleKey)
			if g.Position != nil {
				fmt.Printf("  Position: Row %d, Column %d\n", g.Position.Row, g.Position.Column)
			}
		}
	}

	// Add a gadget (Filter Results gadget)
	fmt.Println("\nAdding a gadget...")
	newGadget, err := client.Dashboard.AddGadget(ctx, dashboardID, &dashboard.DashboardGadget{
		ModuleKey: "com.atlassian.jira.gadgets:filter-results-gadget",
		Title:     "Recent Issues",
		Position: &dashboard.GadgetPosition{
			Row:    0,
			Column: 0,
		},
		Properties: map[string]interface{}{
			"refresh":      "false",
			"isConfigured": "false",
		},
	})
	if err != nil {
		log.Printf("Failed to add gadget: %v", err)
	} else {
		fmt.Printf("Added gadget: %s (ID: %d)\n", newGadget.Title, newGadget.ID)
		gadgetID := newGadget.ID

		// Update gadget position
		fmt.Println("\nUpdating gadget position...")
		updatedGadget, err := client.Dashboard.UpdateGadget(ctx, dashboardID, gadgetID, &dashboard.DashboardGadget{
			Position: &dashboard.GadgetPosition{
				Row:    0,
				Column: 1,
			},
		})
		if err != nil {
			log.Printf("Failed to update gadget: %v", err)
		} else {
			fmt.Printf("Moved gadget to Row %d, Column %d\n",
				updatedGadget.Position.Row, updatedGadget.Position.Column)
		}

		// Remove gadget
		fmt.Println("\nRemoving gadget...")
		err = client.Dashboard.RemoveGadget(ctx, dashboardID, gadgetID)
		if err != nil {
			log.Printf("Failed to remove gadget: %v", err)
		} else {
			fmt.Println("Gadget removed successfully")
		}
	}

	// Example 6: Copy dashboard
	fmt.Println("\n=== Copying Dashboard ===")
	copiedDash, err := client.Dashboard.Copy(ctx, dashboardID, &dashboard.CreateDashboardInput{
		Name:        "Copy of API Demo Dashboard",
		Description: "Copied dashboard",
		SharePermissions: []*dashboard.SharePermission{
			{Type: "private"}, // Private to current user
		},
	})
	if err != nil {
		log.Printf("Failed to copy dashboard: %v", err)
	} else {
		fmt.Printf("Copied dashboard: %s (ID: %s)\n", copiedDash.Name, copiedDash.ID)

		// Clean up: Delete the copied dashboard
		fmt.Println("\nCleaning up: Deleting copied dashboard...")
		err = client.Dashboard.Delete(ctx, copiedDash.ID)
		if err != nil {
			log.Printf("Failed to delete copied dashboard: %v", err)
		} else {
			fmt.Println("Copied dashboard deleted successfully")
		}
	}

	// Example 7: Dashboard Statistics
	fmt.Println("\n=== Dashboard Statistics ===")
	fmt.Printf("Total dashboards accessible: %d\n", len(dashboards))

	// Count by ownership
	ownedCount := 0
	sharedCount := 0
	for _, dash := range dashboards {
		if dash.Owner != nil && dash.Owner.AccountID == "" {
			// Owned by current user (simplified check)
			ownedCount++
		} else {
			sharedCount++
		}
	}

	fmt.Printf("Dashboards: %d total\n", len(dashboards))

	// Count gadgets across all dashboards
	totalGadgets := 0
	for _, dash := range dashboards {
		gadgets, err := client.Dashboard.GetGadgets(ctx, dash.ID)
		if err == nil {
			totalGadgets += len(gadgets)
			if len(gadgets) > 0 {
				fmt.Printf("  %s: %d gadgets\n", dash.Name, len(gadgets))
			}
		}
	}
	fmt.Printf("\nTotal gadgets across all dashboards: %d\n", totalGadgets)

	// Example 8: Search dashboards by name
	fmt.Println("\n=== Searching Dashboards ===")
	searchTerm := "Demo"
	fmt.Printf("Searching for dashboards containing '%s':\n", searchTerm)
	for _, dash := range dashboards {
		if strings.Contains(strings.ToLower(dash.Name), strings.ToLower(searchTerm)) {
			fmt.Printf("- %s (ID: %s)\n", dash.Name, dash.ID)
		}
	}

	// Clean up: Delete the demo dashboard
	fmt.Println("\n=== Cleaning Up ===")
	fmt.Println("Deleting demo dashboard...")
	err = client.Dashboard.Delete(ctx, dashboardID)
	if err != nil {
		log.Printf("Failed to delete dashboard: %v", err)
	} else {
		fmt.Println("Demo dashboard deleted successfully")
	}

	fmt.Println("\n=== Dashboard Example Complete ===")
}
