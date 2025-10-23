// Package issue provides Issue resource management for Jira.
package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/felixgeelhaar/jirasdk/core/project"
	"github.com/felixgeelhaar/jirasdk/core/resolution"
)

// Service provides operations for Issue resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Issue service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Issue represents a Jira issue.
type Issue struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Self   string       `json:"self"`
	Fields *IssueFields `json:"fields,omitempty"`
	Expand string       `json:"expand,omitempty"`
}

// SafeFields returns the issue fields, or an empty IssueFields struct if nil.
// This method prevents nil pointer dereferences when accessing issue fields.
//
// Example:
//
//	// Safe - will not panic even if Fields is nil
//	summary := issue.SafeFields().Summary
//
//	// Unsafe - could panic if Fields is nil
//	summary := issue.Fields.Summary
func (i *Issue) SafeFields() *IssueFields {
	if i.Fields == nil {
		return &IssueFields{}
	}
	return i.Fields
}

// GetSummary safely retrieves the issue summary.
// Returns an empty string if Fields is nil.
func (i *Issue) GetSummary() string {
	if i.Fields == nil {
		return ""
	}
	return i.Fields.Summary
}

// GetDescription safely retrieves the issue description as ADF.
// Returns nil if Fields or Description is nil.
func (i *Issue) GetDescription() *ADF {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Description
}

// GetDescriptionText safely retrieves the issue description as plain text.
// This extracts text from the ADF format.
// Returns an empty string if Fields or Description is nil.
func (i *Issue) GetDescriptionText() string {
	adf := i.GetDescription()
	if adf == nil {
		return ""
	}
	return adf.ToText()
}

// GetEnvironment safely retrieves the issue environment as ADF.
// Returns nil if Fields or Environment is nil.
func (i *Issue) GetEnvironment() *ADF {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Environment
}

// GetEnvironmentText safely retrieves the issue environment as plain text.
// This extracts text from the ADF format.
// Returns an empty string if Fields or Environment is nil.
func (i *Issue) GetEnvironmentText() string {
	adf := i.GetEnvironment()
	if adf == nil {
		return ""
	}
	return adf.ToText()
}

// GetStatus safely retrieves the issue status.
// Returns nil if Fields or Status is nil.
func (i *Issue) GetStatus() *Status {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Status
}

// GetStatusName safely retrieves the status name.
// Returns an empty string if Fields, Status, or Status.Name is not available.
func (i *Issue) GetStatusName() string {
	status := i.GetStatus()
	if status == nil {
		return ""
	}
	return status.Name
}

// GetPriority safely retrieves the issue priority.
// Returns nil if Fields or Priority is nil.
func (i *Issue) GetPriority() *Priority {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Priority
}

// GetPriorityName safely retrieves the priority name.
// Returns an empty string if Fields, Priority, or Priority.Name is not available.
func (i *Issue) GetPriorityName() string {
	priority := i.GetPriority()
	if priority == nil {
		return ""
	}
	return priority.Name
}

// GetAssignee safely retrieves the issue assignee.
// Returns nil if Fields or Assignee is nil.
func (i *Issue) GetAssignee() *User {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Assignee
}

// GetAssigneeName safely retrieves the assignee display name.
// Returns an empty string if Fields, Assignee, or Assignee.DisplayName is not available.
func (i *Issue) GetAssigneeName() string {
	assignee := i.GetAssignee()
	if assignee == nil {
		return ""
	}
	return assignee.DisplayName
}

// GetReporter safely retrieves the issue reporter.
// Returns nil if Fields or Reporter is nil.
func (i *Issue) GetReporter() *User {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Reporter
}

// GetReporterName safely retrieves the reporter display name.
// Returns an empty string if Fields, Reporter, or Reporter.DisplayName is not available.
func (i *Issue) GetReporterName() string {
	reporter := i.GetReporter()
	if reporter == nil {
		return ""
	}
	return reporter.DisplayName
}

// GetParent safely retrieves the issue parent.
// Returns nil if Fields or Parent is nil.
// This is useful for subtasks to reference their parent issue.
//
// Example usage:
//
//	if parent := issue.GetParent(); parent != nil {
//	    fmt.Printf("Parent: %s\n", parent.Key)
//	}
//
// WARNING: Never directly access issue.Fields.Parent as it may be nil and cause a panic.
// Always use this safe accessor method or GetParentKey() instead.
func (i *Issue) GetParent() *IssueRef {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Parent
}

// GetParentKey safely retrieves the parent issue key.
// Returns an empty string if Fields, Parent, or Parent.Key is not available.
//
// Example usage:
//
//	if parentKey := issue.GetParentKey(); parentKey != "" {
//	    fmt.Printf("Parent Issue: %s\n", parentKey)
//	}
//
// This is the recommended method for most use cases when you only need the parent key.
func (i *Issue) GetParentKey() string {
	parent := i.GetParent()
	if parent == nil {
		return ""
	}
	return parent.Key
}

// GetResolution safely retrieves the issue resolution.
// Returns nil if Fields or Resolution is nil.
//
// Example usage:
//
//	if resolution := issue.GetResolution(); resolution != nil {
//	    fmt.Printf("Resolution: %s\n", resolution.Name)
//	}
//
// WARNING: Never directly access issue.Fields.Resolution as it may be nil and cause a panic.
// Always use this safe accessor method or GetResolutionName() instead.
func (i *Issue) GetResolution() *resolution.Resolution {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Resolution
}

// GetResolutionName safely retrieves the resolution name.
// Returns an empty string if Fields, Resolution, or Resolution.Name is not available.
//
// Example usage:
//
//	if resolutionName := issue.GetResolutionName(); resolutionName != "" {
//	    fmt.Printf("Resolution: %s\n", resolutionName)
//	}
//
// This is the recommended method for most use cases when you only need the resolution name.
func (i *Issue) GetResolutionName() string {
	resolution := i.GetResolution()
	if resolution == nil {
		return ""
	}
	return resolution.Name
}

