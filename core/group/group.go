// Package group provides Group resource management for Jira.
//
// Groups are used to manage collections of users for permission and
// notification purposes.
package group

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Group resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Group service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Group represents a Jira group.
type Group struct {
	Name    string         `json:"name"`
	Self    string         `json:"self,omitempty"`
	GroupID string         `json:"groupId,omitempty"`
	Expand  string         `json:"expand,omitempty"`
	Users   *GroupUsers    `json:"users,omitempty"`
}

// GroupUsers represents users in a group.
type GroupUsers struct {
	Size       int     `json:"size"`
	Items      []*User `json:"items"`
	MaxResults int     `json:"maxResults"`
	StartAt    int     `json:"startAt"`
}

// User represents a Jira user.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	AccountType  string `json:"accountType,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Self         string `json:"self,omitempty"`
}

// FindOptions represents options for finding groups.
type FindOptions struct {
	Query      string `json:"query,omitempty"`
	Exclude    string `json:"exclude,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	UserName   string `json:"userName,omitempty"`
}

// Find searches for groups.
//
// Example:
//
//	groups, err := client.Group.Find(ctx, &group.FindOptions{
//		Query:      "jira",
//		MaxResults: 50,
//	})
func (s *Service) Find(ctx context.Context, opts *FindOptions) ([]*Group, error) {
	path := "/rest/api/3/groups/picker"

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
		if opts.Exclude != "" {
			q.Set("exclude", opts.Exclude)
		}
		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		if opts.UserName != "" {
			q.Set("userName", opts.UserName)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Groups []*Group `json:"groups"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Groups, nil
}

// GetOptions represents options for getting a group.
type GetOptions struct {
	GroupName  string   `json:"groupname,omitempty"`
	GroupID    string   `json:"groupId,omitempty"`
	Expand     []string `json:"expand,omitempty"`
}

