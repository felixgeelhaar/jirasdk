// Package project provides Project resource management for Jira.
package project

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/felixgeelhaar/jirasdk/internal/pagination"
)

// Service provides operations for Project resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Project service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Project represents a Jira project.
type Project struct {
	ID              string           `json:"id"`
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	Description     string           `json:"description,omitempty"`
	Self            string           `json:"self,omitempty"`
	ProjectTypeKey  string           `json:"projectTypeKey,omitempty"`
	Lead            *User            `json:"lead,omitempty"`
	AvatarURLs      *AvatarURLs      `json:"avatarUrls,omitempty"`
	IssueTypes      []*IssueType     `json:"issueTypes,omitempty"`
	Components      []*Component     `json:"components,omitempty"`
	Versions        []*Version       `json:"versions,omitempty"`
	Archived        bool             `json:"archived,omitempty"`
	Deleted         bool             `json:"deleted,omitempty"`
	Simplified      bool             `json:"simplified,omitempty"`
	Style           string           `json:"style,omitempty"`
	Insight         *ProjectInsight  `json:"insight,omitempty"`
}

// User represents a Jira user.
type User struct {
	Self         string      `json:"self,omitempty"`
	AccountID    string      `json:"accountId,omitempty"`
	EmailAddress string      `json:"emailAddress,omitempty"`
	DisplayName  string      `json:"displayName,omitempty"`
	Active       bool        `json:"active,omitempty"`
	TimeZone     string      `json:"timeZone,omitempty"`
	AccountType  string      `json:"accountType,omitempty"`
	AvatarURLs   *AvatarURLs `json:"avatarUrls,omitempty"`
}

// AvatarURLs contains URLs for different sizes of avatars.
type AvatarURLs struct {
	Size16 string `json:"16x16,omitempty"`
	Size24 string `json:"24x24,omitempty"`
	Size32 string `json:"32x32,omitempty"`
	Size48 string `json:"48x48,omitempty"`
}

// IssueType represents an issue type in a project.
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// Component represents a project component.
type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lead        *User  `json:"lead,omitempty"`
}

// Version represents a project version.
type Version struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Archived    bool       `json:"archived,omitempty"`
	Released    bool       `json:"released,omitempty"`
	StartDate   string     `json:"startDate,omitempty"`
	ReleaseDate string     `json:"releaseDate,omitempty"`
}

// ProjectInsight contains project insights.
type ProjectInsight struct {
	TotalIssueCount     int        `json:"totalIssueCount,omitempty"`
	LastIssueUpdateTime *time.Time `json:"lastIssueUpdateTime,omitempty"`
}

// GetOptions configures the Get operation.
type GetOptions struct {
	// Expand specifies additional information to include
	Expand []string
}

// ListOptions configures the List operation.
type ListOptions struct {
	// Expand specifies additional information to include
	Expand []string

	// Recent limits to projects the user has recently interacted with
	Recent int

	// Properties specifies project properties to include
	Properties []string

	pagination.Options
}

