// Package search provides JQL search functionality for Jira.
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/felixgeelhaar/jirasdk/internal/pagination"
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

// SearchOptions configures the search operation for the legacy /rest/api/3/search endpoint.
//
// Deprecated: Use SearchJQLOptions with SearchJQL method instead.
// The /rest/api/3/search endpoint will be removed by Atlassian on October 31, 2025.
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

// SearchJQLOptions configures the search operation for the new /rest/api/3/search/jql endpoint.
// This endpoint uses token-based pagination instead of offset-based pagination.
type SearchJQLOptions struct {
	// JQL is the JQL query string (required)
	JQL string

	// Fields specifies which fields to include in the response.
	// If not specified, only the issue ID is returned by default.
	// Use ["*all"] to get all fields or ["*navigable"] for navigable fields.
	Fields []string

	// Expand specifies which additional information to include
	Expand []string

	// MaxResults is the maximum number of results per page.
	// Can be set up to 5000 when fewer fields are requested.
	// Defaults to 100 when using NewSearchJQLIterator; when calling SearchJQL directly and omitted, the server default applies (typically 50).
	MaxResults int

	// NextPageToken is used for pagination. Leave empty for the first page.
	// Use the token returned from the previous response to get the next page.
	NextPageToken string

	// FieldsByKeys specifies if fields should be referenced by keys instead of IDs
	FieldsByKeys bool

	// Properties specifies which issue properties to include
	Properties []string

	// ValidateQuery validates the JQL query before executing
	ValidateQuery bool
}

// SearchResult contains the search results and pagination info for the legacy endpoint.
//
// Deprecated: Use SearchJQLResult with SearchJQL method instead.
type SearchResult struct {
	Issues     []*issue.Issue      `json:"issues"`
	StartAt    int                 `json:"startAt"`
	MaxResults int                 `json:"maxResults"`
	Total      int                 `json:"total"`
	PageInfo   pagination.PageInfo `json:"-"`
}

