// Package filter provides Filter resource management for Jira.
//
// Filters (also known as saved searches) are JQL queries that can be saved,
// shared, and used across Jira. This package provides operations for managing
// filters, including CRUD operations, sharing, and favoriting.
package filter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Service provides operations for Filter resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Filter service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Filter represents a Jira filter (saved search).
type Filter struct {
	ID                  string         `json:"id"`
	Self                string         `json:"self,omitempty"`
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	Owner               *User          `json:"owner,omitempty"`
	JQL                 string         `json:"jql"`
	ViewURL             string         `json:"viewUrl,omitempty"`
	SearchURL           string         `json:"searchUrl,omitempty"`
	Favorite            bool           `json:"favourite"`
	FavouritedCount     int            `json:"favouritedCount,omitempty"`
	SharePermissions    []*Permission  `json:"sharePermissions,omitempty"`
	EditPermissions     []*Permission  `json:"editPermissions,omitempty"`
	Subscriptions       *Subscriptions `json:"subscriptions,omitempty"`
	ApproximateLastUsed string         `json:"approximateLastUsed,omitempty"`
}

// User represents a Jira user (simplified for filter context).
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Self         string `json:"self,omitempty"`
}

// Permission represents a share or edit permission for a filter.
type Permission struct {
	ID      int64    `json:"id,omitempty"`
	Type    string   `json:"type"` // "global", "project", "group", "user", "authenticated", "loggedin"
	Project *Project `json:"project,omitempty"`
	Group   *Group   `json:"group,omitempty"`
	User    *User    `json:"user,omitempty"`
	Role    *Role    `json:"role,omitempty"`
	View    bool     `json:"view,omitempty"`
	Edit    bool     `json:"edit,omitempty"`
}

// Project represents a minimal project reference.
type Project struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// Group represents a user group.
type Group struct {
	Name string `json:"name"`
	Self string `json:"self,omitempty"`
}

// Role represents a project role.
type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// Subscriptions contains filter subscription information.
type Subscriptions struct {
	Size       int             `json:"size,omitempty"`
	Items      []*Subscription `json:"items,omitempty"`
	MaxResults int             `json:"max-results,omitempty"`
	StartIndex int             `json:"start-index,omitempty"`
	EndIndex   int             `json:"end-index,omitempty"`
}

// Subscription represents a filter subscription.
type Subscription struct {
	ID    int64  `json:"id,omitempty"`
	User  *User  `json:"user,omitempty"`
	Group *Group `json:"group,omitempty"`
}

// CreateFilterInput contains the data for creating a filter.
type CreateFilterInput struct {
	Name             string        `json:"name"`
	Description      string        `json:"description,omitempty"`
	JQL              string        `json:"jql"`
	Favorite         bool          `json:"favourite,omitempty"`
	SharePermissions []*Permission `json:"sharePermissions,omitempty"`
	EditPermissions  []*Permission `json:"editPermissions,omitempty"`
}

// UpdateFilterInput contains the data for updating a filter.
type UpdateFilterInput struct {
	Name             string        `json:"name,omitempty"`
	Description      string        `json:"description,omitempty"`
	JQL              string        `json:"jql,omitempty"`
	Favorite         *bool         `json:"favourite,omitempty"`
	SharePermissions []*Permission `json:"sharePermissions,omitempty"`
	EditPermissions  []*Permission `json:"editPermissions,omitempty"`
}

// ListOptions configures filter listing operations.
type ListOptions struct {
	// Expand specifies additional information to retrieve
	Expand []string

	// IncludeFavourites includes favorite filters
	IncludeFavourites bool

	// OrderBy specifies the sort field (name, -name, +name, id, -id, +id, description, -description, +description, owner, -owner, +owner, favourite_count, -favourite_count, +favourite_count)
	OrderBy string

	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results (max 100)
	MaxResults int
}

