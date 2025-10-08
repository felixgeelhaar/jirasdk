// Package agile provides Agile/Scrum resource management for Jira Software.
package agile

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Agile resources (boards, sprints, epics).
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Agile service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Board represents a Jira Software board (Scrum or Kanban).
type Board struct {
	ID       int64     `json:"id"`
	Self     string    `json:"self,omitempty"`
	Name     string    `json:"name"`
	Type     string    `json:"type"` // "scrum" or "kanban"
	Location *Location `json:"location,omitempty"`
	Filter   *Filter   `json:"filter,omitempty"`
	CanEdit  bool      `json:"canEdit,omitempty"`
}

// Location represents the board's project location.
type Location struct {
	ProjectID      int64  `json:"projectId"`
	ProjectKey     string `json:"projectKey,omitempty"`
	ProjectName    string `json:"projectName,omitempty"`
	ProjectTypeKey string `json:"projectTypeKey,omitempty"`
	DisplayName    string `json:"displayName,omitempty"`
	Name           string `json:"name,omitempty"`
}

// Filter represents a board filter.
type Filter struct {
	ID   int64  `json:"id"`
	Self string `json:"self,omitempty"`
}

// Sprint represents a Scrum sprint.
type Sprint struct {
	ID            int64  `json:"id"`
	Self          string `json:"self,omitempty"`
	State         string `json:"state"` // "future", "active", "closed"
	Name          string `json:"name"`
	StartDate     string `json:"startDate,omitempty"`    // ISO 8601 format
	EndDate       string `json:"endDate,omitempty"`      // ISO 8601 format
	CompleteDate  string `json:"completeDate,omitempty"` // ISO 8601 format
	OriginBoardID int64  `json:"originBoardId,omitempty"`
	Goal          string `json:"goal,omitempty"`
}

// Epic represents an Epic issue.
type Epic struct {
	ID      int64  `json:"id"`
	Self    string `json:"self,omitempty"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Color   *Color `json:"color,omitempty"`
	Done    bool   `json:"done"`
}

// Color represents an epic color.
type Color struct {
	Key string `json:"key"`
}

// BoardsOptions configures the GetBoards operation.
type BoardsOptions struct {
	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results
	MaxResults int

	// Type filters by board type ("scrum" or "kanban")
	Type string

	// Name filters by board name
	Name string

	// ProjectKeyOrID filters by project
	ProjectKeyOrID string
}

// GetBoards retrieves all boards with optional filtering.
//
// Example:
//
//	boards, err := client.Agile.GetBoards(ctx, &agile.BoardsOptions{
//	    Type: "scrum",
//	    MaxResults: 50,
//	})
func (s *Service) GetBoards(ctx context.Context, opts *BoardsOptions) ([]*Board, error) {
	path := "/rest/agile/1.0/board"

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

		if opts.Type != "" {
			q.Set("type", opts.Type)
		}

		if opts.Name != "" {
			q.Set("name", opts.Name)
		}

		if opts.ProjectKeyOrID != "" {
			q.Set("projectKeyOrId", opts.ProjectKeyOrID)
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
		MaxResults int      `json:"maxResults"`
		StartAt    int      `json:"startAt"`
		Total      int      `json:"total"`
		IsLast     bool     `json:"isLast"`
		Values     []*Board `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// GetBoard retrieves a board by ID.
//
// Example:
//
//	board, err := client.Agile.GetBoard(ctx, 123)
func (s *Service) GetBoard(ctx context.Context, boardID int64) (*Board, error) {
	if boardID <= 0 {
		return nil, fmt.Errorf("board ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/board/%d", boardID)

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
	var board Board
	if err := s.transport.DecodeResponse(resp, &board); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &board, nil
}

// CreateBoardInput contains the data for creating a board.
type CreateBoardInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "scrum" or "kanban"
	FilterID int64  `json:"filterId"`
}

// CreateBoard creates a new board.
//
// Example:
//
//	board, err := client.Agile.CreateBoard(ctx, &agile.CreateBoardInput{
//	    Name:     "Sprint Board",
//	    Type:     "scrum",
//	    FilterID: 10000,
//	})
func (s *Service) CreateBoard(ctx context.Context, input *CreateBoardInput) (*Board, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("board name is required")
	}

	if input.Type == "" {
		return nil, fmt.Errorf("board type is required")
	}

	if input.FilterID <= 0 {
		return nil, fmt.Errorf("filter ID is required")
	}

	path := "/rest/agile/1.0/board"

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
	var board Board
	if err := s.transport.DecodeResponse(resp, &board); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &board, nil
}

