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
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	To            *Status              `json:"to,omitempty"`
	HasScreen     bool                 `json:"hasScreen,omitempty"`
	IsGlobal      bool                 `json:"isGlobal,omitempty"`
	IsInitial     bool                 `json:"isInitial,omitempty"`
	IsAvailable   bool                 `json:"isAvailable,omitempty"`
	IsConditional bool                 `json:"isConditional,omitempty"`
	Fields        map[string]FieldInfo `json:"fields,omitempty"`
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
	Required        bool          `json:"required"`
	Schema          Schema        `json:"schema,omitempty"`
	Name            string        `json:"name,omitempty"`
	Key             string        `json:"key,omitempty"`
	HasDefaultValue bool          `json:"hasDefaultValue,omitempty"`
	Operations      []string      `json:"operations,omitempty"`
	AllowedValues   []interface{} `json:"allowedValues,omitempty"`
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
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Statuses    []*Status     `json:"statuses,omitempty"`
	IsDefault   bool          `json:"isDefault,omitempty"`
	Transitions []*Transition `json:"transitions,omitempty"`
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

// WorkflowScheme represents a workflow scheme that maps issue types to workflows.
type WorkflowScheme struct {
	ID                int64             `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	DefaultWorkflow   string            `json:"defaultWorkflow,omitempty"`
	IssueTypeMappings map[string]string `json:"issueTypeMappings,omitempty"`
	Draft             bool              `json:"draft"`
	LastModifiedUser  *User             `json:"lastModifiedUser,omitempty"`
	LastModified      string            `json:"lastModified,omitempty"`
	Self              string            `json:"self,omitempty"`
}

// User represents a minimal user reference.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active"`
}

