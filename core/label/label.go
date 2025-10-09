// Package label provides Label resource management for Jira.
//
// Labels are simple tags that can be applied to issues for categorization.
// This package provides operations for retrieving and managing labels.
package label

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Label resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Label service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// ListOptions represents options for listing labels.
type ListOptions struct {
	StartAt    int    `json:"startAt,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	Query      string `json:"query,omitempty"`
}

// List retrieves all labels, optionally filtered by a query string.
//
// Example:
//
//	labels, err := client.Label.List(ctx, &label.ListOptions{
//	    Query:      "bug",
//	    MaxResults: 50,
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]string, error) {
	path := "/rest/api/3/label"

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
		if opts.Query != "" {
			q.Set("query", opts.Query)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []string `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// Suggest retrieves label suggestions based on a query string.
//
// Example:
//
//	suggestions, err := client.Label.Suggest(ctx, "bu")
func (s *Service) Suggest(ctx context.Context, query string) ([]string, error) {
	path := "/rest/api/3/label/suggest"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter
	if query != "" {
		q := req.URL.Query()
		q.Set("query", query)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Suggestions []string `json:"suggestions"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Suggestions, nil
}