// Get retrieves a group.
//
// Example:
//
//	group, err := client.Group.Get(ctx, &group.GetOptions{
//		GroupName: "jira-administrators",
//		Expand:    []string{"users"},
//	})
func (s *Service) Get(ctx context.Context, opts *GetOptions) (*Group, error) {
	if opts == nil || (opts.GroupName == "" && opts.GroupID == "") {
		return nil, fmt.Errorf("group name or ID is required")
	}

	path := "/rest/api/3/group"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if opts.GroupName != "" {
		q.Set("groupname", opts.GroupName)
	}
	if opts.GroupID != "" {
		q.Set("groupId", opts.GroupID)
	}
	if len(opts.Expand) > 0 {
		for _, expand := range opts.Expand {
			q.Add("expand", expand)
		}
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var group Group
	if err := s.transport.DecodeResponse(resp, &group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &group, nil
}

// CreateGroupInput represents input for creating a group.
type CreateGroupInput struct {
	Name string `json:"name"`
}

// Create creates a new group.
//
// Example:
//
//	group, err := client.Group.Create(ctx, &group.CreateGroupInput{
//		Name: "my-new-group",
//	})
func (s *Service) Create(ctx context.Context, input *CreateGroupInput) (*Group, error) {
	if input == nil || input.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	path := "/rest/api/3/group"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var group Group
	if err := s.transport.DecodeResponse(resp, &group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &group, nil
}

// DeleteOptions represents options for deleting a group.
type DeleteOptions struct {
	GroupName string `json:"groupname,omitempty"`
	GroupID   string `json:"groupId,omitempty"`
}

// Delete removes a group.
//
// Example:
//
//	err := client.Group.Delete(ctx, &group.DeleteOptions{
//		GroupName: "my-group",
//	})
func (s *Service) Delete(ctx context.Context, opts *DeleteOptions) error {
	if opts == nil || (opts.GroupName == "" && opts.GroupID == "") {
		return fmt.Errorf("group name or ID is required")
	}

	path := "/rest/api/3/group"

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if opts.GroupName != "" {
		q.Set("groupname", opts.GroupName)
	}
	if opts.GroupID != "" {
		q.Set("groupId", opts.GroupID)
	}
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// GetMembersOptions represents options for getting group members.
type GetMembersOptions struct {
	GroupName       string `json:"groupname,omitempty"`
	GroupID         string `json:"groupId,omitempty"`
	IncludeInactive bool   `json:"includeInactiveUsers,omitempty"`
	StartAt         int    `json:"startAt,omitempty"`
	MaxResults      int    `json:"maxResults,omitempty"`
}

// GetMembers retrieves members of a group.
//
// Example:
//
//	members, err := client.Group.GetMembers(ctx, &group.GetMembersOptions{
//		GroupName:  "jira-users",
//		MaxResults: 50,
//	})
func (s *Service) GetMembers(ctx context.Context, opts *GetMembersOptions) ([]*User, error) {
	if opts == nil || (opts.GroupName == "" && opts.GroupID == "") {
		return nil, fmt.Errorf("group name or ID is required")
	}

	path := "/rest/api/3/group/member"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if opts.GroupName != "" {
		q.Set("groupname", opts.GroupName)
	}
	if opts.GroupID != "" {
		q.Set("groupId", opts.GroupID)
	}
	if opts.IncludeInactive {
		q.Set("includeInactiveUsers", "true")
	}
	if opts.StartAt > 0 {
		q.Set("startAt", strconv.Itoa(opts.StartAt))
	}
	if opts.MaxResults > 0 {
		q.Set("maxResults", strconv.Itoa(opts.MaxResults))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*User `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// AddUserOptions represents options for adding a user to a group.
type AddUserOptions struct {
	GroupName string `json:"groupname,omitempty"`
	GroupID   string `json:"groupId,omitempty"`
	AccountID string `json:"accountId"`
}

// AddUser adds a user to a group.
//
// Example:
//
//	group, err := client.Group.AddUser(ctx, &group.AddUserOptions{
//		GroupName: "jira-users",
//		AccountID: "5b10a2844c20165700ede21g",
//	})
func (s *Service) AddUser(ctx context.Context, opts *AddUserOptions) (*Group, error) {
	if opts == nil || (opts.GroupName == "" && opts.GroupID == "") {
		return nil, fmt.Errorf("group name or ID is required")
	}

	if opts.AccountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	path := "/rest/api/3/group/user"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, map[string]string{
		"accountId": opts.AccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if opts.GroupName != "" {
		q.Set("groupname", opts.GroupName)
	}
	if opts.GroupID != "" {
		q.Set("groupId", opts.GroupID)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var group Group
	if err := s.transport.DecodeResponse(resp, &group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &group, nil
}

// RemoveUserOptions represents options for removing a user from a group.
type RemoveUserOptions struct {
	GroupName string `json:"groupname,omitempty"`
	GroupID   string `json:"groupId,omitempty"`
	AccountID string `json:"accountId"`
	Username  string `json:"username,omitempty"`
}

// RemoveUser removes a user from a group.
//
// Example:
//
//	err := client.Group.RemoveUser(ctx, &group.RemoveUserOptions{
//		GroupName: "jira-users",
//		AccountID: "5b10a2844c20165700ede21g",
//	})
func (s *Service) RemoveUser(ctx context.Context, opts *RemoveUserOptions) error {
	if opts == nil || (opts.GroupName == "" && opts.GroupID == "") {
		return fmt.Errorf("group name or ID is required")
	}

	if opts.AccountID == "" && opts.Username == "" {
		return fmt.Errorf("account ID or username is required")
	}

	path := "/rest/api/3/group/user"

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if opts.GroupName != "" {
		q.Set("groupname", opts.GroupName)
	}
	if opts.GroupID != "" {
		q.Set("groupId", opts.GroupID)
	}
	if opts.AccountID != "" {
		q.Set("accountId", opts.AccountID)
	}
	if opts.Username != "" {
		q.Set("username", opts.Username)
	}
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// BulkOptions represents options for bulk group operations.
type BulkOptions struct {
	GroupNames []string `json:"groupNames,omitempty"`
	GroupIDs   []string `json:"groupIds,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	StartAt    int      `json:"startAt,omitempty"`
}

// BulkGet retrieves multiple groups.
//
// Example:
//
//	groups, err := client.Group.BulkGet(ctx, &group.BulkOptions{
//		GroupNames: []string{"jira-administrators", "jira-users"},
//	})
func (s *Service) BulkGet(ctx context.Context, opts *BulkOptions) ([]*Group, error) {
	if opts == nil || (len(opts.GroupNames) == 0 && len(opts.GroupIDs) == 0) {
		return nil, fmt.Errorf("group names or IDs are required")
	}

	path := "/rest/api/3/group/bulk"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	for _, name := range opts.GroupNames {
		q.Add("groupName", name)
	}
	for _, id := range opts.GroupIDs {
		q.Add("groupId", id)
	}
	if opts.MaxResults > 0 {
		q.Set("maxResults", strconv.Itoa(opts.MaxResults))
	}
	if opts.StartAt > 0 {
		q.Set("startAt", strconv.Itoa(opts.StartAt))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*Group `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}
