package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
		bodyReader = io.NopCloser(&bytesReader{data: data})
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

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		opts           *GetOptions
		responseStatus int
		responseBody   *User
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			accountID:      "5b10a2844c20165700ede21g",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: &User{
				AccountID:   "5b10a2844c20165700ede21g",
				DisplayName: "John Doe",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name:      "with expand options",
			accountID: "5b10a2844c20165700ede21g",
			opts: &GetOptions{
				Expand: []string{"groups", "applicationRoles"},
			},
			responseStatus: http.StatusOK,
			responseBody: &User{
				AccountID:   "5b10a2844c20165700ede21g",
				DisplayName: "John Doe",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
			errMsg:    "account ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.accountID == "" {
				service := NewService(nil)
				_, err := service.Get(context.Background(), tt.accountID, tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user")
				assert.Equal(t, tt.accountID, r.URL.Query().Get("accountId"))

				if tt.opts != nil && len(tt.opts.Expand) > 0 {
					assert.NotEmpty(t, r.URL.Query().Get("expand"))
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			user, err := service.Get(context.Background(), tt.accountID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.responseBody.AccountID, user.AccountID)
			}
		})
	}
}

func TestGetMyself(t *testing.T) {
	transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/myself")

		user := &User{
			AccountID:   "current-user",
			DisplayName: "Current User",
			Active:      true,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	})
	defer transport.Close()

	service := NewService(transport)
	user, err := service.GetMyself(context.Background())

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "current-user", user.AccountID)
	assert.Equal(t, "Current User", user.DisplayName)
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name           string
		opts           *SearchOptions
		responseStatus int
		responseBody   []*User
		wantErr        bool
	}{
		{
			name: "successful search",
			opts: &SearchOptions{
				Query:      "john",
				MaxResults: 50,
			},
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe"},
				{AccountID: "user2", DisplayName: "John Smith"},
			},
			wantErr: false,
		},
		{
			name: "search with filters",
			opts: &SearchOptions{
				Query:           "john",
				IncludeActive:   true,
				IncludeInactive: false,
				Property:        "email",
			},
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe", Active: true},
			},
			wantErr: false,
		},
		{
			name:           "search with no options",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody:   []*User{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/search")

				if tt.opts != nil {
					if tt.opts.Query != "" {
						assert.Equal(t, tt.opts.Query, r.URL.Query().Get("query"))
					}
					if tt.opts.IncludeActive {
						assert.Equal(t, "true", r.URL.Query().Get("includeActive"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			users, err := service.Search(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody), len(users))
			}
		})
	}
}

func TestFindUsers(t *testing.T) {
	tests := []struct {
		name           string
		opts           *FindOptions
		responseStatus int
		responseBody   []*User
		wantErr        bool
	}{
		{
			name: "successful find",
			opts: &FindOptions{
				Query:      "john",
				MaxResults: 25,
			},
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe"},
			},
			wantErr: false,
		},
		{
			name:           "find with no options",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody:   []*User{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/search")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			users, err := service.FindUsers(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody), len(users))
			}
		})
	}
}

func TestFindAssignableUsers(t *testing.T) {
	tests := []struct {
		name           string
		opts           *FindAssignableOptions
		responseStatus int
		responseBody   []*User
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful find with project",
			opts: &FindAssignableOptions{
				Project: "PROJ",
				Query:   "john",
			},
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe"},
			},
			wantErr: false,
		},
		{
			name: "successful find with issue key",
			opts: &FindAssignableOptions{
				IssueKey: "PROJ-123",
				Query:    "john",
			},
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe"},
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
			errMsg:  "options are required",
		},
		{
			name: "missing project and issue key",
			opts: &FindAssignableOptions{
				Query: "john",
			},
			wantErr: true,
			errMsg:  "either project or issue key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil || (tt.opts.Project == "" && tt.opts.IssueKey == "") {
				service := NewService(nil)
				_, err := service.FindAssignableUsers(context.Background(), tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/assignable/search")

				if tt.opts.Project != "" {
					assert.Equal(t, tt.opts.Project, r.URL.Query().Get("project"))
				}

				if tt.opts.IssueKey != "" {
					assert.Equal(t, tt.opts.IssueKey, r.URL.Query().Get("issueKey"))
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			users, err := service.FindAssignableUsers(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody), len(users))
			}
		})
	}
}

