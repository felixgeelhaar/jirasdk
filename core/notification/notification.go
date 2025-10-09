// Package notification provides Notification resource management for Jira.
//
// Notifications inform users about events in Jira (issue updates, mentions, assignments).
// This package provides operations for managing notification schemes and sending notifications.
package notification

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Notification resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Notification service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// NotificationScheme represents a Jira notification scheme.
type NotificationScheme struct {
	ID                  int64                     `json:"id,omitempty"`
	Self                string                    `json:"self,omitempty"`
	Name                string                    `json:"name"`
	Description         string                    `json:"description,omitempty"`
	NotificationSchemes []*NotificationSchemeItem `json:"notificationSchemeEvents,omitempty"`
}

// NotificationSchemeItem represents a notification event within a scheme.
type NotificationSchemeItem struct {
	Event         *NotificationEvent `json:"event,omitempty"`
	Notifications []*Notification    `json:"notifications,omitempty"`
}

// NotificationEvent represents a notification event type.
type NotificationEvent struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Notification represents a notification configuration.
type Notification struct {
	ID          int64        `json:"id,omitempty"`
	Type        string       `json:"type"`
	Parameter   string       `json:"parameter,omitempty"`
	Group       *Group       `json:"group,omitempty"`
	Field       *Field       `json:"field,omitempty"`
	User        *User        `json:"user,omitempty"`
	ProjectRole *ProjectRole `json:"projectRole,omitempty"`
}

// Group represents a Jira user group.
type Group struct {
	Name string `json:"name"`
	Self string `json:"self,omitempty"`
}

// Field represents a Jira field.
type Field struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// User represents a Jira user.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
}

// ProjectRole represents a project role.
type ProjectRole struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// SendNotificationInput represents input for sending a notification.
type SendNotificationInput struct {
	Subject  string              `json:"subject,omitempty"`
	TextBody string              `json:"textBody,omitempty"`
	HTMLBody string              `json:"htmlBody,omitempty"`
	To       *NotificationTarget `json:"to,omitempty"`
	Restrict *NotificationTarget `json:"restrict,omitempty"`
}

// NotificationTarget represents notification recipients.
type NotificationTarget struct {
	Reporter    bool     `json:"reporter,omitempty"`
	Assignee    bool     `json:"assignee,omitempty"`
	Watchers    bool     `json:"watchers,omitempty"`
	Voters      bool     `json:"voters,omitempty"`
	Users       []*User  `json:"users,omitempty"`
	Groups      []*Group `json:"groups,omitempty"`
	GroupIDs    []string `json:"groupIds,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// CreateNotificationSchemeInput represents input for creating a notification scheme.
type CreateNotificationSchemeInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateNotificationSchemeInput represents input for updating a notification scheme.
type UpdateNotificationSchemeInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// AddNotificationInput represents input for adding a notification to a scheme.
type AddNotificationInput struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter,omitempty"`
}

// ListOptions represents options for listing notification schemes.
type ListOptions struct {
	StartAt    int    `json:"startAt,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	Expand     string `json:"expand,omitempty"`
}

// List retrieves all notification schemes with pagination.
//
// Example:
//
//	schemes, err := client.Notification.List(ctx, &notification.ListOptions{MaxResults: 50})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*NotificationScheme, error) {
	path := "/rest/api/3/notificationscheme"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.StartAt > 0 {
			q.Set("startAt", strconv.Itoa(opts.StartAt))
		}
		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*NotificationScheme `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// Get retrieves a specific notification scheme by ID.
//
// Example:
//
//	scheme, err := client.Notification.Get(ctx, 10000)
func (s *Service) Get(ctx context.Context, schemeID int64) (*NotificationScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("notification scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/notificationscheme/%d", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var scheme NotificationScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// Create creates a new notification scheme.
//
// Example:
//
//	scheme, err := client.Notification.Create(ctx, &notification.CreateNotificationSchemeInput{
//	    Name:        "Custom Notification Scheme",
//	    Description: "Notifications for project team",
//	})
func (s *Service) Create(ctx context.Context, input *CreateNotificationSchemeInput) (*NotificationScheme, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("notification scheme name is required")
	}

	path := "/rest/api/3/notificationscheme"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var scheme NotificationScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// Update updates a notification scheme.
//
// Example:
//
//	scheme, err := client.Notification.Update(ctx, 10000, &notification.UpdateNotificationSchemeInput{
//	    Name:        "Updated Notification Scheme",
//	    Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, schemeID int64, input *UpdateNotificationSchemeInput) (*NotificationScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("notification scheme ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/notificationscheme/%d", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var scheme NotificationScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// Delete deletes a notification scheme.
//
// Example:
//
//	err := client.Notification.Delete(ctx, 10000)
func (s *Service) Delete(ctx context.Context, schemeID int64) error {
	if schemeID <= 0 {
		return fmt.Errorf("notification scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/notificationscheme/%d", schemeID)

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

// AddNotification adds a notification to a notification scheme event.
//
// Example:
//
//	notification, err := client.Notification.AddNotification(ctx, 10000, 1, &notification.AddNotificationInput{
//	    Type:      "Group",
//	    Parameter: "jira-administrators",
//	})
func (s *Service) AddNotification(ctx context.Context, schemeID int64, eventID int64, input *AddNotificationInput) (*Notification, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("notification scheme ID is required")
	}

	if eventID <= 0 {
		return nil, fmt.Errorf("event ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Type == "" {
		return nil, fmt.Errorf("notification type is required")
	}

	path := fmt.Sprintf("/rest/api/3/notificationscheme/%d/notification", schemeID)

	// Add event ID as query parameter
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("eventTypeId", strconv.FormatInt(eventID, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var notification Notification
	if err := s.transport.DecodeResponse(resp, &notification); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &notification, nil
}

// RemoveNotification removes a notification from a notification scheme event.
//
// Example:
//
//	err := client.Notification.RemoveNotification(ctx, 10000, 12345)
func (s *Service) RemoveNotification(ctx context.Context, schemeID int64, notificationID int64) error {
	if schemeID <= 0 {
		return fmt.Errorf("notification scheme ID is required")
	}

	if notificationID <= 0 {
		return fmt.Errorf("notification ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/notificationscheme/%d/notification/%d", schemeID, notificationID)

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

// SendIssueNotification sends a notification for an issue.
//
// Example:
//
//	err := client.Notification.SendIssueNotification(ctx, "PROJ-123", &notification.SendNotificationInput{
//	    Subject:  "Important Update",
//	    TextBody: "This issue has been updated",
//	    To: &notification.NotificationTarget{
//	        Assignee: true,
//	        Watchers: true,
//	    },
//	})
func (s *Service) SendIssueNotification(ctx context.Context, issueKey string, input *SendNotificationInput) error {
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}

	if input == nil {
		return fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/notify", issueKey)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