// GetFixVersions safely retrieves the issue fix versions.
// Always returns a slice (never nil, but may be empty) for safe iteration.
//
// Example usage:
//
//	versions := issue.GetFixVersions()  // Safe: never nil, always iterable
//	for _, version := range versions {
//	    fmt.Printf("Fix Version: %s\n", version.Name)
//	}
//
// WARNING: Never directly access issue.Fields.FixVersions as Fields may be nil.
// Always use this safe accessor method.
func (i *Issue) GetFixVersions() []*project.Version {
	if i.Fields == nil || i.Fields.FixVersions == nil {
		return []*project.Version{}
	}
	return i.Fields.FixVersions
}

// GetAffectsVersions safely retrieves the versions affected by the issue.
// Always returns a slice (never nil, but may be empty) for safe iteration.
//
// Example usage:
//
//	versions := issue.GetAffectsVersions()  // Safe: never nil, always iterable
//	for _, version := range versions {
//	    fmt.Printf("Affects Version: %s\n", version.Name)
//	}
//
// WARNING: Never directly access issue.Fields.AffectsVersions as Fields may be nil.
// Always use this safe accessor method.
func (i *Issue) GetAffectsVersions() []*project.Version {
	if i.Fields == nil || i.Fields.AffectsVersions == nil {
		return []*project.Version{}
	}
	return i.Fields.AffectsVersions
}

// GetProject safely retrieves the issue project.
// Returns nil if Fields or Project is nil.
func (i *Issue) GetProject() *Project {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Project
}

// GetProjectKey safely retrieves the project key.
// Returns an empty string if Fields, Project, or Project.Key is not available.
func (i *Issue) GetProjectKey() string {
	project := i.GetProject()
	if project == nil {
		return ""
	}
	return project.Key
}

// GetIssueType safely retrieves the issue type.
// Returns nil if Fields or IssueType is nil.
func (i *Issue) GetIssueType() *IssueType {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.IssueType
}

// GetIssueTypeName safely retrieves the issue type name.
// Returns an empty string if Fields, IssueType, or IssueType.Name is not available.
func (i *Issue) GetIssueTypeName() string {
	issueType := i.GetIssueType()
	if issueType == nil {
		return ""
	}
	return issueType.Name
}

// GetLabels safely retrieves the issue labels.
// Returns an empty slice if Fields or Labels is nil.
func (i *Issue) GetLabels() []string {
	if i.Fields == nil || i.Fields.Labels == nil {
		return []string{}
	}
	return i.Fields.Labels
}

// GetComponents safely retrieves the issue components.
// Returns an empty slice if Fields or Components is nil.
func (i *Issue) GetComponents() []*Component {
	if i.Fields == nil || i.Fields.Components == nil {
		return []*Component{}
	}
	return i.Fields.Components
}

// GetCreated safely retrieves the issue creation timestamp as a pointer.
// Returns nil if Fields or Created is nil.
//
// Example usage:
//
//	if created := issue.GetCreated(); created != nil {
//	    fmt.Printf("Created: %s\n", created.Format(time.RFC3339))
//	}
//
// WARNING: Never directly access issue.Fields.Created as it may be nil and cause a panic.
// Always use this safe accessor method or GetCreatedTime() instead.
func (i *Issue) GetCreated() *time.Time {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Created
}

// GetCreatedTime safely retrieves the issue creation timestamp as a value.
// Returns zero time (time.Time{}) if Fields or Created is nil.
// Use this method when you need a time.Time value instead of a pointer.
//
// Example usage:
//
//	if created := issue.GetCreatedTime(); !created.IsZero() {
//	    fmt.Printf("Created: %s\n", created.Format(time.RFC3339))
//	}
//
// This is the recommended method for most use cases as it avoids nil pointer checks.
// WARNING: Never directly access issue.Fields.Created as it may be nil and cause a panic.
func (i *Issue) GetCreatedTime() time.Time {
	created := i.GetCreated()
	if created == nil {
		return time.Time{}
	}
	return *created
}

// GetUpdated safely retrieves the issue last update timestamp as a pointer.
// Returns nil if Fields or Updated is nil.
//
// Example usage:
//
//	if updated := issue.GetUpdated(); updated != nil {
//	    fmt.Printf("Updated: %s\n", updated.Format(time.RFC3339))
//	}
//
// WARNING: Never directly access issue.Fields.Updated as it may be nil and cause a panic.
// Always use this safe accessor method or GetUpdatedTime() instead.
func (i *Issue) GetUpdated() *time.Time {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.Updated
}

// GetUpdatedTime safely retrieves the issue last update timestamp as a value.
// Returns zero time (time.Time{}) if Fields or Updated is nil.
// Use this method when you need a time.Time value instead of a pointer.
//
// Example usage:
//
//	if updated := issue.GetUpdatedTime(); !updated.IsZero() {
//	    fmt.Printf("Updated: %s\n", updated.Format(time.RFC3339))
//	}
//
// This is the recommended method for most use cases as it avoids nil pointer checks.
// WARNING: Never directly access issue.Fields.Updated as it may be nil and cause a panic.
func (i *Issue) GetUpdatedTime() time.Time {
	updated := i.GetUpdated()
	if updated == nil {
		return time.Time{}
	}
	return *updated
}

