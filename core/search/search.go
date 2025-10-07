// Package search provides JQL search functionality for Jira.
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/felixgeelhaar/jira-connect/core/issue"
	"github.com/felixgeelhaar/jira-connect/internal/pagination"
)

// Service provides JQL search operations.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new Search service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// SearchOptions configures the search operation.
type SearchOptions struct {
	// JQL is the JQL query string
	JQL string

	// Fields specifies which fields to include in the response
	Fields []string

	// Expand specifies which additional information to include
	Expand []string

	// MaxResults is the maximum number of results per page
	MaxResults int

	// StartAt is the index of the first result to return
	StartAt int

	// ValidateQuery validates the JQL query before executing
	ValidateQuery bool
}

// SearchResult contains the search results and pagination info.
type SearchResult struct {
	Issues     []*issue.Issue     `json:"issues"`
	StartAt    int                `json:"startAt"`
	MaxResults int                `json:"maxResults"`
	Total      int                `json:"total"`
	PageInfo   pagination.PageInfo `json:"-"`
}

// Search executes a JQL search query.
//
// Example:
//
//	results, err := client.Search.Search(ctx, &search.SearchOptions{
//		JQL: "project = PROJ AND status = Open",
//		MaxResults: 50,
//	})
func (s *Service) Search(ctx context.Context, opts *SearchOptions) (*SearchResult, error) {
	if opts == nil || opts.JQL == "" {
		return nil, fmt.Errorf("JQL query is required")
	}

	path := "/rest/api/3/search"

	// Build request body
	body := map[string]interface{}{
		"jql":        opts.JQL,
		"startAt":    opts.StartAt,
		"maxResults": opts.MaxResults,
	}

	if len(opts.Fields) > 0 {
		body["fields"] = opts.Fields
	}

	if len(opts.Expand) > 0 {
		body["expand"] = opts.Expand
	}

	if opts.ValidateQuery {
		body["validateQuery"] = "strict"
	}

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result SearchResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Populate PageInfo
	result.PageInfo = pagination.PageInfo{
		StartAt:    result.StartAt,
		MaxResults: result.MaxResults,
		Total:      result.Total,
		IsLast:     result.StartAt+len(result.Issues) >= result.Total,
	}

	return &result, nil
}

// QueryBuilder provides a fluent API for building JQL queries.
type QueryBuilder struct {
	parts []string
}

// NewQueryBuilder creates a new JQL query builder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		parts: make([]string, 0),
	}
}

// Project adds a project filter.
func (qb *QueryBuilder) Project(key string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("project = %s", quote(key)))
	return qb
}

// Status adds a status filter.
func (qb *QueryBuilder) Status(status string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("status = %s", quote(status)))
	return qb
}

// IssueType adds an issue type filter.
func (qb *QueryBuilder) IssueType(issueType string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("issuetype = %s", quote(issueType)))
	return qb
}

// Assignee adds an assignee filter.
func (qb *QueryBuilder) Assignee(assignee string) *QueryBuilder {
	if assignee == "" {
		qb.parts = append(qb.parts, "assignee is EMPTY")
	} else {
		qb.parts = append(qb.parts, fmt.Sprintf("assignee = %s", quote(assignee)))
	}
	return qb
}

// Reporter adds a reporter filter.
func (qb *QueryBuilder) Reporter(reporter string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("reporter = %s", quote(reporter)))
	return qb
}

// Priority adds a priority filter.
func (qb *QueryBuilder) Priority(priority string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("priority = %s", quote(priority)))
	return qb
}

// Labels adds a labels filter.
func (qb *QueryBuilder) Labels(labels ...string) *QueryBuilder {
	for i, label := range labels {
		if i > 0 {
			qb.And()
		}
		qb.parts = append(qb.parts, fmt.Sprintf("labels = %s", quote(label)))
	}
	return qb
}

// Text adds a text search filter.
func (qb *QueryBuilder) Text(text string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("text ~ %s", quote(text)))
	return qb
}

// Summary adds a summary search filter.
func (qb *QueryBuilder) Summary(text string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("summary ~ %s", quote(text)))
	return qb
}

// Description adds a description search filter.
func (qb *QueryBuilder) Description(text string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("description ~ %s", quote(text)))
	return qb
}

// CreatedAfter adds a created date filter (after).
func (qb *QueryBuilder) CreatedAfter(date string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("created >= %s", quote(date)))
	return qb
}

