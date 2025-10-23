// Package main demonstrates version and resolution management with the jirasdk library.
//
// This example shows how to:
//   - Create issues with version fields (AffectsVersions, FixVersions)
//   - Set and retrieve resolution information
//   - Use safe accessor methods to prevent nil pointer panics
//   - Work with version references (by Name or by ID)
//
// IMPORTANT TYPE USAGE PATTERNS:
//
// 1. VERSION REFERENCES - Use Name for simplicity, ID for precision:
//    - By Name (recommended for most cases):
//      {Name: "1.0.0"}
//    - By ID (when you have the version ID from API):
//      {ID: "10001"}
//    - Both (ID takes precedence):
//      {ID: "10001", Name: "1.0.0"}
//
// 2. RESOLUTION REFERENCES - Same pattern as versions:
//    - By Name (most common):
//      {Name: "Done"}
//    - By ID:
//      {ID: "1"}
//
// 3. SAFE ACCESSORS - Always use these to avoid nil pointer panics:
//    - issue.GetAffectsVersions() - returns []*project.Version (never nil)
//    - issue.GetFixVersions() - returns []*project.Version (never nil)
//    - issue.GetResolutionName() - returns string (empty if unresolved)
//    - issue.GetResolution() - returns *resolution.Resolution (may be nil)
//
// 4. COMMON MISTAKES TO AVOID:
//    - ❌ Don't access issue.Fields.FixVersions directly (may cause panic)
//    - ✅ Use issue.GetFixVersions() instead
//    - ❌ Don't assume resolution is always set
//    - ✅ Check if GetResolutionName() returns empty string
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/core/project"
)

