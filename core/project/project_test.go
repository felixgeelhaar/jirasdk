package project

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

func TestProjectGet(t *testing.T) {
	tests := []struct {
		name           string
		projectKeyOrID string
		opts           *GetOptions
		responseStatus int
		responseBody   *Project
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			projectKeyOrID: "PROJ",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: &Project{
				ID:   "10000",
				Key:  "PROJ",
				Name: "Test Project",
			},
			wantErr: false,
		},
		{
			name:           "with expand options",
			projectKeyOrID: "PROJ",
			opts: &GetOptions{
				Expand: []string{"lead", "description"},
			},
			responseStatus: http.StatusOK,
			responseBody: &Project{
				ID:          "10000",
				Key:         "PROJ",
				Name:        "Test Project",
				Description: "A test project",
			},
			wantErr: false,
		},
		{
			name:           "empty project key",
			projectKeyOrID: "",
			wantErr:        true,
			errMsg:         "project key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKeyOrID == "" {
				service := NewService(nil)
				_, err := service.Get(context.Background(), tt.projectKeyOrID, tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKeyOrID)

				if tt.opts != nil && len(tt.opts.Expand) > 0 {
					for _, expand := range tt.opts.Expand {
						assert.Contains(t, r.URL.Query()["expand"], expand)
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			project, err := service.Get(context.Background(), tt.projectKeyOrID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, project)
				assert.Equal(t, tt.responseBody.ID, project.ID)
				assert.Equal(t, tt.responseBody.Key, project.Key)
				assert.Equal(t, tt.responseBody.Name, project.Name)
			}
		})
	}
}

func TestProjectList(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		responseStatus int
		responseBody   struct {
			Values []*Project `json:"values"`
			Total  int        `json:"total"`
		}
		wantErr bool
	}{
		{
			name:           "successful list",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: struct {
				Values []*Project `json:"values"`
				Total  int        `json:"total"`
			}{
				Values: []*Project{
					{ID: "10000", Key: "PROJ1", Name: "Project 1"},
					{ID: "10001", Key: "PROJ2", Name: "Project 2"},
				},
				Total: 2,
			},
			wantErr: false,
		},
		{
			name: "with options",
			opts: &ListOptions{
				Expand: []string{"lead"},
				Recent: 5,
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Values []*Project `json:"values"`
				Total  int        `json:"total"`
			}{
				Values: []*Project{
					{ID: "10000", Key: "PROJ1", Name: "Project 1"},
				},
				Total: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/search")

				if tt.opts != nil {
					if len(tt.opts.Expand) > 0 {
						for _, expand := range tt.opts.Expand {
							assert.Contains(t, r.URL.Query()["expand"], expand)
						}
					}
					if tt.opts.Recent > 0 {
						assert.Equal(t, "5", r.URL.Query().Get("recent"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			projects, err := service.List(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, projects)
				assert.Equal(t, len(tt.responseBody.Values), len(projects))
			}
		})
	}
}

func TestProjectCreate(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateInput
		responseStatus int
		responseBody   *Project
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateInput{
				Key:            "NEWPROJ",
				Name:           "New Project",
				ProjectTypeKey: "software",
			},
			responseStatus: http.StatusCreated,
			responseBody: &Project{
				ID:   "10000",
				Key:  "NEWPROJ",
				Name: "New Project",
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
			name: "missing project key",
			input: &CreateInput{
				Name:           "New Project",
				ProjectTypeKey: "software",
			},
			wantErr: true,
			errMsg:  "project key is required",
		},
		{
			name: "missing name",
			input: &CreateInput{
				Key:            "NEWPROJ",
				ProjectTypeKey: "software",
			},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "missing project type",
			input: &CreateInput{
				Key:  "NEWPROJ",
				Name: "New Project",
			},
			wantErr: true,
			errMsg:  "project type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Key == "" || tt.input.Name == "" || tt.input.ProjectTypeKey == "" {
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
				assert.Contains(t, r.URL.Path, "/rest/api/3/project")

				var body CreateInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Key, body.Key)
				assert.Equal(t, tt.input.Name, body.Name)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			project, err := service.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, project)
				assert.Equal(t, tt.responseBody.ID, project.ID)
				assert.Equal(t, tt.responseBody.Key, project.Key)
			}
		})
	}
}

func TestProjectUpdate(t *testing.T) {
	tests := []struct {
		name           string
		projectKeyOrID string
		input          *UpdateInput
		responseStatus int
		responseBody   *Project
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful update",
			projectKeyOrID: "PROJ",
			input: &UpdateInput{
				Name:        "Updated Name",
				Description: "Updated description",
			},
			responseStatus: http.StatusOK,
			responseBody: &Project{
				ID:          "10000",
				Key:         "PROJ",
				Name:        "Updated Name",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:           "empty project key",
			projectKeyOrID: "",
			input:          &UpdateInput{Name: "Updated"},
			wantErr:        true,
			errMsg:         "project key or ID is required",
		},
		{
			name:           "nil input",
			projectKeyOrID: "PROJ",
			input:          nil,
			wantErr:        true,
			errMsg:         "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKeyOrID == "" || tt.input == nil {
				service := NewService(nil)
				_, err := service.Update(context.Background(), tt.projectKeyOrID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKeyOrID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			project, err := service.Update(context.Background(), tt.projectKeyOrID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, project)
				assert.Equal(t, tt.responseBody.Name, project.Name)
			}
		})
	}
}

func TestProjectDelete(t *testing.T) {
	tests := []struct {
		name           string
		projectKeyOrID string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "empty project key",
			projectKeyOrID: "",
			wantErr:        true,
			errMsg:         "project key or ID is required",
		},
		{
			name:           "unexpected status code",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKeyOrID == "" {
				service := NewService(nil)
				err := service.Delete(context.Background(), tt.projectKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKeyOrID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Delete(context.Background(), tt.projectKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectArchive(t *testing.T) {
	tests := []struct {
		name           string
		projectKeyOrID string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful archive",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "empty project key",
			projectKeyOrID: "",
			wantErr:        true,
			errMsg:         "project key or ID is required",
		},
		{
			name:           "unexpected status code",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKeyOrID == "" {
				service := NewService(nil)
				err := service.Archive(context.Background(), tt.projectKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKeyOrID+"/archive")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Archive(context.Background(), tt.projectKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectRestore(t *testing.T) {
	tests := []struct {
		name           string
		projectKeyOrID string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful restore",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty project key",
			projectKeyOrID: "",
			wantErr:        true,
			errMsg:         "project key or ID is required",
		},
		{
			name:           "unexpected status code",
			projectKeyOrID: "PROJ",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKeyOrID == "" {
				service := NewService(nil)
				err := service.Restore(context.Background(), tt.projectKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKeyOrID+"/restore")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.Restore(context.Background(), tt.projectKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
