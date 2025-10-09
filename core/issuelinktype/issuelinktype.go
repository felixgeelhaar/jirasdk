// Package issuelinktype provides Issue Link Type management for Jira.
//
// Issue link types define the relationships between issues.
package issuelinktype

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Issue Link Type resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Issue Link Type service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// IssueLinkType represents a type of link between issues.
type IssueLinkType struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
	Self    string `json:"self,omitempty"`
}

// List retrieves all issue link types.
//
// Example:
//
//	linkTypes, err := client.IssueLinkType.List(ctx)
func (s *Service) List(ctx context.Context) ([]*IssueLinkType, error) {
	path := "/rest/api/3/issueLinkType"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		IssueLinkTypes []*IssueLinkType `json:"issueLinkTypes"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.IssueLinkTypes, nil
}

// Get retrieves a specific issue link type.
//
// Example:
//
//	linkType, err := client.IssueLinkType.Get(ctx, "10000")
func (s *Service) Get(ctx context.Context, issueLinkTypeID string) (*IssueLinkType, error) {
	if issueLinkTypeID == "" {
		return nil, fmt.Errorf("issue link type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issueLinkType/%s", issueLinkTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var linkType IssueLinkType
	if err := s.transport.DecodeResponse(resp, &linkType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkType, nil
}

// CreateInput represents input for creating an issue link type.
type CreateInput struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// Create creates a new issue link type.
//
// Example:
//
//	linkType, err := client.IssueLinkType.Create(ctx, &issuelinktype.CreateInput{
//		Name:    "Dependency",
//		Inward:  "depends on",
//		Outward: "is depended on by",
//	})
func (s *Service) Create(ctx context.Context, input *CreateInput) (*IssueLinkType, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" || input.Inward == "" || input.Outward == "" {
		return nil, fmt.Errorf("name, inward, and outward are required")
	}

	path := "/rest/api/3/issueLinkType"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var linkType IssueLinkType
	if err := s.transport.DecodeResponse(resp, &linkType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkType, nil
}

// UpdateInput represents input for updating an issue link type.
type UpdateInput struct {
	Name    string `json:"name,omitempty"`
	Inward  string `json:"inward,omitempty"`
	Outward string `json:"outward,omitempty"`
}

// Update updates an issue link type.
//
// Example:
//
//	linkType, err := client.IssueLinkType.Update(ctx, "10000", &issuelinktype.UpdateInput{
//		Name: "Updated Dependency",
//	})
func (s *Service) Update(ctx context.Context, issueLinkTypeID string, input *UpdateInput) (*IssueLinkType, error) {
	if issueLinkTypeID == "" {
		return nil, fmt.Errorf("issue link type ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/issueLinkType/%s", issueLinkTypeID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var linkType IssueLinkType
	if err := s.transport.DecodeResponse(resp, &linkType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkType, nil
}

// Delete deletes an issue link type.
//
// Example:
//
//	err := client.IssueLinkType.Delete(ctx, "10000")
func (s *Service) Delete(ctx context.Context, issueLinkTypeID string) error {
	if issueLinkTypeID == "" {
		return fmt.Errorf("issue link type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issueLinkType/%s", issueLinkTypeID)

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