// GetDueDate safely retrieves the issue due date as a pointer.
// Returns nil if Fields or DueDate is nil.
//
// Example usage:
//
//	if dueDate := issue.GetDueDate(); dueDate != nil {
//	    fmt.Printf("Due: %s\n", dueDate.Format("2006-01-02"))
//	}
//
// WARNING: Never directly access issue.Fields.DueDate as it may be nil and cause a panic.
// Always use this safe accessor method or GetDueDateValue() instead.
func (i *Issue) GetDueDate() *time.Time {
	if i.Fields == nil {
		return nil
	}
	return i.Fields.DueDate
}

// GetDueDateValue safely retrieves the issue due date as a value.
// Returns zero time (time.Time{}) if Fields or DueDate is nil.
// Use this method when you need a time.Time value instead of a pointer.
//
// Example usage:
//
//	if dueDate := issue.GetDueDateValue(); !dueDate.IsZero() {
//	    fmt.Printf("Due: %s\n", dueDate.Format("2006-01-02"))
//	    if time.Now().After(dueDate) {
//	        fmt.Println("Task is OVERDUE!")
//	    }
//	}
//
// This is the recommended method for most use cases as it avoids nil pointer checks.
// WARNING: Never directly access issue.Fields.DueDate as it may be nil and cause a panic.
func (i *Issue) GetDueDateValue() time.Time {
	dueDate := i.GetDueDate()
	if dueDate == nil {
		return time.Time{}
	}
	return *dueDate
}

