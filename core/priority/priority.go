// Package priority provides Priority resource management for Jira.
//
// Priorities indicate the importance or urgency of issues (e.g., Highest, High, Medium, Low, Lowest).
// This package provides operations for managing priorities.
package priority

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Priority resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Priority service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Priority represents a Jira priority.
type Priority struct {
	ID          string `json:"id"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// CreatePriorityInput represents input for creating a priority.
type CreatePriorityInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
}

// UpdatePriorityInput represents input for updating a priority.
type UpdatePriorityInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
}

// List retrieves all priorities.
//
// Example:
//
//	priorities, err := client.Priority.List(ctx)
func (s *Service) List(ctx context.Context) ([]*Priority, error) {
	path := "/rest/api/3/priority"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var priorities []*Priority
	if err := s.transport.DecodeResponse(resp, &priorities); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return priorities, nil
}

// Get retrieves a specific priority by ID.
//
// Example:
//
//	priority, err := client.Priority.Get(ctx, "1")
func (s *Service) Get(ctx context.Context, priorityID string) (*Priority, error) {
	if priorityID == "" {
		return nil, fmt.Errorf("priority ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/priority/%s", priorityID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var priority Priority
	if err := s.transport.DecodeResponse(resp, &priority); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &priority, nil
}

// Create creates a new priority.
//
// Example:
//
//	priority, err := client.Priority.Create(ctx, &priority.CreatePriorityInput{
//	    Name:        "Critical",
//	    Description: "Critical priority for urgent issues",
//	    StatusColor: "#FF0000",
//	})
func (s *Service) Create(ctx context.Context, input *CreatePriorityInput) (*Priority, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("priority name is required")
	}

	path := "/rest/api/3/priority"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var priority Priority
	if err := s.transport.DecodeResponse(resp, &priority); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &priority, nil
}

// Update updates a priority.
//
// Example:
//
//	priority, err := client.Priority.Update(ctx, "1", &priority.UpdatePriorityInput{
//	    Name:        "Updated Critical",
//	    Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, priorityID string, input *UpdatePriorityInput) (*Priority, error) {
	if priorityID == "" {
		return nil, fmt.Errorf("priority ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/priority/%s", priorityID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var priority Priority
	if err := s.transport.DecodeResponse(resp, &priority); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &priority, nil
}

// Delete deletes a priority.
//
// Example:
//
//	err := client.Priority.Delete(ctx, "1", "2")
func (s *Service) Delete(ctx context.Context, priorityID, replacementID string) error {
	if priorityID == "" {
		return fmt.Errorf("priority ID is required")
	}

	if replacementID == "" {
		return fmt.Errorf("replacement priority ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/priority/%s", priorityID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add replacement ID query parameter
	q := req.URL.Query()
	q.Set("replaceWith", replacementID)
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// SetDefault sets a priority as the default.
//
// Example:
//
//	err := client.Priority.SetDefault(ctx, "3")
func (s *Service) SetDefault(ctx context.Context, priorityID string) error {
	if priorityID == "" {
		return fmt.Errorf("priority ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/priority/default")

	input := struct {
		ID string `json:"id"`
	}{
		ID: priorityID,
	}

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
