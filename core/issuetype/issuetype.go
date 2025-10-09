// Package issuetype provides Issue Type resource management for Jira.
//
// Issue types categorize different types of work items (e.g., Bug, Story, Task).
// This package provides operations for managing issue types and their schemes.
package issuetype

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Issue Type resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Issue Type service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// IssueType represents a Jira issue type.
type IssueType struct {
	ID             string `json:"id"`
	Self           string `json:"self,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	IconURL        string `json:"iconUrl,omitempty"`
	Subtask        bool   `json:"subtask"`
	AvatarID       int64  `json:"avatarId,omitempty"`
	HierarchyLevel int    `json:"hierarchyLevel,omitempty"`
	Scope          *Scope `json:"scope,omitempty"`
}

// Scope represents the scope of an issue type.
type Scope struct {
	Type    string   `json:"type"`
	Project *Project `json:"project,omitempty"`
}

// Project represents a simplified Jira project.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

// CreateIssueTypeInput represents input for creating an issue type.
type CreateIssueTypeInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"` // "subtask" or "standard"
}

// UpdateIssueTypeInput represents input for updating an issue type.
type UpdateIssueTypeInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	AvatarID    int64  `json:"avatarId,omitempty"`
}

// List retrieves all issue types.
//
// Example:
//
//	issueTypes, err := client.IssueType.List(ctx)
func (s *Service) List(ctx context.Context) ([]*IssueType, error) {
	path := "/rest/api/3/issuetype"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var issueTypes []*IssueType
	if err := s.transport.DecodeResponse(resp, &issueTypes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return issueTypes, nil
}

// Get retrieves a specific issue type by ID.
//
// Example:
//
//	issueType, err := client.IssueType.Get(ctx, "10001")
func (s *Service) Get(ctx context.Context, issueTypeID string) (*IssueType, error) {
	if issueTypeID == "" {
		return nil, fmt.Errorf("issue type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetype/%s", issueTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var issueType IssueType
	if err := s.transport.DecodeResponse(resp, &issueType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issueType, nil
}

// Create creates a new issue type.
//
// Example:
//
//	issueType, err := client.IssueType.Create(ctx, &issuetype.CreateIssueTypeInput{
//	    Name:        "Incident",
//	    Description: "Production incident",
//	    Type:        "standard",
//	})
func (s *Service) Create(ctx context.Context, input *CreateIssueTypeInput) (*IssueType, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("issue type name is required")
	}

	path := "/rest/api/3/issuetype"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var issueType IssueType
	if err := s.transport.DecodeResponse(resp, &issueType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issueType, nil
}

// Update updates an existing issue type.
//
// Example:
//
//	issueType, err := client.IssueType.Update(ctx, "10001", &issuetype.UpdateIssueTypeInput{
//	    Name:        "Updated Incident",
//	    Description: "Updated production incident type",
//	})
func (s *Service) Update(ctx context.Context, issueTypeID string, input *UpdateIssueTypeInput) (*IssueType, error) {
	if issueTypeID == "" {
		return nil, fmt.Errorf("issue type ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetype/%s", issueTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var issueType IssueType
	if err := s.transport.DecodeResponse(resp, &issueType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issueType, nil
}

// Delete deletes an issue type.
//
// Example:
//
//	err := client.IssueType.Delete(ctx, "10001", "10002")
func (s *Service) Delete(ctx context.Context, issueTypeID, alternativeIssueTypeID string) error {
	if issueTypeID == "" {
		return fmt.Errorf("issue type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetype/%s", issueTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add alternative issue type ID if provided
	if alternativeIssueTypeID != "" {
		q := req.URL.Query()
		q.Set("alternativeIssueTypeId", alternativeIssueTypeID)
		req.URL.RawQuery = q.Encode()
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
