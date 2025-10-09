// Package main demonstrates Filter resource management with the jira-connect library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/filter"
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

	// Example 1: Create a new filter
	fmt.Println("=== Creating a New Filter ===")
	newFilter, err := client.Filter.Create(ctx, &filter.CreateFilterInput{
		Name:        "My Unresolved Bugs",
		Description: "All unresolved bugs assigned to me",
		JQL:         "assignee = currentUser() AND type = Bug AND resolution = Unresolved",
		Favorite:    true,
	})
	if err != nil {
		log.Printf("Failed to create filter: %v", err)
	} else {
		fmt.Printf("Created filter: %s (ID: %s)\n", newFilter.Name, newFilter.ID)
		fmt.Printf("  JQL: %s\n", newFilter.JQL)
		fmt.Printf("  Favorite: %t\n", newFilter.Favorite)
		fmt.Printf("  View URL: %s\n", newFilter.ViewURL)
	}

	// Example 2: List all filters
	fmt.Println("\n=== Listing All Filters ===")
	filters, err := client.Filter.List(ctx, &filter.ListOptions{
		StartAt:    0,
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Failed to list filters: %v", err)
	} else {
		fmt.Printf("Found %d filters:\n", len(filters))
		for i, f := range filters {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, f.Name, f.ID)
			fmt.Printf("   JQL: %s\n", f.JQL)
			if f.Favorite {
				fmt.Printf("   ⭐ Favorite\n")
			}
		}
	}

	// Example 3: Get favorite filters
	fmt.Println("\n=== Getting Favorite Filters ===")
	favorites, err := client.Filter.GetFavorites(ctx)
	if err != nil {
		log.Printf("Failed to get favorites: %v", err)
	} else {
		fmt.Printf("Found %d favorite filters:\n", len(favorites))
		for i, f := range favorites {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, f.Name, f.ID)
		}
	}

	// Example 4: Get my filters
	fmt.Println("\n=== Getting My Filters ===")
	myFilters, err := client.Filter.GetMyFilters(ctx, []string{"sharePermissions"}, true)
	if err != nil {
		log.Printf("Failed to get my filters: %v", err)
	} else {
		fmt.Printf("Found %d filters owned by me:\n", len(myFilters))
		for i, f := range myFilters {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, f.Name, f.ID)
			if len(f.SharePermissions) > 0 {
				fmt.Printf("   Shared with: %d permissions\n", len(f.SharePermissions))
			}
		}
	}

	// Example 5: Get a specific filter
	if len(filters) > 0 {
		filterID := filters[0].ID
		fmt.Printf("\n=== Getting Filter %s ===\n", filterID)
		specificFilter, err := client.Filter.Get(ctx, filterID, []string{"sharePermissions", "subscriptions"})
		if err != nil {
			log.Printf("Failed to get filter: %v", err)
		} else {
			fmt.Printf("Filter: %s\n", specificFilter.Name)
			fmt.Printf("  Description: %s\n", specificFilter.Description)
			fmt.Printf("  JQL: %s\n", specificFilter.JQL)
			if specificFilter.Owner != nil {
				fmt.Printf("  Owner: %s\n", specificFilter.Owner.DisplayName)
			}
			fmt.Printf("  Favorited by %d users\n", specificFilter.FavouritedCount)
		}
	}

	// Example 6: Update a filter
	if newFilter != nil {
		fmt.Printf("\n=== Updating Filter %s ===\n", newFilter.ID)
		updatedFilter, err := client.Filter.Update(ctx, newFilter.ID, &filter.UpdateFilterInput{
			Name:        "My Critical Bugs",
			Description: "All critical bugs assigned to me",
			JQL:         "assignee = currentUser() AND type = Bug AND priority = Highest",
		})
		if err != nil {
			log.Printf("Failed to update filter: %v", err)
		} else {
			fmt.Printf("Updated filter: %s\n", updatedFilter.Name)
			fmt.Printf("  New JQL: %s\n", updatedFilter.JQL)
		}
	}

	// Example 7: Set filter as favorite
	if len(filters) > 0 && !filters[0].Favorite {
		filterID := filters[0].ID
		fmt.Printf("\n=== Setting Filter %s as Favorite ===\n", filterID)
		favFilter, err := client.Filter.SetFavorite(ctx, filterID)
		if err != nil {
			log.Printf("Failed to set favorite: %v", err)
		} else {
			fmt.Printf("Filter '%s' is now a favorite: %t\n", favFilter.Name, favFilter.Favorite)
		}
	}

	// Example 8: Get default share scope
	fmt.Println("\n=== Getting Default Share Scope ===")
	scope, err := client.Filter.GetDefaultShareScope(ctx)
	if err != nil {
		log.Printf("Failed to get default share scope: %v", err)
	} else {
		fmt.Printf("Default share scope: %s\n", scope)
	}

	// Example 9: Add share permission
	if newFilter != nil {
		fmt.Printf("\n=== Adding Share Permission to Filter %s ===\n", newFilter.ID)
		permissions, err := client.Filter.AddSharePermission(ctx, newFilter.ID, &filter.Permission{
			Type: "group",
			Group: &filter.Group{
				Name: "jira-users",
			},
		})
		if err != nil {
			log.Printf("Failed to add share permission: %v", err)
		} else if len(permissions) > 0 {
			fmt.Printf("Added %s permission (ID: %d)\n", permissions[0].Type, permissions[0].ID)
			if permissions[0].Group != nil {
				fmt.Printf("  Group: %s\n", permissions[0].Group.Name)
			}
		}
	}

	// Example 10: Search filters
	fmt.Println("\n=== Searching Filters ===")
	searchResults, err := client.Filter.List(ctx, &filter.ListOptions{
		StartAt:    0,
		MaxResults: 5,
		OrderBy:    "name",
	})
	if err != nil {
		log.Printf("Failed to search filters: %v", err)
	} else {
		fmt.Printf("Found %d filters:\n", len(searchResults))
		for i, f := range searchResults {
			fmt.Printf("%d. %s\n", i+1, f.Name)
		}
	}

	// Example 11: Remove favorite
	if newFilter != nil && newFilter.Favorite {
		fmt.Printf("\n=== Removing Filter %s from Favorites ===\n", newFilter.ID)
		unfavFilter, err := client.Filter.RemoveFavorite(ctx, newFilter.ID)
		if err != nil {
			log.Printf("Failed to remove favorite: %v", err)
		} else {
			fmt.Printf("Filter '%s' is no longer a favorite: %t\n", unfavFilter.Name, !unfavFilter.Favorite)
		}
	}

	// Example 12: Delete filter (cleanup)
	if newFilter != nil {
		fmt.Printf("\n=== Deleting Filter %s ===\n", newFilter.ID)
		err := client.Filter.Delete(ctx, newFilter.ID)
		if err != nil {
			log.Printf("Failed to delete filter: %v", err)
		} else {
			fmt.Printf("Successfully deleted filter '%s'\n", newFilter.Name)
		}
	}

	fmt.Println("\n=== Filter Management Example Complete ===")
}
