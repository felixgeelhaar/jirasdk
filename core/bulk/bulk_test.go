package bulk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		bodyReader = strings.NewReader(string(data))
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
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func (m *mockTransport) Close() {
	m.server.Close()
}

func TestCreateIssues(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateIssuesInput
		responseStatus int
		responseBody   *CreateIssuesResult
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateIssuesInput{
				IssueUpdates: []IssueUpdate{
					{
						Fields: map[string]interface{}{
							"project":   map[string]string{"key": "PROJ"},
							"summary":   "Test issue 1",
							"issuetype": map[string]string{"name": "Task"},
						},
					},
					{
						Fields: map[string]interface{}{
							"project":   map[string]string{"key": "PROJ"},
							"summary":   "Test issue 2",
							"issuetype": map[string]string{"name": "Task"},
						},
					},
				},
			},
			responseStatus: http.StatusCreated,
			responseBody: &CreateIssuesResult{
				Issues: []CreatedIssue{
					{ID: "10001", Key: "PROJ-1", Self: "https://example.atlassian.net/rest/api/3/issue/10001"},
					{ID: "10002", Key: "PROJ-2", Self: "https://example.atlassian.net/rest/api/3/issue/10002"},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "input is required",
		},
		{
			name: "empty issue updates",
			input: &CreateIssuesInput{
				IssueUpdates: []IssueUpdate{},
			},
			wantErr: true,
			errMsg:  "at least one issue update is required",
		},
		{
			name: "too many issues",
			input: &CreateIssuesInput{
				IssueUpdates: make([]IssueUpdate, 1001),
			},
			wantErr: true,
			errMsg:  "cannot create more than 1000 issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && tt.input != nil && (len(tt.input.IssueUpdates) == 0 || len(tt.input.IssueUpdates) > MaxBulkIssues) {
				service := NewService(nil)
				_, err := service.CreateIssues(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			if tt.input == nil {
				service := NewService(nil)
				_, err := service.CreateIssues(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/bulk")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.CreateIssues(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result.Issues, 2)
			}
		})
	}
}

func TestDeleteIssues(t *testing.T) {
	tests := []struct {
		name           string
		input          *DeleteIssuesInput
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful delete",
			input: &DeleteIssuesInput{
				IssueIDs: []string{"PROJ-1", "PROJ-2", "PROJ-3"},
			},
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "input is required",
		},
		{
			name: "empty issue IDs",
			input: &DeleteIssuesInput{
				IssueIDs: []string{},
			},
			wantErr: true,
			errMsg:  "at least one issue ID is required",
		},
		{
			name: "too many issues",
			input: &DeleteIssuesInput{
				IssueIDs: make([]string, 1001),
			},
			wantErr: true,
			errMsg:  "cannot delete more than 1000 issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && tt.input != nil && (len(tt.input.IssueIDs) == 0 || len(tt.input.IssueIDs) > MaxBulkIssues) {
				service := NewService(nil)
				err := service.DeleteIssues(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			if tt.input == nil {
				service := NewService(nil)
				err := service.DeleteIssues(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/bulk")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteIssues(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetProgress(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		responseStatus int
		responseBody   *BulkOperationProgress
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "operation running",
			taskID:         "task-123",
			responseStatus: http.StatusOK,
			responseBody: &BulkOperationProgress{
				TaskID:          "task-123",
				Status:          BulkOperationStatusRunning,
				ProgressPercent: 50,
				Created:         time.Now().Unix(),
				Started:         time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name:           "operation complete",
			taskID:         "task-456",
			responseStatus: http.StatusOK,
			responseBody: &BulkOperationProgress{
				TaskID:          "task-456",
				Status:          BulkOperationStatusComplete,
				ProgressPercent: 100,
				Result: &BulkOperationResult{
					SuccessCount: 10,
					ErrorCount:   0,
				},
				Created:   time.Now().Unix(),
				Started:   time.Now().Unix(),
				Completed: time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name:    "empty task ID",
			taskID:  "",
			wantErr: true,
			errMsg:  "task ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.taskID == "" {
				service := NewService(nil)
				_, err := service.GetProgress(context.Background(), tt.taskID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/task/")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			progress, err := service.GetProgress(context.Background(), tt.taskID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, progress)
				assert.Equal(t, tt.taskID, progress.TaskID)
			}
		})
	}
}

func TestWaitForCompletion(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		pollInterval time.Duration
		responses    []*BulkOperationProgress
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "successful completion",
			taskID:       "task-123",
			pollInterval: 100 * time.Millisecond,
			responses: []*BulkOperationProgress{
				{
					TaskID:          "task-123",
					Status:          BulkOperationStatusRunning,
					ProgressPercent: 25,
				},
				{
					TaskID:          "task-123",
					Status:          BulkOperationStatusRunning,
					ProgressPercent: 50,
				},
				{
					TaskID:          "task-123",
					Status:          BulkOperationStatusComplete,
					ProgressPercent: 100,
					Result: &BulkOperationResult{
						SuccessCount: 100,
						ErrorCount:   0,
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "operation failed",
			taskID:       "task-456",
			pollInterval: 100 * time.Millisecond,
			responses: []*BulkOperationProgress{
				{
					TaskID:          "task-456",
					Status:          BulkOperationStatusRunning,
					ProgressPercent: 30,
				},
				{
					TaskID:          "task-456",
					Status:          BulkOperationStatusFailed,
					ProgressPercent: 50,
					Result: &BulkOperationResult{
						SuccessCount: 50,
						ErrorCount:   50,
					},
				},
			},
			wantErr: false, // Function returns the failed progress, not an error
		},
		{
			name:         "empty task ID",
			taskID:       "",
			pollInterval: time.Second,
			wantErr:      true,
			errMsg:       "task ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.taskID == "" {
				service := NewService(nil)
				_, err := service.WaitForCompletion(context.Background(), tt.taskID, tt.pollInterval)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			callCount := 0
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/task/")

				if callCount < len(tt.responses) {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tt.responses[callCount])
					callCount++
				}
			})
			defer transport.Close()

			service := NewService(transport)

			// Use context with timeout to prevent infinite waiting in tests
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			progress, err := service.WaitForCompletion(ctx, tt.taskID, tt.pollInterval)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, progress)
				// Should have reached terminal state
				assert.Contains(t, []string{BulkOperationStatusComplete, BulkOperationStatusFailed, BulkOperationStatusCancelled}, progress.Status)
			}
		})
	}
}

func TestWaitForCompletionContextCancelled(t *testing.T) {
	transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
		// Always return running status
		progress := &BulkOperationProgress{
			TaskID:          "task-123",
			Status:          BulkOperationStatusRunning,
			ProgressPercent: 50,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(progress)
	})
	defer transport.Close()

	service := NewService(transport)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err := service.WaitForCompletion(ctx, "task-123", 100*time.Millisecond)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
