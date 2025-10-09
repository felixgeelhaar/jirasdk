// Package screen provides Screen resource management for Jira.
//
// Screens define which fields are displayed during issue operations (create, edit, view).
// This package provides operations for managing screens, screen tabs, and screen schemes.
package screen

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Screen resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Screen service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Screen represents a Jira screen.
type Screen struct {
	ID          int64        `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Scope       *ScreenScope `json:"scope,omitempty"`
	Tabs        []*ScreenTab `json:"tabs,omitempty"`
}

// ScreenScope represents the scope of a screen.
type ScreenScope struct {
	Type    string   `json:"type"`
	Project *Project `json:"project,omitempty"`
}

// Project represents a simplified Jira project.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

// ScreenTab represents a tab within a screen.
type ScreenTab struct {
	ID       int64          `json:"id,omitempty"`
	Name     string         `json:"name"`
	Position int            `json:"position,omitempty"`
	Fields   []*ScreenField `json:"fields,omitempty"`
}

// ScreenField represents a field within a screen tab.
type ScreenField struct {
	ID       string `json:"id"`
	Position int    `json:"position,omitempty"`
}

// CreateScreenInput represents input for creating a screen.
type CreateScreenInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateScreenInput represents input for updating a screen.
type UpdateScreenInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateTabInput represents input for creating a screen tab.
type CreateTabInput struct {
	Name string `json:"name"`
}

// UpdateTabInput represents input for updating a screen tab.
type UpdateTabInput struct {
	Name string `json:"name,omitempty"`
}

// AddFieldInput represents input for adding a field to a screen tab.
type AddFieldInput struct {
	FieldID string `json:"fieldId"`
}

// ListOptions represents options for listing screens.
type ListOptions struct {
	StartAt    int `json:"startAt,omitempty"`
	MaxResults int `json:"maxResults,omitempty"`
}

// List retrieves all screens with pagination.
//
// Example:
//
//	screens, err := client.Screen.List(ctx, &screen.ListOptions{MaxResults: 50})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Screen, error) {
	path := "/rest/api/3/screens"

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
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*Screen `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// Get retrieves a specific screen by ID.
//
// Example:
//
//	screen, err := client.Screen.Get(ctx, 10000)
func (s *Service) Get(ctx context.Context, screenID int64) (*Screen, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d", screenID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var screen Screen
	if err := s.transport.DecodeResponse(resp, &screen); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &screen, nil
}

// Create creates a new screen.
//
// Example:
//
//	screen, err := client.Screen.Create(ctx, &screen.CreateScreenInput{
//	    Name:        "Bug Tracking Screen",
//	    Description: "Screen for bug tracking workflows",
//	})
func (s *Service) Create(ctx context.Context, input *CreateScreenInput) (*Screen, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("screen name is required")
	}

	path := "/rest/api/3/screens"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var screen Screen
	if err := s.transport.DecodeResponse(resp, &screen); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &screen, nil
}

// Update updates a screen.
//
// Example:
//
//	screen, err := client.Screen.Update(ctx, 10000, &screen.UpdateScreenInput{
//	    Name:        "Updated Bug Screen",
//	    Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, screenID int64, input *UpdateScreenInput) (*Screen, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d", screenID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var screen Screen
	if err := s.transport.DecodeResponse(resp, &screen); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &screen, nil
}

// Delete deletes a screen.
//
// Example:
//
//	err := client.Screen.Delete(ctx, 10000)
func (s *Service) Delete(ctx context.Context, screenID int64) error {
	if screenID <= 0 {
		return fmt.Errorf("screen ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d", screenID)

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

// GetAvailableFields retrieves available fields that can be added to a screen.
//
// Example:
//
//	fields, err := client.Screen.GetAvailableFields(ctx, 10000)
func (s *Service) GetAvailableFields(ctx context.Context, screenID int64) ([]*ScreenField, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/availableFields", screenID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var fields []*ScreenField
	if err := s.transport.DecodeResponse(resp, &fields); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return fields, nil
}

// ListTabs retrieves all tabs for a screen.
//
// Example:
//
//	tabs, err := client.Screen.ListTabs(ctx, 10000)
func (s *Service) ListTabs(ctx context.Context, screenID int64) ([]*ScreenTab, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs", screenID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var tabs []*ScreenTab
	if err := s.transport.DecodeResponse(resp, &tabs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tabs, nil
}

// CreateTab creates a new tab for a screen.
//
// Example:
//
//	tab, err := client.Screen.CreateTab(ctx, 10000, &screen.CreateTabInput{
//	    Name: "Details Tab",
//	})
func (s *Service) CreateTab(ctx context.Context, screenID int64, input *CreateTabInput) (*ScreenTab, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("tab name is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs", screenID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var tab ScreenTab
	if err := s.transport.DecodeResponse(resp, &tab); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tab, nil
}

// UpdateTab updates a screen tab.
//
// Example:
//
//	tab, err := client.Screen.UpdateTab(ctx, 10000, 10100, &screen.UpdateTabInput{
//	    Name: "Updated Tab Name",
//	})
func (s *Service) UpdateTab(ctx context.Context, screenID, tabID int64, input *UpdateTabInput) (*ScreenTab, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	if tabID <= 0 {
		return nil, fmt.Errorf("tab ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs/%d", screenID, tabID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var tab ScreenTab
	if err := s.transport.DecodeResponse(resp, &tab); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tab, nil
}

// DeleteTab deletes a screen tab.
//
// Example:
//
//	err := client.Screen.DeleteTab(ctx, 10000, 10100)
func (s *Service) DeleteTab(ctx context.Context, screenID, tabID int64) error {
	if screenID <= 0 {
		return fmt.Errorf("screen ID is required")
	}

	if tabID <= 0 {
		return fmt.Errorf("tab ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs/%d", screenID, tabID)

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

// AddField adds a field to a screen tab.
//
// Example:
//
//	field, err := client.Screen.AddField(ctx, 10000, 10100, &screen.AddFieldInput{
//	    FieldID: "summary",
//	})
func (s *Service) AddField(ctx context.Context, screenID, tabID int64, input *AddFieldInput) (*ScreenField, error) {
	if screenID <= 0 {
		return nil, fmt.Errorf("screen ID is required")
	}

	if tabID <= 0 {
		return nil, fmt.Errorf("tab ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.FieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs/%d/fields", screenID, tabID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var field ScreenField
	if err := s.transport.DecodeResponse(resp, &field); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &field, nil
}

// RemoveField removes a field from a screen tab.
//
// Example:
//
//	err := client.Screen.RemoveField(ctx, 10000, 10100, "summary")
func (s *Service) RemoveField(ctx context.Context, screenID, tabID int64, fieldID string) error {
	if screenID <= 0 {
		return fmt.Errorf("screen ID is required")
	}

	if tabID <= 0 {
		return fmt.Errorf("tab ID is required")
	}

	if fieldID == "" {
		return fmt.Errorf("field ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/screens/%d/tabs/%d/fields/%s", screenID, tabID, fieldID)

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