func main() {
	// Create client from environment variables
	// Required: JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN, JIRA_PROJECT_KEY
	client, err := jira.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	projectKey := os.Getenv("JIRA_PROJECT_KEY")
	if projectKey == "" {
		log.Fatal("JIRA_PROJECT_KEY environment variable is required")
	}

	fmt.Println("=== Version and Resolution Management Demo ===")
	fmt.Println("This example demonstrates proper type usage for versions and resolutions")
	fmt.Println()

	// Example 1: Create a bug with affected versions
	// ================================================
	// AffectsVersions tracks which versions have the issue.
	// PATTERN: Use slice of *project.Version with Name field
	fmt.Println("1. Creating a bug with affected versions...")
	fmt.Println("   Pattern: AffectsVersions: []*project.Version{{Name: \"1.0.0\"}}")

	bug, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			// Required fields for any issue
			Project:   &issue.Project{Key: projectKey},
			Summary:   "Critical bug in authentication",
			IssueType: &issue.IssueType{Name: "Bug"},
			Priority:  &issue.Priority{Name: "High"},

			// AffectsVersions: Versions where this bug exists
			// Use Name field to reference versions (SDK will resolve to ID)
			AffectsVersions: []*project.Version{
				{Name: "1.0.0"}, // Reference by version name
				{Name: "1.1.0"}, // Multiple versions can be affected
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create bug: %v", err)
	}
	fmt.Printf("   ✓ Created bug: %s\n", bug.Key)
	fmt.Println()

	// Example 2: Create a story with fix versions
	// ==============================================
	// FixVersions tracks which versions will contain the fix/feature.
	// PATTERN: Same as AffectsVersions - slice of *project.Version
	fmt.Println("2. Creating a story with fix versions...")
	fmt.Println("   Pattern: FixVersions: []*project.Version{{Name: \"2.0.0\"}}")

	story, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project:   &issue.Project{Key: projectKey},
			Summary:   "Implement user profile feature",
			IssueType: &issue.IssueType{Name: "Story"},
			Priority:  &issue.Priority{Name: "Medium"},

			// FixVersions: Versions that will include this feature
			FixVersions: []*project.Version{
				{Name: "2.0.0"}, // Will be released in version 2.0.0
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create story: %v", err)
	}
	fmt.Printf("   ✓ Created story: %s\n", story.Key)
	fmt.Println()

	// Example 3: Retrieve issue and display version information
	// ===========================================================
	// IMPORTANT: Always use safe accessor methods to read version data
	// ❌ NEVER: issue.Fields.FixVersions (can cause nil panic)
	// ✅ ALWAYS: issue.GetFixVersions() (safe, returns empty slice, never nil)
	fmt.Println("3. Retrieving issue and displaying version information...")
	fmt.Println("   Pattern: Use GetAffectsVersions() and GetFixVersions()")

	retrieved, err := client.Issue.Get(ctx, bug.Key, &issue.GetOptions{
		Fields: []string{"summary", "versions", "fixVersions", "status", "resolution"},
	})
	if err != nil {
		log.Printf("   Warning: Failed to get issue: %v", err)
	} else {
		fmt.Printf("   Issue: %s\n", retrieved.Key)
		fmt.Printf("   Summary: %s\n", retrieved.GetSummary())

		// SAFE ACCESSOR PATTERN: GetAffectsVersions()
		// Returns []*project.Version (never nil, but may be empty slice)
		affectsVersions := retrieved.GetAffectsVersions()
		if len(affectsVersions) > 0 {
			fmt.Printf("   Affects Versions:\n")
			for _, version := range affectsVersions {
				// Each version has Name and ID fields
				fmt.Printf("     - %s", version.Name)
				if version.ID != "" {
					fmt.Printf(" (ID: %s)", version.ID)
				}
				fmt.Println()
			}
		} else {
			fmt.Println("   Affects Versions: None set")
		}

		// SAFE ACCESSOR PATTERN: GetFixVersions()
		// Same behavior as GetAffectsVersions()
		fixVersions := retrieved.GetFixVersions()
		if len(fixVersions) > 0 {
			fmt.Printf("   Fix Versions:\n")
			for _, version := range fixVersions {
				fmt.Printf("     - %s", version.Name)
				if version.ID != "" {
					fmt.Printf(" (ID: %s)", version.ID)
				}
				fmt.Println()
			}
		} else {
			fmt.Println("   Fix Versions: None set")
		}
	}
	fmt.Println()

	// Example 4: Update issue to add fix versions
	// ==============================================
	// When updating, use map[string]interface{} for Fields
	// PATTERN for versions: []map[string]string with "name" or "id" keys
	fmt.Println("4. Updating bug to add fix versions...")
	fmt.Println("   Pattern: Fields: map[string]interface{}{\"fixVersions\": []map[string]string{{\"name\": \"1.2.0\"}}}")

	err = client.Issue.Update(ctx, bug.Key, &issue.UpdateInput{
		Fields: map[string]interface{}{
			// Update pattern for version arrays
			// Each version is a map with "name" or "id" key
			"fixVersions": []map[string]string{
				{"name": "1.2.0"}, // Reference by name
				{"name": "2.0.0"}, // Can set multiple versions
			},
			// NOTE: This REPLACES all fix versions, doesn't append
			// To append, you'd need to get current versions first
		},
	})
	if err != nil {
		log.Printf("   Warning: Failed to update issue: %v", err)
	} else {
		fmt.Printf("   ✓ Updated %s with fix versions\n", bug.Key)
	}
	fmt.Println()

	// Example 5: Get available resolutions
	// =====================================
	// Query Jira for all available resolutions in your instance
	// Useful to know what resolution values you can use
	fmt.Println("5. Listing available resolutions...")
	fmt.Println("   API: client.Resolution.List(ctx)")

	resolutions, err := client.Resolution.List(ctx)
	if err != nil {
		log.Printf("   Warning: Failed to list resolutions: %v", err)
	} else {
		fmt.Printf("   Available resolutions (%d):\n", len(resolutions))
		for _, res := range resolutions {
			// Each resolution has ID, Name, and Description
			fmt.Printf("     - %s (ID: %s)\n", res.Name, res.ID)
		}
		fmt.Println("   💡 Use these names when setting resolution on issues")
	}
	fmt.Println()

	// Example 6: Set resolution on an issue
	// ======================================
	// PATTERN for resolution: map[string]string with "name" or "id" key
	// Common resolutions: "Done", "Won't Fix", "Duplicate", "Cannot Reproduce"
	fmt.Println("6. Setting resolution on the bug...")
	fmt.Println("   Pattern: Fields: map[string]interface{}{\"resolution\": map[string]string{\"name\": \"Done\"}}")

	// Note: Some Jira workflows may require a status transition before/with resolution
	// This example just sets the resolution field directly
	err = client.Issue.Update(ctx, bug.Key, &issue.UpdateInput{
		Fields: map[string]interface{}{
			// Resolution pattern: map with "name" key
			"resolution": map[string]string{
				"name": "Done", // Use a resolution name from the list above
			},
			// Alternative: use "id" instead of "name"
			// "resolution": map[string]string{"id": "1"},
		},
	})
	if err != nil {
		log.Printf("   Warning: Failed to set resolution: %v", err)
	} else {
		fmt.Printf("   ✓ Set resolution for %s\n", bug.Key)
	}
	fmt.Println()

	// Example 7: Retrieve resolved issue and display resolution
	// ===========================================================
	// SAFE ACCESSOR PATTERNS for resolution:
	// 1. GetResolutionName() - returns string (empty if unresolved)
	// 2. GetResolution() - returns *resolution.Resolution (nil if unresolved)
	fmt.Println("7. Retrieving resolved issue and displaying resolution...")
	fmt.Println("   Pattern: Use GetResolutionName() or GetResolution()")

	resolved, err := client.Issue.Get(ctx, bug.Key, &issue.GetOptions{
		Fields: []string{"summary", "status", "resolution"},
	})
	if err != nil {
		log.Printf("   Warning: Failed to get resolved issue: %v", err)
	} else {
		fmt.Printf("   Issue: %s\n", resolved.Key)
		fmt.Printf("   Status: %s\n", resolved.GetStatusName())

		// PATTERN 1: GetResolutionName() - Simple string check
		// ✅ Recommended for most cases when you just need the name
		resolutionName := resolved.GetResolutionName()
		if resolutionName != "" {
			fmt.Printf("   Resolution: %s\n", resolutionName)
		} else {
			fmt.Printf("   Resolution: Unresolved\n")
		}

		// PATTERN 2: GetResolution() - Full resolution object
		// ✅ Use when you need ID, Description, or other fields
		if res := resolved.GetResolution(); res != nil {
			fmt.Printf("   Resolution Details:\n")
			fmt.Printf("     - ID: %s\n", res.ID)
			fmt.Printf("     - Name: %s\n", res.Name)
			if res.Description != "" {
				fmt.Printf("     - Description: %s\n", res.Description)
			}
		}

		// ❌ NEVER: resolved.Fields.Resolution.Name (can panic!)
		// ✅ ALWAYS: resolved.GetResolutionName() (safe)
	}
	fmt.Println()

	// Example 8: Create a task with both affected and fix versions
	// ==============================================================
	// Real-world scenario: A security issue that needs backporting
	// Shows using both AffectsVersions and FixVersions together
	fmt.Println("8. Creating a task with both affected and fix versions...")
	fmt.Println("   Scenario: Security fix needs backporting to multiple versions")

	task, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project:   &issue.Project{Key: projectKey},
			Summary:   "Backport security fix to older versions",
			IssueType: &issue.IssueType{Name: "Task"},
			Priority:  &issue.Priority{Name: "Critical"},

			// Bug exists in these versions
			AffectsVersions: []*project.Version{
				{Name: "1.0.0"},
				{Name: "1.1.0"},
				{Name: "1.2.0"},
			},

			// Will be fixed in these patch versions
			FixVersions: []*project.Version{
				{Name: "1.0.1"}, // Patch for 1.0.x line
				{Name: "1.1.1"}, // Patch for 1.1.x line
				{Name: "1.2.1"}, // Patch for 1.2.x line
			},
		},
	})
	if err != nil {
		log.Printf("   Warning: Failed to create task: %v", err)
	} else {
		fmt.Printf("   ✓ Created task: %s\n", task.Key)

		// Retrieve and display version counts
		retrieved, err := client.Issue.Get(ctx, task.Key, &issue.GetOptions{
			Fields: []string{"versions", "fixVersions"},
		})
		if err == nil {
			affectsVersions := retrieved.GetAffectsVersions()
			fixVersions := retrieved.GetFixVersions()

			fmt.Printf("   📊 Affects %d versions, will be fixed in %d versions\n",
				len(affectsVersions), len(fixVersions))
		}
	}
	fmt.Println()

	// Example 9: Demonstrating safe accessors with nil/empty fields
	// ==============================================================
	// Shows what happens when fields are not set - safe accessors never panic!
	fmt.Println("9. Demonstrating safe accessor methods...")
	fmt.Println("   Creating issue WITHOUT versions/resolution to test safe accessors")

	// Create an issue without version or resolution fields
	simple, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			Project:   &issue.Project{Key: projectKey},
			Summary:   "Simple task without versions",
			IssueType: &issue.IssueType{Name: "Task"},
			// Note: NO AffectsVersions, FixVersions, or Resolution set
		},
	})
	if err == nil {
		fmt.Printf("   ✓ Created simple task: %s\n", simple.Key)

		retrieved, err := client.Issue.Get(ctx, simple.Key, nil)
		if err == nil {
			// SAFE ACCESSORS: These NEVER panic, even when fields aren't set

			// Returns empty slice (never nil) for unset version fields
			affectsVersions := retrieved.GetAffectsVersions()
			fixVersions := retrieved.GetFixVersions()

			// Returns empty string for unset resolution
			resolutionName := retrieved.GetResolutionName()

			fmt.Println("   Results when fields are not set:")
			fmt.Printf("     - GetAffectsVersions(): %v (length: %d)\n", affectsVersions, len(affectsVersions))
			fmt.Printf("     - GetFixVersions(): %v (length: %d)\n", fixVersions, len(fixVersions))
			fmt.Printf("     - GetResolutionName(): '%s' (empty string)\n", resolutionName)
			fmt.Println("   ✅ No panics! Safe accessors handle missing fields gracefully")
		}
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println()
	fmt.Printf("Created Issues:\n")
	fmt.Printf("  - %s: Bug with affected versions\n", bug.Key)
	fmt.Printf("  - %s: Story with fix versions\n", story.Key)
	if task != nil {
		fmt.Printf("  - %s: Task with both affected and fix versions\n", task.Key)
	}
	if simple != nil {
		fmt.Printf("  - %s: Simple task (no versions)\n", simple.Key)
	}
	fmt.Println()

	fmt.Println("📚 KEY TAKEAWAYS - TYPE USAGE PATTERNS:")
	fmt.Println()
	fmt.Println("1. CREATING ISSUES WITH VERSIONS:")
	fmt.Println("   AffectsVersions: []*project.Version{{Name: \"1.0.0\"}}")
	fmt.Println("   FixVersions: []*project.Version{{Name: \"2.0.0\"}}")
	fmt.Println()
	fmt.Println("2. UPDATING ISSUES:")
	fmt.Println("   Fields: map[string]interface{}{")
	fmt.Println("     \"fixVersions\": []map[string]string{{\"name\": \"2.0.0\"}},")
	fmt.Println("     \"resolution\": map[string]string{\"name\": \"Done\"},")
	fmt.Println("   }")
	fmt.Println()
	fmt.Println("3. READING DATA - ALWAYS USE SAFE ACCESSORS:")
	fmt.Println("   ✅ issue.GetAffectsVersions()  // Returns []*project.Version (never nil, may be empty)")
	fmt.Println("   ✅ issue.GetFixVersions()      // Returns []*project.Version (never nil, may be empty)")
	fmt.Println("   ✅ issue.GetResolutionName()   // Returns string (empty if unresolved)")
	fmt.Println("   ✅ issue.GetResolution()       // Returns *resolution.Resolution or nil")
	fmt.Println("   ❌ issue.Fields.FixVersions    // NEVER! Can panic if Fields is nil")
	fmt.Println()
	fmt.Println("4. VERSION REFERENCES:")
	fmt.Println("   - By Name (recommended): {Name: \"1.0.0\"}")
	fmt.Println("   - By ID (if you have it): {ID: \"10001\"}")
	fmt.Println("   - Both: {ID: \"10001\", Name: \"1.0.0\"} (ID takes precedence)")
	fmt.Println()
	fmt.Println("5. COMMON RESOLUTIONS:")
	fmt.Println("   - \"Done\", \"Won't Fix\", \"Duplicate\", \"Cannot Reproduce\"")
	fmt.Println("   - Use client.Resolution.List(ctx) to see all available")
	fmt.Println()
	fmt.Println("💡 For more examples, see:")
	fmt.Println("   - examples/subtasks/ - Parent-child relationships")
	fmt.Println("   - examples/basic/ - General issue operations")
	fmt.Println("   - examples/projects/ - Version management at project level")
	fmt.Println()
	fmt.Println("🧹 To clean up, delete the created issues from your Jira instance.")
}