// IssueFields contains all fields of a Jira issue.
//
// IMPORTANT TYPE USAGE PATTERNS:
//
// 1. CREATING ISSUES - Required and Optional Fields:
//
//	Required for all issues:
//	  - Summary: Short description (string)
//	  - Project: Project reference with Key
//	  - IssueType: Issue type with Name (e.g., "Bug", "Story", "Task")
//
//	Required for specific issue types:
//	  - Parent: Required for subtasks (use &IssueRef{Key: "PARENT-KEY"})
//
//	Optional common fields:
//	  - Description: Detailed description (use SetDescriptionText() or ADF)
//	  - Priority: Priority level (e.g., &Priority{Name: "High"})
//	  - Assignee: User assignment (use &User{AccountID: "..."})
//	  - Labels: Tags for categorization
//	  - FixVersions: Versions where issue will be/was fixed
//	  - AffectsVersions: Versions affected by the issue (typically for bugs)
//	  - DueDate: Deadline for the issue
//
//	Example - Creating a bug:
//	  fields := &IssueFields{
//	      Summary:   "Critical login bug",
//	      Project:   &Project{Key: "PROJ"},
//	      IssueType: &IssueType{Name: "Bug"},
//	      Priority:  &Priority{Name: "High"},
//	      AffectsVersions: []*project.Version{{Name: "1.0.0"}},
//	  }
//	  fields.SetDescriptionText("Users cannot log in after update")
//
//	Example - Creating a subtask:
//	  fields := &IssueFields{
//	      Summary:   "Implement login API",
//	      Project:   &Project{Key: "PROJ"},
//	      IssueType: &IssueType{Name: "Sub-task"},
//	      Parent:    &IssueRef{Key: "PROJ-123"},  // Link to parent issue
//	  }
//
// 2. READING ISSUE DATA - ALWAYS Use Safe Accessor Methods:
//
//	Safe accessors prevent nil pointer panics and provide sensible defaults:
//	  - issue.GetSummary() - returns string (empty if nil)
//	  - issue.GetDescription() - returns *ADF (nil if not set)
//	  - issue.GetStatusName() - returns string (empty if nil)
//	  - issue.GetPriorityName() - returns string (empty if nil)
//	  - issue.GetAssignee() - returns *User (nil if not assigned)
//	  - issue.GetParent() - returns *IssueRef (nil if not a subtask)
//	  - issue.GetFixVersions() - returns []*Version (never nil, may be empty)
//	  - issue.GetAffectsVersions() - returns []*Version (never nil, may be empty)
//	  - issue.GetCreatedTime() - returns time.Time (zero if nil)
//
//	❌ NEVER access fields directly: issue.Fields.Status
//	✅ ALWAYS use safe accessors: issue.GetStatusName()
//
// 3. UPDATING ISSUES - Use UpdateInput with map[string]interface{}:
//
//	When updating issues, you must use lowercase field names (Jira API convention):
//
//	Example - Update summary and priority:
//	  err := client.Issue.Update(ctx, issueKey, &UpdateInput{
//	      Fields: map[string]interface{}{
//	          "summary":  "Updated summary",
//	          "priority": map[string]string{"name": "Critical"},
//	      },
//	  })
//
//	Example - Update fix versions:
//	  err := client.Issue.Update(ctx, issueKey, &UpdateInput{
//	      Fields: map[string]interface{}{
//	          "fixVersions": []map[string]string{
//	              {"name": "2.0.0"},
//	              {"name": "2.1.0"},
//	          },
//	      },
//	  })
//
//	Example - Set resolution (when closing):
//	  err := client.Issue.Update(ctx, issueKey, &UpdateInput{
//	      Fields: map[string]interface{}{
//	          "resolution": map[string]string{"name": "Done"},
//	      },
//	  })
//
// 4. ADF (Atlassian Document Format) Fields:
//
//	Description and Environment fields require ADF format (Jira Cloud API v3):
//
//	Simple text - Use convenience methods:
//	  fields.SetDescriptionText("Plain text description")
//	  fields.SetEnvironmentText("Production environment")
//
//	Rich formatting - Use ADF builder:
//	  adf := issue.NewADF().
//	      AddHeading("Problem", 2).
//	      AddParagraph("Detailed description...").
//	      AddBulletList([]string{"Item 1", "Item 2"})
//	  fields.SetDescription(adf)
//
// 5. VERSION FIELDS:
//
//	FixVersions: Versions where the issue is/will be fixed (used for tracking releases)
//	  - Set when planning which release will contain the fix
//	  - Common for all issue types when release planning is needed
//	  - Use []*project.Version{{Name: "2.0.0"}}
//
//	AffectsVersions: Versions where the issue exists (typically for bugs)
//	  - Set when reporting bugs to indicate which versions have the problem
//	  - Helps determine backporting needs
//	  - Use []*project.Version{{Name: "1.0.0"}, {Name: "1.1.0"}}
//
// 6. READ-ONLY FIELDS:
//
//	These fields are set by Jira and cannot be modified directly:
//	  - Status: Updated via transitions (see client.Issue.Transition)
//	  - Reporter: Set automatically to the creating user
//	  - Created: Set automatically when issue is created
//	  - Updated: Set automatically when issue is modified
//
// 7. CUSTOM FIELDS:
//
//	For custom fields defined in your Jira instance:
//	  - Use the Custom field with CustomFields helper methods
//	  - Access via UnknownFields for dynamic handling
type IssueFields struct {
	// Core Fields
	// ===========

	// Summary is the issue title/summary (required for creation).
	// This is a short, one-line description of the issue.
	//
	// Example: "Fix critical login bug" or "Implement user authentication"
	//
	// Access: Direct access is safe (string type, never nil)
	// Update: Use lowercase "summary" in UpdateInput
	Summary string `json:"summary,omitempty"`

	// Description is the detailed issue description in ADF format.
	// Jira Cloud API v3 requires ADF (Atlassian Document Format) for rich text.
	//
	// RECOMMENDED: Use SetDescriptionText() for plain text:
	//   fields.SetDescriptionText("This is a description")
	//
	// ADVANCED: Use ADF builder for rich formatting:
	//   fields.SetDescription(issue.NewADF().AddParagraph("Text"))
	//
	// Access: Use issue.GetDescription() (safe, returns nil if not set)
	// Update: Use lowercase "description" in UpdateInput with ADF structure
	Description *ADF `json:"description,omitempty"`

	// Environment describes the issue environment in ADF format.
	// Commonly used for bugs to specify where the issue occurs.
	//
	// Example: "Production server: web-01, OS: Ubuntu 22.04, Browser: Chrome 120"
	//
	// RECOMMENDED: Use SetEnvironmentText() for plain text:
	//   fields.SetEnvironmentText("Production: web-01, Ubuntu 22.04")
	//
	// Access: Use issue.GetEnvironment() (safe, returns nil if not set)
	// Update: Use lowercase "environment" in UpdateInput with ADF structure
	Environment *ADF `json:"environment,omitempty"`

	// Metadata Fields
	// ===============

	// IssueType defines the type of issue (required for creation).
	// Common types: "Bug", "Story", "Task", "Epic", "Sub-task"
	//
	// Example: &IssueType{Name: "Bug"}
	//
	// IMPORTANT: For subtasks, must use "Sub-task" (exact name varies by Jira config)
	//
	// Access: Use issue.GetIssueType() (safe, returns nil if not set)
	// Update: Cannot be changed after creation in most Jira configurations
	IssueType *IssueType `json:"issuetype,omitempty"`

	// Project is the project this issue belongs to (required for creation).
	// Reference by Key (recommended) or ID.
	//
	// Example: &Project{Key: "PROJ"}
	//
	// Access: Direct access or use issue.SafeFields().Project
	// Update: Cannot be changed after creation (issues are tied to projects)
	Project *Project `json:"project,omitempty"`

	// Status is the current workflow status of the issue (read-only).
	// Examples: "To Do", "In Progress", "Done", "Closed"
	//
	// IMPORTANT: Status cannot be set directly. Use client.Issue.Transition() to change status.
	//
	// Access: Use issue.GetStatus() or issue.GetStatusName() (safe)
	// Update: Use Transition API (not UpdateInput)
	Status *Status `json:"status,omitempty"`

	// Resolution indicates how the issue was resolved (set when closing).
	// Examples: "Done", "Won't Fix", "Duplicate", "Cannot Reproduce"
	//
	// Typically set when transitioning to a closed/resolved status.
	//
	// Access: Use issue.GetResolution() or issue.GetResolutionName() (safe)
	// Update: Use lowercase "resolution" with map[string]string{"name": "Done"}
	Resolution *resolution.Resolution `json:"resolution,omitempty"`

	// Priority indicates the issue importance (optional).
	// Common values: "Highest", "High", "Medium", "Low", "Lowest"
	//
	// Example: &Priority{Name: "High"}
	//
	// Access: Use issue.GetPriority() or issue.GetPriorityName() (safe)
	// Update: Use lowercase "priority" with map[string]string{"name": "Critical"}
	Priority *Priority `json:"priority,omitempty"`

	// People Fields
	// =============

	// Assignee is the user assigned to work on the issue (optional).
	// Reference by AccountID (Jira Cloud) or Name (Jira Server).
	//
	// Example: &User{AccountID: "5b10a..."} (Cloud)
	//          &User{Name: "jsmith"} (Server)
	//
	// Access: Use issue.GetAssignee() or issue.GetAssigneeName() (safe)
	// Update: Use lowercase "assignee" with map[string]string{"accountId": "..."}
	Assignee *User `json:"assignee,omitempty"`

	// Reporter is the user who created the issue (read-only, set automatically).
	// This is set to the authenticated user when creating an issue.
	//
	// Access: Use issue.GetReporter() or issue.GetReporterName() (safe)
	// Update: Cannot be changed after creation in most configurations
	Reporter *User `json:"reporter,omitempty"`

	// Relationship Fields
	// ===================

	// Parent is the parent issue reference for subtasks (required for subtasks).
	// Links a subtask to its parent issue.
	//
	// Example: &IssueRef{Key: "PROJ-123"}
	//
	// IMPORTANT: Only set this when IssueType is "Sub-task" or similar.
	//
	// Access: Use issue.GetParent() or issue.GetParentKey() (safe)
	// Update: Use lowercase "parent" with map[string]string{"key": "PROJ-456"}
	Parent *IssueRef `json:"parent,omitempty"`

	// Version Fields
	// ==============

	// FixVersions are versions where the issue is/will be fixed (optional).
	// Used for release planning and tracking which version contains the fix.
	//
	// Example: []*project.Version{{Name: "2.0.0"}, {Name: "2.1.0"}}
	//
	// Common usage:
	//   - Set when planning which release will contain the fix
	//   - Can have multiple versions for backporting
	//   - Updated as release plans change
	//
	// Access: Use issue.GetFixVersions() (safe, never nil but may be empty)
	// Update: Use lowercase "fixVersions" with []map[string]string{{"name": "2.0.0"}}
	FixVersions []*project.Version `json:"fixVersions,omitempty"`

	// AffectsVersions are versions where the issue exists (optional, typically for bugs).
	// Indicates which versions have this problem, helps determine backporting needs.
	//
	// Example: []*project.Version{{Name: "1.0.0"}, {Name: "1.1.0"}}
	//
	// Common usage:
	//   - Set when reporting bugs to indicate affected releases
	//   - Helps prioritize fixes for supported versions
	//   - Used to determine which versions need backports
	//
	// Access: Use issue.GetAffectsVersions() (safe, never nil but may be empty)
	// Update: Use lowercase "versions" with []map[string]string{{"name": "1.0.0"}}
	AffectsVersions []*project.Version `json:"versions,omitempty"`

	// Date/Time Fields
	// ================
	// IMPORTANT: All date/time fields are pointers and may be nil.
	// ALWAYS use the safe accessor methods to avoid nil pointer panics!

	// Created is when the issue was created (read-only, set automatically).
	//
	// Access: Use issue.GetCreated() or issue.GetCreatedTime() (safe)
	// ❌ NEVER: issue.Fields.Created (can panic if nil)
	// ✅ ALWAYS: issue.GetCreatedTime() (returns zero time if nil)
	Created *time.Time `json:"created,omitempty"`

	// Updated is when the issue was last modified (read-only, set automatically).
	//
	// Access: Use issue.GetUpdated() or issue.GetUpdatedTime() (safe)
	// ❌ NEVER: issue.Fields.Updated (can panic if nil)
	// ✅ ALWAYS: issue.GetUpdatedTime() (returns zero time if nil)
	Updated *time.Time `json:"updated,omitempty"`

	// DueDate is the deadline for completing the issue (optional).
	//
	// Example: Set to 7 days from now:
	//   dueDate := time.Now().Add(7 * 24 * time.Hour)
	//   fields.DueDate = &dueDate
	//
	// Access: Use issue.GetDueDate() or issue.GetDueDateValue() (safe)
	// Update: Use lowercase "duedate" with RFC3339 formatted string
	// ❌ NEVER: issue.Fields.DueDate (can panic if nil)
	// ✅ ALWAYS: issue.GetDueDateValue() (returns zero time if nil)
	DueDate *time.Time `json:"duedate,omitempty"`

	// Organization Fields
	// ===================

	// Labels are tags for categorizing and filtering issues (optional).
	// Example: []string{"backend", "security", "urgent"}
	//
	// Access: Direct access is safe (slice type, never nil but may be empty)
	// Update: Use lowercase "labels" with []string{"label1", "label2"}
	Labels []string `json:"labels,omitempty"`

	// Components are project-specific groupings (optional).
	// Used to organize issues within a project (e.g., "API", "UI", "Database").
	//
	// Example: []*Component{{Name: "Backend API"}}
	//
	// Access: Direct access is safe (slice type, never nil but may be empty)
	// Update: Use lowercase "components" with []map[string]string{{"name": "API"}}
	Components []*Component `json:"components,omitempty"`

	// Advanced Fields
	// ===============

	// Custom contains type-safe custom field values.
	// Use the CustomFields methods for setting and getting values.
	//
	// Example:
	//   fields.Custom = CustomFields{
	//       "customfield_10001": "Custom value",
	//   }
	Custom CustomFields `json:"-"`

	// UnknownFields stores any additional fields from the API response.
	// This includes both custom fields and any future fields added by Jira.
	// Used for forward compatibility and dynamic field handling.
	UnknownFields map[string]interface{} `json:"-"`
}

