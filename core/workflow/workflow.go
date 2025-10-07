// Package workflow provides Workflow resource management for Jira.
package workflow

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Workflow resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Workflow service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Transition represents a workflow transition.
type Transition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	To          *Status                `json:"to,omitempty"`
	HasScreen   bool                   `json:"hasScreen,omitempty"`
	IsGlobal    bool                   `json:"isGlobal,omitempty"`
	IsInitial   bool                   `json:"isInitial,omitempty"`
	IsAvailable bool                   `json:"isAvailable,omitempty"`
	IsConditional bool                 `json:"isConditional,omitempty"`
	Fields      map[string]FieldInfo   `json:"fields,omitempty"`
}

// Status represents an issue status.
type Status struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	IconURL        string          `json:"iconUrl,omitempty"`
	StatusCategory *StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a status category.
type StatusCategory struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	ColorName string `json:"colorName,omitempty"`
	Name      string `json:"name"`
}

// FieldInfo contains information about a field in a transition.
type FieldInfo struct {
	Required     bool     `json:"required"`
	Schema       Schema   `json:"schema,omitempty"`
	Name         string   `json:"name,omitempty"`
	Key          string   `json:"key,omitempty"`
	HasDefaultValue bool  `json:"hasDefaultValue,omitempty"`
	Operations   []string `json:"operations,omitempty"`
	AllowedValues []interface{} `json:"allowedValues,omitempty"`
}

// Schema represents a field schema.
type Schema struct {
	Type     string `json:"type,omitempty"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Custom   string `json:"custom,omitempty"`
	CustomID int64  `json:"customId,omitempty"`
}

// GetTransitionsOptions configures the GetTransitions operation.
type GetTransitionsOptions struct {
	// Expand specifies additional information to include
	Expand []string

	// TransitionID filters to a specific transition
	TransitionID string

	// SkipRemoteOnlyCondition whether transitions with conditions only for remote applications should be skipped
	SkipRemoteOnlyCondition bool
}

// GetTransitions retrieves available transitions for an issue.
//
// Example:
//
//	transitions, err := client.Workflow.GetTransitions(ctx, "PROJ-123", &workflow.GetTransitionsOptions{
//		Expand: []string{"transitions.fields"},
//	})
func (s *Service) GetTransitions(ctx context.Context, issueKeyOrID string, opts *GetTransitionsOptions) ([]*Transition, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKeyOrID)

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

		if opts.TransitionID != "" {
			q.Set("transitionId", opts.TransitionID)
		}

		if opts.SkipRemoteOnlyCondition {
			q.Set("skipRemoteOnlyCondition", "true")
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
		Expand      string        `json:"expand,omitempty"`
		Transitions []*Transition `json:"transitions"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Transitions, nil
}

// Workflow represents a Jira workflow.
type Workflow struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Statuses    []*Status            `json:"statuses,omitempty"`
	IsDefault   bool                 `json:"isDefault,omitempty"`
	Transitions []*Transition        `json:"transitions,omitempty"`
}

// ListOptions configures the List operation.
type ListOptions struct {
	// WorkflowName filters by workflow name
	WorkflowName string

	// MaxResults is the maximum number of results
	MaxResults int

	// StartAt is the index of the first result
	StartAt int
}

// List retrieves all workflows.
//
// Example:
//
//	workflows, err := client.Workflow.List(ctx, &workflow.ListOptions{
//		MaxResults: 50,
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Workflow, error) {
	path := "/rest/api/3/workflow/search"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()

		if opts.WorkflowName != "" {
			q.Set("workflowName", opts.WorkflowName)
		}

		if opts.MaxResults > 0 {
			q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
		}

		if opts.StartAt > 0 {
			q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
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
		Values     []*Workflow `json:"values"`
		MaxResults int         `json:"maxResults"`
		StartAt    int         `json:"startAt"`
		Total      int         `json:"total"`
		IsLast     bool        `json:"isLast"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// Get retrieves a workflow by ID or name.
//
// Example:
//
//	workflow, err := client.Workflow.Get(ctx, "classic-default-workflow")
func (s *Service) Get(ctx context.Context, workflowID string) (*Workflow, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflow/%s", workflowID)

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
	var workflow Workflow
	if err := s.transport.DecodeResponse(resp, &workflow); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workflow, nil
}

// GetAllStatuses retrieves all statuses.
//
// Example:
//
//	statuses, err := client.Workflow.GetAllStatuses(ctx)
func (s *Service) GetAllStatuses(ctx context.Context) ([]*Status, error) {
	path := "/rest/api/3/status"

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
	var statuses []*Status
	if err := s.transport.DecodeResponse(resp, &statuses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return statuses, nil
}

// GetStatus retrieves a status by ID.
//
// Example:
//
//	status, err := client.Workflow.GetStatus(ctx, "10000")
func (s *Service) GetStatus(ctx context.Context, statusID string) (*Status, error) {
	if statusID == "" {
		return nil, fmt.Errorf("status ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/status/%s", statusID)

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
	var status Status
	if err := s.transport.DecodeResponse(resp, &status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}
