package issue

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Worklog represents time logged against an issue.
type Worklog struct {
	ID               string                `json:"id"`
	Self             string                `json:"self,omitempty"`
	Author           *User                 `json:"author,omitempty"`
	UpdateAuthor     *User                 `json:"updateAuthor,omitempty"`
	Comment          string                `json:"comment,omitempty"`
	Created          *time.Time            `json:"created,omitempty"`
	Updated          *time.Time            `json:"updated,omitempty"`
	Started          *time.Time            `json:"started,omitempty"`
	TimeSpent        string                `json:"timeSpent,omitempty"`        // e.g., "3h 20m"
	TimeSpentSeconds int64                 `json:"timeSpentSeconds,omitempty"` // Duration in seconds
	IssueID          string                `json:"issueId,omitempty"`
	Visibility       *WorklogVisibility    `json:"visibility,omitempty"`
}

// WorklogVisibility controls who can see the worklog.
type WorklogVisibility struct {
	Type  string `json:"type"`  // "group" or "role"
	Value string `json:"value"` // group name or role name
}

// AddWorklogInput contains the data for logging work on an issue.
type AddWorklogInput struct {
	// TimeSpent is a human-readable duration (e.g., "3h 20m", "1d 4h")
	TimeSpent string `json:"timeSpent,omitempty"`

	// TimeSpentSeconds is the duration in seconds (alternative to TimeSpent)
	TimeSpentSeconds int64 `json:"timeSpentSeconds,omitempty"`

	// Started is when the work was started
	Started *time.Time `json:"started,omitempty"`

	// Comment describes the work performed
	Comment string `json:"comment,omitempty"`

	// Visibility controls who can see this worklog
	Visibility *WorklogVisibility `json:"visibility,omitempty"`

	// AdjustEstimate controls how the remaining estimate is affected
	AdjustEstimate *AdjustEstimate `json:"-"`
}

// AdjustEstimate specifies how to adjust the remaining estimate.
type AdjustEstimate struct {
	// Type is one of: "new", "leave", "manual", "auto"
	Type string

	// NewEstimate is the new remaining estimate (for type="new")
	NewEstimate string

	// ReduceBy reduces the estimate by this amount (for type="manual")
	ReduceBy string
}

// UpdateWorklogInput contains the data for updating a worklog.
type UpdateWorklogInput struct {
	TimeSpent        string             `json:"timeSpent,omitempty"`
	TimeSpentSeconds int64              `json:"timeSpentSeconds,omitempty"`
	Started          *time.Time         `json:"started,omitempty"`
	Comment          string             `json:"comment,omitempty"`
	Visibility       *WorklogVisibility `json:"visibility,omitempty"`
}

// ListWorklogsOptions contains options for listing worklogs.
type ListWorklogsOptions struct {
	// StartedAfter filters worklogs started after this date
	StartedAfter *time.Time

	// StartedBefore filters worklogs started before this date
	StartedBefore *time.Time

	// MaxResults limits the number of results
	MaxResults int

	// StartAt is the starting index for pagination
	StartAt int
}

// AddWorklog logs work on an issue.
//
// Example:
//
//	worklog, err := client.Issue.AddWorklog(ctx, "PROJ-123", &issue.AddWorklogInput{
//	    TimeSpent: "3h 20m",
//	    Started:   time.Now(),
//	    Comment:   "Implemented new feature",
//	})
func (s *Service) AddWorklog(ctx context.Context, issueKeyOrID string, input *AddWorklogInput) (*Worklog, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("worklog input is required")
	}

	if input.TimeSpent == "" && input.TimeSpentSeconds == 0 {
		return nil, fmt.Errorf("either timeSpent or timeSpentSeconds is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", issueKeyOrID)

	// Add query parameters for estimate adjustment
	if input.AdjustEstimate != nil {
		// TODO: Add query parameters based on AdjustEstimate type
		// For example: ?adjustEstimate=new&newEstimate=2d
	}

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
	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// GetWorklog retrieves a specific worklog.
//
// Example:
//
//	worklog, err := client.Issue.GetWorklog(ctx, "PROJ-123", "10000")
func (s *Service) GetWorklog(ctx context.Context, issueKeyOrID, worklogID string) (*Worklog, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if worklogID == "" {
		return nil, fmt.Errorf("worklog ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKeyOrID, worklogID)

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
	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// ListWorklogs retrieves all worklogs for an issue.
//
// Example:
//
//	worklogs, err := client.Issue.ListWorklogs(ctx, "PROJ-123", nil)
func (s *Service) ListWorklogs(ctx context.Context, issueKeyOrID string, opts *ListWorklogsOptions) ([]*Worklog, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", issueKeyOrID)

	// TODO: Add query parameters from opts

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
		Worklogs []*Worklog `json:"worklogs"`
	}
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Worklogs, nil
}

// UpdateWorklog updates an existing worklog.
//
// Example:
//
//	err := client.Issue.UpdateWorklog(ctx, "PROJ-123", "10000", &issue.UpdateWorklogInput{
//	    TimeSpent: "4h",
//	    Comment:   "Updated time estimate",
//	})
func (s *Service) UpdateWorklog(ctx context.Context, issueKeyOrID, worklogID string, input *UpdateWorklogInput) (*Worklog, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if worklogID == "" {
		return nil, fmt.Errorf("worklog ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("update worklog input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKeyOrID, worklogID)

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
	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// DeleteWorklog removes a worklog from an issue.
//
// Example:
//
//	err := client.Issue.DeleteWorklog(ctx, "PROJ-123", "10000")
func (s *Service) DeleteWorklog(ctx context.Context, issueKeyOrID, worklogID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if worklogID == "" {
		return fmt.Errorf("worklog ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKeyOrID, worklogID)

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

// Helper functions for time formatting

// ParseDuration converts a Jira time string to seconds.
// Examples: "3h 20m" -> 12000, "1d 4h" -> 100800
func ParseDuration(timeStr string) (int64, error) {
	// This is a simplified implementation
	// A full implementation would parse weeks, days, hours, minutes
	return 0, fmt.Errorf("not implemented")
}

// FormatDuration converts seconds to a Jira time string.
// Example: 12000 -> "3h 20m"
func FormatDuration(seconds int64) string {
	if seconds == 0 {
		return "0m"
	}

	weeks := seconds / (7 * 24 * 3600)
	seconds %= 7 * 24 * 3600

	days := seconds / (24 * 3600)
	seconds %= 24 * 3600

	hours := seconds / 3600
	seconds %= 3600

	minutes := seconds / 60

	result := ""
	if weeks > 0 {
		result += fmt.Sprintf("%dw ", weeks)
	}
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm", minutes)
	}

	return result
}
