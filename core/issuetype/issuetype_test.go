package issuetype

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

// mockTransport implements the RoundTripper interface for testing.
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

func TestList(t *testing.T) {
	transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/rest/api/3/issuetype", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*IssueType{
			{ID: "10001", Name: "Bug", Subtask: false},
			{ID: "10002", Name: "Story", Subtask: false},
		})
	})
	defer transport.Close()

	service := NewService(transport)
	issueTypes, err := service.List(context.Background())

	require.NoError(t, err)
	require.Len(t, issueTypes, 2)
	assert.Equal(t, "10001", issueTypes[0].ID)
	assert.Equal(t, "Bug", issueTypes[0].Name)
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		issueTypeID string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "success",
			issueTypeID: "10001",
			wantErr:     false,
		},
		{
			name:        "empty ID",
			issueTypeID: "",
			wantErr:     true,
			errMsg:      "issue type ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueTypeID == "" {
				service := NewService(nil)
				_, err := service.Get(context.Background(), tt.issueTypeID)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issuetype/")

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&IssueType{
					ID:   "10001",
					Name: "Bug",
				})
			})
			defer transport.Close()

			service := NewService(transport)
			issueType, err := service.Get(context.Background(), tt.issueTypeID)

			require.NoError(t, err)
			require.NotNil(t, issueType)
			assert.Equal(t, "10001", issueType.ID)
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateIssueTypeInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			input: &CreateIssueTypeInput{
				Name:        "Incident",
				Description: "Production incident",
				Type:        "standard",
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
			name: "empty name",
			input: &CreateIssueTypeInput{
				Type: "standard",
			},
			wantErr: true,
			errMsg:  "issue type name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				_, err := service.Create(context.Background(), tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/rest/api/3/issuetype", r.URL.Path)

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(&IssueType{
					ID:   "10003",
					Name: tt.input.Name,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			issueType, err := service.Create(context.Background(), tt.input)

			require.NoError(t, err)
			require.NotNil(t, issueType)
			assert.Equal(t, tt.input.Name, issueType.Name)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		issueTypeID string
		input       *UpdateIssueTypeInput
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "success",
			issueTypeID: "10001",
			input: &UpdateIssueTypeInput{
				Name: "Updated Bug",
			},
			wantErr: false,
		},
		{
			name:        "empty ID",
			issueTypeID: "",
			input:       &UpdateIssueTypeInput{Name: "Updated"},
			wantErr:     true,
			errMsg:      "issue type ID is required",
		},
		{
			name:        "nil input",
			issueTypeID: "10001",
			input:       nil,
			wantErr:     true,
			errMsg:      "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				_, err := service.Update(context.Background(), tt.issueTypeID, tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issuetype/")

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&IssueType{
					ID:   tt.issueTypeID,
					Name: tt.input.Name,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			issueType, err := service.Update(context.Background(), tt.issueTypeID, tt.input)

			require.NoError(t, err)
			require.NotNil(t, issueType)
			assert.Equal(t, tt.input.Name, issueType.Name)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name                   string
		issueTypeID            string
		alternativeIssueTypeID string
		wantErr                bool
		errMsg                 string
	}{
		{
			name:                   "success",
			issueTypeID:            "10001",
			alternativeIssueTypeID: "",
			wantErr:                false,
		},
		{
			name:                   "with alternative ID",
			issueTypeID:            "10001",
			alternativeIssueTypeID: "10002",
			wantErr:                false,
		},
		{
			name:        "empty ID",
			issueTypeID: "",
			wantErr:     true,
			errMsg:      "issue type ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.Delete(context.Background(), tt.issueTypeID, tt.alternativeIssueTypeID)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issuetype/")

				if tt.alternativeIssueTypeID != "" {
					assert.Equal(t, tt.alternativeIssueTypeID, r.URL.Query().Get("alternativeIssueTypeId"))
				}

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Delete(context.Background(), tt.issueTypeID, tt.alternativeIssueTypeID)

			require.NoError(t, err)
		})
	}
}