// GetWorkflowScheme retrieves a workflow scheme by ID.
//
// Example:
//
//	scheme, err := client.Workflow.GetWorkflowScheme(ctx, 10000)
func (s *Service) GetWorkflowScheme(ctx context.Context, schemeID int64) (*WorkflowScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("workflow scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflowscheme/%d", schemeID)

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
	var scheme WorkflowScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// ListWorkflowSchemesOptions configures the ListWorkflowSchemes operation.
type ListWorkflowSchemesOptions struct {
	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results
	MaxResults int
}

// ListWorkflowSchemes retrieves all workflow schemes.
//
// Example:
//
//	schemes, err := client.Workflow.ListWorkflowSchemes(ctx, nil)
func (s *Service) ListWorkflowSchemes(ctx context.Context, opts *ListWorkflowSchemesOptions) ([]*WorkflowScheme, error) {
	path := "/rest/api/3/workflowscheme"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.StartAt > 0 {
			q.Set("startAt", fmt.Sprintf("%d", opts.StartAt))
		}
		if opts.MaxResults > 0 {
			q.Set("maxResults", fmt.Sprintf("%d", opts.MaxResults))
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
		Values []*WorkflowScheme `json:"values"`
	}
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// CreateWorkflowSchemeInput represents input for creating a workflow scheme.
type CreateWorkflowSchemeInput struct {
	Name              string            `json:"name"`
	Description       string            `json:"description,omitempty"`
	DefaultWorkflow   string            `json:"defaultWorkflow,omitempty"`
	IssueTypeMappings map[string]string `json:"issueTypeMappings,omitempty"`
}

// CreateWorkflowScheme creates a new workflow scheme.
//
// Example:
//
//	scheme, err := client.Workflow.CreateWorkflowScheme(ctx, &workflow.CreateWorkflowSchemeInput{
//		Name:            "My Workflow Scheme",
//		Description:     "Custom workflow scheme",
//		DefaultWorkflow: "classic-default-workflow",
//		IssueTypeMappings: map[string]string{
//			"10001": "software-simplified-workflow",
//		},
//	})
func (s *Service) CreateWorkflowScheme(ctx context.Context, input *CreateWorkflowSchemeInput) (*WorkflowScheme, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	path := "/rest/api/3/workflowscheme"

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
	var scheme WorkflowScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// UpdateWorkflowSchemeInput represents input for updating a workflow scheme.
type UpdateWorkflowSchemeInput struct {
	Name              string            `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	DefaultWorkflow   string            `json:"defaultWorkflow,omitempty"`
	IssueTypeMappings map[string]string `json:"issueTypeMappings,omitempty"`
}

// UpdateWorkflowScheme updates a workflow scheme.
//
// Example:
//
//	scheme, err := client.Workflow.UpdateWorkflowScheme(ctx, 10000, &workflow.UpdateWorkflowSchemeInput{
//		Description: "Updated description",
//	})
func (s *Service) UpdateWorkflowScheme(ctx context.Context, schemeID int64, input *UpdateWorkflowSchemeInput) (*WorkflowScheme, error) {
	if schemeID <= 0 {
		return nil, fmt.Errorf("workflow scheme ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflowscheme/%d", schemeID)

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
	var scheme WorkflowScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}

// DeleteWorkflowScheme deletes a workflow scheme.
//
// Example:
//
//	err := client.Workflow.DeleteWorkflowScheme(ctx, 10000)
func (s *Service) DeleteWorkflowScheme(ctx context.Context, schemeID int64) error {
	if schemeID <= 0 {
		return fmt.Errorf("workflow scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflowscheme/%d", schemeID)

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

// GetStatusCategories retrieves all status categories.
//
// Example:
//
//	categories, err := client.Workflow.GetStatusCategories(ctx)
func (s *Service) GetStatusCategories(ctx context.Context) ([]*StatusCategory, error) {
	path := "/rest/api/3/statuscategory"

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
	var categories []*StatusCategory
	if err := s.transport.DecodeResponse(resp, &categories); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return categories, nil
}

// GetStatusCategory retrieves a status category by ID or key.
//
// Example:
//
//	category, err := client.Workflow.GetStatusCategory(ctx, "2")
func (s *Service) GetStatusCategory(ctx context.Context, idOrKey string) (*StatusCategory, error) {
	if idOrKey == "" {
		return nil, fmt.Errorf("status category ID or key is required")
	}

	path := fmt.Sprintf("/rest/api/3/statuscategory/%s", idOrKey)

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
	var category StatusCategory
	if err := s.transport.DecodeResponse(resp, &category); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &category, nil
}

// DoTransitionInput represents input for performing a transition.
type DoTransitionInput struct {
	Transition *TransitionInput       `json:"transition"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Update     map[string]interface{} `json:"update,omitempty"`
}

// TransitionInput contains transition details.
type TransitionInput struct {
	ID string `json:"id"`
}

// DoTransition performs a workflow transition on an issue.
//
// Example:
//
//	err := client.Workflow.DoTransition(ctx, "PROJ-123", &workflow.DoTransitionInput{
//		Transition: &workflow.TransitionInput{
//			ID: "31",
//		},
//		Fields: map[string]interface{}{
//			"resolution": map[string]string{
//				"name": "Fixed",
//			},
//		},
//	})
func (s *Service) DoTransition(ctx context.Context, issueKeyOrID string, input *DoTransitionInput) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if input == nil || input.Transition == nil {
		return fmt.Errorf("transition input is required")
	}

	if input.Transition.ID == "" {
		return fmt.Errorf("transition ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
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

// WorkflowSchemeIssueType represents the issue type workflow mapping.
type WorkflowSchemeIssueType struct {
	IssueType    string `json:"issueType"`
	Workflow     string `json:"workflow,omitempty"`
	UpdateDraft  bool   `json:"updateDraftIfNeeded,omitempty"`
}

// SetWorkflowSchemeIssueType sets the workflow for an issue type in a workflow scheme.
//
// Example:
//
//	err := client.Workflow.SetWorkflowSchemeIssueType(ctx, 10000, &workflow.WorkflowSchemeIssueType{
//		IssueType: "10001",
//		Workflow:  "software-simplified-workflow",
//	})
func (s *Service) SetWorkflowSchemeIssueType(ctx context.Context, schemeID int64, input *WorkflowSchemeIssueType) error {
	if schemeID <= 0 {
		return fmt.Errorf("workflow scheme ID is required")
	}

	if input == nil || input.IssueType == "" {
		return fmt.Errorf("issue type is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflowscheme/%d/issuetype/%s", schemeID, input.IssueType)

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

// DeleteWorkflowSchemeIssueType removes the workflow for an issue type in a workflow scheme.
//
// Example:
//
//	err := client.Workflow.DeleteWorkflowSchemeIssueType(ctx, 10000, "10001")
func (s *Service) DeleteWorkflowSchemeIssueType(ctx context.Context, schemeID int64, issueType string) error {
	if schemeID <= 0 {
		return fmt.Errorf("workflow scheme ID is required")
	}

	if issueType == "" {
		return fmt.Errorf("issue type is required")
	}

	path := fmt.Sprintf("/rest/api/3/workflowscheme/%d/issuetype/%s", schemeID, issueType)

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