// DeleteBoard deletes a board.
//
// Example:
//
//	err := client.Agile.DeleteBoard(ctx, 123)
func (s *Service) DeleteBoard(ctx context.Context, boardID int64) error {
	if boardID <= 0 {
		return fmt.Errorf("board ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/board/%d", boardID)

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

// SprintsOptions configures the GetBoardSprints operation.
type SprintsOptions struct {
	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results
	MaxResults int

	// State filters by sprint state (comma-separated: "future,active,closed")
	State string
}

// GetBoardSprints retrieves all sprints for a board.
//
// Example:
//
//	sprints, err := client.Agile.GetBoardSprints(ctx, 123, &agile.SprintsOptions{
//	    State: "active,future",
//	    MaxResults: 50,
//	})
func (s *Service) GetBoardSprints(ctx context.Context, boardID int64, opts *SprintsOptions) ([]*Sprint, error) {
	if boardID <= 0 {
		return nil, fmt.Errorf("board ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint", boardID)

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

		if opts.State != "" {
			q.Set("state", opts.State)
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
		MaxResults int       `json:"maxResults"`
		StartAt    int       `json:"startAt"`
		IsLast     bool      `json:"isLast"`
		Values     []*Sprint `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// GetSprint retrieves a sprint by ID.
//
// Example:
//
//	sprint, err := client.Agile.GetSprint(ctx, 456)
func (s *Service) GetSprint(ctx context.Context, sprintID int64) (*Sprint, error) {
	if sprintID <= 0 {
		return nil, fmt.Errorf("sprint ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID)

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
	var sprint Sprint
	if err := s.transport.DecodeResponse(resp, &sprint); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sprint, nil
}

// CreateSprintInput contains the data for creating a sprint.
type CreateSprintInput struct {
	Name          string `json:"name"`
	StartDate     string `json:"startDate,omitempty"` // ISO 8601 format
	EndDate       string `json:"endDate,omitempty"`   // ISO 8601 format
	OriginBoardID int64  `json:"originBoardId"`
	Goal          string `json:"goal,omitempty"`
}

// CreateSprint creates a new sprint.
//
// Example:
//
//	sprint, err := client.Agile.CreateSprint(ctx, &agile.CreateSprintInput{
//	    Name:          "Sprint 25",
//	    OriginBoardID: 123,
//	    StartDate:     "2024-06-01T09:00:00.000Z",
//	    EndDate:       "2024-06-14T17:00:00.000Z",
//	    Goal:          "Complete user authentication",
//	})
func (s *Service) CreateSprint(ctx context.Context, input *CreateSprintInput) (*Sprint, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("sprint name is required")
	}

	if input.OriginBoardID <= 0 {
		return nil, fmt.Errorf("origin board ID is required")
	}

	path := "/rest/agile/1.0/sprint"

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
	var sprint Sprint
	if err := s.transport.DecodeResponse(resp, &sprint); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sprint, nil
}

// UpdateSprintInput contains the data for updating a sprint.
type UpdateSprintInput struct {
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"` // "future", "active", "closed"
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

// UpdateSprint updates a sprint.
//
// Example:
//
//	sprint, err := client.Agile.UpdateSprint(ctx, 456, &agile.UpdateSprintInput{
//	    State: "active",
//	    Goal:  "Updated goal",
//	})
func (s *Service) UpdateSprint(ctx context.Context, sprintID int64, input *UpdateSprintInput) (*Sprint, error) {
	if sprintID <= 0 {
		return nil, fmt.Errorf("sprint ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID)

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
	var sprint Sprint
	if err := s.transport.DecodeResponse(resp, &sprint); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sprint, nil
}

// DeleteSprint deletes a sprint.
//
// Example:
//
//	err := client.Agile.DeleteSprint(ctx, 456)
func (s *Service) DeleteSprint(ctx context.Context, sprintID int64) error {
	if sprintID <= 0 {
		return fmt.Errorf("sprint ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprintID)

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

// EpicsOptions configures the GetBoardEpics operation.
type EpicsOptions struct {
	// StartAt is the starting index for pagination
	StartAt int

	// MaxResults limits the number of results
	MaxResults int

	// Done filters by completion status
	Done *bool
}

// GetBoardEpics retrieves all epics for a board.
//
// Example:
//
//	epics, err := client.Agile.GetBoardEpics(ctx, 123, &agile.EpicsOptions{
//	    MaxResults: 50,
//	})
func (s *Service) GetBoardEpics(ctx context.Context, boardID int64, opts *EpicsOptions) ([]*Epic, error) {
	if boardID <= 0 {
		return nil, fmt.Errorf("board ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/board/%d/epic", boardID)

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

		if opts.Done != nil {
			q.Set("done", fmt.Sprintf("%t", *opts.Done))
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
		MaxResults int     `json:"maxResults"`
		StartAt    int     `json:"startAt"`
		Total      int     `json:"total"`
		IsLast     bool    `json:"isLast"`
		Values     []*Epic `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// GetEpic retrieves an epic by ID.
//
// Example:
//
//	epic, err := client.Agile.GetEpic(ctx, 789)
func (s *Service) GetEpic(ctx context.Context, epicID int64) (*Epic, error) {
	if epicID <= 0 {
		return nil, fmt.Errorf("epic ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/epic/%d", epicID)

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
	var epic Epic
	if err := s.transport.DecodeResponse(resp, &epic); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &epic, nil
}

// MoveIssuesToSprintInput contains the data for moving issues to a sprint.
type MoveIssuesToSprintInput struct {
	Issues []string `json:"issues"` // Issue keys or IDs
}

// MoveIssuesToSprint moves issues to a sprint.
//
// Example:
//
//	err := client.Agile.MoveIssuesToSprint(ctx, 456, &agile.MoveIssuesToSprintInput{
//	    Issues: []string{"PROJ-123", "PROJ-124"},
//	})
func (s *Service) MoveIssuesToSprint(ctx context.Context, sprintID int64, input *MoveIssuesToSprintInput) error {
	if sprintID <= 0 {
		return fmt.Errorf("sprint ID is required")
	}

	if input == nil || len(input.Issues) == 0 {
		return fmt.Errorf("at least one issue is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue", sprintID)

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

// GetBacklog retrieves backlog issues for a board.
//
// Example:
//
//	issues, err := client.Agile.GetBacklog(ctx, 123, nil)
func (s *Service) GetBacklog(ctx context.Context, boardID int64, opts *BoardsOptions) ([]interface{}, error) {
	if boardID <= 0 {
		return nil, fmt.Errorf("board ID is required")
	}

	path := fmt.Sprintf("/rest/agile/1.0/board/%d/backlog", boardID)

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
		MaxResults int           `json:"maxResults"`
		StartAt    int           `json:"startAt"`
		Total      int           `json:"total"`
		Issues     []interface{} `json:"issues"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Issues, nil
}
