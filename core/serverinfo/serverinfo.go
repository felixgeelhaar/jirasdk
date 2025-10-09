// Package serverinfo provides Server Information for Jira.
//
// Server info provides details about the Jira instance including version,
// build number, and deployment type.
package serverinfo

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Server Information.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Server Info service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// ServerInfo represents Jira server information.
type ServerInfo struct {
	BaseURL        string         `json:"baseUrl"`
	Version        string         `json:"version"`
	VersionNumbers []int          `json:"versionNumbers"`
	DeploymentType string         `json:"deploymentType"`
	BuildNumber    int            `json:"buildNumber"`
	BuildDate      string         `json:"buildDate"`
	ServerTime     string         `json:"serverTime"`
	ScmInfo        string         `json:"scmInfo"`
	ServerTitle    string         `json:"serverTitle"`
	HealthChecks   []*HealthCheck `json:"healthChecks,omitempty"`
}

// HealthCheck represents a health check result.
type HealthCheck struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Passed      bool   `json:"passed"`
}

// Get retrieves server information.
//
// Example:
//
//	info, err := client.ServerInfo.Get(ctx)
func (s *Service) Get(ctx context.Context) (*ServerInfo, error) {
	path := "/rest/api/3/serverInfo"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var info ServerInfo
	if err := s.transport.DecodeResponse(resp, &info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

// Configuration represents Jira configuration.
type Configuration struct {
	VotingEnabled             bool                       `json:"votingEnabled"`
	WatchingEnabled           bool                       `json:"watchingEnabled"`
	UnassignedIssuesAllowed   bool                       `json:"unassignedIssuesAllowed"`
	SubTasksEnabled           bool                       `json:"subTasksEnabled"`
	IssueLinkingEnabled       bool                       `json:"issueLinkingEnabled"`
	TimeTrackingEnabled       bool                       `json:"timeTrackingEnabled"`
	AttachmentsEnabled        bool                       `json:"attachmentsEnabled"`
	TimeTrackingConfiguration *TimeTrackingConfiguration `json:"timeTrackingConfiguration,omitempty"`
}

// TimeTrackingConfiguration represents time tracking settings.
type TimeTrackingConfiguration struct {
	WorkingHoursPerDay float64 `json:"workingHoursPerDay"`
	WorkingDaysPerWeek float64 `json:"workingDaysPerWeek"`
	TimeFormat         string  `json:"timeFormat"`
	DefaultUnit        string  `json:"defaultUnit"`
}

// GetConfiguration retrieves Jira configuration.
//
// Example:
//
//	config, err := client.ServerInfo.GetConfiguration(ctx)
func (s *Service) GetConfiguration(ctx context.Context) (*Configuration, error) {
	path := "/rest/api/3/configuration"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var config Configuration
	if err := s.transport.DecodeResponse(resp, &config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}
