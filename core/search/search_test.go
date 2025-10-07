package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/felixgeelhaar/jira-connect/core/issue"
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
