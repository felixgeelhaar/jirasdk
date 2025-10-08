package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/felixgeelhaar/jira-connect"
	"github.com/felixgeelhaar/jira-connect/auth"
	"github.com/felixgeelhaar/jira-connect/core/project"
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

	// Example 1: List all projects
	fmt.Println("=== Listing All Projects ===")
	projects, err := client.Project.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}

	fmt.Printf("Found %d projects:\n", len(projects))
	for _, p := range projects {
		fmt.Printf("- %s (%s)\n", p.Name, p.Key)
		if p.Description != "" {
			fmt.Printf("  Description: %s\n", p.Description)
		}
		if p.Lead != nil {
			fmt.Printf("  Lead: %s\n", p.Lead.DisplayName)
		}
	}

	// For the rest of the examples, we'll use the first project
	if len(projects) == 0 {
		log.Fatal("No projects found. Please create a project in Jira first.")
	}
	projectKey := projects[0].Key

	// Example 2: Get specific project details
	fmt.Printf("\n=== Getting Project Details: %s ===\n", projectKey)
	projectDetails, err := client.Project.Get(ctx, projectKey)
	if err != nil {
		log.Fatalf("Failed to get project: %v", err)
	}

	fmt.Printf("ID: %s\n", projectDetails.ID)
	fmt.Printf("Key: %s\n", projectDetails.Key)
	fmt.Printf("Name: %s\n", projectDetails.Name)
	fmt.Printf("Description: %s\n", projectDetails.Description)
	fmt.Printf("Project Type: %s\n", projectDetails.ProjectTypeKey)

	// Example 3: Component Management
	fmt.Printf("\n=== Component Management (%s) ===\n", projectKey)

	// List existing components
	components, err := client.Project.ListProjectComponents(ctx, projectKey)
	if err != nil {
		log.Printf("Failed to list components: %v", err)
	} else {
		fmt.Printf("\nExisting components: %d\n", len(components))
		for _, comp := range components {
			fmt.Printf("- %s (ID: %s)\n", comp.Name, comp.ID)
			if comp.Description != "" {
				fmt.Printf("  Description: %s\n", comp.Description)
			}
		}
	}

	// Create a new component
	fmt.Println("\nCreating new component...")
	newComponent, err := client.Project.CreateComponent(ctx, &project.CreateComponentInput{
		Name:         "API Services",
		Description:  "Backend API services and microservices",
		Project:      projectKey,
		AssigneeType: "PROJECT_DEFAULT",
	})
	if err != nil {
		log.Printf("Failed to create component: %v", err)
	} else {
		fmt.Printf("Created component: %s (ID: %s)\n", newComponent.Name, newComponent.ID)

		// Get the component details
		componentDetails, err := client.Project.GetComponent(ctx, newComponent.ID)
		if err != nil {
			log.Printf("Failed to get component: %v", err)
		} else {
			fmt.Printf("Component details:\n")
			fmt.Printf("  Name: %s\n", componentDetails.Name)
			fmt.Printf("  Description: %s\n", componentDetails.Description)
			if componentDetails.Lead != nil {
				fmt.Printf("  Lead: %s\n", componentDetails.Lead.DisplayName)
			}
		}

		// Update the component
		fmt.Println("\nUpdating component...")
		updatedComponent, err := client.Project.UpdateComponent(ctx, newComponent.ID, &project.UpdateComponentInput{
			Description: "Updated: Backend API services and microservices infrastructure",
		})
		if err != nil {
			log.Printf("Failed to update component: %v", err)
		} else {
			fmt.Printf("Updated component description: %s\n", updatedComponent.Description)
		}

		// Clean up: Delete the component
		fmt.Println("\nCleaning up: Deleting test component...")
		err = client.Project.DeleteComponent(ctx, newComponent.ID)
		if err != nil {
			log.Printf("Failed to delete component: %v", err)
		} else {
			fmt.Println("Component deleted successfully")
		}
	}

	// Example 4: Version Management
	fmt.Printf("\n=== Version Management (%s) ===\n", projectKey)

	// List existing versions
	versions, err := client.Project.ListProjectVersions(ctx, projectKey)
	if err != nil {
		log.Printf("Failed to list versions: %v", err)
	} else {
		fmt.Printf("\nExisting versions: %d\n", len(versions))

		// Group versions by status
		var released, unreleased, archived []*project.Version
		for _, v := range versions {
			if v.Archived {
				archived = append(archived, v)
			} else if v.Released {
				released = append(released, v)
			} else {
				unreleased = append(unreleased, v)
			}
		}

		if len(unreleased) > 0 {
			fmt.Println("\nUnreleased:")
			for _, v := range unreleased {
				fmt.Printf("  - %s (ID: %s)\n", v.Name, v.ID)
				if v.ReleaseDate != "" {
					fmt.Printf("    Release date: %s\n", v.ReleaseDate)
				}
			}
		}

		if len(released) > 0 {
			fmt.Println("\nReleased:")
			for _, v := range released {
				fmt.Printf("  - %s (ID: %s)\n", v.Name, v.ID)
				if v.ReleaseDate != "" {
					fmt.Printf("    Released: %s\n", v.ReleaseDate)
				}
			}
		}

		if len(archived) > 0 {
			fmt.Println("\nArchived:")
			for _, v := range archived {
				fmt.Printf("  - %s (ID: %s)\n", v.Name, v.ID)
			}
		}
	}

	// Create a new version
	fmt.Println("\nCreating new version...")
	newVersion, err := client.Project.CreateVersion(ctx, &project.CreateVersionInput{
		Name:        "v2.5.0",
		Description: "Sprint 25 release",
		Project:     projectKey,
		StartDate:   "2024-06-01",
		ReleaseDate: "2024-06-30",
		Released:    false,
		Archived:    false,
	})
	if err != nil {
		log.Printf("Failed to create version: %v", err)
	} else {
		fmt.Printf("Created version: %s (ID: %s)\n", newVersion.Name, newVersion.ID)

		// Get the version details
		versionDetails, err := client.Project.GetVersion(ctx, newVersion.ID)
		if err != nil {
			log.Printf("Failed to get version: %v", err)
		} else {
			fmt.Printf("Version details:\n")
			fmt.Printf("  Name: %s\n", versionDetails.Name)
			fmt.Printf("  Description: %s\n", versionDetails.Description)
			fmt.Printf("  Start date: %s\n", versionDetails.StartDate)
			fmt.Printf("  Release date: %s\n", versionDetails.ReleaseDate)
			fmt.Printf("  Released: %t\n", versionDetails.Released)
			fmt.Printf("  Archived: %t\n", versionDetails.Archived)
		}

		// Update the version to mark it as released
		fmt.Println("\nMarking version as released...")
		released := true
		updatedVersion, err := client.Project.UpdateVersion(ctx, newVersion.ID, &project.UpdateVersionInput{
			Released: &released,
		})
		if err != nil {
			log.Printf("Failed to update version: %v", err)
		} else {
			fmt.Printf("Version %s is now released: %t\n", updatedVersion.Name, updatedVersion.Released)
		}

		// Clean up: Delete the version
		fmt.Println("\nCleaning up: Deleting test version...")
		err = client.Project.DeleteVersion(ctx, newVersion.ID)
		if err != nil {
			log.Printf("Failed to delete version: %v", err)
		} else {
			fmt.Println("Version deleted successfully")
		}
	}

	// Example 5: Create a new project
	fmt.Println("\n=== Creating New Project ===")
	newProject, err := client.Project.Create(ctx, &project.CreateInput{
		Key:            "DEMO",
		Name:           "Demo Project",
		Description:    "A demonstration project created via API",
		ProjectTypeKey: "software",
		LeadAccountID:  projectDetails.Lead.AccountID, // Use current project lead
	})
	if err != nil {
		log.Printf("Failed to create project: %v", err)
	} else {
		fmt.Printf("Created project: %s (%s)\n", newProject.Name, newProject.Key)

		// Update the project
		fmt.Println("\nUpdating project...")
		_, err = client.Project.Update(ctx, newProject.Key, &project.UpdateInput{
			Name:        "Demo Project - Updated",
			Description: "Updated description for demo project",
		})
		if err != nil {
			log.Printf("Failed to update project: %v", err)
		} else {
			fmt.Println("Project updated successfully")
		}

		// Delete the project
		fmt.Println("\nCleaning up: Deleting demo project...")
		err = client.Project.Delete(ctx, newProject.Key)
		if err != nil {
			log.Printf("Failed to delete project: %v", err)
		} else {
			fmt.Println("Project deleted successfully")
		}
	}

	// Example 6: Archive and restore project
	fmt.Println("\n=== Archive and Restore Operations ===")

	// Note: Be careful with these operations on real projects
	// We'll demonstrate with a test project if you have one
	testProjectKey := "TEST" // Replace with your test project key

	fmt.Printf("Archiving project %s...\n", testProjectKey)
	err = client.Project.Archive(ctx, testProjectKey)
	if err != nil {
		log.Printf("Failed to archive project (this is expected if project doesn't exist or is already archived): %v", err)
	} else {
		fmt.Println("Project archived successfully")

		// Restore the project
		fmt.Printf("\nRestoring project %s...\n", testProjectKey)
		err = client.Project.Restore(ctx, testProjectKey)
		if err != nil {
			log.Printf("Failed to restore project: %v", err)
		} else {
			fmt.Println("Project restored successfully")
		}
	}

	// Example 7: Project analysis
	fmt.Println("\n=== Project Analysis ===")
	for _, p := range projects {
		details, err := client.Project.Get(ctx, p.Key)
		if err != nil {
			continue
		}

		fmt.Printf("\n%s (%s):\n", details.Name, details.Key)
		fmt.Printf("  Components: %d\n", len(details.Components))
		fmt.Printf("  Versions: %d\n", len(details.Versions))
		fmt.Printf("  Issue Types: %d\n", len(details.IssueTypes))

		// Count released vs unreleased versions
		var releasedCount, unreleasedCount int
		for _, v := range details.Versions {
			if v.Released {
				releasedCount++
			} else {
				unreleasedCount++
			}
		}
		fmt.Printf("  Released versions: %d\n", releasedCount)
		fmt.Printf("  Unreleased versions: %d\n", unreleasedCount)

		// Show project structure
		if len(details.Components) > 0 {
			fmt.Println("  Component structure:")
			for _, comp := range details.Components {
				fmt.Printf("    - %s", comp.Name)
				if comp.Lead != nil {
					fmt.Printf(" (Lead: %s)", comp.Lead.DisplayName)
				}
				fmt.Println()
			}
		}
	}

	fmt.Println("\n=== Projects Example Complete ===")
}
