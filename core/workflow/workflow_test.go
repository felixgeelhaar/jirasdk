package workflow

import (
	"context"
	"encoding/json"
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

func TestGetTransitions(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		opts           *GetTransitionsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		errMsg         string
		checkResult    func(*testing.T, []*Transition)
	}{
		{
			name:           "successful get transitions",
			issueKeyOrID:   "PROJ-123",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"transitions": []map[string]interface{}{
					{
						"id":   "11",
						"name": "To Do",
						"to": map[string]interface{}{
							"id":   "10000",
							"name": "To Do",
						},
					},
					{
						"id":   "21",
						"name": "In Progress",
						"to": map[string]interface{}{
							"id":   "10001",
							"name": "In Progress",
						},
					},
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, transitions []*Transition) {
				assert.Len(t, transitions, 2)
				assert.Equal(t, "11", transitions[0].ID)
				assert.Equal(t, "To Do", transitions[0].Name)
				assert.NotNil(t, transitions[0].To)
				assert.Equal(t, "10000", transitions[0].To.ID)
			},
		},
		{
			name:         "with expand options",
			issueKeyOrID: "PROJ-123",
			opts: &GetTransitionsOptions{
				Expand:                  []string{"transitions.fields"},
				TransitionID:            "11",
				SkipRemoteOnlyCondition: true,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"transitions": []map[string]interface{}{
					{
						"id":   "11",
						"name": "To Do",
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				_, err := service.GetTransitions(context.Background(), tt.issueKeyOrID, tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/")
				assert.Contains(t, r.URL.Path, "/transitions")

				if tt.opts != nil {
					if len(tt.opts.Expand) > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("expand"))
					}
					if tt.opts.TransitionID != "" {
						assert.Equal(t, tt.opts.TransitionID, r.URL.Query().Get("transitionId"))
					}
					if tt.opts.SkipRemoteOnlyCondition {
						assert.Equal(t, "true", r.URL.Query().Get("skipRemoteOnlyCondition"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			transitions, err := service.GetTransitions(context.Background(), tt.issueKeyOrID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, transitions)
				if tt.checkResult != nil {
					tt.checkResult(t, transitions)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		checkResult    func(*testing.T, []*Workflow)
	}{
		{
			name:           "successful list",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":   "workflow-1",
						"name": "Classic Workflow",
					},
					{
						"id":   "workflow-2",
						"name": "Simplified Workflow",
					},
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, workflows []*Workflow) {
				assert.Len(t, workflows, 2)
				assert.Equal(t, "workflow-1", workflows[0].ID)
				assert.Equal(t, "Classic Workflow", workflows[0].Name)
			},
		},
		{
			name: "with workflow name filter",
			opts: &ListOptions{
				WorkflowName: "Classic",
				MaxResults:   50,
				StartAt:      10,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":   "workflow-1",
						"name": "Classic Workflow",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflow/search")

				if tt.opts != nil {
					if tt.opts.WorkflowName != "" {
						assert.Equal(t, tt.opts.WorkflowName, r.URL.Query().Get("workflowName"))
					}
					if tt.opts.MaxResults > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("maxResults"))
					}
					if tt.opts.StartAt > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("startAt"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			workflows, err := service.List(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, workflows)
				if tt.checkResult != nil {
					tt.checkResult(t, workflows)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		workflowID     string
		responseStatus int
		responseBody   *Workflow
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			workflowID:     "classic-default-workflow",
			responseStatus: http.StatusOK,
			responseBody: &Workflow{
				ID:          "classic-default-workflow",
				Name:        "Classic Default Workflow",
				Description: "The default workflow",
			},
			wantErr: false,
		},
		{
			name:       "empty workflow ID",
			workflowID: "",
			wantErr:    true,
			errMsg:     "workflow ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.workflowID == "" {
				service := NewService(nil)
				_, err := service.Get(context.Background(), tt.workflowID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflow/")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			workflow, err := service.Get(context.Background(), tt.workflowID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, workflow)
				assert.Equal(t, tt.responseBody.ID, workflow.ID)
				assert.Equal(t, tt.responseBody.Name, workflow.Name)
			}
		})
	}
}

func TestGetAllStatuses(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   []*Status
		wantErr        bool
		checkResult    func(*testing.T, []*Status)
	}{
		{
			name:           "successful get all statuses",
			responseStatus: http.StatusOK,
			responseBody: []*Status{
				{
					ID:   "10000",
					Name: "To Do",
					StatusCategory: &StatusCategory{
						ID:   1,
						Key:  "new",
						Name: "To Do",
					},
				},
				{
					ID:   "10001",
					Name: "In Progress",
					StatusCategory: &StatusCategory{
						ID:   2,
						Key:  "indeterminate",
						Name: "In Progress",
					},
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, statuses []*Status) {
				assert.Len(t, statuses, 2)
				assert.Equal(t, "10000", statuses[0].ID)
				assert.Equal(t, "To Do", statuses[0].Name)
				assert.NotNil(t, statuses[0].StatusCategory)
				assert.Equal(t, "new", statuses[0].StatusCategory.Key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/status")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			statuses, err := service.GetAllStatuses(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, statuses)
				if tt.checkResult != nil {
					tt.checkResult(t, statuses)
				}
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		statusID       string
		responseStatus int
		responseBody   *Status
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get status",
			statusID:       "10000",
			responseStatus: http.StatusOK,
			responseBody: &Status{
				ID:          "10000",
				Name:        "To Do",
				Description: "Work to be done",
				StatusCategory: &StatusCategory{
					ID:   1,
					Key:  "new",
					Name: "To Do",
				},
			},
			wantErr: false,
		},
		{
			name:     "empty status ID",
			statusID: "",
			wantErr:  true,
			errMsg:   "status ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.statusID == "" {
				service := NewService(nil)
				_, err := service.GetStatus(context.Background(), tt.statusID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/status/")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			status, err := service.GetStatus(context.Background(), tt.statusID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, status)
				assert.Equal(t, tt.responseBody.ID, status.ID)
				assert.Equal(t, tt.responseBody.Name, status.Name)
			}
		})
	}
}