// SetDescriptionText is a convenience method to set the description from plain text.
// It automatically converts the text to ADF format required by Jira Cloud API v3.
//
// Example:
//
//	fields := &issue.IssueFields{
//		Summary: "My Issue",
//	}
//	fields.SetDescriptionText("This is the issue description.\n\nMultiple paragraphs supported.")
func (f *IssueFields) SetDescriptionText(text string) {
	f.Description = ADFFromText(text)
}

// SetDescription sets the description using an ADF document.
//
// Example:
//
//	adf := issue.NewADF().
//		AddHeading("Problem", 2).
//		AddParagraph("The application crashes when...").
//		AddBulletList([]string{"Step 1", "Step 2", "Step 3"})
//	fields.SetDescription(adf)
func (f *IssueFields) SetDescription(adf *ADF) {
	f.Description = adf
}

// SetEnvironmentText is a convenience method to set the environment from plain text.
// It automatically converts the text to ADF format required by Jira Cloud API v3.
//
// Example:
//
//	fields := &issue.IssueFields{
//		Summary: "Production Issue",
//	}
//	fields.SetEnvironmentText("Production server: web-prod-01\nOS: Ubuntu 22.04")
func (f *IssueFields) SetEnvironmentText(text string) {
	f.Environment = ADFFromText(text)
}