// SearchJQLResult contains the search results and pagination info for the new JQL endpoint.
// Note: Unlike the legacy endpoint, this does not include a total count due to performance considerations.
type SearchJQLResult struct {
	Issues        []*issue.Issue `json:"issues"`
	MaxResults    int            `json:"maxResults,omitempty"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

// Search executes a JQL search query using the legacy endpoint.
//
// Deprecated: Use SearchJQL instead. The /rest/api/3/search endpoint will be removed
// by Atlassian on October 31, 2025. Migrate to SearchJQL which uses the new
// /rest/api/3/search/jql endpoint with token-based pagination.
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

// SearchJQL executes a JQL search query using the new Enhanced JQL Search API.
// This method uses token-based pagination and should be used instead of Search().
//
// Key differences from Search():
//   - Uses token-based pagination with NextPageToken instead of StartAt
//   - No total count returned (for performance reasons)
//   - Default fields is ["id"] instead of ["*navigable"]
//   - Supports up to 5000 results per page (vs 100 in legacy endpoint)
//
// Example:
//
//	results, err := client.Search.SearchJQL(ctx, &search.SearchJQLOptions{
//		JQL: "project = PROJ AND status = Open",
//		Fields: []string{"summary", "status", "assignee"},
//		MaxResults: 100,
//	})
//
//	// Paginate through results using NextPageToken
//	for results.NextPageToken != "" {
//		results, err = client.Search.SearchJQL(ctx, &search.SearchJQLOptions{
//			JQL: "project = PROJ AND status = Open",
//			Fields: []string{"summary", "status", "assignee"},
//			MaxResults: 100,
//			NextPageToken: results.NextPageToken,
//		})
//		// Process results...
//	}
func (s *Service) SearchJQL(ctx context.Context, opts *SearchJQLOptions) (*SearchJQLResult, error) {
	if opts == nil || opts.JQL == "" {
		return nil, fmt.Errorf("JQL query is required")
	}

	path := "/rest/api/3/search/jql"

	// Build request body
	body := map[string]interface{}{
		"jql": opts.JQL,
	}

	// Set maxResults (defaults to 50 if not specified)
	if opts.MaxResults > 0 {
		body["maxResults"] = opts.MaxResults
	}

	// Add nextPageToken for pagination
	if opts.NextPageToken != "" {
		body["nextPageToken"] = opts.NextPageToken
	}

	// Add fields if specified
	if len(opts.Fields) > 0 {
		body["fields"] = opts.Fields
	}

	// Add expand if specified
	if len(opts.Expand) > 0 {
		body["expand"] = opts.Expand
	}

	// Add fieldsByKeys if specified
	if opts.FieldsByKeys {
		body["fieldsByKeys"] = true
	}

	// Add properties if specified
	if len(opts.Properties) > 0 {
		body["properties"] = opts.Properties
	}

	// Add validation if specified
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
	var result SearchJQLResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HasNextPage returns true if there are more results available.
func (r *SearchJQLResult) HasNextPage() bool {
	return r.NextPageToken != ""
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

// SearchIterator provides an iterator for paginated search results using the legacy endpoint.
//
// Deprecated: Use SearchJQLIterator instead. The underlying /rest/api/3/search endpoint
// will be removed by Atlassian on October 31, 2025.
type SearchIterator struct {
	service *Service
	opts    *SearchOptions
	current *SearchResult
	index   int
	ctx     context.Context
}

// NewSearchIterator creates a new search iterator using the legacy endpoint.
//
// Deprecated: Use NewSearchJQLIterator instead.
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

// SearchJQLIterator provides an iterator for paginated search results using the new JQL endpoint.
// This iterator automatically handles token-based pagination.
type SearchJQLIterator struct {
	service *Service
	opts    *SearchJQLOptions
	current *SearchJQLResult
	index   int
	ctx     context.Context
	err     error
}

// NewSearchJQLIterator creates a new search iterator using the Enhanced JQL Search API.
//
// Example:
//
//	iter := client.Search.NewSearchJQLIterator(ctx, &search.SearchJQLOptions{
//		JQL: "project = PROJ AND status = Open",
//		Fields: []string{"summary", "status", "assignee"},
//		MaxResults: 100,
//	})
//
//	for iter.Next() {
//		issue := iter.Issue()
//		fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
//	}
//
//	if err := iter.Err(); err != nil {
//		log.Fatal(err)
//	}
func (s *Service) NewSearchJQLIterator(ctx context.Context, opts *SearchJQLOptions) *SearchJQLIterator {
	if opts == nil {
		opts = &SearchJQLOptions{}
	}
	if opts.MaxResults == 0 {
		opts.MaxResults = 100 // Default to 100 for better performance
	}

	return &SearchJQLIterator{
		service: s,
		opts:    opts,
		ctx:     ctx,
		index:   -1,
	}
}

// Next advances the iterator to the next issue.
// Returns true if an issue is available, false if iteration is complete or an error occurred.
func (it *SearchJQLIterator) Next() bool {
	it.index++

	// Check if we need to fetch the next page
	if it.current == nil || it.index >= len(it.current.Issues) {
		// Check if there are more pages
		if it.current != nil && !it.current.HasNextPage() {
			return false
		}

		// Fetch next page
		if it.current != nil {
			it.opts.NextPageToken = it.current.NextPageToken
		}

		result, err := it.service.SearchJQL(it.ctx, it.opts)
		if err != nil {
			it.err = err
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
// Returns nil if there is no current issue (before first Next() call or after iteration completes).
func (it *SearchJQLIterator) Issue() *issue.Issue {
	if it.current == nil || it.index < 0 || it.index >= len(it.current.Issues) {
		return nil
	}
	return it.current.Issues[it.index]
}

// Err returns any error encountered during iteration.
func (it *SearchJQLIterator) Err() error {
	return it.err
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
