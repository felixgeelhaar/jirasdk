// Package dashboard provides Dashboard resource management for Jira.
//
// Dashboards allow users to create customizable views with gadgets displaying
// project and issue information.
package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Dashboard resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Dashboard service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Dashboard represents a Jira dashboard.
type Dashboard struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Owner           *User             `json:"owner,omitempty"`
	SharePermissions []*SharePermission `json:"sharePermissions,omitempty"`
	EditPermissions  []*SharePermission `json:"editPermissions,omitempty"`
	Self            string            `json:"self,omitempty"`
	IsFavourite     bool              `json:"isFavourite,omitempty"`
	Rank            int               `json:"rank,omitempty"`
	View            string            `json:"view,omitempty"`
	IsWritable      bool              `json:"isWritable,omitempty"`
	SystemDashboard bool              `json:"systemDashboard,omitempty"`
}

// User represents a Jira user.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Self         string `json:"self,omitempty"`
}

// SharePermission represents sharing permissions for a dashboard.
type SharePermission struct {
	ID      int64  `json:"id,omitempty"`
	Type    string `json:"type"`
	Project *struct {
		ID  string `json:"id,omitempty"`
		Key string `json:"key,omitempty"`
	} `json:"project,omitempty"`
	Role *struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"role,omitempty"`
	Group *struct {
		Name string `json:"name,omitempty"`
	} `json:"group,omitempty"`
}

// DashboardGadget represents a gadget on a dashboard.
type DashboardGadget struct {
	ID         int64                  `json:"id,omitempty"`
	ModuleKey  string                 `json:"moduleKey"`
	URI        string                 `json:"uri,omitempty"`
	Color      string                 `json:"color,omitempty"`
	Position   *GadgetPosition        `json:"position,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GadgetPosition represents the position of a gadget on a dashboard.
type GadgetPosition struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

// ListOptions represents options for listing dashboards.
type ListOptions struct {
	Filter     string `json:"filter,omitempty"`
	StartAt    int    `json:"startAt,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
}

// List retrieves all dashboards visible to the user.
//
// Example:
//
//	dashboards, err := client.Dashboard.List(ctx, &dashboard.ListOptions{
//		Filter:     "favourite",
//		MaxResults: 50,
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Dashboard, error) {
	path := "/rest/api/3/dashboard"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.Filter != "" {
			q.Set("filter", opts.Filter)
		}
		if opts.StartAt > 0 {
			q.Set("startAt", strconv.Itoa(opts.StartAt))
		}
		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Dashboards []*Dashboard `json:"dashboards"`
		StartAt    int          `json:"startAt"`
		MaxResults int          `json:"maxResults"`
		Total      int          `json:"total"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Dashboards, nil
}

// Get retrieves a dashboard by ID.
//
// Example:
//
//	dashboard, err := client.Dashboard.Get(ctx, "10000")
func (s *Service) Get(ctx context.Context, dashboardID string) (*Dashboard, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s", dashboardID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var dashboard Dashboard
	if err := s.transport.DecodeResponse(resp, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// CreateDashboardInput represents input for creating a dashboard.
type CreateDashboardInput struct {
	Name             string             `json:"name"`
	Description      string             `json:"description,omitempty"`
	SharePermissions []*SharePermission `json:"sharePermissions,omitempty"`
	EditPermissions  []*SharePermission `json:"editPermissions,omitempty"`
}

// Create creates a new dashboard.
//
// Example:
//
//	dashboard, err := client.Dashboard.Create(ctx, &dashboard.CreateDashboardInput{
//		Name:        "My Dashboard",
//		Description: "Custom dashboard for project tracking",
//		SharePermissions: []*dashboard.SharePermission{
//			{Type: "global"},
//		},
//	})
func (s *Service) Create(ctx context.Context, input *CreateDashboardInput) (*Dashboard, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	path := "/rest/api/3/dashboard"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var dashboard Dashboard
	if err := s.transport.DecodeResponse(resp, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// UpdateDashboardInput represents input for updating a dashboard.
type UpdateDashboardInput struct {
	Name             string             `json:"name,omitempty"`
	Description      string             `json:"description,omitempty"`
	SharePermissions []*SharePermission `json:"sharePermissions,omitempty"`
	EditPermissions  []*SharePermission `json:"editPermissions,omitempty"`
}

// Update updates a dashboard.
//
// Example:
//
//	dashboard, err := client.Dashboard.Update(ctx, "10000", &dashboard.UpdateDashboardInput{
//		Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, dashboardID string, input *UpdateDashboardInput) (*Dashboard, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s", dashboardID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var dashboard Dashboard
	if err := s.transport.DecodeResponse(resp, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// Delete deletes a dashboard.
//
// Example:
//
//	err := client.Dashboard.Delete(ctx, "10000")
func (s *Service) Delete(ctx context.Context, dashboardID string) error {
	if dashboardID == "" {
		return fmt.Errorf("dashboard ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s", dashboardID)

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

// Copy copies a dashboard.
//
// Example:
//
//	dashboard, err := client.Dashboard.Copy(ctx, "10000", &dashboard.CreateDashboardInput{
//		Name:        "Copy of Dashboard",
//		Description: "Copied dashboard",
//	})
func (s *Service) Copy(ctx context.Context, dashboardID string, input *CreateDashboardInput) (*Dashboard, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s/copy", dashboardID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var dashboard Dashboard
	if err := s.transport.DecodeResponse(resp, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// GetGadgets retrieves all gadgets for a dashboard.
//
// Example:
//
//	gadgets, err := client.Dashboard.GetGadgets(ctx, "10000")
func (s *Service) GetGadgets(ctx context.Context, dashboardID string) ([]*DashboardGadget, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s/gadget", dashboardID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Gadgets []*DashboardGadget `json:"gadgets"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Gadgets, nil
}

// AddGadget adds a gadget to a dashboard.
//
// Example:
//
//	gadget, err := client.Dashboard.AddGadget(ctx, "10000", &dashboard.DashboardGadget{
//		ModuleKey: "com.atlassian.jira.gadgets:filter-results",
//		Position: &dashboard.GadgetPosition{Row: 0, Column: 0},
//		Title:    "My Filter",
//	})
func (s *Service) AddGadget(ctx context.Context, dashboardID string, gadget *DashboardGadget) (*DashboardGadget, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	if gadget == nil || gadget.ModuleKey == "" {
		return nil, fmt.Errorf("gadget with module key is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s/gadget", dashboardID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, gadget)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result DashboardGadget
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateGadget updates a gadget on a dashboard.
//
// Example:
//
//	gadget, err := client.Dashboard.UpdateGadget(ctx, "10000", 20000, &dashboard.DashboardGadget{
//		Title: "Updated Title",
//		Position: &dashboard.GadgetPosition{Row: 1, Column: 0},
//	})
func (s *Service) UpdateGadget(ctx context.Context, dashboardID string, gadgetID int64, gadget *DashboardGadget) (*DashboardGadget, error) {
	if dashboardID == "" {
		return nil, fmt.Errorf("dashboard ID is required")
	}

	if gadgetID <= 0 {
		return nil, fmt.Errorf("gadget ID is required")
	}

	if gadget == nil {
		return nil, fmt.Errorf("gadget is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s/gadget/%d", dashboardID, gadgetID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, gadget)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result DashboardGadget
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// RemoveGadget removes a gadget from a dashboard.
//
// Example:
//
//	err := client.Dashboard.RemoveGadget(ctx, "10000", 20000)
func (s *Service) RemoveGadget(ctx context.Context, dashboardID string, gadgetID int64) error {
	if dashboardID == "" {
		return fmt.Errorf("dashboard ID is required")
	}

	if gadgetID <= 0 {
		return fmt.Errorf("gadget ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/dashboard/%s/gadget/%d", dashboardID, gadgetID)

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
