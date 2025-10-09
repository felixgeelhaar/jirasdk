// Package audit provides Audit Log resource management for Jira.
//
// Audit logs track administrative and security-related events in Jira.
// This package provides operations for retrieving and searching audit records.
package audit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Service provides operations for Audit Log resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Audit Log service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// AuditRecord represents an audit log record.
type AuditRecord struct {
	ID               int64                  `json:"id"`
	Summary          string                 `json:"summary"`
	RemoteAddress    string                 `json:"remoteAddress,omitempty"`
	AuthorKey        string                 `json:"authorKey,omitempty"`
	Created          string                 `json:"created"`
	Category         string                 `json:"category,omitempty"`
	EventSource      string                 `json:"eventSource,omitempty"`
	Description      string                 `json:"description,omitempty"`
	ObjectItem       *AuditObjectItem       `json:"objectItem,omitempty"`
	ChangedValues    []*AuditChangedValue   `json:"changedValues,omitempty"`
	AssociatedItems  []*AuditAssociatedItem `json:"associatedItems,omitempty"`
}

// AuditObjectItem represents an object affected by an audit event.
type AuditObjectItem struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	TypeName   string `json:"typeName,omitempty"`
	ParentID   string `json:"parentId,omitempty"`
	ParentName string `json:"parentName,omitempty"`
}

// AuditChangedValue represents a changed value in an audit record.
type AuditChangedValue struct {
	FieldName   string `json:"fieldName"`
	ChangedFrom string `json:"changedFrom,omitempty"`
	ChangedTo   string `json:"changedTo,omitempty"`
}

// AuditAssociatedItem represents an item associated with an audit event.
type AuditAssociatedItem struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	TypeName   string `json:"typeName,omitempty"`
	ParentID   string `json:"parentId,omitempty"`
	ParentName string `json:"parentName,omitempty"`
}

// ListOptions represents options for listing audit records.
type ListOptions struct {
	Offset     int       `json:"offset,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Filter     string    `json:"filter,omitempty"`
	From       time.Time `json:"from,omitempty"`
	To         time.Time `json:"to,omitempty"`
}

// List retrieves audit records with optional filtering.
//
// Example:
//
//	records, err := client.Audit.List(ctx, &audit.ListOptions{
//	    Limit:  100,
//	    Filter: "user_management",
//	})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*AuditRecord, error) {
	path := "/rest/api/3/auditing/record"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.Offset > 0 {
			q.Set("offset", strconv.Itoa(opts.Offset))
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Filter != "" {
			q.Set("filter", opts.Filter)
		}
		if !opts.From.IsZero() {
			q.Set("from", opts.From.Format(time.RFC3339))
		}
		if !opts.To.IsZero() {
			q.Set("to", opts.To.Format(time.RFC3339))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Offset  int            `json:"offset"`
		Limit   int            `json:"limit"`
		Total   int            `json:"total"`
		Records []*AuditRecord `json:"records"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Records, nil
}
