// Package timetracking provides Time Tracking resource management for Jira.
//
// Time tracking allows teams to estimate and log work on issues.
// This package provides operations for managing time tracking settings and worklog entries.
package timetracking

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Time Tracking resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Time Tracking service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// TimeTrackingProvider represents a time tracking provider configuration.
type TimeTrackingProvider struct {
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// TimeTrackingConfiguration represents time tracking settings.
type TimeTrackingConfiguration struct {
	WorkingHoursPerDay float64 `json:"workingHoursPerDay"`
	WorkingDaysPerWeek float64 `json:"workingDaysPerWeek"`
	TimeFormat         string  `json:"timeFormat"`
	DefaultUnit        string  `json:"defaultUnit"`
}

// UpdateTimeTrackingConfigurationInput represents input for updating time tracking configuration.
type UpdateTimeTrackingConfigurationInput struct {
	WorkingHoursPerDay float64 `json:"workingHoursPerDay,omitempty"`
	WorkingDaysPerWeek float64 `json:"workingDaysPerWeek,omitempty"`
	TimeFormat         string  `json:"timeFormat,omitempty"`
	DefaultUnit        string  `json:"defaultUnit,omitempty"`
}

// Worklog represents a work log entry on an issue.
type Worklog struct {
	Self             string            `json:"self,omitempty"`
	Author           *User             `json:"author,omitempty"`
	UpdateAuthor     *User             `json:"updateAuthor,omitempty"`
	Comment          string            `json:"comment,omitempty"`
	Created          string            `json:"created,omitempty"`
	Updated          string            `json:"updated,omitempty"`
	Started          string            `json:"started,omitempty"`
	TimeSpent        string            `json:"timeSpent,omitempty"`
	TimeSpentSeconds int64             `json:"timeSpentSeconds,omitempty"`
	ID               string            `json:"id,omitempty"`
	IssueID          string            `json:"issueId,omitempty"`
	Properties       []WorklogProperty `json:"properties,omitempty"`
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

// WorklogProperty represents a property on a worklog.
type WorklogProperty struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// CreateWorklogInput represents input for creating a worklog.
type CreateWorklogInput struct {
	Comment          string            `json:"comment,omitempty"`
	Started          string            `json:"started,omitempty"`
	TimeSpent        string            `json:"timeSpent,omitempty"`
	TimeSpentSeconds int64             `json:"timeSpentSeconds,omitempty"`
	Properties       []WorklogProperty `json:"properties,omitempty"`
}

// UpdateWorklogInput represents input for updating a worklog.
type UpdateWorklogInput struct {
	Comment          string            `json:"comment,omitempty"`
	Started          string            `json:"started,omitempty"`
	TimeSpent        string            `json:"timeSpent,omitempty"`
	TimeSpentSeconds int64             `json:"timeSpentSeconds,omitempty"`
	Properties       []WorklogProperty `json:"properties,omitempty"`
}

// WorklogListOptions represents options for listing worklogs.
type WorklogListOptions struct {
	StartAt       int    `json:"startAt,omitempty"`
	MaxResults    int    `json:"maxResults,omitempty"`
	StartedAfter  int64  `json:"startedAfter,omitempty"`
	StartedBefore int64  `json:"startedBefore,omitempty"`
	Expand        string `json:"expand,omitempty"`
}

// GetAvailableProviders retrieves available time tracking providers.
//
// Example:
//
//	providers, err := client.TimeTracking.GetAvailableProviders(ctx)
func (s *Service) GetAvailableProviders(ctx context.Context) ([]*TimeTrackingProvider, error) {
	path := "/rest/api/3/configuration/timetracking/list"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var providers []*TimeTrackingProvider
	if err := s.transport.DecodeResponse(resp, &providers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return providers, nil
}

// GetSelectedProvider retrieves the selected time tracking provider.
//
// Example:
//
//	provider, err := client.TimeTracking.GetSelectedProvider(ctx)
func (s *Service) GetSelectedProvider(ctx context.Context) (*TimeTrackingProvider, error) {
	path := "/rest/api/3/configuration/timetracking"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var provider TimeTrackingProvider
	if err := s.transport.DecodeResponse(resp, &provider); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &provider, nil
}

// SelectProvider sets the time tracking provider.
//
// Example:
//
//	provider, err := client.TimeTracking.SelectProvider(ctx, "JIRA")
func (s *Service) SelectProvider(ctx context.Context, key string) (*TimeTrackingProvider, error) {
	if key == "" {
		return nil, fmt.Errorf("provider key is required")
	}

	path := "/rest/api/3/configuration/timetracking"

	input := TimeTrackingProvider{
		Key: key,
	}

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var provider TimeTrackingProvider
	if err := s.transport.DecodeResponse(resp, &provider); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &provider, nil
}

// GetConfiguration retrieves time tracking configuration settings.
//
// Example:
//
//	config, err := client.TimeTracking.GetConfiguration(ctx)
func (s *Service) GetConfiguration(ctx context.Context) (*TimeTrackingConfiguration, error) {
	path := "/rest/api/3/configuration/timetracking/options"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var config TimeTrackingConfiguration
	if err := s.transport.DecodeResponse(resp, &config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// UpdateConfiguration updates time tracking configuration settings.
//
// Example:
//
//	config, err := client.TimeTracking.UpdateConfiguration(ctx, &timetracking.UpdateTimeTrackingConfigurationInput{
//	    WorkingHoursPerDay: 8.0,
//	    WorkingDaysPerWeek: 5.0,
//	    TimeFormat:         "pretty",
//	    DefaultUnit:        "hour",
//	})
func (s *Service) UpdateConfiguration(ctx context.Context, input *UpdateTimeTrackingConfigurationInput) (*TimeTrackingConfiguration, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := "/rest/api/3/configuration/timetracking/options"

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var config TimeTrackingConfiguration
	if err := s.transport.DecodeResponse(resp, &config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// GetIssueWorklogs retrieves all worklogs for an issue.
//
// Example:
//
//	worklogs, err := client.TimeTracking.GetIssueWorklogs(ctx, "PROJ-123", &timetracking.WorklogListOptions{MaxResults: 50})
func (s *Service) GetIssueWorklogs(ctx context.Context, issueKey string, opts *WorklogListOptions) ([]*Worklog, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", issueKey)

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
		if opts.StartedAfter > 0 {
			q.Set("startedAfter", strconv.FormatInt(opts.StartedAfter, 10))
		}
		if opts.StartedBefore > 0 {
			q.Set("startedBefore", strconv.FormatInt(opts.StartedBefore, 10))
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
		Worklogs []*Worklog `json:"worklogs"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Worklogs, nil
}

// GetWorklog retrieves a specific worklog by ID.
//
// Example:
//
//	worklog, err := client.TimeTracking.GetWorklog(ctx, "PROJ-123", "10000")
func (s *Service) GetWorklog(ctx context.Context, issueKey, worklogID string) (*Worklog, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	if worklogID == "" {
		return nil, fmt.Errorf("worklog ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKey, worklogID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// CreateWorklog adds a worklog to an issue.
//
// Example:
//
//	worklog, err := client.TimeTracking.CreateWorklog(ctx, "PROJ-123", &timetracking.CreateWorklogInput{
//	    Comment:          "Worked on bug fix",
//	    TimeSpent:        "3h 30m",
//	    Started:          "2024-01-15T09:00:00.000+0000",
//	})
func (s *Service) CreateWorklog(ctx context.Context, issueKey string, input *CreateWorklogInput) (*Worklog, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.TimeSpent == "" && input.TimeSpentSeconds == 0 {
		return nil, fmt.Errorf("either timeSpent or timeSpentSeconds is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", issueKey)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// UpdateWorklog updates a worklog.
//
// Example:
//
//	worklog, err := client.TimeTracking.UpdateWorklog(ctx, "PROJ-123", "10000", &timetracking.UpdateWorklogInput{
//	    Comment:   "Updated work description",
//	    TimeSpent: "4h",
//	})
func (s *Service) UpdateWorklog(ctx context.Context, issueKey, worklogID string, input *UpdateWorklogInput) (*Worklog, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	if worklogID == "" {
		return nil, fmt.Errorf("worklog ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKey, worklogID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var worklog Worklog
	if err := s.transport.DecodeResponse(resp, &worklog); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &worklog, nil
}

// DeleteWorklog deletes a worklog.
//
// Example:
//
//	err := client.TimeTracking.DeleteWorklog(ctx, "PROJ-123", "10000")
func (s *Service) DeleteWorklog(ctx context.Context, issueKey, worklogID string) error {
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}

	if worklogID == "" {
		return fmt.Errorf("worklog ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog/%s", issueKey, worklogID)

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
