// Package main demonstrates parent/subtask relationships with the jirasdk library.
//
// This example shows how to:
//   - Create parent issues (Stories, Epics, Tasks)
//   - Create subtasks with parent references
//   - Retrieve and display parent information
//   - Query for all subtasks of a parent
//   - Move subtasks between parents
//   - Use safe accessor methods for parent relationships
//
// IMPORTANT TYPE USAGE PATTERNS:
//
// 1. PARENT REFERENCES - Use IssueRef with Key or ID:
//   - By Key (recommended for most cases):
//     Parent: &issue.IssueRef{Key: "PROJ-123"}
//   - By ID (when you have the issue ID from API):
//     Parent: &issue.IssueRef{ID: "10001"}
//   - Both (Key takes precedence):
//     Parent: &issue.IssueRef{Key: "PROJ-123", ID: "10001"}
//
// 2. CREATING SUBTASKS - Three required steps:
//
//   - Set IssueType to "Sub-task" (exact name varies by Jira config)
//
//   - Set Parent field to reference the parent issue
//
//   - Include all standard required fields (Project, Summary, etc.)
//
//     Example:
//     IssueType: &issue.IssueType{Name: "Sub-task"}
//     Parent:    &issue.IssueRef{Key: "PROJ-123"}
//
// 3. SAFE ACCESSORS - Always use these to avoid nil pointer panics:
//   - issue.GetParent() - returns *issue.IssueRef (may be nil)
//   - issue.GetParentKey() - returns string (empty if no parent)
//
// 4. QUERYING SUBTASKS - Use JQL:
//   - All subtasks of a parent: "parent = PROJ-123"
//   - Subtasks in specific status: "parent = PROJ-123 AND status = 'In Progress'"
//
// 5. COMMON MISTAKES TO AVOID:
//   - ❌ Don't access issue.Fields.Parent directly (may cause panic)
//   - ✅ Use issue.GetParent() or issue.GetParentKey() instead
//   - ❌ Don't forget to set IssueType to "Sub-task"
//   - ✅ Verify your Jira's exact subtask type name
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/core/search"
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

	fmt.Println("=== Parent/Subtask Relationships Demo ===")
	fmt.Println("This example demonstrates proper type usage for parent-child relationships")
	fmt.Println()

	// Example 1: Create a parent story
	// ==================================
	// First, create a regular issue that will be the parent
	// This can be any issue type: Story, Epic, Task, Bug, etc.
	fmt.Println("1. Creating a parent story...")
	fmt.Println("   Creating a regular Story issue (will be the parent)")

	projectKey := os.Getenv("JIRA_PROJECT_KEY")
	parent, err := client.Issue.Create(ctx, &issue.CreateInput{
		Fields: &issue.IssueFields{
			// Standard required fields for any issue
			Project:   &issue.Project{Key: projectKey},
			Summary:   "Implement User Authentication System",
			IssueType: &issue.IssueType{Name: "Story"},
			Priority:  &issue.Priority{Name: "High"},
			// Note: No Parent field - this is a top-level issue
		},
	})
	if err != nil {
		log.Fatalf("Failed to create parent story: %v", err)
	}
	fmt.Printf("   ✓ Created parent story: %s\n", parent.Key)
	fmt.Println()

	// Example 2: Create subtasks with parent reference
	// =================================================
	// PATTERN: To create a subtask, you must:
	// 1. Set IssueType to "Sub-task" (check your Jira config for exact name)
	// 2. Set Parent to reference the parent issue
	fmt.Println("2. Creating subtasks with parent reference...")
	fmt.Println("   Pattern: IssueType: {{Name: \"Sub-task\"}}, Parent: {{Key: parent.Key}}")

	subtasks := []string{
		"Design authentication database schema",
		"Implement user registration endpoint",
		"Implement login endpoint with JWT",
		"Add password reset functionality",
		"Write integration tests",
	}

	createdSubtasks := make([]*issue.Issue, 0, len(subtasks))
	for i, summary := range subtasks {
		subtask, err := client.Issue.Create(ctx, &issue.CreateInput{
			Fields: &issue.IssueFields{
				// Standard required fields
				Project: &issue.Project{Key: projectKey},
				Summary: summary,

				// CRITICAL: Set IssueType to "Sub-task"
				// Note: The exact name depends on your Jira configuration
				// Common names: "Sub-task", "Subtask", "Sub Task"
				IssueType: &issue.IssueType{Name: "Sub-task"},

				// CRITICAL: Set Parent to link this subtask to parent issue
				// Use the Key of the parent issue created above
				Parent: &issue.IssueRef{Key: parent.Key},

				// Optional fields (vary by project configuration)
				Priority: &issue.Priority{Name: "Medium"},
			},
		})
		if err != nil {
			log.Printf("   Warning: Failed to create subtask %d: %v", i+1, err)
			continue
		}
		createdSubtasks = append(createdSubtasks, subtask)
		fmt.Printf("   ✓ Created subtask %d/%d: %s - %s\n", i+1, len(subtasks), subtask.Key, summary)
	}
	fmt.Println()

	// Example 3: Retrieve subtask and display parent information
	// ============================================================
	// IMPORTANT: Always use safe accessor methods to read parent data
	// ❌ NEVER: issue.Fields.Parent.Key (can cause nil panic)
	// ✅ ALWAYS: issue.GetParentKey() (safe, returns empty if no parent)
	if len(createdSubtasks) > 0 {
		fmt.Println("3. Retrieving subtask and displaying parent information...")
		fmt.Println("   Pattern: Use GetParent() and GetParentKey()")

		firstSubtask, err := client.Issue.Get(ctx, createdSubtasks[0].Key, &issue.GetOptions{
			Fields: []string{"summary", "parent", "status", "issuetype"},
		})
		if err != nil {
			log.Printf("   Warning: Failed to get subtask: %v", err)
		} else {
			fmt.Printf("   Subtask: %s\n", firstSubtask.Key)
			fmt.Printf("   Summary: %s\n", firstSubtask.GetSummary())

			// PATTERN 1: GetParentKey() - Simple string check
			// ✅ Recommended for most cases when you just need the key
			parentKey := firstSubtask.GetParentKey()
			if parentKey != "" {
				fmt.Printf("   Parent Issue: %s\n", parentKey)
			} else {
				fmt.Println("   Parent Issue: None (not a subtask)")
			}

			// PATTERN 2: GetParent() - Full IssueRef object
			// ✅ Use when you need both Key and ID
			if parent := firstSubtask.GetParent(); parent != nil {
				fmt.Printf("   Parent Details:\n")
				fmt.Printf("     - Key: %s\n", parent.Key)
				if parent.ID != "" {
					fmt.Printf("     - ID: %s\n", parent.ID)
				}
			}

			// ❌ NEVER: firstSubtask.Fields.Parent.Key (can panic!)
			// ✅ ALWAYS: firstSubtask.GetParentKey() (safe)
		}
		fmt.Println()
	}

	// Example 4: Query all subtasks of a parent issue
	// ==================================================
	// PATTERN: Use JQL with "parent = PARENT-KEY" to find all subtasks
	// This is the most efficient way to retrieve all subtasks of a parent
	fmt.Println("4. Querying all subtasks of parent issue...")
	fmt.Println("   Pattern: JQL query \"parent = PARENT-KEY\"")

	// PATTERN: SearchJQL with parent filter
	// ✅ Use "parent = KEY" in JQL to find subtasks
	// ✅ Add ORDER BY to control result ordering
	// ✅ Specify Fields to limit data returned (improves performance)
	searchResult, err := client.Search.SearchJQL(ctx, &search.SearchJQLOptions{
		// JQL query: "parent = PROJ-123" finds all subtasks of PROJ-123
		JQL: fmt.Sprintf("parent = %s ORDER BY created ASC", parent.Key),

		// MaxResults: Limit number of results (use pagination for large sets)
		MaxResults: 50,

		// Fields: Only fetch the fields you need
		// This improves performance by reducing payload size
		Fields: []string{"summary", "status", "priority"},
	})
	if err != nil {
		log.Printf("   Warning: Failed to search subtasks: %v", err)
	} else {
		fmt.Printf("   Found %d subtasks:\n", len(searchResult.Issues))
		for i, subtask := range searchResult.Issues {
			// Use safe accessors to read issue data
			// ✅ GetSummary() - safe, never panics
			// ✅ GetStatusName() - safe, returns empty if status not set
			fmt.Printf("   %d. %s - %s (Status: %s)\n",
				i+1,
				subtask.Key,
				subtask.GetSummary(),
				subtask.GetStatusName(),
			)
		}
	}
	fmt.Println()

	// Additional JQL patterns for subtasks:
	// - "parent = PROJ-123 AND status = 'In Progress'" - Filter by status
	// - "parent = PROJ-123 AND assignee = currentUser()" - My subtasks
	// - "parent in (PROJ-123, PROJ-124)" - Subtasks of multiple parents

	// Example 5: Move a subtask to a different parent
	// ==================================================
	// PATTERN: Update the parent field to move a subtask to a different parent
	// This is useful for reorganizing work or correcting parent assignments
	if len(createdSubtasks) > 1 {
		fmt.Println("5. Creating new parent and moving subtask...")
		fmt.Println("   Pattern: Update parent field with new parent key")

		// Create another parent story
		// This is just a regular issue creation (any type can be a parent)
		newParent, err := client.Issue.Create(ctx, &issue.CreateInput{
			Fields: &issue.IssueFields{
				Project:   &issue.Project{Key: os.Getenv("JIRA_PROJECT_KEY")},
				Summary:   "Improve User Authentication",
				IssueType: &issue.IssueType{Name: "Story"},
			},
		})
		if err != nil {
			log.Printf("   Warning: Failed to create new parent: %v", err)
		} else {
			fmt.Printf("   Created new parent: %s\n", newParent.Key)

			// Move the last subtask to the new parent
			// PATTERN: Use Update with parent field in map[string]interface{}
			lastSubtask := createdSubtasks[len(createdSubtasks)-1]

			// CRITICAL UPDATE PATTERN:
			// ✅ Use map[string]interface{} for Fields in UpdateInput
			// ✅ Use lowercase "parent" (Jira API field name)
			// ✅ Use map[string]string{"key": "PARENT-KEY"} for parent value
			err = client.Issue.Update(ctx, lastSubtask.Key, &issue.UpdateInput{
				Fields: map[string]interface{}{
					// Use lowercase field name and nested map structure
					"parent": map[string]string{
						"key": newParent.Key, // Reference parent by key
					},
				},
			})
			if err != nil {
				log.Printf("   Warning: Failed to move subtask: %v", err)
			} else {
				fmt.Printf("   Moved subtask %s from %s to %s\n",
					lastSubtask.Key,
					parent.Key,
					newParent.Key,
				)
			}
		}
		fmt.Println()
	}

	// Note: You can also reference parent by ID:
	// "parent": map[string]string{"id": "10001"}
	// But using Key is more readable and maintainable

	// Example 6: Checking if an issue is a subtask
	// ==============================================
	// PATTERN: Two ways to detect subtasks
	// 1. Check if issue has a parent (GetParent() != nil)
	// 2. Check if issue type has Subtask flag set
	if len(createdSubtasks) > 0 {
		fmt.Println("6. Checking if an issue is a subtask...")
		fmt.Println("   Pattern 1: Check if GetParent() != nil")
		fmt.Println("   Pattern 2: Check if IssueType.Subtask == true")

		// Retrieve issue with minimal fields
		// ✅ Only fetch fields needed for the check (performance optimization)
		subtask, err := client.Issue.Get(ctx, createdSubtasks[0].Key, &issue.GetOptions{
			Fields: []string{"issuetype", "parent"},
		})
		if err == nil {
			// PATTERN 1: Check if issue has a parent
			// ✅ Use GetParent() to safely check for parent relationship
			// This is the most reliable way to detect subtasks
			hasParent := subtask.GetParent() != nil
			fmt.Printf("   Issue %s has parent: %v\n", subtask.Key, hasParent)

			// PATTERN 2: Check if issue type is marked as subtask
			// ✅ Use GetIssueType() to safely access issue type
			// The Subtask boolean field indicates if this is a subtask type
			if issueType := subtask.GetIssueType(); issueType != nil {
				fmt.Printf("   Issue type is subtask: %v\n", issueType.Subtask)
				fmt.Printf("   Issue type name: %s\n", issueType.Name)
			}

			// RECOMMENDED: Use both checks for robustness
			// An issue is a subtask if it has a parent AND its type is subtask
			isSubtask := hasParent && subtask.GetIssueType() != nil && subtask.GetIssueType().Subtask
			fmt.Printf("   ✅ Confirmed subtask: %v\n", isSubtask)
		}
		fmt.Println()
	}

	fmt.Println("=== Demo Complete ===")
	fmt.Printf("\nCreated:\n")
	fmt.Printf("  - 1 parent story: %s\n", parent.Key)
	fmt.Printf("  - %d subtasks\n", len(createdSubtasks))
	fmt.Println()

	// Key Takeaways Section
	// =====================
	// Comprehensive summary of all parent/subtask patterns demonstrated above
	fmt.Println("📚 KEY TAKEAWAYS - PARENT/SUBTASK TYPE USAGE PATTERNS:")
	fmt.Println()

	fmt.Println("1. CREATING SUBTASKS - Three Critical Requirements:")
	fmt.Println("   ✅ Set IssueType to \"Sub-task\" (check your Jira config for exact name)")
	fmt.Println("   ✅ Set Parent field: &issue.IssueRef{Key: \"PARENT-KEY\"}")
	fmt.Println("   ✅ Include all standard required fields (Project, Summary, etc.)")
	fmt.Println()
	fmt.Println("   Example:")
	fmt.Println("     IssueType: &issue.IssueType{Name: \"Sub-task\"}")
	fmt.Println("     Parent:    &issue.IssueRef{Key: \"PROJ-123\"}")
	fmt.Println()

	fmt.Println("2. PARENT REFERENCES - Use IssueRef with Key or ID:")
	fmt.Println("   ✅ By Key (recommended):    Parent: &issue.IssueRef{Key: \"PROJ-123\"}")
	fmt.Println("   ✅ By ID (when available):  Parent: &issue.IssueRef{ID: \"10001\"}")
	fmt.Println("   ✅ Both (Key takes precedence)")
	fmt.Println()

	fmt.Println("3. READING PARENT DATA - ALWAYS USE SAFE ACCESSORS:")
	fmt.Println("   ✅ issue.GetParent()      // Returns *IssueRef (may be nil)")
	fmt.Println("   ✅ issue.GetParentKey()   // Returns string (empty if no parent)")
	fmt.Println("   ❌ issue.Fields.Parent    // NEVER! Can panic if Fields is nil")
	fmt.Println()
	fmt.Println("   Pattern 1 - Simple key check:")
	fmt.Println("     if parentKey := issue.GetParentKey(); parentKey != \"\" {")
	fmt.Printf("       fmt.Printf(\"Parent: %%s\\n\", parentKey)\n")
	fmt.Println("     }")
	fmt.Println()
	fmt.Println("   Pattern 2 - Full parent details:")
	fmt.Println("     if parent := issue.GetParent(); parent != nil {")
	fmt.Printf("       fmt.Printf(\"Key: %%s, ID: %%s\\n\", parent.Key, parent.ID)\n")
	fmt.Println("     }")
	fmt.Println()

	fmt.Println("4. QUERYING SUBTASKS - Use JQL:")
	fmt.Println("   ✅ All subtasks:            \"parent = PROJ-123\"")
	fmt.Println("   ✅ Filter by status:        \"parent = PROJ-123 AND status = 'In Progress'\"")
	fmt.Println("   ✅ Current user's subtasks: \"parent = PROJ-123 AND assignee = currentUser()\"")
	fmt.Println("   ✅ Multiple parents:        \"parent in (PROJ-123, PROJ-124)\"")
	fmt.Println()
	fmt.Println("   Example:")
	fmt.Println("     searchResult, err := client.Search.SearchJQL(ctx, &search.SearchJQLOptions{")
	fmt.Println("       JQL:        \"parent = PROJ-123 ORDER BY created ASC\",")
	fmt.Println("       MaxResults: 50,")
	fmt.Println("       Fields:     []string{\"summary\", \"status\"},")
	fmt.Println("     })")
	fmt.Println()

	fmt.Println("5. UPDATING PARENT - Use map[string]interface{} for UpdateInput:")
	fmt.Println("   ✅ Use lowercase field name: \"parent\" (Jira API convention)")
	fmt.Println("   ✅ Use nested map structure: map[string]string{\"key\": \"NEW-PARENT\"}")
	fmt.Println()
	fmt.Println("   Example:")
	fmt.Println("     err := client.Issue.Update(ctx, subtaskKey, &issue.UpdateInput{")
	fmt.Println("       Fields: map[string]interface{}{")
	fmt.Println("         \"parent\": map[string]string{\"key\": newParent.Key},")
	fmt.Println("       },")
	fmt.Println("     })")
	fmt.Println()

	fmt.Println("6. DETECTING SUBTASKS - Two Methods:")
	fmt.Println("   ✅ Check parent:     issue.GetParent() != nil")
	fmt.Println("   ✅ Check type flag:  issue.GetIssueType().Subtask == true")
	fmt.Println()
	fmt.Println("   Recommended - Use both for robustness:")
	fmt.Println("     isSubtask := issue.GetParent() != nil && ")
	fmt.Println("                  issue.GetIssueType() != nil && ")
	fmt.Println("                  issue.GetIssueType().Subtask")
	fmt.Println()

	fmt.Println("7. COMMON MISTAKES TO AVOID:")
	fmt.Println("   ❌ Accessing Fields.Parent directly → Use GetParent() or GetParentKey()")
	fmt.Println("   ❌ Forgetting to set IssueType to \"Sub-task\"")
	fmt.Println("   ❌ Using wrong subtask type name (verify in your Jira config)")
	fmt.Println("   ❌ Not checking for nil when accessing parent details")
	fmt.Println("   ❌ Using uppercase \"Parent\" in UpdateInput → Use lowercase \"parent\"")
	fmt.Println()

	fmt.Println("📖 For more information, see:")
	fmt.Println("   - Jira REST API: https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-group-issues")
	fmt.Println("   - SDK Documentation: https://pkg.go.dev/github.com/felixgeelhaar/jirasdk")
	fmt.Println()

	fmt.Println("To clean up, delete the issues from your Jira instance.")
}
