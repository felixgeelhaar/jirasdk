// Package issue provides Issue resource management for Jira.
package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
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

// GetDescription safely retrieves the issue description.
// Returns an empty string if Fields is nil.
func (i *Issue) GetDescription() string {
	if i.Fields == nil {
		return ""
	}
	return i.Fields.Description
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

// IssueFields contains the fields of an issue.
type IssueFields struct {
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	IssueType   *IssueType   `json:"issuetype,omitempty"`
	Project     *Project     `json:"project,omitempty"`
	Status      *Status      `json:"status,omitempty"`
	Priority    *Priority    `json:"priority,omitempty"`
	Assignee    *User        `json:"assignee,omitempty"`
	Reporter    *User        `json:"reporter,omitempty"`
	Created     *time.Time   `json:"created,omitempty"`
	Updated     *time.Time   `json:"updated,omitempty"`
	DueDate     *time.Time   `json:"duedate,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	Components  []*Component `json:"components,omitempty"`

	// Custom contains custom field values
	// Use the type-safe CustomFields methods for setting/getting values
	Custom CustomFields `json:"-"`

	// UnknownFields stores any additional fields from the API response
	// This includes both custom fields and any future fields added by Jira
	UnknownFields map[string]interface{} `json:"-"`
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

// UnmarshalJSON implements custom JSON unmarshaling for IssueFields.
// It extracts custom fields from the API response.
func (f *IssueFields) UnmarshalJSON(data []byte) error {
	// Unmarshal standard fields
	type Alias IssueFields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Unmarshal to map to extract custom fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Initialize Custom if nil
	if f.Custom == nil {
		f.Custom = NewCustomFields()
	}

	// Extract custom fields (fields starting with "customfield_")
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
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// Project represents a Jira project.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// Status represents an issue status.
type Status struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Category    *StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a status category.
type StatusCategory struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	ColorName string `json:"colorName,omitempty"`
}

// Priority represents an issue priority.
type Priority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
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
	ID          string `json:"id"`
	Name        string `json:"name"`
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
// Example:
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123", nil)
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

// Create creates a new issue.
//
// Example:
//
//	input := &issue.CreateInput{
//		Fields: &issue.IssueFields{
//			Project:   &issue.Project{Key: "PROJ"},
//			Summary:   "New issue",
//			IssueType: &issue.IssueType{Name: "Task"},
//		},
//	}
//	created, err := client.Issue.Create(ctx, input)
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
// Example:
//
//	input := &issue.UpdateInput{
//		Fields: map[string]interface{}{
//			"summary": "Updated summary",
//		},
//	}
//	err := client.Issue.Update(ctx, "PROJ-123", input)
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