// Get retrieves a project by key or ID.
//
// Example:
//
//	project, err := client.Project.Get(ctx, "PROJ", &project.GetOptions{
//		Expand: []string{"lead", "issueTypes", "description"},
//	})
func (s *Service) Get(ctx context.Context, projectKeyOrID string, opts *GetOptions) (*Project, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s", projectKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add expand query parameter
	if opts != nil && len(opts.Expand) > 0 {
		q := req.URL.Query()
		for _, expand := range opts.Expand {
			q.Add("expand", expand)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var project Project
	if err := s.transport.DecodeResponse(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// List retrieves all projects.
//
// Example:
//
//	projects, err := client.Project.List(ctx, &project.ListOptions{
//		Expand: []string{"lead", "description"},
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Project, error) {
	path := "/rest/api/3/project/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()

		if len(opts.Expand) > 0 {
			for _, expand := range opts.Expand {
				q.Add("expand", expand)
			}
		}

		if opts.Recent > 0 {
			q.Set("recent", fmt.Sprintf("%d", opts.Recent))
		}

		if len(opts.Properties) > 0 {
			for _, prop := range opts.Properties {
				q.Add("properties", prop)
			}
		}

		opts.Options.ApplyToURL(req.URL)
		req.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result struct {
		Values     []*Project          `json:"values"`
		StartAt    int                 `json:"startAt"`
		MaxResults int                 `json:"maxResults"`
		Total      int                 `json:"total"`
		IsLast     bool                `json:"isLast"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// CreateInput contains the data for creating a project.
type CreateInput struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	ProjectTypeKey  string `json:"projectTypeKey"`
	Description     string `json:"description,omitempty"`
	ProjectTemplate string `json:"projectTemplateKey,omitempty"`
	LeadAccountID   string `json:"leadAccountId,omitempty"`
	URL             string `json:"url,omitempty"`
	AssigneeType    string `json:"assigneeType,omitempty"`
}

// Create creates a new project.
//
// Example:
//
//	project, err := client.Project.Create(ctx, &project.CreateInput{
//		Key:            "NEWPROJ",
//		Name:           "New Project",
//		ProjectTypeKey: "software",
//		Description:    "A new software project",
//	})
func (s *Service) Create(ctx context.Context, input *CreateInput) (*Project, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Key == "" {
		return nil, fmt.Errorf("project key is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	if input.ProjectTypeKey == "" {
		return nil, fmt.Errorf("project type is required")
	}

	path := "/rest/api/3/project"

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
	var project Project
	if err := s.transport.DecodeResponse(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// UpdateInput contains the data for updating a project.
type UpdateInput struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	LeadAccountID string `json:"leadAccountId,omitempty"`
	URL           string `json:"url,omitempty"`
	AssigneeType  string `json:"assigneeType,omitempty"`
}

// Update updates an existing project.
//
// Example:
//
//	project, err := client.Project.Update(ctx, "PROJ", &project.UpdateInput{
//		Name:        "Updated Project Name",
//		Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, projectKeyOrID string, input *UpdateInput) (*Project, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s", projectKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var project Project
	if err := s.transport.DecodeResponse(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// Delete deletes a project.
//
// Example:
//
//	err := client.Project.Delete(ctx, "PROJ")
func (s *Service) Delete(ctx context.Context, projectKeyOrID string) error {
	if projectKeyOrID == "" {
		return fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s", projectKeyOrID)

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

// Archive archives a project.
//
// Example:
//
//	err := client.Project.Archive(ctx, "PROJ")
func (s *Service) Archive(ctx context.Context, projectKeyOrID string) error {
	if projectKeyOrID == "" {
		return fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/archive", projectKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, nil)
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

	// Archive returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Restore restores an archived project.
//
// Example:
//
//	err := client.Project.Restore(ctx, "PROJ")
func (s *Service) Restore(ctx context.Context, projectKeyOrID string) error {
	if projectKeyOrID == "" {
		return fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/restore", projectKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, nil)
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

	// Restore returns 200 OK on success
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Component Management

// CreateComponentInput contains the data for creating a component.
type CreateComponentInput struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	LeadAccountID   string `json:"leadAccountId,omitempty"`
	AssigneeType    string `json:"assigneeType,omitempty"` // "PROJECT_DEFAULT", "COMPONENT_LEAD", "PROJECT_LEAD", "UNASSIGNED"
	Project         string `json:"project"`                // Project ID or key
}

// CreateComponent creates a new component in a project.
//
// Example:
//
//	component, err := client.Project.CreateComponent(ctx, &project.CreateComponentInput{
//		Name:        "Frontend",
//		Description: "Frontend components",
//		Project:     "PROJ",
//	})
func (s *Service) CreateComponent(ctx context.Context, input *CreateComponentInput) (*Component, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("component name is required")
	}

	if input.Project == "" {
		return nil, fmt.Errorf("project is required")
	}

	path := "/rest/api/3/component"

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
	var component Component
	if err := s.transport.DecodeResponse(resp, &component); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &component, nil
}

// UpdateComponentInput contains the data for updating a component.
type UpdateComponentInput struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	LeadAccountID string `json:"leadAccountId,omitempty"`
	AssigneeType  string `json:"assigneeType,omitempty"`
}

// UpdateComponent updates an existing component.
//
// Example:
//
//	component, err := client.Project.UpdateComponent(ctx, "10000", &project.UpdateComponentInput{
//		Name:        "Updated Frontend",
//		Description: "Updated description",
//	})
func (s *Service) UpdateComponent(ctx context.Context, componentID string, input *UpdateComponentInput) (*Component, error) {
	if componentID == "" {
		return nil, fmt.Errorf("component ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/component/%s", componentID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var component Component
	if err := s.transport.DecodeResponse(resp, &component); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &component, nil
}

// GetComponent retrieves a component by ID.
//
// Example:
//
//	component, err := client.Project.GetComponent(ctx, "10000")
func (s *Service) GetComponent(ctx context.Context, componentID string) (*Component, error) {
	if componentID == "" {
		return nil, fmt.Errorf("component ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/component/%s", componentID)

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
	var component Component
	if err := s.transport.DecodeResponse(resp, &component); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &component, nil
}

// DeleteComponent deletes a component.
//
// Example:
//
//	err := client.Project.DeleteComponent(ctx, "10000")
func (s *Service) DeleteComponent(ctx context.Context, componentID string) error {
	if componentID == "" {
		return fmt.Errorf("component ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/component/%s", componentID)

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

// Version Management

// CreateVersionInput contains the data for creating a version.
type CreateVersionInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Project     string `json:"project"`            // Project ID or key (required for creation, becomes projectId in response)
	ProjectID   int64  `json:"projectId,omitempty"` // Returned by API, don't set on creation
	Archived    bool   `json:"archived,omitempty"`
	Released    bool   `json:"released,omitempty"`
	StartDate   string `json:"startDate,omitempty"`   // Format: YYYY-MM-DD
	ReleaseDate string `json:"releaseDate,omitempty"` // Format: YYYY-MM-DD
}

// CreateVersion creates a new version in a project.
//
// Example:
//
//	version, err := client.Project.CreateVersion(ctx, &project.CreateVersionInput{
//		Name:        "v1.0.0",
//		Description: "First major release",
//		Project:     "PROJ",
//		ReleaseDate: "2024-12-31",
//	})
func (s *Service) CreateVersion(ctx context.Context, input *CreateVersionInput) (*Version, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("version name is required")
	}

	if input.Project == "" {
		return nil, fmt.Errorf("project is required")
	}

	path := "/rest/api/3/version"

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
	var version Version
	if err := s.transport.DecodeResponse(resp, &version); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &version, nil
}

// UpdateVersionInput contains the data for updating a version.
type UpdateVersionInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Archived    *bool  `json:"archived,omitempty"`
	Released    *bool  `json:"released,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// UpdateVersion updates an existing version.
//
// Example:
//
//	released := true
//	version, err := client.Project.UpdateVersion(ctx, "10000", &project.UpdateVersionInput{
//		Released:    &released,
//		ReleaseDate: "2024-06-15",
//	})
func (s *Service) UpdateVersion(ctx context.Context, versionID string, input *UpdateVersionInput) (*Version, error) {
	if versionID == "" {
		return nil, fmt.Errorf("version ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/version/%s", versionID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var version Version
	if err := s.transport.DecodeResponse(resp, &version); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &version, nil
}

// GetVersion retrieves a version by ID.
//
// Example:
//
//	version, err := client.Project.GetVersion(ctx, "10000")
func (s *Service) GetVersion(ctx context.Context, versionID string) (*Version, error) {
	if versionID == "" {
		return nil, fmt.Errorf("version ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/version/%s", versionID)

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
	var version Version
	if err := s.transport.DecodeResponse(resp, &version); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &version, nil
}

// DeleteVersion deletes a version.
//
// Example:
//
//	err := client.Project.DeleteVersion(ctx, "10000")
func (s *Service) DeleteVersion(ctx context.Context, versionID string) error {
	if versionID == "" {
		return fmt.Errorf("version ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/version/%s", versionID)

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

// ListProjectVersions retrieves all versions for a project.
//
// Example:
//
//	versions, err := client.Project.ListProjectVersions(ctx, "PROJ")
func (s *Service) ListProjectVersions(ctx context.Context, projectKeyOrID string) ([]*Version, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/versions", projectKeyOrID)

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
	var versions []*Version
	if err := s.transport.DecodeResponse(resp, &versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return versions, nil
}

// ListProjectComponents retrieves all components for a project.
//
// Example:
//
//	components, err := client.Project.ListProjectComponents(ctx, "PROJ")
func (s *Service) ListProjectComponents(ctx context.Context, projectKeyOrID string) ([]*Component, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/components", projectKeyOrID)

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
	var components []*Component
	if err := s.transport.DecodeResponse(resp, &components); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return components, nil
}
