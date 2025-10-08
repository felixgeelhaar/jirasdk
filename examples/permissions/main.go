package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/auth"
	"github.com/felixgeelhaar/jirasdk/core/permission"
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
		jira.WithAuth(auth.NewBasicAuth(email, apiToken)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Get all available permissions
	fmt.Println("=== Getting All Available Permissions ===")
	allPermissions, err := client.Permission.GetAllPermissions(ctx)
	if err != nil {
		log.Fatalf("Failed to get permissions: %v", err)
	}

	fmt.Printf("Found %d permissions in Jira:\n", len(allPermissions))
	for _, perm := range allPermissions {
		fmt.Printf("- %s (%s)\n", perm.Name, perm.Key)
		if perm.Description != "" {
			fmt.Printf("  Description: %s\n", perm.Description)
		}
	}

	// Example 2: Check current user's permissions
	fmt.Println("\n=== Checking My Permissions ===")
	myPerms, err := client.Permission.GetMyPermissions(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get my permissions: %v", err)
	}

	fmt.Println("\nMy Global Permissions:")
	for key, status := range myPerms.Permissions {
		if status.HavePermission {
			fmt.Printf("✓ %s - %s\n", key, status.Name)
		}
	}

	// Example 3: Check permissions for a specific project
	projectKey := os.Getenv("JIRA_PROJECT_KEY")
	if projectKey != "" {
		fmt.Printf("\n=== Checking Permissions for Project %s ===\n", projectKey)
		projectPerms, err := client.Permission.GetMyPermissions(ctx, &permission.MyPermissionsOptions{
			ProjectKey:  projectKey,
			Permissions: "BROWSE_PROJECTS,CREATE_ISSUES,EDIT_ISSUES,DELETE_ISSUES,ASSIGN_ISSUES",
		})
		if err != nil {
			log.Printf("Failed to get project permissions: %v", err)
		} else {
			fmt.Println("\nProject Permissions:")
			for key, status := range projectPerms.Permissions {
				symbol := "✓"
				if !status.HavePermission {
					symbol = "✗"
				}
				fmt.Printf("%s %s - %s\n", symbol, key, status.Name)
			}
		}
	}

	// Example 4: List all permission schemes
	fmt.Println("\n=== Listing Permission Schemes ===")
	schemes, err := client.Permission.ListPermissionSchemes(ctx, nil)
	if err != nil {
		log.Printf("Failed to list permission schemes: %v", err)
	} else {
		fmt.Printf("Found %d permission schemes:\n", len(schemes))
		for _, scheme := range schemes {
			fmt.Printf("- %s (ID: %d)\n", scheme.Name, scheme.ID)
			if scheme.Description != "" {
				fmt.Printf("  Description: %s\n", scheme.Description)
			}
		}

		// Get detailed scheme information for the first scheme
		if len(schemes) > 0 {
			schemeID := schemes[0].ID
			fmt.Printf("\n=== Getting Detailed Scheme Information (ID: %d) ===\n", schemeID)

			detailedScheme, err := client.Permission.GetPermissionScheme(ctx, schemeID, &permission.GetPermissionSchemeOptions{
				Expand: []string{"permissions", "user", "group", "projectRole"},
			})
			if err != nil {
				log.Printf("Failed to get scheme details: %v", err)
			} else {
				fmt.Printf("Scheme: %s\n", detailedScheme.Name)
				fmt.Printf("Description: %s\n", detailedScheme.Description)

				if detailedScheme.Permissions != nil {
					fmt.Printf("\nPermission Grants (%d):\n", len(detailedScheme.Permissions))
					for _, grant := range detailedScheme.Permissions {
						fmt.Printf("  - Permission: %s\n", grant.Permission)
						if grant.Holder != nil {
							fmt.Printf("    Holder Type: %s\n", grant.Holder.Type)
						}
					}
				}
			}
		}
	}

	// Example 5: Create a new permission scheme
	fmt.Println("\n=== Creating New Permission Scheme ===")
	newScheme, err := client.Permission.CreatePermissionScheme(ctx, &permission.CreatePermissionSchemeInput{
		Name:        "Demo Permission Scheme",
		Description: "Temporary scheme for demonstration purposes",
	})
	if err != nil {
		log.Printf("Failed to create permission scheme: %v", err)
	} else {
		fmt.Printf("Created scheme: %s (ID: %d)\n", newScheme.Name, newScheme.ID)

		// Update the scheme
		fmt.Println("\n=== Updating Permission Scheme ===")
		updatedScheme, err := client.Permission.UpdatePermissionScheme(ctx, newScheme.ID, &permission.UpdatePermissionSchemeInput{
			Name:        "Updated Demo Scheme",
			Description: "Updated description for demonstration",
		})
		if err != nil {
			log.Printf("Failed to update permission scheme: %v", err)
		} else {
			fmt.Printf("Updated scheme: %s\n", updatedScheme.Name)
			fmt.Printf("New description: %s\n", updatedScheme.Description)
		}

		// Clean up: Delete the scheme
		fmt.Println("\n=== Cleaning Up: Deleting Demo Scheme ===")
		err = client.Permission.DeletePermissionScheme(ctx, newScheme.ID)
		if err != nil {
			log.Printf("Failed to delete permission scheme: %v", err)
		} else {
			fmt.Println("Successfully deleted demo scheme")
		}
	}

	// Example 6: Project role management
	if projectKey != "" {
		fmt.Printf("\n=== Managing Project Roles for %s ===\n", projectKey)

		// Get all roles for the project
		roles, err := client.Permission.GetProjectRoles(ctx, projectKey)
		if err != nil {
			log.Printf("Failed to get project roles: %v", err)
		} else {
			fmt.Printf("Found %d roles:\n", len(roles))
			for name, url := range roles {
				fmt.Printf("- %s: %s\n", name, url)
			}
		}

		// Get specific role details
		fmt.Println("\n=== Getting Role Details ===")
		// Note: You need to know a valid role ID for your project
		// Common role IDs: 10002 (Administrators), 10001 (Developers)
		roleID := int64(10002) // Administrators role

		roleDetails, err := client.Permission.GetProjectRole(ctx, projectKey, roleID)
		if err != nil {
			log.Printf("Failed to get role details: %v", err)
		} else {
			fmt.Printf("Role: %s (ID: %d)\n", roleDetails.Name, roleDetails.ID)
			if roleDetails.Description != "" {
				fmt.Printf("Description: %s\n", roleDetails.Description)
			}

			if roleDetails.Actors != nil {
				fmt.Printf("\nActors (%d):\n", len(roleDetails.Actors))
				for _, actor := range roleDetails.Actors {
					fmt.Printf("  - Type: %s, Display: %s\n", actor.Type, actor.DisplayName)
				}
			}
		}

		// Demonstrate adding/removing actors (commented out to avoid modifying real data)
		/*
		// Add a user to a role
		fmt.Println("\n=== Adding User to Role ===")
		updatedRole, err := client.Permission.AddActorsToProjectRole(ctx, projectKey, roleID, &permission.AddActorInput{
			User: []string{"accountId123"},
		})
		if err != nil {
			log.Printf("Failed to add user to role: %v", err)
		} else {
			fmt.Printf("Added user to role: %s\n", updatedRole.Name)
		}

		// Add a group to a role
		fmt.Println("\n=== Adding Group to Role ===")
		updatedRole, err = client.Permission.AddActorsToProjectRole(ctx, projectKey, roleID, &permission.AddActorInput{
			Group: []string{"developers"},
		})
		if err != nil {
			log.Printf("Failed to add group to role: %v", err)
		} else {
			fmt.Printf("Added group to role: %s\n", updatedRole.Name)
		}

		// Remove a user from a role
		fmt.Println("\n=== Removing User from Role ===")
		err = client.Permission.RemoveActorFromProjectRole(ctx, projectKey, roleID, "user", "accountId123")
		if err != nil {
			log.Printf("Failed to remove user from role: %v", err)
		} else {
			fmt.Println("Successfully removed user from role")
		}

		// Remove a group from a role
		fmt.Println("\n=== Removing Group from Role ===")
		err = client.Permission.RemoveActorFromProjectRole(ctx, projectKey, roleID, "group", "developers")
		if err != nil {
			log.Printf("Failed to remove group from role: %v", err)
		} else {
			fmt.Println("Successfully removed group from role")
		}
		*/
	}

	fmt.Println("\n=== Permission Management Example Complete ===")
}
