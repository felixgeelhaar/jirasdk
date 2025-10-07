// Package workflow provides Workflow resource management for Jira.
package workflow

import (
	"context"
	"fmt"
)

// Service provides operations for Workflow resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	// Methods will be implemented
}

// NewService creates a new Workflow service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Transition represents a workflow transition.
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetTransitions retrieves available transitions for an issue.
func (s *Service) GetTransitions(ctx context.Context, issueKeyOrID string) ([]*Transition, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	// TODO: Implement

	return nil, fmt.Errorf("not implemented yet")
}
