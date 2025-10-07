// Package user provides User resource management for Jira.
package user

import (
	"context"
	"fmt"
)

// Service provides operations for User resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	// Methods will be implemented
}

// NewService creates a new User service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// User represents a Jira user.
type User struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName"`
	Active       bool   `json:"active"`
	Self         string `json:"self,omitempty"`
}

// Get retrieves a user by account ID.
func (s *Service) Get(ctx context.Context, accountID string) (*User, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	// TODO: Implement

	return nil, fmt.Errorf("not implemented yet")
}