// SetEnvironment sets the environment using an ADF document.
//
// Example:
//
//	adf := issue.NewADF().
//		AddHeading("Server Environment", 3).
//		AddBulletList([]string{
//			"Server: web-prod-01",
//			"OS: Ubuntu 22.04",
//			"Database: PostgreSQL 15",
//		})
//	fields.SetEnvironment(adf)
func (f *IssueFields) SetEnvironment(adf *ADF) {
	f.Environment = adf
}

// MarshalJSON implements custom JSON marshaling for IssueFields.
// It merges standard fields with custom fields for API requests.
func (f *IssueFields) MarshalJSON() ([]byte, error) {
	// Create a map with standard fields
	type Alias IssueFields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	// Marshal standard fields
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// If no custom fields, return standard fields
	if len(f.Custom) == 0 {
		return data, nil
	}

	// Unmarshal to map to merge with custom fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Merge custom fields
	for fieldID, field := range f.Custom {
		result[fieldID] = field.Value
	}

	return json.Marshal(result)
}

// tryParseDateTime attempts to intelligently parse a string value as a date/time.
// It returns (parsedTime, true) if successful, (zero, false) if not a date/time string.
//
// Jira returns dates in various formats:
//   - Date only: "2025-10-30"
//   - DateTime with timezone: "2024-01-01T10:30:00.000+0000" (non-standard)
//   - RFC3339: "2024-01-01T10:30:00.000Z"
//   - Time only: "15:30:00" (for time-tracking fields)
func tryParseDateTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	// List of formats to try, in order of likelihood
	formats := []string{
		"2006-01-02",                   // Date only (YYYY-MM-DD)
		"2006-01-02T15:04:05.000-0700", // Jira format with timezone
		time.RFC3339,                   // Standard RFC3339
		"2006-01-02T15:04:05Z",         // RFC3339 without milliseconds
		time.RFC3339Nano,               // RFC3339 with nanoseconds
		"15:04:05",                     // Time only (HH:MM:SS)
		"15:04",                        // Time without seconds (HH:MM)
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

// normalizeFieldValue attempts to normalize a field value, converting date/time strings
// to a format that Go's time.Time can unmarshal.
// Returns the original value if it's not a date/time string.
func normalizeFieldValue(value interface{}) interface{} {
	// Only process string values
	str, ok := value.(string)
	if !ok || str == "" {
		return value
	}

	// Try to parse as date/time
	if parsed, isDateTime := tryParseDateTime(str); isDateTime {
		// Convert to RFC3339 format which time.Time can unmarshal
		return parsed.Format(time.RFC3339)
	}

	return value
}

// UnmarshalJSON implements custom JSON unmarshaling for IssueFields.
// It handles flexible date/time formats from Jira and extracts custom fields.
func (f *IssueFields) UnmarshalJSON(data []byte) error {
	// First, unmarshal to map for preprocessing
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Normalize all field values - this handles both standard and custom date fields
	// by converting non-standard date formats to RFC3339
	for key, value := range raw {
		raw[key] = normalizeFieldValue(value)
	}

	// Re-marshal the normalized data
	normalizedData, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	// Unmarshal standard fields with type alias to avoid recursion
	type Alias IssueFields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(normalizedData, aux); err != nil {
		return err
	}

	// Initialize Custom if nil
	if f.Custom == nil {
		f.Custom = NewCustomFields()
	}

	// Extract custom fields (fields starting with "customfield_")
	// The values have already been normalized, so custom date fields will work
	for key, value := range raw {
		if strings.HasPrefix(key, "customfield_") {
			f.Custom[key] = &CustomField{
				ID:    key,
				Value: value,
			}
		}
	}

	return nil
}

// IssueType represents an issue type.
type IssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// Project represents a Jira project.
type Project struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// Status represents an issue status.
type Status struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Category    *StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a status category.
type StatusCategory struct {
	ID        int    `json:"id,omitempty"`
	Key       string `json:"key,omitempty"`
	Name      string `json:"name,omitempty"`
	ColorName string `json:"colorName,omitempty"`
}

// Priority represents an issue priority.
type Priority struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	IconURL string `json:"iconUrl,omitempty"`
}

// User represents a Jira user.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	TimeZone     string `json:"timeZone,omitempty"`
	Self         string `json:"self,omitempty"`
}

// Component represents a project component.
type Component struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

// GetOptions configures the Get operation.
type GetOptions struct {
	// Fields specifies which fields to include in the response
	Fields []string

	// Expand specifies which additional information to include
	Expand []string

	// Properties specifies which properties to include
	Properties []string
}

