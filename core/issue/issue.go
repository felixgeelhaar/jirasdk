// Package issue provides Issue resource management for Jira.
package issue

import (
	"context"
	"fmt"
	"net/http"
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
	ID     string        `json:"id"`
	Key    string        `json:"key"`
	Self   string        `json:"self"`
	Fields *IssueFields  `json:"fields,omitempty"`
	Expand string        `json:"expand,omitempty"`
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
	// Custom fields are handled as map[string]interface{}
	// Users can type assert to specific types as needed
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
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
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