func TestListIssueTypeSchemes(t *testing.T) {
	tests := []struct {
		name         string
		opts         *ListIssueTypeSchemesOptions
		responseBody interface{}
		checkResult  func(*testing.T, []*IssueTypeScheme)
	}{
		{
			name: "success",
			opts: nil,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{"id": "10000", "name": "Default Issue Type Scheme", "isDefault": true},
					{"id": "10001", "name": "Custom Scheme", "isDefault": false},
				},
			},
			checkResult: func(t *testing.T, schemes []*IssueTypeScheme) {
				assert.Len(t, schemes, 2)
				assert.Equal(t, "10000", schemes[0].ID)
				assert.Equal(t, "Default Issue Type Scheme", schemes[0].Name)
				assert.True(t, schemes[0].IsDefault)
			},
		},
		{
			name: "with pagination",
			opts: &ListIssueTypeSchemesOptions{
				StartAt:    10,
				MaxResults: 25,
			},
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{"id": "10002", "name": "Another Scheme"},
				},
			},
			checkResult: func(t *testing.T, schemes []*IssueTypeScheme) {
				assert.Len(t, schemes, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/rest/api/3/issuetypescheme", r.URL.Path)

				if tt.opts != nil {
					if tt.opts.StartAt > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("startAt"))
					}
					if tt.opts.MaxResults > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("maxResults"))
					}
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			schemes, err := service.ListIssueTypeSchemes(context.Background(), tt.opts)

			require.NoError(t, err)
			require.NotNil(t, schemes)
			if tt.checkResult != nil {
				tt.checkResult(t, schemes)
			}
		})
	}
}

func TestCreateIssueTypeScheme(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateIssueTypeSchemeInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			input: &CreateIssueTypeSchemeInput{
				Name:               "Software Development",
				Description:        "Issue types for software projects",
				DefaultIssueTypeID: "10001",
				IssueTypeIDs:       []string{"10001", "10002"},
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
			name: "empty name",
			input: &CreateIssueTypeSchemeInput{
				IssueTypeIDs: []string{"10001"},
			},
			wantErr: true,
			errMsg:  "scheme name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				_, err := service.CreateIssueTypeScheme(context.Background(), tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/rest/api/3/issuetypescheme", r.URL.Path)

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(&IssueTypeScheme{
					ID:   "10010",
					Name: tt.input.Name,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.CreateIssueTypeScheme(context.Background(), tt.input)

			require.NoError(t, err)
			require.NotNil(t, scheme)
			assert.Equal(t, tt.input.Name, scheme.Name)
		})
	}
}

func TestUpdateIssueTypeScheme(t *testing.T) {
	tests := []struct {
		name     string
		schemeID string
		input    *UpdateIssueTypeSchemeInput
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: "10000",
			input: &UpdateIssueTypeSchemeInput{
				Name:        "Updated Scheme",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:     "empty ID",
			schemeID: "",
			input:    &UpdateIssueTypeSchemeInput{Name: "Updated"},
			wantErr:  true,
			errMsg:   "scheme ID is required",
		},
		{
			name:     "nil input",
			schemeID: "10000",
			input:    nil,
			wantErr:  true,
			errMsg:   "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.UpdateIssueTypeScheme(context.Background(), tt.schemeID, tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issuetypescheme/")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.UpdateIssueTypeScheme(context.Background(), tt.schemeID, tt.input)

			require.NoError(t, err)
		})
	}
}

func TestDeleteIssueTypeScheme(t *testing.T) {
	tests := []struct {
		name     string
		schemeID string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: "10000",
			wantErr:  false,
		},
		{
			name:     "empty ID",
			schemeID: "",
			wantErr:  true,
			errMsg:   "scheme ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.DeleteIssueTypeScheme(context.Background(), tt.schemeID)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issuetypescheme/")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteIssueTypeScheme(context.Background(), tt.schemeID)

			require.NoError(t, err)
		})
	}
}