// SearchOptions configures filter search operations.
type SearchOptions struct {
	// FilterName filters by name
	FilterName string

	// AccountID filters by owner account ID
	AccountID string

	// GroupName filters by group with view permissions
	GroupName string

	// ProjectID filters by associated project
	ProjectID string

	// OrderBy specifies the sort order
	OrderBy string

	// Expand specifies additional information
	Expand []string

	// OverrideSharePermissions includes filters user can't edit
	OverrideSharePermissions bool

	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results
	MaxResults int
}

// Get retrieves a filter by ID.
//
// Example:
//
//	filter, err := client.Filter.Get(ctx, "10000", nil)
func (s *Service) Get(ctx context.Context, filterID string, expand []string) (*Filter, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s", filterID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add expand query parameter
	if len(expand) > 0 {
		q := req.URL.Query()
		q.Set("expand", strings.Join(expand, ","))
		req.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var filter Filter
	if err := s.transport.DecodeResponse(resp, &filter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &filter, nil
}

// Create creates a new filter.
//
// Example:
//
//	filter, err := client.Filter.Create(ctx, &filter.CreateFilterInput{
//	    Name: "My Bugs",
//	    JQL:  "project = PROJ AND type = Bug AND resolution = Unresolved",
//	    Description: "All unresolved bugs in PROJ",
//	    Favorite: true,
//	})
func (s *Service) Create(ctx context.Context, input *CreateFilterInput) (*Filter, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("filter name is required")
	}

	if input.JQL == "" {
		return nil, fmt.Errorf("JQL query is required")
	}

	path := "/rest/api/3/filter"

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
	var filter Filter
	if err := s.transport.DecodeResponse(resp, &filter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &filter, nil
}

// Update updates an existing filter.
//
// Example:
//
//	filter, err := client.Filter.Update(ctx, "10000", &filter.UpdateFilterInput{
//	    Name: "Updated Filter Name",
//	    JQL:  "project = PROJ AND type = Bug",
//	})
func (s *Service) Update(ctx context.Context, filterID string, input *UpdateFilterInput) (*Filter, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s", filterID)

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
	var filter Filter
	if err := s.transport.DecodeResponse(resp, &filter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &filter, nil
}

// Delete deletes a filter.
//
// Example:
//
//	err := client.Filter.Delete(ctx, "10000")
func (s *Service) Delete(ctx context.Context, filterID string) error {
	if filterID == "" {
		return fmt.Errorf("filter ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s", filterID)

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

// List retrieves filters with optional filtering and pagination.
//
// Example:
//
//	filters, err := client.Filter.List(ctx, &filter.ListOptions{
//	    IncludeFavourites: true,
//	    MaxResults: 50,
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Filter, error) {
	path := "/rest/api/3/filter/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()

		if len(opts.Expand) > 0 {
			q.Set("expand", strings.Join(opts.Expand, ","))
		}

		if opts.IncludeFavourites {
			q.Set("includeFavourites", "true")
		}

		if opts.OrderBy != "" {
			q.Set("orderBy", opts.OrderBy)
		}

		if opts.StartAt > 0 {
			q.Set("startAt", strconv.Itoa(opts.StartAt))
		}

		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
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
		StartAt    int       `json:"startAt"`
		MaxResults int       `json:"maxResults"`
		Total      int       `json:"total"`
		IsLast     bool      `json:"isLast"`
		Values     []*Filter `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// GetFavorites retrieves the current user's favorite filters.
//
// Example:
//
//	favorites, err := client.Filter.GetFavorites(ctx)
func (s *Service) GetFavorites(ctx context.Context) ([]*Filter, error) {
	path := "/rest/api/3/filter/favourite"

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
	var filters []*Filter
	if err := s.transport.DecodeResponse(resp, &filters); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return filters, nil
}

// GetMyFilters retrieves filters owned by the current user.
//
// Example:
//
//	myFilters, err := client.Filter.GetMyFilters(ctx, nil)
func (s *Service) GetMyFilters(ctx context.Context, expand []string, includeFavourites bool) ([]*Filter, error) {
	path := "/rest/api/3/filter/my"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()

	if len(expand) > 0 {
		q.Set("expand", strings.Join(expand, ","))
	}

	if includeFavourites {
		q.Set("includeFavourites", "true")
	}

	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var filters []*Filter
	if err := s.transport.DecodeResponse(resp, &filters); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return filters, nil
}

// SetFavorite marks a filter as favorite for the current user.
//
// Example:
//
//	filter, err := client.Filter.SetFavorite(ctx, "10000")
func (s *Service) SetFavorite(ctx context.Context, filterID string) (*Filter, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s/favourite", filterID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var filter Filter
	if err := s.transport.DecodeResponse(resp, &filter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &filter, nil
}

// RemoveFavorite removes a filter from the current user's favorites.
//
// Example:
//
//	filter, err := client.Filter.RemoveFavorite(ctx, "10000")
func (s *Service) RemoveFavorite(ctx context.Context, filterID string) (*Filter, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s/favourite", filterID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var filter Filter
	if err := s.transport.DecodeResponse(resp, &filter); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &filter, nil
}

// GetDefaultShareScope retrieves the default share scope for the current user.
//
// Example:
//
//	scope, err := client.Filter.GetDefaultShareScope(ctx)
func (s *Service) GetDefaultShareScope(ctx context.Context) (string, error) {
	path := "/rest/api/3/filter/defaultShareScope"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result struct {
		Scope string `json:"scope"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Scope, nil
}

// SetDefaultShareScope sets the default share scope for the current user.
//
// Scope can be: "GLOBAL", "AUTHENTICATED", "PRIVATE"
//
// Example:
//
//	err := client.Filter.SetDefaultShareScope(ctx, "PRIVATE")
func (s *Service) SetDefaultShareScope(ctx context.Context, scope string) error {
	if scope == "" {
		return fmt.Errorf("scope is required")
	}

	path := "/rest/api/3/filter/defaultShareScope"

	input := struct {
		Scope string `json:"scope"`
	}{
		Scope: scope,
	}

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
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

// GetSharePermission retrieves a single share permission by ID.
//
// Example:
//
//	permission, err := client.Filter.GetSharePermission(ctx, "10000", 12345)
func (s *Service) GetSharePermission(ctx context.Context, filterID string, permissionID int64) (*Permission, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	if permissionID <= 0 {
		return nil, fmt.Errorf("permission ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s/permission/%d", filterID, permissionID)

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
	var permission Permission
	if err := s.transport.DecodeResponse(resp, &permission); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &permission, nil
}

// AddSharePermission adds a share permission to a filter.
//
// Example:
//
//	permission, err := client.Filter.AddSharePermission(ctx, "10000", &filter.Permission{
//	    Type: "group",
//	    Group: &filter.Group{Name: "jira-users"},
//	})
func (s *Service) AddSharePermission(ctx context.Context, filterID string, permission *Permission) ([]*Permission, error) {
	if filterID == "" {
		return nil, fmt.Errorf("filter ID is required")
	}

	if permission == nil {
		return nil, fmt.Errorf("permission is required")
	}

	if permission.Type == "" {
		return nil, fmt.Errorf("permission type is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s/permission", filterID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var permissions []*Permission
	if err := s.transport.DecodeResponse(resp, &permissions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return permissions, nil
}

// DeleteSharePermission removes a share permission from a filter.
//
// Example:
//
//	err := client.Filter.DeleteSharePermission(ctx, "10000", 12345)
func (s *Service) DeleteSharePermission(ctx context.Context, filterID string, permissionID int64) error {
	if filterID == "" {
		return fmt.Errorf("filter ID is required")
	}

	if permissionID <= 0 {
		return fmt.Errorf("permission ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/filter/%s/permission/%d", filterID, permissionID)

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
