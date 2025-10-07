// Package project provides Project resource management for Jira.
package project

import (
	"context"
	"fmt"
)

// Service provides operations for Project resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	// Methods will be implemented
}

// NewService creates a new Project service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Project represents a Jira project.
type Project struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
	ProjectType string `json:"projectTypeKey,omitempty"`
}

// Get retrieves a project by key or ID.
func (s *Service) Get(ctx context.Context, projectKeyOrID string) (*Project, error) {
	if projectKeyOrID == "" {
		return nil, fmt.Errorf("project key or ID is required")
	}

	// TODO: Implement

	return nil, fmt.Errorf("not implemented yet")
}

// List retrieves all projects.
func (s *Service) List(ctx context.Context) ([]*Project, error) {
	// TODO: Implement

	return nil, fmt.Errorf("not implemented yet")
}