func TestAddIssueTypesToScheme(t *testing.T) {
	tests := []struct {
		name     string
		schemeID string
		input    *AddIssueTypesToSchemeInput
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: "10000",
			input: &AddIssueTypesToSchemeInput{
				IssueTypeIDs: []string{"10004", "10005"},
			},
			wantErr: false,
		},
		{
			name:     "empty ID",
			schemeID: "",
			input: &AddIssueTypesToSchemeInput{
				IssueTypeIDs: []string{"10004"},
			},
			wantErr: true,
			errMsg:  "scheme ID is required",
		},
		{
			name:     "nil input",
			schemeID: "10000",
			input:    nil,
			wantErr:  true,
			errMsg:   "at least one issue type ID is required",
		},
		{
			name:     "empty list",
			schemeID: "10000",
			input:    &AddIssueTypesToSchemeInput{},
			wantErr:  true,
			errMsg:   "at least one issue type ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.AddIssueTypesToScheme(context.Background(), tt.schemeID, tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/issuetype")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.AddIssueTypesToScheme(context.Background(), tt.schemeID, tt.input)

			require.NoError(t, err)
		})
	}
}

func TestRemoveIssueTypeFromScheme(t *testing.T) {
	tests := []struct {
		name        string
		schemeID    string
		issueTypeID string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "success",
			schemeID:    "10000",
			issueTypeID: "10004",
			wantErr:     false,
		},
		{
			name:        "empty scheme ID",
			schemeID:    "",
			issueTypeID: "10004",
			wantErr:     true,
			errMsg:      "scheme ID is required",
		},
		{
			name:        "empty issue type ID",
			schemeID:    "10000",
			issueTypeID: "",
			wantErr:     true,
			errMsg:      "issue type ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.RemoveIssueTypeFromScheme(context.Background(), tt.schemeID, tt.issueTypeID)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/issuetype/")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.RemoveIssueTypeFromScheme(context.Background(), tt.schemeID, tt.issueTypeID)

			require.NoError(t, err)
		})
	}
}

func TestGetIssueTypeSchemeMappings(t *testing.T) {
	tests := []struct {
		name         string
		opts         *GetIssueTypeSchemeMappingsOptions
		responseBody interface{}
		checkResult  func(*testing.T, []*IssueTypeSchemeMapping)
	}{
		{
			name: "success",
			opts: nil,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{"issueTypeSchemeId": "10000", "issueTypeId": "10001"},
					{"issueTypeSchemeId": "10000", "issueTypeId": "10002"},
				},
			},
			checkResult: func(t *testing.T, mappings []*IssueTypeSchemeMapping) {
				assert.Len(t, mappings, 2)
				assert.Equal(t, "10000", mappings[0].IssueTypeSchemeID)
				assert.Equal(t, "10001", mappings[0].IssueTypeID)
			},
		},
		{
			name: "with scheme IDs filter",
			opts: &GetIssueTypeSchemeMappingsOptions{
				IssueTypeSchemeIDs: []string{"10000", "10001"},
				MaxResults:         50,
			},
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{"issueTypeSchemeId": "10000", "issueTypeId": "10001"},
				},
			},
			checkResult: func(t *testing.T, mappings []*IssueTypeSchemeMapping) {
				assert.Len(t, mappings, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/rest/api/3/issuetypescheme/mapping", r.URL.Path)

				if tt.opts != nil && len(tt.opts.IssueTypeSchemeIDs) > 0 {
					ids := r.URL.Query()["issueTypeSchemeId"]
					assert.Equal(t, len(tt.opts.IssueTypeSchemeIDs), len(ids))
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			mappings, err := service.GetIssueTypeSchemeMappings(context.Background(), tt.opts)

			require.NoError(t, err)
			require.NotNil(t, mappings)
			if tt.checkResult != nil {
				tt.checkResult(t, mappings)
			}
		})
	}
}
