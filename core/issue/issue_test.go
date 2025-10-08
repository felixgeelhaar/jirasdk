package issue

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		bodyReader = io.NopCloser(newBytesReader(data))
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

	if resp.StatusCode >= 400 {
		return &mockError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return json.Unmarshal(body, target)
}

func (m *mockTransport) Close() {
	m.server.Close()
}

type mockError struct {
	StatusCode int
	Message    string
}

func (e *mockError) Error() string {
	return e.Message
}

// Helper to create bytes reader from byte slice
func newBytesReader(data []byte) io.Reader {
	return io.NopCloser(io.Reader(&bytesReader{data: data}))
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

func TestIssueGet(t *testing.T) {
	tests := []struct {
		name           string
		issueKey       string
		responseStatus int
		responseBody   *Issue
		wantErr        bool
	}{
		{
			name:           "successful get",
			issueKey:       "PROJ-123",
			responseStatus: http.StatusOK,
			responseBody: &Issue{
				ID:  "10001",
				Key: "PROJ-123",
				Fields: &IssueFields{
					Summary:     "Test issue",
					Description: "Test description",
				},
			},
			wantErr: false,
		},
		{
			name:           "issue not found",
			issueKey:       "PROJ-999",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:     "empty issue key",
			issueKey: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKey == "" {
				// Test validation without making HTTP call
				service := NewService(nil)
				_, err := service.Get(context.Background(), tt.issueKey, nil)
				require.Error(t, err)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, tt.issueKey)

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			})
			defer transport.Close()

			service := NewService(transport)
			issue, err := service.Get(context.Background(), tt.issueKey, nil)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, issue)
			} else {
				require.NoError(t, err)
				require.NotNil(t, issue)
				assert.Equal(t, tt.responseBody.Key, issue.Key)
				assert.Equal(t, tt.responseBody.Fields.Summary, issue.Fields.Summary)
			}
		})
	}
}

func TestIssueCreate(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateInput
		responseStatus int
		responseBody   map[string]string
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateInput{
				Fields: &IssueFields{
					Project:   &Project{Key: "PROJ"},
					Summary:   "New issue",
					IssueType: &IssueType{Name: "Task"},
				},
			},
			responseStatus: http.StatusCreated,
			responseBody: map[string]string{
				"id":   "10001",
				"key":  "PROJ-124",
				"self": "https://example.atlassian.net/rest/api/3/issue/10001",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "create input is required",
		},
		{
			name: "missing project",
			input: &CreateInput{
				Fields: &IssueFields{
					Summary:   "New issue",
					IssueType: &IssueType{Name: "Task"},
				},
			},
			wantErr: true,
			errMsg:  "project is required",
		},
		{
			name: "missing summary",
			input: &CreateInput{
				Fields: &IssueFields{
					Project:   &Project{Key: "PROJ"},
					IssueType: &IssueType{Name: "Task"},
				},
			},
			wantErr: true,
			errMsg:  "summary is required",
		},
		{
			name: "missing issue type",
			input: &CreateInput{
				Fields: &IssueFields{
					Project: &Project{Key: "PROJ"},
					Summary: "New issue",
				},
			},
			wantErr: true,
			errMsg:  "issue type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Fields.Project == nil || tt.input.Fields.Summary == "" || tt.input.Fields.IssueType == nil {
				// Test validation without making HTTP call
				service := NewService(nil)
				_, err := service.Create(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue")

				// Verify request body
				var body CreateInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Fields.Summary, body.Fields.Summary)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			issue, err := service.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, issue)
				assert.Equal(t, tt.responseBody["key"], issue.Key)
			}
		})
	}
}

func TestIssueUpdate(t *testing.T) {
	tests := []struct {
		name           string
		issueKey       string
		input          *UpdateInput
		responseStatus int
		wantErr        bool
	}{
		{
			name:     "successful update",
			issueKey: "PROJ-123",
			input: &UpdateInput{
				Fields: map[string]interface{}{
					"summary": "Updated summary",
				},
			},
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "empty issue key",
			issueKey: "",
			input: &UpdateInput{
				Fields: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name:     "nil input",
			issueKey: "PROJ-123",
			input:    nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKey == "" || tt.input == nil {
				service := NewService(nil)
				err := service.Update(context.Background(), tt.issueKey, tt.input)
				require.Error(t, err)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, tt.issueKey)
				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Update(context.Background(), tt.issueKey, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIssueDelete(t *testing.T) {
	tests := []struct {
		name           string
		issueKey       string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful delete",
			issueKey:       "PROJ-123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "empty issue key",
			issueKey: "",
			wantErr:  true,
		},
		{
			name:           "not found",
			issueKey:       "PROJ-999",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKey == "" {
				service := NewService(nil)
				err := service.Delete(context.Background(), tt.issueKey)
				require.Error(t, err)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, tt.issueKey)
				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Delete(context.Background(), tt.issueKey)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIssueDoTransition(t *testing.T) {
	now := time.Now()
	_ = now // Use the variable to avoid compiler error

	tests := []struct {
		name           string
		issueKey       string
		input          *TransitionInput
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:     "successful transition",
			issueKey: "PROJ-123",
			input: &TransitionInput{
				Transition: &Transition{ID: "11"},
			},
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "empty issue key",
			issueKey: "",
			input: &TransitionInput{
				Transition: &Transition{ID: "11"},
			},
			wantErr: true,
		},
		{
			name:     "nil input",
			issueKey: "PROJ-123",
			input:    nil,
			wantErr:  true,
			errMsg:   "transition input is required",
		},
		{
			name:     "missing transition ID",
			issueKey: "PROJ-123",
			input: &TransitionInput{
				Transition: &Transition{ID: ""},
			},
			wantErr: true,
			errMsg:  "transition ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKey == "" || tt.input == nil || (tt.input.Transition != nil && tt.input.Transition.ID == "") {
				service := NewService(nil)
				err := service.DoTransition(context.Background(), tt.issueKey, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, tt.issueKey)
				assert.Contains(t, r.URL.Path, "transitions")
				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DoTransition(context.Background(), tt.issueKey, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