// Get retrieves an issue by key or ID.
//
// This method fetches an issue from Jira and returns a strongly-typed Issue struct.
// Always use the safe accessor methods when reading issue data to prevent nil pointer panics.
//
// Basic Example:
//
//	// Get an issue with all default fields
//	issue, err := client.Issue.Get(ctx, "PROJ-123", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// ✅ SAFE - Use accessor methods to read data
//	fmt.Printf("Summary: %s\n", issue.GetSummary())
//	fmt.Printf("Status: %s\n", issue.GetStatusName())
//	fmt.Printf("Assignee: %s\n", issue.GetAssigneeName())
//
// Get specific fields only (improves performance):
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123", &issue.GetOptions{
//	    Fields: []string{"summary", "status", "priority", "assignee"},
//	})
//
// Working with versions and resolutions:
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123", &issue.GetOptions{
//	    Fields: []string{"versions", "fixVersions", "resolution"},
//	})
//
//	// ✅ Safe - Use accessor methods
//	affectsVersions := issue.GetAffectsVersions()  // Never nil, but may be empty
//	fixVersions := issue.GetFixVersions()          // Never nil, but may be empty
//	resolutionName := issue.GetResolutionName()    // Returns "" if unresolved
//
//	for _, v := range affectsVersions {
//	    fmt.Printf("Affects version: %s\n", v.Name)
//	}
//
// Working with subtasks and parent issues:
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123", &issue.GetOptions{
//	    Fields: []string{"parent", "issuetype"},
//	})
//
//	// Check if it's a subtask and get parent
//	if parentKey := issue.GetParentKey(); parentKey != "" {
//	    fmt.Printf("This is a subtask of %s\n", parentKey)
//	}
//
// ❌ Common Mistakes to Avoid:
//
//	// DON'T: Direct field access (can panic!)
//	// summary := issue.Fields.Summary      // Panics if Fields is nil
//	// status := issue.Fields.Status.Name   // Panics if Status is nil
//
//	// DO: Use safe accessor methods
//	summary := issue.GetSummary()           // Returns "" if Fields is nil
//	status := issue.GetStatusName()         // Returns "" if Status is nil
func (s *Service) Get(ctx context.Context, issueKeyOrID string, opts *GetOptions) (*Issue, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s", issueKeyOrID)

	// TODO: Add query parameters from opts (fields, expand, properties)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var issue Issue
	if err := s.transport.DecodeResponse(resp, &issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// CreateInput contains the data for creating an issue.
type CreateInput struct {
	Fields *IssueFields `json:"fields"`
}

// Create creates a new issue in Jira.
//
// Required fields for all issues:
//   - Project: Reference to the project (use &Project{Key: "PROJ"})
//   - Summary: Short description of the issue
//   - IssueType: Type of issue (use &IssueType{Name: "Bug"}, "Story", "Task", etc.)
//
// Additional required fields for specific types:
//   - Parent: Required for subtasks (use &IssueRef{Key: "PARENT-KEY"})
//
// Basic Example - Create a simple task:
//
//	created, err := client.Issue.Create(ctx, &issue.CreateInput{
//	    Fields: &issue.IssueFields{
//	        Project:   &issue.Project{Key: "PROJ"},
//	        Summary:   "Implement user authentication",
//	        IssueType: &issue.IssueType{Name: "Task"},
//	        Priority:  &issue.Priority{Name: "High"},
//	    },
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created: %s\n", created.Key)
//
// Create a bug with version tracking:
//
//	fields := &issue.IssueFields{
//	    Project:   &issue.Project{Key: "PROJ"},
//	    Summary:   "Critical login bug",
//	    IssueType: &issue.IssueType{Name: "Bug"},
//	    Priority:  &issue.Priority{Name: "Highest"},
//
//	    // Indicate which versions have this bug
//	    AffectsVersions: []*project.Version{
//	        {Name: "1.0.0"},
//	        {Name: "1.1.0"},
//	    },
//
//	    // Indicate which version will fix it
//	    FixVersions: []*project.Version{
//	        {Name: "1.2.0"},
//	    },
//	}
//	// Set description as plain text
//	fields.SetDescriptionText("Users cannot log in after the latest deployment.")
//	fields.SetEnvironmentText("Production: server-01, Ubuntu 22.04")
//
//	bug, err := client.Issue.Create(ctx, &issue.CreateInput{Fields: fields})
//
// Create a subtask (IMPORTANT: Must set both IssueType and Parent):
//
//	subtask, err := client.Issue.Create(ctx, &issue.CreateInput{
//	    Fields: &issue.IssueFields{
//	        Project:   &issue.Project{Key: "PROJ"},
//	        Summary:   "Implement login API endpoint",
//	        IssueType: &issue.IssueType{Name: "Sub-task"},  // Critical!
//	        Parent:    &issue.IssueRef{Key: "PROJ-123"},    // Link to parent
//	        Priority:  &issue.Priority{Name: "Medium"},
//	    },
//	})
//
// Create with rich formatted description using ADF:
//
//	adf := issue.NewADF().
//	    AddHeading("Problem Description", 2).
//	    AddParagraph("The authentication service fails when users try to log in.").
//	    AddHeading("Steps to Reproduce", 2).
//	    AddBulletList([]string{
//	        "Navigate to login page",
//	        "Enter valid credentials",
//	        "Click submit button",
//	    })
//
//	fields := &issue.IssueFields{
//	    Project:   &issue.Project{Key: "PROJ"},
//	    Summary:   "Login service failure",
//	    IssueType: &issue.IssueType{Name: "Bug"},
//	}
//	fields.SetDescription(adf)
//
//	created, err := client.Issue.Create(ctx, &issue.CreateInput{Fields: fields})
//
// Create with labels and components:
//
//	created, err := client.Issue.Create(ctx, &issue.CreateInput{
//	    Fields: &issue.IssueFields{
//	        Project:    &issue.Project{Key: "PROJ"},
//	        Summary:    "Implement caching layer",
//	        IssueType:  &issue.IssueType{Name: "Story"},
//	        Labels:     []string{"performance", "backend"},
//	        Components: []*issue.Component{{Name: "API"}},
//	    },
//	})
//
// ❌ Common Mistakes to Avoid:
//
//	// DON'T: Forget required fields
//	// created, err := client.Issue.Create(ctx, &issue.CreateInput{
//	//     Fields: &issue.IssueFields{
//	//         Summary: "Bug fix",  // Missing Project and IssueType!
//	//     },
//	// })
//
//	// DON'T: Forget IssueType when creating subtasks
//	// subtask, err := client.Issue.Create(ctx, &issue.CreateInput{
//	//     Fields: &issue.IssueFields{
//	//         Project: &issue.Project{Key: "PROJ"},
//	//         Summary: "Subtask",
//	//         Parent:  &issue.IssueRef{Key: "PROJ-123"},  // Missing IssueType!
//	//     },
//	// })
//
//	// DO: Always include all required fields
//	created, err := client.Issue.Create(ctx, &issue.CreateInput{
//	    Fields: &issue.IssueFields{
//	        Project:   &issue.Project{Key: "PROJ"},
//	        Summary:   "Fix critical bug",
//	        IssueType: &issue.IssueType{Name: "Bug"},  // Required!
//	    },
//	})
func (s *Service) Create(ctx context.Context, input *CreateInput) (*Issue, error) {
	if input == nil || input.Fields == nil {
		return nil, fmt.Errorf("create input is required")
	}

	// Validate required fields
	if input.Fields.Project == nil {
		return nil, fmt.Errorf("project is required")
	}
	if input.Fields.Summary == "" {
		return nil, fmt.Errorf("summary is required")
	}
	if input.Fields.IssueType == nil {
		return nil, fmt.Errorf("issue type is required")
	}

	path := "/rest/api/3/issue"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return created issue with basic info
	// A full GET request would be needed to retrieve all fields
	return &Issue{
		ID:   result.ID,
		Key:  result.Key,
		Self: result.Self,
	}, nil
}

// UpdateInput contains the data for updating an issue.
type UpdateInput struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// Update updates an existing issue.
//
// IMPORTANT: When updating issues, you must use lowercase field names (Jira API convention)
// and wrap object references in nested maps.
//
// Basic Example - Update simple fields:
//
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "summary":  "Updated issue summary",
//	        "priority": map[string]string{"name": "Critical"},
//	        "labels":   []string{"bug", "production", "urgent"},
//	    },
//	})
//
// Update version fields:
//
//	// Set fix versions (which releases will contain the fix)
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "fixVersions": []map[string]string{
//	            {"name": "2.0.0"},
//	            {"name": "2.1.0"},  // Backport to multiple versions
//	        },
//	    },
//	})
//
//	// Set affected versions (which versions have the bug)
//	err = client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "versions": []map[string]string{  // Note: "versions", not "affectsVersions"
//	            {"name": "1.0.0"},
//	            {"name": "1.1.0"},
//	        },
//	    },
//	})
//
// Set resolution when closing an issue:
//
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "resolution": map[string]string{
//	            "name": "Done",  // or "Won't Fix", "Duplicate", etc.
//	        },
//	    },
//	})
//
// Move a subtask to a different parent:
//
//	err := client.Issue.Update(ctx, "PROJ-456", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "parent": map[string]string{
//	            "key": "PROJ-789",  // New parent issue key
//	        },
//	    },
//	})
//
// Update multiple fields at once:
//
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "summary":     "Critical production bug",
//	        "priority":    map[string]string{"name": "Highest"},
//	        "fixVersions": []map[string]string{{"name": "1.2.0"}},
//	        "resolution":  map[string]string{"name": "Done"},
//	        "labels":      []string{"security", "urgent"},
//	    },
//	})
//
// Assign issue to a user:
//
//	// Jira Cloud - Use accountId
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "assignee": map[string]string{
//	            "accountId": "5b10a2844c20165700ede21g",
//	        },
//	    },
//	})
//
//	// Jira Server/DC - Use name
//	err = client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "assignee": map[string]string{
//	            "name": "jsmith",
//	        },
//	    },
//	})
//
// Update components:
//
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "components": []map[string]string{
//	            {"name": "Backend API"},
//	            {"name": "Database"},
//	        },
//	    },
//	})
//
// ❌ Common Mistakes to Avoid:
//
//	// DON'T: Use uppercase field names
//	// err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	//     Fields: map[string]interface{}{
//	//         "Summary": "Wrong",  // Wrong! Use lowercase "summary"
//	//     },
//	// })
//
//	// DON'T: Use direct string values for object references
//	// err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	//     Fields: map[string]interface{}{
//	//         "priority": "High",  // Wrong! Use map[string]string{"name": "High"}
//	//     },
//	// })
//
//	// DON'T: Use "affectsVersions" when updating
//	// err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	//     Fields: map[string]interface{}{
//	//         "affectsVersions": [...],  // Wrong! Use "versions"
//	//     },
//	// })
//
//	// DO: Use lowercase field names and proper nested maps
//	err := client.Issue.Update(ctx, "PROJ-123", &issue.UpdateInput{
//	    Fields: map[string]interface{}{
//	        "summary":  "Correct",
//	        "priority": map[string]string{"name": "High"},
//	        "versions": []map[string]string{{"name": "1.0.0"}},
//	    },
//	})
func (s *Service) Update(ctx context.Context, issueKeyOrID string, input *UpdateInput) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if input == nil {
		return fmt.Errorf("update input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Close response body
	defer resp.Body.Close()

	// Update returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Delete deletes an issue.
//
// Example:
//
//	err := client.Issue.Delete(ctx, "PROJ-123")
func (s *Service) Delete(ctx context.Context, issueKeyOrID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Close response body
	defer resp.Body.Close()

	// Delete returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// TransitionInput contains the data for transitioning an issue.
type TransitionInput struct {
	Transition *Transition            `json:"transition"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// Transition represents a workflow transition.
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// DoTransition transitions an issue to a new status.
//
// Example:
//
//	input := &issue.TransitionInput{
//		Transition: &issue.Transition{ID: "11"},
//	}
//	err := client.Issue.DoTransition(ctx, "PROJ-123", input)
func (s *Service) DoTransition(ctx context.Context, issueKeyOrID string, input *TransitionInput) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if input == nil || input.Transition == nil {
		return fmt.Errorf("transition input is required")
	}

	if input.Transition.ID == "" {
		return fmt.Errorf("transition ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Close response body
	defer resp.Body.Close()

	// Transition returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
