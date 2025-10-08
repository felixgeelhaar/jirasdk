package permission

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetAllPermissions(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "successful get",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissions": []map[string]interface{}{
					{
						"id":          "BROWSE_PROJECTS",
						"key":         "BROWSE_PROJECTS",
						"name":        "Browse Projects",
						"type":        "PROJECT",
						"description": "Ability to browse projects",
					},
					{
						"id":   "EDIT_ISSUES",
						"key":  "EDIT_ISSUES",
						"name": "Edit Issues",
						"type": "PROJECT",
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
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissions")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			permissions, err := service.GetAllPermissions(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, permissions)
				assert.Len(t, permissions, 2)
			}
		})
	}
}

func TestGetMyPermissions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *MyPermissionsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "successful get without options",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissions": map[string]interface{}{
					"EDIT_ISSUES": map[string]interface{}{
						"id":             "EDIT_ISSUES",
						"key":            "EDIT_ISSUES",
						"name":           "Edit Issues",
						"type":           "PROJECT",
						"havePermission": true,
					},
					"DELETE_ISSUES": map[string]interface{}{
						"id":             "DELETE_ISSUES",
						"key":            "DELETE_ISSUES",
						"name":           "Delete Issues",
						"type":           "PROJECT",
						"havePermission": false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with project filter",
			opts: &MyPermissionsOptions{
				ProjectKey:  "PROJ",
				Permissions: "EDIT_ISSUES,DELETE_ISSUES",
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissions": map[string]interface{}{
					"EDIT_ISSUES": map[string]interface{}{
						"id":             "EDIT_ISSUES",
						"havePermission": true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with issue key filter",
			opts: &MyPermissionsOptions{
				IssueKey:    "PROJ-123",
				Permissions: "EDIT_ISSUES",
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissions": map[string]interface{}{
					"EDIT_ISSUES": map[string]interface{}{
						"id":             "EDIT_ISSUES",
						"havePermission": true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with project ID and issue ID filters",
			opts: &MyPermissionsOptions{
				ProjectID:   "10000",
				IssueID:     "10001",
				Permissions: "EDIT_ISSUES,DELETE_ISSUES",
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissions": map[string]interface{}{
					"EDIT_ISSUES": map[string]interface{}{
						"id":             "EDIT_ISSUES",
						"havePermission": true,
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
				assert.Contains(t, r.URL.Path, "/rest/api/3/mypermissions")

				if tt.opts != nil {
					if tt.opts.ProjectKey != "" {
						assert.Equal(t, tt.opts.ProjectKey, r.URL.Query().Get("projectKey"))
					}
					if tt.opts.ProjectID != "" {
						assert.Equal(t, tt.opts.ProjectID, r.URL.Query().Get("projectId"))
					}
					if tt.opts.IssueKey != "" {
						assert.Equal(t, tt.opts.IssueKey, r.URL.Query().Get("issueKey"))
					}
					if tt.opts.IssueID != "" {
						assert.Equal(t, tt.opts.IssueID, r.URL.Query().Get("issueId"))
					}
					if tt.opts.Permissions != "" {
						assert.Equal(t, tt.opts.Permissions, r.URL.Query().Get("permissions"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			perms, err := service.GetMyPermissions(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, perms)
				assert.NotEmpty(t, perms.Permissions)
			}
		})
	}
}

func TestListPermissionSchemes(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListPermissionSchemesOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "successful list",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissionSchemes": []map[string]interface{}{
					{
						"id":          int64(10000),
						"name":        "Default Permission Scheme",
						"description": "Default scheme",
					},
					{
						"id":   int64(10001),
						"name": "Custom Scheme",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with expand",
			opts: &ListPermissionSchemesOptions{
				Expand: []string{"permissions"},
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"permissionSchemes": []map[string]interface{}{
					{
						"id":   int64(10000),
						"name": "Default Permission Scheme",
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
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissionscheme")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			schemes, err := service.ListPermissionSchemes(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schemes)
			}
		})
	}
}

func TestGetPermissionScheme(t *testing.T) {
	tests := []struct {
		name           string
		schemeID       int64
		opts           *GetPermissionSchemeOptions
		responseStatus int
		responseBody   *PermissionScheme
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			schemeID:       10000,
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: &PermissionScheme{
				ID:          10000,
				Name:        "Default Permission Scheme",
				Description: "Default scheme for all projects",
			},
			wantErr: false,
		},
		{
			name:     "invalid scheme ID",
			schemeID: 0,
			wantErr:  true,
			errMsg:   "scheme ID is required",
		},
		{
			name:     "with expand parameter",
			schemeID: 10000,
			opts: &GetPermissionSchemeOptions{
				Expand: []string{"permissions", "user", "group", "projectRole"},
			},
			responseStatus: http.StatusOK,
			responseBody: &PermissionScheme{
				ID:          10000,
				Name:        "Expanded Scheme",
				Description: "Scheme with expanded fields",
				Permissions: []*PermissionGrant{
					{
						ID:         1,
						Permission: "BROWSE_PROJECTS",
						Holder: &PermissionHolder{
							Type: "group",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schemeID <= 0 {
				service := NewService(nil)
				_, err := service.GetPermissionScheme(context.Background(), tt.schemeID, tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissionscheme/10000")

				if tt.opts != nil && len(tt.opts.Expand) > 0 {
					assert.Contains(t, r.URL.Query().Get("expand"), "permissions")
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.GetPermissionScheme(context.Background(), tt.schemeID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scheme)
				assert.Equal(t, tt.responseBody.ID, scheme.ID)
			}
		})
	}
}

func TestCreatePermissionScheme(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreatePermissionSchemeInput
		responseStatus int
		responseBody   *PermissionScheme
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreatePermissionSchemeInput{
				Name:        "New Scheme",
				Description: "A new permission scheme",
			},
			responseStatus: http.StatusCreated,
			responseBody: &PermissionScheme{
				ID:          10002,
				Name:        "New Scheme",
				Description: "A new permission scheme",
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
			name: "missing name",
			input: &CreatePermissionSchemeInput{
				Description: "Description only",
			},
			wantErr: true,
			errMsg:  "scheme name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Name == "" {
				service := NewService(nil)
				_, err := service.CreatePermissionScheme(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissionscheme")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.CreatePermissionScheme(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scheme)
				assert.Equal(t, tt.responseBody.ID, scheme.ID)
			}
		})
	}
}

func TestUpdatePermissionScheme(t *testing.T) {
	tests := []struct {
		name           string
		schemeID       int64
		input          *UpdatePermissionSchemeInput
		responseStatus int
		responseBody   *PermissionScheme
		wantErr        bool
		errMsg         string
	}{
		{
			name:     "successful update",
			schemeID: 10000,
			input: &UpdatePermissionSchemeInput{
				Description: "Updated description",
			},
			responseStatus: http.StatusOK,
			responseBody: &PermissionScheme{
				ID:          10000,
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:     "invalid scheme ID",
			schemeID: 0,
			input:    &UpdatePermissionSchemeInput{},
			wantErr:  true,
			errMsg:   "scheme ID is required",
		},
		{
			name:     "nil input",
			schemeID: 10000,
			input:    nil,
			wantErr:  true,
			errMsg:   "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schemeID <= 0 || tt.input == nil {
				service := NewService(nil)
				_, err := service.UpdatePermissionScheme(context.Background(), tt.schemeID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissionscheme/10000")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.UpdatePermissionScheme(context.Background(), tt.schemeID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scheme)
			}
		})
	}
}

func TestDeletePermissionScheme(t *testing.T) {
	tests := []struct {
		name           string
		schemeID       int64
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			schemeID:       10000,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "invalid scheme ID",
			schemeID: 0,
			wantErr:  true,
			errMsg:   "scheme ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schemeID <= 0 {
				service := NewService(nil)
				err := service.DeletePermissionScheme(context.Background(), tt.schemeID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/permissionscheme/10000")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeletePermissionScheme(context.Background(), tt.schemeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetProjectRoles(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		responseStatus int
		responseBody   map[string]string
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			projectKey:     "PROJ",
			responseStatus: http.StatusOK,
			responseBody: map[string]string{
				"Administrators": "https://example.atlassian.net/rest/api/3/project/PROJ/role/10002",
				"Developers":     "https://example.atlassian.net/rest/api/3/project/PROJ/role/10001",
			},
			wantErr: false,
		},
		{
			name:       "empty project key",
			projectKey: "",
			wantErr:    true,
			errMsg:     "project key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKey == "" {
				service := NewService(nil)
				_, err := service.GetProjectRoles(context.Background(), tt.projectKey)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/PROJ/role")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			roles, err := service.GetProjectRoles(context.Background(), tt.projectKey)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, roles)
				assert.Len(t, roles, 2)
			}
		})
	}
}

func TestGetProjectRole(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		roleID         int64
		responseStatus int
		responseBody   *ProjectRole
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			projectKey:     "PROJ",
			roleID:         10002,
			responseStatus: http.StatusOK,
			responseBody: &ProjectRole{
				ID:          10002,
				Name:        "Administrators",
				Description: "Admin role",
			},
			wantErr: false,
		},
		{
			name:       "empty project key",
			projectKey: "",
			roleID:     10002,
			wantErr:    true,
			errMsg:     "project key or ID is required",
		},
		{
			name:       "invalid role ID",
			projectKey: "PROJ",
			roleID:     0,
			wantErr:    true,
			errMsg:     "role ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKey == "" || tt.roleID <= 0 {
				service := NewService(nil)
				_, err := service.GetProjectRole(context.Background(), tt.projectKey, tt.roleID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/PROJ/role/10002")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			role, err := service.GetProjectRole(context.Background(), tt.projectKey, tt.roleID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, role)
				assert.Equal(t, tt.responseBody.ID, role.ID)
			}
		})
	}
}

func TestAddActorsToProjectRole(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		roleID         int64
		input          *AddActorInput
		responseStatus int
		responseBody   *ProjectRole
		wantErr        bool
		errMsg         string
	}{
		{
			name:       "successful add users",
			projectKey: "PROJ",
			roleID:     10002,
			input: &AddActorInput{
				User: []string{"accountId1", "accountId2"},
			},
			responseStatus: http.StatusOK,
			responseBody: &ProjectRole{
				ID:   10002,
				Name: "Administrators",
			},
			wantErr: false,
		},
		{
			name:       "successful add groups",
			projectKey: "PROJ",
			roleID:     10002,
			input: &AddActorInput{
				Group: []string{"developers", "admins"},
			},
			responseStatus: http.StatusOK,
			responseBody: &ProjectRole{
				ID:   10002,
				Name: "Administrators",
			},
			wantErr: false,
		},
		{
			name:       "empty actors",
			projectKey: "PROJ",
			roleID:     10002,
			input:      &AddActorInput{},
			wantErr:    true,
			errMsg:     "at least one user or group is required",
		},
		{
			name:       "nil input",
			projectKey: "PROJ",
			roleID:     10002,
			input:      nil,
			wantErr:    true,
			errMsg:     "at least one user or group is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || (len(tt.input.User) == 0 && len(tt.input.Group) == 0) {
				service := NewService(nil)
				_, err := service.AddActorsToProjectRole(context.Background(), tt.projectKey, tt.roleID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/PROJ/role/10002")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			role, err := service.AddActorsToProjectRole(context.Background(), tt.projectKey, tt.roleID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, role)
			}
		})
	}
}

func TestRemoveActorFromProjectRole(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		roleID         int64
		actorType      string
		actor          string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful remove user",
			projectKey:     "PROJ",
			roleID:         10002,
			actorType:      "user",
			actor:          "accountId123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "successful remove group",
			projectKey:     "PROJ",
			roleID:         10002,
			actorType:      "group",
			actor:          "developers",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:       "empty actor type",
			projectKey: "PROJ",
			roleID:     10002,
			actorType:  "",
			actor:      "accountId123",
			wantErr:    true,
			errMsg:     "actor type is required",
		},
		{
			name:       "empty actor",
			projectKey: "PROJ",
			roleID:     10002,
			actorType:  "user",
			actor:      "",
			wantErr:    true,
			errMsg:     "actor is required",
		},
		{
			name:       "empty project key",
			projectKey: "",
			roleID:     10002,
			actorType:  "user",
			actor:      "accountId123",
			wantErr:    true,
			errMsg:     "project key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.projectKey == "" || tt.actorType == "" || tt.actor == "" {
				service := NewService(nil)
				err := service.RemoveActorFromProjectRole(context.Background(), tt.projectKey, tt.roleID, tt.actorType, tt.actor)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/PROJ/role/10002")
				assert.Equal(t, tt.actor, r.URL.Query().Get(tt.actorType))

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.RemoveActorFromProjectRole(context.Background(), tt.projectKey, tt.roleID, tt.actorType, tt.actor)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
