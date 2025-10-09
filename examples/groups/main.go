package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/group"
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

	// Example 1: Search for groups
	fmt.Println("=== Searching for Groups ===")
	groups, err := client.Group.Find(ctx, &group.FindOptions{
		Query:      "",
		MaxResults: 50,
	})
	if err != nil {
		log.Fatalf("Failed to search groups: %v", err)
	}

	fmt.Printf("Found %d groups:\n", len(groups))
	for _, grp := range groups {
		fmt.Printf("- %s", grp.Name)
		if grp.GroupID != "" {
			fmt.Printf(" (ID: %s)", grp.GroupID)
		}
		fmt.Println()
	}

	// Example 2: Search for specific groups
	fmt.Println("\n=== Searching for Admin Groups ===")
	adminGroups, err := client.Group.Find(ctx, &group.FindOptions{
		Query:      "admin",
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("Failed to search admin groups: %v", err)
	} else {
		fmt.Printf("Found %d admin-related groups:\n", len(adminGroups))
		for _, grp := range adminGroups {
			fmt.Printf("- %s\n", grp.Name)
		}
	}

	// Example 3: Get group details with expansion
	if len(groups) > 0 {
		groupName := groups[0].Name
		fmt.Printf("\n=== Getting Group Details: %s ===\n", groupName)

		grpDetails, err := client.Group.Get(ctx, &group.GetOptions{
			GroupName: groupName,
			Expand:    []string{"users"},
		})
		if err != nil {
			log.Printf("Failed to get group details: %v", err)
		} else {
			fmt.Printf("Name: %s\n", grpDetails.Name)
			if grpDetails.GroupID != "" {
				fmt.Printf("Group ID: %s\n", grpDetails.GroupID)
			}

			if grpDetails.Users != nil {
				fmt.Printf("\nMembers: %d\n", grpDetails.Users.Size)
				if len(grpDetails.Users.Items) > 0 {
					fmt.Println("Sample members:")
					for i, user := range grpDetails.Users.Items {
						if i >= 5 {
							fmt.Printf("  ... and %d more\n", grpDetails.Users.Size-5)
							break
						}
						fmt.Printf("  - %s", user.DisplayName)
						if user.EmailAddress != "" {
							fmt.Printf(" (%s)", user.EmailAddress)
						}
						fmt.Println()
					}
				}
			}
		}
	}

	// Example 4: Create a new group
	fmt.Println("\n=== Creating New Group ===")
	newGroup, err := client.Group.Create(ctx, &group.CreateGroupInput{
		Name: "jirasdk-demo-group",
	})
	if err != nil {
		log.Printf("Failed to create group: %v", err)
		// If group already exists, try to get it
		fmt.Println("Group might already exist, attempting to retrieve...")
		newGroup, err = client.Group.Get(ctx, &group.GetOptions{
			GroupName: "jirasdk-demo-group",
		})
		if err != nil {
			log.Fatalf("Failed to get existing group: %v", err)
		}
	}

	fmt.Printf("Group: %s\n", newGroup.Name)
	groupName := newGroup.Name

	// Example 5: Group Member Management
	fmt.Println("\n=== Group Member Management ===")

	// Get current user to add to group
	currentUser, err := client.User.GetMyself(ctx)
	if err != nil {
		log.Printf("Failed to get current user: %v", err)
	} else {
		fmt.Printf("Current user: %s\n", currentUser.DisplayName)

		// Add user to group
		fmt.Println("\nAdding current user to demo group...")
		updatedGroup, err := client.Group.AddUser(ctx, &group.AddUserOptions{
			GroupName: groupName,
			AccountID: currentUser.AccountID,
		})
		if err != nil {
			log.Printf("Failed to add user to group: %v", err)
		} else {
			fmt.Printf("Added %s to %s\n", currentUser.DisplayName, updatedGroup.Name)

			// List group members
			fmt.Println("\nListing group members...")
			members, err := client.Group.GetMembers(ctx, &group.GetMembersOptions{
				GroupName:  groupName,
				MaxResults: 50,
			})
			if err != nil {
				log.Printf("Failed to list members: %v", err)
			} else {
				fmt.Printf("Group has %d members:\n", len(members))
				for _, member := range members {
					fmt.Printf("  - %s", member.DisplayName)
					if member.EmailAddress != "" {
						fmt.Printf(" (%s)", member.EmailAddress)
					}
					fmt.Println()
				}
			}

			// Remove user from group
			fmt.Println("\nRemoving user from group...")
			err = client.Group.RemoveUser(ctx, &group.RemoveUserOptions{
				GroupName: groupName,
				AccountID: currentUser.AccountID,
			})
			if err != nil {
				log.Printf("Failed to remove user from group: %v", err)
			} else {
				fmt.Printf("Removed %s from %s\n", currentUser.DisplayName, groupName)
			}
		}
	}

	// Example 6: Bulk get groups
	if len(groups) >= 3 {
		fmt.Println("\n=== Bulk Get Groups ===")
		groupNames := []string{groups[0].Name, groups[1].Name, groups[2].Name}
		fmt.Printf("Fetching details for %d groups in bulk...\n", len(groupNames))

		bulkGroups, err := client.Group.BulkGet(ctx, &group.BulkOptions{
			GroupNames: groupNames,
			MaxResults: 50,
		})
		if err != nil {
			log.Printf("Failed to bulk get groups: %v", err)
		} else {
			fmt.Printf("Retrieved %d groups:\n", len(bulkGroups))
			for _, grp := range bulkGroups {
				fmt.Printf("  - %s\n", grp.Name)
			}
		}
	}

	// Example 7: Group Statistics
	fmt.Println("\n=== Group Statistics ===")
	fmt.Printf("Total groups found: %d\n", len(groups))

	// Count members across all groups (sample)
	fmt.Println("\nGroup membership overview:")
	totalMembers := 0
	groupsWithMembers := 0

	for i, grp := range groups {
		if i >= 10 {
			fmt.Printf("... and %d more groups\n", len(groups)-10)
			break
		}

		members, err := client.Group.GetMembers(ctx, &group.GetMembersOptions{
			GroupName:  grp.Name,
			MaxResults: 1000,
		})
		if err == nil {
			memberCount := len(members)
			if memberCount > 0 {
				fmt.Printf("  %s: %d members\n", grp.Name, memberCount)
				totalMembers += memberCount
				groupsWithMembers++
			}
		}
	}

	if groupsWithMembers > 0 {
		avgMembers := float64(totalMembers) / float64(groupsWithMembers)
		fmt.Printf("\nAverage members per group: %.1f\n", avgMembers)
	}

	// Example 8: Search and filter groups
	fmt.Println("\n=== Advanced Group Search ===")

	// Search for developer groups
	devGroups, err := client.Group.Find(ctx, &group.FindOptions{
		Query:      "dev",
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to search developer groups: %v", err)
	} else {
		fmt.Printf("Developer-related groups: %d\n", len(devGroups))
		for _, grp := range devGroups {
			fmt.Printf("  - %s\n", grp.Name)
		}
	}

	// Search for team groups
	teamGroups, err := client.Group.Find(ctx, &group.FindOptions{
		Query:      "team",
		MaxResults: 50,
	})
	if err != nil {
		log.Printf("Failed to search team groups: %v", err)
	} else {
		fmt.Printf("\nTeam-related groups: %d\n", len(teamGroups))
		for _, grp := range teamGroups {
			fmt.Printf("  - %s\n", grp.Name)
		}
	}

	// Clean up: Delete the demo group
	fmt.Println("\n=== Cleaning Up ===")
	fmt.Println("Deleting demo group...")
	err = client.Group.Delete(ctx, &group.DeleteOptions{
		GroupName: groupName,
	})
	if err != nil {
		log.Printf("Failed to delete group: %v", err)
	} else {
		fmt.Printf("Demo group '%s' deleted successfully\n", groupName)
	}

	fmt.Println("\n=== Group Example Complete ===")
}
