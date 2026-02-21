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

// IssueTypeScheme represents an issue type scheme.
type IssueTypeScheme struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	DefaultIssueTypeID string `json:"defaultIssueTypeId,omitempty"`
	IsDefault          bool   `json:"isDefault,omitempty"`
}

// IssueTypeSchemeMapping represents a mapping between an issue type scheme and an issue type.
type IssueTypeSchemeMapping struct {
	IssueTypeSchemeID string `json:"issueTypeSchemeId"`
	IssueTypeID       string `json:"issueTypeId"`
}

// CreateIssueTypeSchemeInput represents input for creating an issue type scheme.
type CreateIssueTypeSchemeInput struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	DefaultIssueTypeID string   `json:"defaultIssueTypeId,omitempty"`
	IssueTypeIDs       []string `json:"issueTypeIds"`
}

// UpdateIssueTypeSchemeInput represents input for updating an issue type scheme.
type UpdateIssueTypeSchemeInput struct {
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	DefaultIssueTypeID string `json:"defaultIssueTypeId,omitempty"`
}

// AddIssueTypesToSchemeInput represents input for adding issue types to a scheme.
type AddIssueTypesToSchemeInput struct {
	IssueTypeIDs []string `json:"issueTypeIds"`
}

// ListIssueTypeSchemesOptions configures the ListIssueTypeSchemes operation.
type ListIssueTypeSchemesOptions struct {
	StartAt    int
	MaxResults int
}

// GetIssueTypeSchemeMappingsOptions configures the GetIssueTypeSchemeMappings operation.
type GetIssueTypeSchemeMappingsOptions struct {
	IssueTypeSchemeIDs []string
	StartAt            int
	MaxResults         int
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

// ListIssueTypeSchemes retrieves all issue type schemes.
//
// Example:
//
//	schemes, err := client.IssueType.ListIssueTypeSchemes(ctx, nil)
func (s *Service) ListIssueTypeSchemes(ctx context.Context, opts *ListIssueTypeSchemesOptions) ([]*IssueTypeScheme, error) {
	path := "/rest/api/3/issuetypescheme"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if opts != nil {
		q := req.URL.Query()

		if opts.StartAt > 0 {
			q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
		}

		if opts.MaxResults > 0 {
			q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}

		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*IssueTypeScheme `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// CreateIssueTypeScheme creates a new issue type scheme.
//
// This is needed because creating issue types no longer auto-adds them to
// the Default Work Type Scheme (Jira Cloud CHANGE-2999/3000, February 2026).
//
// Example:
//
//	scheme, err := client.IssueType.CreateIssueTypeScheme(ctx, &issuetype.CreateIssueTypeSchemeInput{
//	    Name:               "Software Development",
//	    Description:        "Issue types for software projects",
//	    DefaultIssueTypeID: "10001",
//	    IssueTypeIDs:       []string{"10001", "10002", "10003"},
//	})
func (s *Service) CreateIssueTypeScheme(ctx context.Context, input *CreateIssueTypeSchemeInput) (*IssueTypeScheme, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("scheme name is required")
	}

	path := "/rest/api/3/issuetypescheme"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var scheme IssueTypeScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// UpdateIssueTypeScheme updates an existing issue type scheme.
//
// Example:
//
//	err := client.IssueType.UpdateIssueTypeScheme(ctx, "10000", &issuetype.UpdateIssueTypeSchemeInput{
//	    Name:        "Updated Scheme",
//	    Description: "Updated description",
//	})
func (s *Service) UpdateIssueTypeScheme(ctx context.Context, schemeID string, input *UpdateIssueTypeSchemeInput) error {
	if schemeID == "" {
		return fmt.Errorf("scheme ID is required")
	}

	if input == nil {
		return fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetypescheme/%s", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// DeleteIssueTypeScheme deletes an issue type scheme.
//
// Example:
//
//	err := client.IssueType.DeleteIssueTypeScheme(ctx, "10000")
func (s *Service) DeleteIssueTypeScheme(ctx context.Context, schemeID string) error {
	if schemeID == "" {
		return fmt.Errorf("scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetypescheme/%s", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// AddIssueTypesToScheme adds issue types to an issue type scheme.
//
// Example:
//
//	err := client.IssueType.AddIssueTypesToScheme(ctx, "10000", &issuetype.AddIssueTypesToSchemeInput{
//	    IssueTypeIDs: []string{"10004", "10005"},
//	})
func (s *Service) AddIssueTypesToScheme(ctx context.Context, schemeID string, input *AddIssueTypesToSchemeInput) error {
	if schemeID == "" {
		return fmt.Errorf("scheme ID is required")
	}

	if input == nil || len(input.IssueTypeIDs) == 0 {
		return fmt.Errorf("at least one issue type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetypescheme/%s/issuetype", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// RemoveIssueTypeFromScheme removes an issue type from an issue type scheme.
//
// Example:
//
//	err := client.IssueType.RemoveIssueTypeFromScheme(ctx, "10000", "10004")
func (s *Service) RemoveIssueTypeFromScheme(ctx context.Context, schemeID, issueTypeID string) error {
	if schemeID == "" {
		return fmt.Errorf("scheme ID is required")
	}

	if issueTypeID == "" {
		return fmt.Errorf("issue type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuetypescheme/%s/issuetype/%s", schemeID, issueTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// GetIssueTypeSchemeMappings retrieves the issue type scheme to issue type mappings.
//
// Example:
//
//	mappings, err := client.IssueType.GetIssueTypeSchemeMappings(ctx, &issuetype.GetIssueTypeSchemeMappingsOptions{
//	    IssueTypeSchemeIDs: []string{"10000", "10001"},
//	})
func (s *Service) GetIssueTypeSchemeMappings(ctx context.Context, opts *GetIssueTypeSchemeMappingsOptions) ([]*IssueTypeSchemeMapping, error) {
	path := "/rest/api/3/issuetypescheme/mapping"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if opts != nil {
		q := req.URL.Query()

		for _, id := range opts.IssueTypeSchemeIDs {
			q.Add("issueTypeSchemeId", id)
		}

		if opts.StartAt > 0 {
			q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
		}

		if opts.MaxResults > 0 {
			q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}

		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*IssueTypeSchemeMapping `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}
