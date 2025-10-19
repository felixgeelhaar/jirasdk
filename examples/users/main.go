// Package main demonstrates user search and retrieval operations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/user"
)

func main() {
	// Get configuration from environment
	baseURL := os.Getenv("JIRA_BASE_URL")
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if baseURL == "" || email == "" || apiToken == "" {
		log.Fatal("Please set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables")
	}

	// Create client with API token authentication
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAPIToken(email, apiToken),
		jira.WithTimeout(30*time.Second),
		jira.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== User Search Operations ===")
	fmt.Println()

	// 1. Get current user
	fmt.Println("1. Getting current user...")
	currentUser, err := client.User.GetMyself(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	fmt.Printf("   Logged in as: %s (%s)\n", currentUser.DisplayName, currentUser.EmailAddress)
	fmt.Printf("   Account ID: %s\n\n", currentUser.AccountID)

	// 2. Find users by name (convenience method)
	searchName := "john" // Replace with a name to search for
	fmt.Printf("2. Finding users matching '%s' (using FindByName)...\n", searchName)
	users, err := client.User.FindByName(ctx, searchName, 10)
	if err != nil {
		log.Printf("   Warning: Could not find users: %v\n", err)
	} else {
		fmt.Printf("   Found %d users:\n", len(users))
		for i, u := range users {
			status := "Active"
			if !u.Active {
				status = "Inactive"
			}
			fmt.Printf("   %d. %s (%s) - %s\n", i+1, u.DisplayName, u.EmailAddress, status)
			fmt.Printf("      Account ID: %s\n", u.AccountID)
		}
		fmt.Println()
	}

	// 3. Advanced search with options
	fmt.Println("3. Advanced search (using Search with options)...")
	advancedUsers, err := client.User.Search(ctx, &user.SearchOptions{
		Query:         "admin",
		MaxResults:    5,
		IncludeActive: true,
	})
	if err != nil {
		log.Printf("   Warning: Could not search users: %v\n", err)
	} else {
		fmt.Printf("   Found %d active users matching 'admin':\n", len(advancedUsers))
		for i, u := range advancedUsers {
			fmt.Printf("   %d. %s (%s)\n", i+1, u.DisplayName, u.EmailAddress)
		}
		fmt.Println()
	}

	// 4. Find assignable users for a project
	projectKey := "PROJ" // Replace with your project key
	fmt.Printf("4. Finding assignable users for project %s...\n", projectKey)
	assignableUsers, err := client.User.FindAssignableUsers(ctx, &user.FindAssignableOptions{
		Project:    projectKey,
		Query:      "", // Empty query returns all assignable users
		MaxResults: 10,
	})
	if err != nil {
		log.Printf("   Warning: Could not find assignable users: %v\n", err)
	} else {
		fmt.Printf("   Found %d assignable users:\n", len(assignableUsers))
		for i, u := range assignableUsers {
			fmt.Printf("   %d. %s (%s)\n", i+1, u.DisplayName, u.EmailAddress)
		}
		fmt.Println()
	}

	// 5. Get user by account ID
	if len(users) > 0 {
		accountID := users[0].AccountID
		fmt.Printf("5. Getting user details by account ID (%s)...\n", accountID)
		userDetails, err := client.User.Get(ctx, accountID, &user.GetOptions{
			Expand: []string{"groups", "applicationRoles"},
		})
		if err != nil {
			log.Printf("   Warning: Could not get user details: %v\n", err)
		} else {
			fmt.Printf("   Display Name: %s\n", userDetails.DisplayName)
			fmt.Printf("   Email: %s\n", userDetails.EmailAddress)
			fmt.Printf("   Account Type: %s\n", userDetails.AccountType)
			fmt.Printf("   Time Zone: %s\n", userDetails.TimeZone)
			if userDetails.Groups != nil {
				fmt.Printf("   Groups: %d\n", userDetails.Groups.Size)
			}
			fmt.Println()
		}
	}

	// 6. Bulk get users by account IDs
	if len(users) >= 2 {
		accountIDs := []string{users[0].AccountID, users[1].AccountID}
		fmt.Println("6. Bulk getting users by account IDs...")
		bulkUsers, err := client.User.BulkGet(ctx, &user.BulkGetOptions{
			AccountIDs: accountIDs,
			MaxResults: 10,
		})
		if err != nil {
			log.Printf("   Warning: Could not bulk get users: %v\n", err)
		} else {
			fmt.Printf("   Retrieved %d users:\n", len(bulkUsers))
			for i, u := range bulkUsers {
				fmt.Printf("   %d. %s (%s)\n", i+1, u.DisplayName, u.EmailAddress)
			}
			fmt.Println()
		}
	}

	fmt.Println("=== User operations completed successfully! ===")
}
