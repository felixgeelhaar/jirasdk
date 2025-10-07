// Package user provides User resource management for Jira.
package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/felixgeelhaar/jira-connect/internal/pagination"
)

// Service provides operations for User resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new User service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// User represents a Jira user.
type User struct {
	Self         string      `json:"self,omitempty"`
	AccountID    string      `json:"accountId,omitempty"`
	AccountType  string      `json:"accountType,omitempty"`
	EmailAddress string      `json:"emailAddress,omitempty"`
	DisplayName  string      `json:"displayName,omitempty"`
	Active       bool        `json:"active,omitempty"`
	TimeZone     string      `json:"timeZone,omitempty"`
	Locale       string      `json:"locale,omitempty"`
	AvatarURLs   *AvatarURLs `json:"avatarUrls,omitempty"`
	Groups       *Groups     `json:"groups,omitempty"`
}

// AvatarURLs contains URLs for different sizes of avatars.
type AvatarURLs struct {
	Size16 string `json:"16x16,omitempty"`
	Size24 string `json:"24x24,omitempty"`
	Size32 string `json:"32x32,omitempty"`
	Size48 string `json:"48x48,omitempty"`
}

// Groups represents user group membership.
type Groups struct {
	Size  int           `json:"size,omitempty"`
	Items []*GroupItem  `json:"items,omitempty"`
}

// GroupItem represents a single group.
type GroupItem struct {
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// GetOptions configures the Get operation.
type GetOptions struct {
	// Expand specifies additional information to include
	Expand []string
}

// Get retrieves a user by account ID.
//
// Example:
//
//	user, err := client.User.Get(ctx, "5b10a2844c20165700ede21g", &user.GetOptions{
//		Expand: []string{"groups", "applicationRoles"},
//	})
func (s *Service) Get(ctx context.Context, accountID string, opts *GetOptions) (*User, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	path := "/rest/api/3/user"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Set("accountId", accountID)

	if opts != nil && len(opts.Expand) > 0 {
		q.Set("expand", strings.Join(opts.Expand, ","))
	}

	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var user User
	if err := s.transport.DecodeResponse(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// GetMyself retrieves the currently authenticated user.
//
// Example:
//
//	user, err := client.User.GetMyself(ctx)
func (s *Service) GetMyself(ctx context.Context) (*User, error) {
	path := "/rest/api/3/myself"

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
	var user User
	if err := s.transport.DecodeResponse(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// SearchOptions configures the search operation.
type SearchOptions struct {
	// Query is the search query string
	Query string

	// MaxResults is the maximum number of results
	MaxResults int

	// StartAt is the index of the first result
	StartAt int

	// IncludeActive includes active users in results
	IncludeActive bool

	// IncludeInactive includes inactive users in results
	IncludeInactive bool

	// Property filters users by property
	Property string
}

// Search searches for users.
//
// Example:
//
//	users, err := client.User.Search(ctx, &user.SearchOptions{
//		Query:      "john",
//		MaxResults: 50,
//	})
func (s *Service) Search(ctx context.Context, opts *SearchOptions) ([]*User, error) {
	path := "/rest/api/3/user/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()

	if opts != nil {
		if opts.Query != "" {
			q.Set("query", opts.Query)
		}

		if opts.MaxResults > 0 {
			q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}

		if opts.StartAt > 0 {
			q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
		}

		if opts.IncludeActive {
			q.Set("includeActive", "true")
		}

		if opts.IncludeInactive {
			q.Set("includeInactive", "true")
		}

		if opts.Property != "" {
			q.Set("property", opts.Property)
		}
	}

	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var users []*User
	if err := s.transport.DecodeResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// FindOptions configures the find operation.
type FindOptions struct {
	// Query is the search query string
	Query string

	// Username filters by username (deprecated, use accountId)
	Username string

	// AccountID filters by account ID
	AccountID string

	// MaxResults is the maximum number of results
	MaxResults int

	// StartAt is the index of the first result
	StartAt int

	pagination.Options
}

// FindUsers finds users with browse permission for a project or issue.
//
// Example:
//
//	users, err := client.User.FindUsers(ctx, &user.FindOptions{
//		Query:      "john",
//		MaxResults: 50,
//	})
func (s *Service) FindUsers(ctx context.Context, opts *FindOptions) ([]*User, error) {
	path := "/rest/api/3/user/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()

		if opts.Query != "" {
			q.Set("query", opts.Query)
		}

		if opts.Username != "" {
			q.Set("username", opts.Username)
		}

		if opts.AccountID != "" {
			q.Set("accountId", opts.AccountID)
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
	var users []*User
	if err := s.transport.DecodeResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// FindAssignableOptions configures finding assignable users.
type FindAssignableOptions struct {
	// Project filters by project key
	Project string

	// IssueKey filters by issue key
	IssueKey string

	// Query is the search query string
	Query string

	// MaxResults is the maximum number of results
	MaxResults int

	// StartAt is the index of the first result
	StartAt int

	pagination.Options
}

// FindAssignableUsers finds users that can be assigned to issues.
//
// Example:
//
//	users, err := client.User.FindAssignableUsers(ctx, &user.FindAssignableOptions{
//		Project: "PROJ",
//		Query:   "john",
//	})
func (s *Service) FindAssignableUsers(ctx context.Context, opts *FindAssignableOptions) ([]*User, error) {
	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}

	if opts.Project == "" && opts.IssueKey == "" {
		return nil, fmt.Errorf("either project or issue key is required")
	}

	path := "/rest/api/3/user/assignable/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()

	if opts.Project != "" {
		q.Set("project", opts.Project)
	}

	if opts.IssueKey != "" {
		q.Set("issueKey", opts.IssueKey)
	}

	if opts.Query != "" {
		q.Set("query", opts.Query)
	}

	opts.Options.ApplyToURL(req.URL)
	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var users []*User
	if err := s.transport.DecodeResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}

// BulkGetOptions configures bulk user retrieval.
type BulkGetOptions struct {
	// AccountIDs is the list of account IDs to retrieve
	AccountIDs []string

	// MaxResults is the maximum number of results
	MaxResults int

	// StartAt is the index of the first result
	StartAt int
}

// BulkGet retrieves multiple users by their account IDs.
//
// Example:
//
//	users, err := client.User.BulkGet(ctx, &user.BulkGetOptions{
//		AccountIDs: []string{"5b10a2844c20165700ede21g", "5b10a0effa615349cb016cd8"},
//	})
func (s *Service) BulkGet(ctx context.Context, opts *BulkGetOptions) ([]*User, error) {
	if opts == nil || len(opts.AccountIDs) == 0 {
		return nil, fmt.Errorf("account IDs are required")
	}

	path := "/rest/api/3/user/bulk"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()

	for _, accountID := range opts.AccountIDs {
		q.Add("accountId", accountID)
	}

	if opts.MaxResults > 0 {
		q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
	}

	if opts.StartAt > 0 {
		q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
	}

	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result struct {
		Values []*User `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// GetDefaultColumns retrieves the default issue table columns for the user.
//
// Example:
//
//	columns, err := client.User.GetDefaultColumns(ctx, "5b10a2844c20165700ede21g")
func (s *Service) GetDefaultColumns(ctx context.Context, accountID string) ([]string, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	path := "/rest/api/3/user/columns"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Set("accountId", accountID)
	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var columns []struct {
		Value string `json:"value"`
	}

	if err := s.transport.DecodeResponse(resp, &columns); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]string, len(columns))
	for i, col := range columns {
		result[i] = col.Value
	}

	return result, nil
}
