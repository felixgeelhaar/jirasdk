// Package securitylevel provides Security Level resource management for Jira.
//
// Security levels control who can view issues. This package provides operations
// for managing security levels and security schemes.
package securitylevel

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Security Level resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Security Level service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// SecurityLevel represents a Jira security level.
type SecurityLevel struct {
	ID          string `json:"id"`
	Self        string `json:"self,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SecurityScheme represents a security scheme containing security levels.
type SecurityScheme struct {
	ID             string           `json:"id"`
	Self           string           `json:"self,omitempty"`
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	DefaultLevel   *SecurityLevel   `json:"defaultSecurityLevelId,omitempty"`
	SecurityLevels []*SecurityLevel `json:"levels,omitempty"`
}

// Get retrieves a specific security level by ID.
//
// Example:
//
//	level, err := client.SecurityLevel.Get(ctx, "10000")
func (s *Service) Get(ctx context.Context, levelID string) (*SecurityLevel, error) {
	if levelID == "" {
		return nil, fmt.Errorf("security level ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/securitylevel/%s", levelID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var level SecurityLevel
	if err := s.transport.DecodeResponse(resp, &level); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &level, nil
}

// GetIssueSecuritySchemes retrieves all security schemes.
//
// Example:
//
//	schemes, err := client.SecurityLevel.GetIssueSecuritySchemes(ctx)
func (s *Service) GetIssueSecuritySchemes(ctx context.Context) ([]*SecurityScheme, error) {
	path := "/rest/api/3/issuesecurityschemes"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		IssueSecuritySchemes []*SecurityScheme `json:"issueSecuritySchemes"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.IssueSecuritySchemes, nil
}

// GetIssueSecurityScheme retrieves a specific security scheme by ID.
//
// Example:
//
//	scheme, err := client.SecurityLevel.GetIssueSecurityScheme(ctx, "10000")
func (s *Service) GetIssueSecurityScheme(ctx context.Context, schemeID string) (*SecurityScheme, error) {
	if schemeID == "" {
		return nil, fmt.Errorf("security scheme ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issuesecurityschemes/%s", schemeID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var scheme SecurityScheme
	if err := s.transport.DecodeResponse(resp, &scheme); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &scheme, nil
}
