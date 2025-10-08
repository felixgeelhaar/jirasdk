// Package permission provides Permission resource management for Jira.
package permission

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Permission resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Permission service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Permission represents a Jira permission.
type Permission struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// PermissionHolder represents an entity that holds a permission.
type PermissionHolder struct {
	Type      string `json:"type"` // "user", "group", "projectRole", "applicationRole"
	Parameter string `json:"parameter,omitempty"`
	Value     string `json:"value,omitempty"`
	Expand    string `json:"expand,omitempty"`
}

// PermissionGrant represents a granted permission.
type PermissionGrant struct {
	ID         int64             `json:"id"`
	Self       string            `json:"self,omitempty"`
	Holder     *PermissionHolder `json:"holder,omitempty"`
	Permission string            `json:"permission"`
}

// PermissionScheme represents a permission scheme.
type PermissionScheme struct {
	ID          int64              `json:"id"`
	Self        string             `json:"self,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Permissions []*PermissionGrant `json:"permissions,omitempty"`
	Expand      string             `json:"expand,omitempty"`
}

// MyPermissions represents current user's permissions.
type MyPermissions struct {
	Permissions map[string]*PermissionStatus `json:"permissions"`
}

// PermissionStatus indicates if a permission is granted.
type PermissionStatus struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Description    string `json:"description,omitempty"`
	HavePermission bool   `json:"havePermission"`
}

// ProjectRole represents a project role.
type ProjectRole struct {
	Self        string   `json:"self,omitempty"`
	Name        string   `json:"name"`
	ID          int64    `json:"id"`
	Description string   `json:"description,omitempty"`
	Actors      []*Actor `json:"actors,omitempty"`
	Scope       *Scope   `json:"scope,omitempty"`
}

// Actor represents a role actor (user or group).
type Actor struct {
	ID          int64       `json:"id"`
	DisplayName string      `json:"displayName"`
	Type        string      `json:"type"` // "atlassian-user-role-actor" or "atlassian-group-role-actor"
	Name        string      `json:"name,omitempty"`
	ActorUser   *ActorUser  `json:"actorUser,omitempty"`
	ActorGroup  *ActorGroup `json:"actorGroup,omitempty"`
}

// ActorUser represents a user actor.
type ActorUser struct {
	AccountID string `json:"accountId"`
}

// ActorGroup represents a group actor.
type ActorGroup struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
}

// Scope represents the scope of a role.
type Scope struct {
	Type    string   `json:"type"`
	Project *Project `json:"project,omitempty"`
}

// Project represents a minimal project reference.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

// GetAllPermissions retrieves all permissions in Jira.
//
// Example:
//
//	permissions, err := client.Permission.GetAllPermissions(ctx)
func (s *Service) GetAllPermissions(ctx context.Context) ([]*Permission, error) {
	path := "/rest/api/3/permissions"

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
	var result struct {
		Permissions []*Permission `json:"permissions"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Permissions, nil
}

// MyPermissionsOptions configures the GetMyPermissions operation.
type MyPermissionsOptions struct {
	// ProjectKey or ProjectID to check permissions for
	ProjectKey string
	ProjectID  string

	// IssueKey or IssueID to check permissions for
	IssueKey string
	IssueID  string

	// Permissions to check (comma-separated keys)
	Permissions string
}

// GetMyPermissions retrieves the current user's permissions.
//
// Example:
//
//	permissions, err := client.Permission.GetMyPermissions(ctx, &permission.MyPermissionsOptions{
//	    ProjectKey: "PROJ",
//	    Permissions: "EDIT_ISSUES,DELETE_ISSUES",
//	})
func (s *Service) GetMyPermissions(ctx context.Context, opts *MyPermissionsOptions) (*MyPermissions, error) {
	path := "/rest/api/3/mypermissions"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()

		if opts.ProjectKey != "" {
			q.Set("projectKey", opts.ProjectKey)
		}

		if opts.ProjectID != "" {
			q.Set("projectId", opts.ProjectID)
		}

		if opts.IssueKey != "" {
			q.Set("issueKey", opts.IssueKey)
		}

		if opts.IssueID != "" {
			q.Set("issueId", opts.IssueID)
		}

		if opts.Permissions != "" {
			q.Set("permissions", opts.Permissions)
		}

		req.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var myPerms MyPermissions
	if err := s.transport.DecodeResponse(resp, &myPerms); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &myPerms, nil
}

// ListPermissionSchemesOptions configures the ListPermissionSchemes operation.
type ListPermissionSchemesOptions struct {
	// Expand specifies additional information to include
	Expand []string
}