// CreatedBefore adds a created date filter (before).
func (qb *QueryBuilder) CreatedBefore(date string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("created <= %s", quote(date)))
	return qb
}

// UpdatedAfter adds an updated date filter (after).
func (qb *QueryBuilder) UpdatedAfter(date string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("updated >= %s", quote(date)))
	return qb
}

// UpdatedBefore adds an updated date filter (before).
func (qb *QueryBuilder) UpdatedBefore(date string) *QueryBuilder {
	qb.parts = append(qb.parts, fmt.Sprintf("updated <= %s", quote(date)))
	return qb
}

// And adds an AND operator.
func (qb *QueryBuilder) And() *QueryBuilder {
	if len(qb.parts) > 0 {
		qb.parts = append(qb.parts, "AND")
	}
	return qb
}

// Or adds an OR operator.
func (qb *QueryBuilder) Or() *QueryBuilder {
	if len(qb.parts) > 0 {
		qb.parts = append(qb.parts, "OR")
	}
	return qb
}

// OrderBy adds an ORDER BY clause.
func (qb *QueryBuilder) OrderBy(field, direction string) *QueryBuilder {
	order := "ASC"
	if strings.ToUpper(direction) == "DESC" {
		order = "DESC"
	}
	qb.parts = append(qb.parts, fmt.Sprintf("ORDER BY %s %s", field, order))
	return qb
}

// Raw adds a raw JQL fragment.
func (qb *QueryBuilder) Raw(jql string) *QueryBuilder {
	qb.parts = append(qb.parts, jql)
	return qb
}

// Build constructs the final JQL query string.
func (qb *QueryBuilder) Build() string {
	return strings.Join(qb.parts, " ")
}

// quote quotes a string value for JQL.
func quote(value string) string {
	// If the value contains spaces or special characters, quote it
	if strings.ContainsAny(value, " \t\n,()[]{}") || value == "" {
		// Escape quotes in the value
		escaped := strings.ReplaceAll(value, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped)
	}
	return value
}

// SearchIterator provides an iterator for paginated search results.
type SearchIterator struct {
	service *Service
	opts    *SearchOptions
	current *SearchResult
	index   int
	ctx     context.Context
}

// NewSearchIterator creates a new search iterator.
func (s *Service) NewSearchIterator(ctx context.Context, opts *SearchOptions) *SearchIterator {
	if opts == nil {
		opts = &SearchOptions{}
	}
	if opts.MaxResults == 0 {
		opts.MaxResults = pagination.DefaultMaxResults
	}

	return &SearchIterator{
		service: s,
		opts:    opts,
		ctx:     ctx,
		index:   -1,
	}
}

// Next advances the iterator to the next issue.
func (it *SearchIterator) Next() bool {
	it.index++

	// Check if we need to fetch the next page
	if it.current == nil || it.index >= len(it.current.Issues) {
		// Check if there are more pages
		if it.current != nil && !it.current.PageInfo.HasNextPage() {
			return false
		}

		// Fetch next page
		if it.current != nil {
			it.opts.StartAt = it.current.PageInfo.NextStartAt()
		}

		result, err := it.service.Search(it.ctx, it.opts)
		if err != nil {
			return false
		}

		it.current = result
		it.index = 0

		// Check if we got any results
		if len(result.Issues) == 0 {
			return false
		}
	}

	return it.index < len(it.current.Issues)
}

// Issue returns the current issue.
func (it *SearchIterator) Issue() *issue.Issue {
	if it.current == nil || it.index < 0 || it.index >= len(it.current.Issues) {
		return nil
	}
	return it.current.Issues[it.index]
}

// Err returns any error encountered during iteration.
func (it *SearchIterator) Err() error {
	// TODO: Store and return errors
	return nil
}

// ParseURL parses a Jira issue URL and extracts the issue key.
//
// Example:
//
//	key, err := search.ParseURL("https://example.atlassian.net/browse/PROJ-123")
//	// Returns: "PROJ-123", nil
func ParseURL(issueURL string) (string, error) {
	u, err := url.Parse(issueURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Handle /browse/ISSUE-KEY URLs
	if strings.Contains(u.Path, "/browse/") {
		parts := strings.Split(u.Path, "/browse/")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("unable to extract issue key from URL")
}
