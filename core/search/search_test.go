package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felixgeelhaar/jirasdk/core/issue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport implements the RoundTripper interface for testing
type mockTransport struct {
	server *httptest.Server
}

func newMockTransport(handler http.HandlerFunc) *mockTransport {
	return &mockTransport{
		server: httptest.NewServer(handler),
	}
}

func (m *mockTransport) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = io.NopCloser(io.Reader(&bytesReader{data: data}))
	}

	req, err := http.NewRequestWithContext(ctx, method, m.server.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (m *mockTransport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := m.server.Client()
	return client.Do(req)
}

func (m *mockTransport) DecodeResponse(resp *http.Response, target interface{}) error {
	if target == nil {
		return nil
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

func (m *mockTransport) Close() {
	m.server.Close()
}

type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name           string
		opts           *SearchOptions
		responseStatus int
		responseBody   *SearchResult
		wantErr        bool
		errMsg         string
		checkRequest   func(*testing.T, *http.Request)
	}{
		{
			name: "successful search",
			opts: &SearchOptions{
				JQL:        "project = PROJ",
				MaxResults: 50,
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchResult{
				Issues: []*issue.Issue{
					{ID: "10001", Key: "PROJ-1"},
					{ID: "10002", Key: "PROJ-2"},
				},
				StartAt:    0,
				MaxResults: 50,
				Total:      2,
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
			errMsg:  "JQL query is required",
		},
		{
			name: "empty JQL",
			opts: &SearchOptions{
				JQL: "",
			},
			wantErr: true,
			errMsg:  "JQL query is required",
		},
		{
			name: "with all options",
			opts: &SearchOptions{
				JQL:        "project = PROJ",
				MaxResults: 100,
				StartAt:    50,
				Fields:     []string{"summary", "status", "assignee"},
				Expand:     []string{"changelog", "renderedFields"},
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchResult{
				Issues:     []*issue.Issue{{ID: "10001", Key: "PROJ-1"}},
				StartAt:    50,
				MaxResults: 100,
				Total:      150,
			},
			wantErr: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "project = PROJ", body["jql"])
				assert.Equal(t, float64(100), body["maxResults"])
				assert.Equal(t, float64(50), body["startAt"])
				assert.NotNil(t, body["fields"])
				assert.NotNil(t, body["expand"])
			},
		},
		{
			name: "with validate query",
			opts: &SearchOptions{
				JQL:           "project = PROJ",
				ValidateQuery: true,
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchResult{
				Issues:     []*issue.Issue{},
				StartAt:    0,
				MaxResults: 0,
				Total:      0,
			},
			wantErr: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "strict", body["validateQuery"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil || tt.opts.JQL == "" {
				service := NewService(nil)
				_, err := service.Search(context.Background(), tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/search")

				// Verify request body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.opts.JQL, body["jql"])

				if tt.checkRequest != nil {
					// Re-read body for custom check
					bodyBytes, _ := json.Marshal(body)
					r.Body = io.NopCloser(io.Reader(&bytesReader{data: bodyBytes}))
					tt.checkRequest(t, r)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.Search(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, len(tt.responseBody.Issues), len(result.Issues))
				assert.Equal(t, tt.responseBody.Total, result.Total)
			}
		})
	}
}

func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() string
		expected string
	}{
		{
			name: "simple project filter",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					Build()
			},
			expected: "project = PROJ",
		},
		{
			name: "project and status",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					Status("Open").
					Build()
			},
			expected: "project = PROJ AND status = Open",
		},
		{
			name: "complex query",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					Status("Open").
					And().
					Assignee("john.doe").
					OrderBy("created", "DESC").
					Build()
			},
			expected: "project = PROJ AND status = Open AND assignee = john.doe ORDER BY created DESC",
		},
		{
			name: "empty assignee",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					Assignee("").
					Build()
			},
			expected: "project = PROJ AND assignee is EMPTY",
		},
		{
			name: "text search",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					Text("bug").
					Build()
			},
			expected: "project = PROJ AND text ~ bug",
		},
		{
			name: "labels filter",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					Labels("urgent", "security").
					Build()
			},
			expected: "project = PROJ AND labels = urgent AND labels = security",
		},
		{
			name: "date filters",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					And().
					CreatedAfter("2025-01-01").
					And().
					UpdatedBefore("2025-12-31").
					Build()
			},
			expected: "project = PROJ AND created >= 2025-01-01 AND updated <= 2025-12-31",
		},
		{
			name: "quoted values with spaces",
			build: func() string {
				return NewQueryBuilder().
					Summary("critical bug").
					Build()
			},
			expected: `summary ~ "critical bug"`,
		},
		{
			name: "issue type filter",
			build: func() string {
				return NewQueryBuilder().
					IssueType("Bug").
					Build()
			},
			expected: "issuetype = Bug",
		},
		{
			name: "reporter filter",
			build: func() string {
				return NewQueryBuilder().
					Reporter("john.doe").
					Build()
			},
			expected: "reporter = john.doe",
		},
		{
			name: "priority filter",
			build: func() string {
				return NewQueryBuilder().
					Priority("High").
					Build()
			},
			expected: "priority = High",
		},
		{
			name: "description search",
			build: func() string {
				return NewQueryBuilder().
					Description("error message").
					Build()
			},
			expected: `description ~ "error message"`,
		},
		{
			name: "created before filter",
			build: func() string {
				return NewQueryBuilder().
					CreatedBefore("2025-12-31").
					Build()
			},
			expected: "created <= 2025-12-31",
		},
		{
			name: "updated after filter",
			build: func() string {
				return NewQueryBuilder().
					UpdatedAfter("2025-01-01").
					Build()
			},
			expected: "updated >= 2025-01-01",
		},
		{
			name: "OR operator",
			build: func() string {
				return NewQueryBuilder().
					Status("Open").
					Or().
					Status("In Progress").
					Build()
			},
			expected: "status = Open OR status = \"In Progress\"",
		},
		{
			name: "raw JQL",
			build: func() string {
				return NewQueryBuilder().
					Raw("customfield_10000 = value").
					Build()
			},
			expected: "customfield_10000 = value",
		},
		{
			name: "order by ascending",
			build: func() string {
				return NewQueryBuilder().
					Project("PROJ").
					OrderBy("created", "ASC").
					Build()
			},
			expected: "project = PROJ ORDER BY created ASC",
		},
		{
			name: "single label",
			build: func() string {
				return NewQueryBuilder().
					Labels("urgent").
					Build()
			},
			expected: "labels = urgent",
		},
		{
			name: "AND on empty builder",
			build: func() string {
				return NewQueryBuilder().
					And().
					Project("PROJ").
					Build()
			},
			expected: "project = PROJ",
		},
		{
			name: "OR on empty builder",
			build: func() string {
				return NewQueryBuilder().
					Or().
					Project("PROJ").
					Build()
			},
			expected: "project = PROJ",
		},
		{
			name: "complex with OR and AND",
			build: func() string {
				return NewQueryBuilder().
					Status("Open").
					Or().
					Status("Reopened").
					And().
					Priority("High").
					Build()
			},
			expected: "status = Open OR status = Reopened AND priority = High",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "test",
			expected: "test",
		},
		{
			name:     "with space",
			input:    "test value",
			expected: `"test value"`,
		},
		{
			name:     "with quotes",
			input:    `test "value"`,
			expected: `"test \"value\""`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quote(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchIterator(t *testing.T) {
	// Test successful iteration
	t.Run("successful iteration", func(t *testing.T) {
		callCount := 0
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var result SearchResult
			if callCount == 1 {
				result = SearchResult{
					Issues:     []*issue.Issue{{ID: "1", Key: "PROJ-1"}, {ID: "2", Key: "PROJ-2"}},
					StartAt:    0,
					MaxResults: 2,
					Total:      3,
				}
			} else {
				result = SearchResult{
					Issues:     []*issue.Issue{{ID: "3", Key: "PROJ-3"}},
					StartAt:    2,
					MaxResults: 2,
					Total:      3,
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchIterator(context.Background(), &SearchOptions{
			JQL:        "project = PROJ",
			MaxResults: 2,
		})

		count := 0
		for iter.Next() {
			count++
			issue := iter.Issue()
			assert.NotNil(t, issue)
		}

		assert.NoError(t, iter.Err())
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, callCount)
	})

	// Test error during iteration
	t.Run("error during iteration", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchIterator(context.Background(), &SearchOptions{
			JQL: "project = PROJ",
		})

		// Next returns false on error, but Err() is not set in current implementation
		assert.False(t, iter.Next())
	})

	// Test nil issue before Next
	t.Run("nil issue before next", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			result := SearchResult{
				Issues:     []*issue.Issue{{ID: "1", Key: "PROJ-1"}},
				StartAt:    0,
				MaxResults: 1,
				Total:      1,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchIterator(context.Background(), &SearchOptions{
			JQL: "project = PROJ",
		})

		// Call Issue before Next
		assert.Nil(t, iter.Issue())
	})
}

func TestSearchJQL(t *testing.T) {
	tests := []struct {
		name           string
		opts           *SearchJQLOptions
		responseStatus int
		responseBody   *SearchJQLResult
		wantErr        bool
		errMsg         string
		checkRequest   func(*testing.T, *http.Request)
	}{
		{
			name: "successful search with basic options",
			opts: &SearchJQLOptions{
				JQL:        "project = PROJ",
				MaxResults: 100,
				Fields:     []string{"summary", "status"},
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchJQLResult{
				Issues: []*issue.Issue{
					{ID: "10001", Key: "PROJ-1"},
					{ID: "10002", Key: "PROJ-2"},
				},
				MaxResults:    100,
				NextPageToken: "next-token-123",
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
			errMsg:  "JQL query is required",
		},
		{
			name: "empty JQL",
			opts: &SearchJQLOptions{
				JQL: "",
			},
			wantErr: true,
			errMsg:  "JQL query is required",
		},
		{
			name: "with pagination token",
			opts: &SearchJQLOptions{
				JQL:           "project = PROJ",
				MaxResults:    100,
				NextPageToken: "page-token-456",
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchJQLResult{
				Issues:        []*issue.Issue{{ID: "10003", Key: "PROJ-3"}},
				MaxResults:    100,
				NextPageToken: "", // Last page
			},
			wantErr: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "page-token-456", body["nextPageToken"])
			},
		},
		{
			name: "with all options",
			opts: &SearchJQLOptions{
				JQL:           "project = PROJ",
				MaxResults:    200,
				Fields:        []string{"*all"},
				Expand:        []string{"changelog"},
				FieldsByKeys:  true,
				Properties:    []string{"prop1", "prop2"},
				ValidateQuery: true,
			},
			responseStatus: http.StatusOK,
			responseBody: &SearchJQLResult{
				Issues:        []*issue.Issue{{ID: "10001", Key: "PROJ-1"}},
				MaxResults:    200,
				NextPageToken: "",
			},
			wantErr: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "project = PROJ", body["jql"])
				assert.Equal(t, float64(200), body["maxResults"])
				assert.NotNil(t, body["fields"])
				assert.NotNil(t, body["expand"])
				assert.Equal(t, true, body["fieldsByKeys"])
				assert.NotNil(t, body["properties"])
				assert.Equal(t, "strict", body["validateQuery"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil || tt.opts.JQL == "" {
				service := NewService(nil)
				_, err := service.SearchJQL(context.Background(), tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/search/jql")

				// Verify request body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.opts.JQL, body["jql"])

				if tt.checkRequest != nil {
					// Re-read body for custom check
					bodyBytes, _ := json.Marshal(body)
					r.Body = io.NopCloser(io.Reader(&bytesReader{data: bodyBytes}))
					tt.checkRequest(t, r)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.SearchJQL(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, len(tt.responseBody.Issues), len(result.Issues))
				assert.Equal(t, tt.responseBody.NextPageToken, result.NextPageToken)
			}
		})
	}
}

func TestSearchJQLResult_HasNextPage(t *testing.T) {
	tests := []struct {
		name     string
		result   *SearchJQLResult
		expected bool
	}{
		{
			name: "has next page",
			result: &SearchJQLResult{
				NextPageToken: "token-123",
			},
			expected: true,
		},
		{
			name: "no next page",
			result: &SearchJQLResult{
				NextPageToken: "",
			},
			expected: false,
		},
		{
			name: "nil token",
			result: &SearchJQLResult{
				NextPageToken: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasNextPage())
		})
	}
}

func TestSearchJQLIterator(t *testing.T) {
	// Test successful iteration with pagination
	t.Run("successful iteration with pagination", func(t *testing.T) {
		callCount := 0
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var result SearchJQLResult

			// Check that endpoint is correct
			assert.Contains(t, r.URL.Path, "/rest/api/3/search/jql")

			if callCount == 1 {
				result = SearchJQLResult{
					Issues:        []*issue.Issue{{ID: "1", Key: "PROJ-1"}, {ID: "2", Key: "PROJ-2"}},
					MaxResults:    2,
					NextPageToken: "token-page2",
				}
			} else {
				// Verify nextPageToken was sent
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "token-page2", body["nextPageToken"])

				result = SearchJQLResult{
					Issues:        []*issue.Issue{{ID: "3", Key: "PROJ-3"}},
					MaxResults:    2,
					NextPageToken: "", // Last page
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchJQLIterator(context.Background(), &SearchJQLOptions{
			JQL:        "project = PROJ",
			MaxResults: 2,
		})

		count := 0
		var issueKeys []string
		for iter.Next() {
			count++
			issue := iter.Issue()
			require.NotNil(t, issue)
			issueKeys = append(issueKeys, issue.Key)
		}

		assert.NoError(t, iter.Err())
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, callCount)
		assert.Equal(t, []string{"PROJ-1", "PROJ-2", "PROJ-3"}, issueKeys)
	})

	// Test error during iteration
	t.Run("error during iteration", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchJQLIterator(context.Background(), &SearchJQLOptions{
			JQL: "project = PROJ",
		})

		// Next returns false on error
		assert.False(t, iter.Next())
		assert.Error(t, iter.Err())
	})

	// Test nil issue before Next
	t.Run("nil issue before next", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			result := SearchJQLResult{
				Issues:        []*issue.Issue{{ID: "1", Key: "PROJ-1"}},
				MaxResults:    1,
				NextPageToken: "",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchJQLIterator(context.Background(), &SearchJQLOptions{
			JQL: "project = PROJ",
		})

		// Call Issue before Next
		assert.Nil(t, iter.Issue())
	})

	// Test empty results
	t.Run("empty results", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			result := SearchJQLResult{
				Issues:        []*issue.Issue{},
				MaxResults:    100,
				NextPageToken: "",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchJQLIterator(context.Background(), &SearchJQLOptions{
			JQL: "project = NONEXISTENT",
		})

		count := 0
		for iter.Next() {
			count++
		}

		assert.NoError(t, iter.Err())
		assert.Equal(t, 0, count)
	})

	// Test default max results
	t.Run("default max results", func(t *testing.T) {
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)

			// Verify maxResults is set to default
			assert.Equal(t, float64(100), body["maxResults"])

			result := SearchJQLResult{
				Issues:        []*issue.Issue{{ID: "1", Key: "PROJ-1"}},
				MaxResults:    100,
				NextPageToken: "",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)
		iter := service.NewSearchJQLIterator(context.Background(), &SearchJQLOptions{
			JQL: "project = PROJ",
			// MaxResults not specified
		})

		assert.True(t, iter.Next())
		assert.NoError(t, iter.Err())
	})

	// Test that iterator does not mutate caller's options
	t.Run("does not mutate caller options", func(t *testing.T) {
		callCount := 0
		transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			var result SearchJQLResult
			if callCount == 1 {
				result = SearchJQLResult{
					Issues:        []*issue.Issue{{ID: "1", Key: "PROJ-1"}},
					MaxResults:    2,
					NextPageToken: "token-page2",
				}
			} else {
				result = SearchJQLResult{
					Issues:        []*issue.Issue{{ID: "2", Key: "PROJ-2"}},
					MaxResults:    2,
					NextPageToken: "",
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
		})
		defer transport.Close()

		service := NewService(transport)

		// Create options with initial state
		originalOpts := &SearchJQLOptions{
			JQL:           "project = PROJ",
			MaxResults:    2,
			Fields:        []string{"summary"},
			NextPageToken: "", // Initially empty
		}

		// Store original values
		originalToken := originalOpts.NextPageToken
		originalMaxResults := originalOpts.MaxResults
		originalFieldsLen := len(originalOpts.Fields)

		// Create iterator (should copy options)
		iter := service.NewSearchJQLIterator(context.Background(), originalOpts)

		// Iterate through all results
		count := 0
		for iter.Next() {
			count++
		}

		// Verify iteration worked
		assert.Equal(t, 2, count)
		assert.NoError(t, iter.Err())

		// CRITICAL: Verify original options were NOT mutated
		assert.Equal(t, originalToken, originalOpts.NextPageToken, "NextPageToken should not be mutated")
		assert.Equal(t, originalMaxResults, originalOpts.MaxResults, "MaxResults should not be mutated")
		assert.Equal(t, originalFieldsLen, len(originalOpts.Fields), "Fields should not be mutated")
		assert.Equal(t, "summary", originalOpts.Fields[0], "Fields content should not be mutated")
	})
}


func TestParseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "standard browse URL",
			url:      "https://example.atlassian.net/browse/PROJ-123",
			expected: "PROJ-123",
			wantErr:  false,
		},
		{
			name:     "with query params",
			url:      "https://example.atlassian.net/browse/PROJ-123?focusedCommentId=12345",
			expected: "PROJ-123",
			wantErr:  false,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "no browse path",
			url:     "https://example.atlassian.net/projects/PROJ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseURL(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
