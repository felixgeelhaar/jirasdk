// Package resolution provides Resolution resource management for Jira.
//
// Resolutions indicate how an issue was closed (e.g., Fixed, Won't Fix, Duplicate, Cannot Reproduce).
// This package provides operations for managing resolutions.
package resolution

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Resolution resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Resolution service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Resolution represents a Jira resolution.
type Resolution struct {
	ID          string `json:"id"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// CreateResolutionInput represents input for creating a resolution.
type CreateResolutionInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateResolutionInput represents input for updating a resolution.
type UpdateResolutionInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// List retrieves all resolutions.
//
// Example:
//
//	resolutions, err := client.Resolution.List(ctx)
func (s *Service) List(ctx context.Context) ([]*Resolution, error) {
	path := "/rest/api/3/resolution"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var resolutions []*Resolution
	if err := s.transport.DecodeResponse(resp, &resolutions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resolutions, nil
}

// Get retrieves a specific resolution by ID.
//
// Example:
//
//	resolution, err := client.Resolution.Get(ctx, "1")
func (s *Service) Get(ctx context.Context, resolutionID string) (*Resolution, error) {
	if resolutionID == "" {
		return nil, fmt.Errorf("resolution ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/resolution/%s", resolutionID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var resolution Resolution
	if err := s.transport.DecodeResponse(resp, &resolution); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resolution, nil
}

// Create creates a new resolution.
//
// Example:
//
//	resolution, err := client.Resolution.Create(ctx, &resolution.CreateResolutionInput{
//	    Name:        "Deferred",
//	    Description: "Issue deferred to future release",
//	})
func (s *Service) Create(ctx context.Context, input *CreateResolutionInput) (*Resolution, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("resolution name is required")
	}

	path := "/rest/api/3/resolution"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var resolution Resolution
	if err := s.transport.DecodeResponse(resp, &resolution); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resolution, nil
}

// Update updates a resolution.
//
// Example:
//
//	resolution, err := client.Resolution.Update(ctx, "1", &resolution.UpdateResolutionInput{
//	    Name:        "Updated Deferred",
//	    Description: "Updated description",
//	})
func (s *Service) Update(ctx context.Context, resolutionID string, input *UpdateResolutionInput) (*Resolution, error) {
	if resolutionID == "" {
		return nil, fmt.Errorf("resolution ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/resolution/%s", resolutionID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var resolution Resolution
	if err := s.transport.DecodeResponse(resp, &resolution); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resolution, nil
}

// Delete deletes a resolution.
//
// Example:
//
//	err := client.Resolution.Delete(ctx, "1", "2")
func (s *Service) Delete(ctx context.Context, resolutionID, replacementID string) error {
	if resolutionID == "" {
		return fmt.Errorf("resolution ID is required")
	}

	if replacementID == "" {
		return fmt.Errorf("replacement resolution ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/resolution/%s", resolutionID)

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

// SetDefault sets a resolution as the default.
//
// Example:
//
//	err := client.Resolution.SetDefault(ctx, "1")
func (s *Service) SetDefault(ctx context.Context, resolutionID string) error {
	if resolutionID == "" {
		return fmt.Errorf("resolution ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/resolution/default")

	input := struct {
		ID string `json:"id"`
	}{
		ID: resolutionID,
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