func TestBulkGet(t *testing.T) {
	tests := []struct {
		name           string
		opts           *BulkGetOptions
		responseStatus int
		responseBody   struct {
			Values []*User `json:"values"`
		}
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful bulk get",
			opts: &BulkGetOptions{
				AccountIDs: []string{"user1", "user2"},
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Values []*User `json:"values"`
			}{
				Values: []*User{
					{AccountID: "user1", DisplayName: "User One"},
					{AccountID: "user2", DisplayName: "User Two"},
				},
			},
			wantErr: false,
		},
		{
			name: "bulk get with pagination",
			opts: &BulkGetOptions{
				AccountIDs: []string{"user1", "user2", "user3"},
				MaxResults: 2,
				StartAt:    0,
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Values []*User `json:"values"`
			}{
				Values: []*User{
					{AccountID: "user1", DisplayName: "User One"},
					{AccountID: "user2", DisplayName: "User Two"},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
			errMsg:  "account IDs are required",
		},
		{
			name: "empty account IDs",
			opts: &BulkGetOptions{
				AccountIDs: []string{},
			},
			wantErr: true,
			errMsg:  "account IDs are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil || len(tt.opts.AccountIDs) == 0 {
				service := NewService(nil)
				_, err := service.BulkGet(context.Background(), tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/bulk")

				accountIDs := r.URL.Query()["accountId"]
				assert.Equal(t, len(tt.opts.AccountIDs), len(accountIDs))

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			users, err := service.BulkGet(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody.Values), len(users))
			}
		})
	}
}

func TestFindByName(t *testing.T) {
	tests := []struct {
		name           string
		userName       string
		maxResults     int
		responseStatus int
		responseBody   []*User
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful find by name",
			userName:       "john",
			maxResults:     10,
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe", EmailAddress: "john.doe@example.com"},
				{AccountID: "user2", DisplayName: "John Smith", EmailAddress: "john.smith@example.com"},
			},
			wantErr: false,
		},
		{
			name:           "find by email",
			userName:       "john.doe@example.com",
			maxResults:     5,
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user1", DisplayName: "John Doe", EmailAddress: "john.doe@example.com"},
			},
			wantErr: false,
		},
		{
			name:           "default max results",
			userName:       "alice",
			maxResults:     0, // Should default to 50
			responseStatus: http.StatusOK,
			responseBody: []*User{
				{AccountID: "user3", DisplayName: "Alice Johnson"},
			},
			wantErr: false,
		},
		{
			name:       "empty name",
			userName:   "",
			maxResults: 10,
			wantErr:    true,
			errMsg:     "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.userName == "" {
				service := NewService(nil)
				_, err := service.FindByName(context.Background(), tt.userName, tt.maxResults)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/search")
				assert.Equal(t, tt.userName, r.URL.Query().Get("query"))

				// Check that maxResults is properly set (defaults to 50 if 0 or negative)
				expectedMaxResults := tt.maxResults
				if expectedMaxResults <= 0 {
					expectedMaxResults = 50
				}
				assert.Equal(t, fmt.Sprintf("%d", expectedMaxResults), r.URL.Query().Get("maxResults"))

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			users, err := service.FindByName(context.Background(), tt.userName, tt.maxResults)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody), len(users))
			}
		})
	}
}

func TestGetDefaultColumns(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		responseStatus int
		responseBody   []struct {
			Value string `json:"value"`
		}
		wantErr bool
		errMsg  string
	}{
		{
			name:           "successful get",
			accountID:      "user123",
			responseStatus: http.StatusOK,
			responseBody: []struct {
				Value string `json:"value"`
			}{
				{Value: "issuetype"},
				{Value: "issuekey"},
				{Value: "summary"},
				{Value: "assignee"},
				{Value: "status"},
			},
			wantErr: false,
		},
		{
			name:      "empty account ID",
			accountID: "",
			wantErr:   true,
			errMsg:    "account ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.accountID == "" {
				service := NewService(nil)
				_, err := service.GetDefaultColumns(context.Background(), tt.accountID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/user/columns")
				assert.Equal(t, tt.accountID, r.URL.Query().Get("accountId"))

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			columns, err := service.GetDefaultColumns(context.Background(), tt.accountID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.responseBody), len(columns))
				for i, col := range columns {
					assert.Equal(t, tt.responseBody[i].Value, col)
				}
			}
		})
	}
}