// ListPermissionSchemes retrieves all permission schemes.
//
// Example:
//
//	schemes, err := client.Permission.ListPermissionSchemes(ctx, nil)
func (s *Service) ListPermissionSchemes(ctx context.Context, opts *ListPermissionSchemesOptions) ([]*PermissionScheme, error) {
	path := "/rest/api/3/permissionscheme"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
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
	var result struct {
		PermissionSchemes []*PermissionScheme `json:"permissionSchemes"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.PermissionSchemes, nil
}

// GetPermissionSchemeOptions configures the GetPermissionScheme operation.
type GetPermissionSchemeOptions struct {
	// Expand specifies additional information to include
	Expand []string
}

// GetPermissionScheme retrieves a permission scheme by ID.
//
// Example:
//
//	scheme, err := client.Permission.GetPermissionScheme(ctx, 10000, &permission.GetPermissionSchemeOptions{
//	    Expand: []string{"permissions", "user"},
//	})
func (s *Service) GetPermissionScheme(ctx context.Context, schemeID int64, opts *GetPermissionSchemeOptions) (*PermissionScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/permissionscheme/%d", schemeID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
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
	var scheme PermissionScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// CreatePermissionSchemeInput contains the data for creating a permission scheme.
type CreatePermissionSchemeInput struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Permissions []*PermissionGrant `json:"permissions,omitempty"`
}

// CreatePermissionScheme creates a new permission scheme.
//
// Example:
//
//	scheme, err := client.Permission.CreatePermissionScheme(ctx, &permission.CreatePermissionSchemeInput{
//	    Name:        "Custom Scheme",
//	    Description: "Custom permission scheme for projects",
//	})
func (s *Service) CreatePermissionScheme(ctx context.Context, input *CreatePermissionSchemeInput) (*PermissionScheme, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("scheme name is required")
	}

	path := "/rest/api/3/permissionscheme"

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
	var scheme PermissionScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// UpdatePermissionSchemeInput contains the data for updating a permission scheme.
type UpdatePermissionSchemeInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdatePermissionScheme updates a permission scheme.
//
// Example:
//
//	scheme, err := client.Permission.UpdatePermissionScheme(ctx, 10000, &permission.UpdatePermissionSchemeInput{
//	    Description: "Updated description",
//	})
func (s *Service) UpdatePermissionScheme(ctx context.Context, schemeID int64, input *UpdatePermissionSchemeInput) (*PermissionScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("scheme ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/permissionscheme/%d", schemeID)

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
	var scheme PermissionScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// DeletePermissionScheme deletes a permission scheme.
//
// Example:
//
//	err := client.Permission.DeletePermissionScheme(ctx, 10000)
func (s *Service) DeletePermissionScheme(ctx context.Context, schemeID int64) error {
	if schemeID <= 0 {
		return fmt.Errorf("scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/permissionscheme/%d", schemeID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// GetProjectRoles retrieves all roles for a project.
//
// Example:
//
//	roles, err := client.Permission.GetProjectRoles(ctx, "PROJ")
func (s *Service) GetProjectRoles(ctx context.Context, projectKeyOrID string) (map[string]string, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/role", projectKeyOrID)

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

	// Decode response (returns a map of role names to URLs)
	var roles map[string]string
	if err := s.transport.DecodeResponse(resp, &roles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return roles, nil
}

// GetProjectRole retrieves details of a specific role in a project.
//
// Example:
//
//	role, err := client.Permission.GetProjectRole(ctx, "PROJ", 10002)
func (s *Service) GetProjectRole(ctx context.Context, projectKeyOrID string, roleID int64) (*ProjectRole, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	if roleID <= 0 {
		return nil, fmt.Errorf("role ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/role/%d", projectKeyOrID, roleID)

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
	var role ProjectRole
	if err := s.transport.DecodeResponse(resp, &role); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &role, nil
}

// AddActorInput contains the data for adding actors to a project role.
type AddActorInput struct {
	User  []string `json:"user,omitempty"`  // Account IDs
	Group []string `json:"group,omitempty"` // Group names
}

// AddActorsToProjectRole adds actors (users or groups) to a project role.
//
// Example:
//
//	role, err := client.Permission.AddActorsToProjectRole(ctx, "PROJ", 10002, &permission.AddActorInput{
//	    User: []string{"accountId1", "accountId2"},
//	    Group: []string{"developers", "admins"},
//	})
func (s *Service) AddActorsToProjectRole(ctx context.Context, projectKeyOrID string, roleID int64, input *AddActorInput) (*ProjectRole, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	if roleID <= 0 {
		return nil, fmt.Errorf("role ID is required")
	}

	if input == nil || (len(input.User) == 0 && len(input.Group) == 0) {
		return nil, fmt.Errorf("at least one user or group is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/role/%d", projectKeyOrID, roleID)

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
	var role ProjectRole
	if err := s.transport.DecodeResponse(resp, &role); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &role, nil
}

// RemoveActorFromProjectRole removes an actor from a project role.
//
// Example:
//
//	// Remove user
//	err := client.Permission.RemoveActorFromProjectRole(ctx, "PROJ", 10002, "user", "accountId123")
//
//	// Remove group
//	err = client.Permission.RemoveActorFromProjectRole(ctx, "PROJ", 10002, "group", "developers")
func (s *Service) RemoveActorFromProjectRole(ctx context.Context, projectKeyOrID string, roleID int64, actorType, actor string) error {
	if projectKeyOrID == "" {
		return fmt.Errorf("project key or ID is required")
	}

	if roleID <= 0 {
		return fmt.Errorf("role ID is required")
	}

	if actorType == "" {
		return fmt.Errorf("actor type is required (user or group)")
	}

	if actor == "" {
		return fmt.Errorf("actor is required")
	}

	path := fmt.Sprintf("/rest/api/3/project/%s/role/%d", projectKeyOrID, roleID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Set(actorType, actor)
	req.URL.RawQuery = q.Encode()

	// Execute request
	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
