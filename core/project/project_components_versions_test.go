package project

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateComponent(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateComponentInput
		responseStatus int
		responseBody   *Component
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateComponentInput{
				Name:          "Backend Services",
				Description:   "All backend microservices",
				Project:       "PROJ",
				AssigneeType:  "PROJECT_DEFAULT",
				LeadAccountID: "123456:abcdef",
			},
			responseStatus: http.StatusCreated,
			responseBody: &Component{
				ID:          "10000",
				Name:        "Backend Services",
				Description: "All backend microservices",
				Lead: &User{
					AccountID:   "123456:abcdef",
					DisplayName: "John Doe",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			input: &CreateComponentInput{
				Project: "PROJ",
			},
			wantErr: true,
			errMsg:  "component name is required",
		},
		{
			name: "missing project",
			input: &CreateComponentInput{
				Name: "Backend Services",
			},
			wantErr: true,
			errMsg:  "project is required",
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Name == "" || tt.input.Project == "" {
				service := NewService(nil)
				_, err := service.CreateComponent(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/component")

				var received CreateComponentInput
				err := json.NewDecoder(r.Body).Decode(&received)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Name, received.Name)
				assert.Equal(t, tt.input.Project, received.Project)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			component, err := service.CreateComponent(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, component)
				assert.Equal(t, tt.responseBody.ID, component.ID)
				assert.Equal(t, tt.responseBody.Name, component.Name)
			}
		})
	}
}

func TestUpdateComponent(t *testing.T) {
	tests := []struct {
		name           string
		componentID    string
		input          *UpdateComponentInput
		responseStatus int
		responseBody   *Component
		wantErr        bool
		errMsg         string
	}{
		{
			name:        "successful update",
			componentID: "10000",
			input: &UpdateComponentInput{
				Name:        "Updated Component",
				Description: "Updated description",
			},
			responseStatus: http.StatusOK,
			responseBody: &Component{
				ID:          "10000",
				Name:        "Updated Component",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:        "empty component ID",
			componentID: "",
			input:       &UpdateComponentInput{Name: "Test"},
			wantErr:     true,
			errMsg:      "component ID is required",
		},
		{
			name:        "nil input",
			componentID: "10000",
			input:       nil,
			wantErr:     true,
			errMsg:      "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.componentID == "" || tt.input == nil {
				service := NewService(nil)
				_, err := service.UpdateComponent(context.Background(), tt.componentID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/component/"+tt.componentID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			component, err := service.UpdateComponent(context.Background(), tt.componentID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, component)
				assert.Equal(t, tt.responseBody.Name, component.Name)
			}
		})
	}
}

func TestGetComponent(t *testing.T) {
	tests := []struct {
		name           string
		componentID    string
		responseStatus int
		responseBody   *Component
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			componentID:    "10000",
			responseStatus: http.StatusOK,
			responseBody: &Component{
				ID:          "10000",
				Name:        "Backend Services",
				Description: "All backend microservices",
			},
			wantErr: false,
		},
		{
			name:        "empty component ID",
			componentID: "",
			wantErr:     true,
			errMsg:      "component ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.componentID == "" {
				service := NewService(nil)
				_, err := service.GetComponent(context.Background(), tt.componentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/component/"+tt.componentID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			component, err := service.GetComponent(context.Background(), tt.componentID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, component)
				assert.Equal(t, tt.responseBody.ID, component.ID)
				assert.Equal(t, tt.responseBody.Name, component.Name)
			}
		})
	}
}

func TestDeleteComponent(t *testing.T) {
	tests := []struct {
		name           string
		componentID    string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			componentID:    "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:        "empty component ID",
			componentID: "",
			wantErr:     true,
			errMsg:      "component ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.componentID == "" {
				service := NewService(nil)
				err := service.DeleteComponent(context.Background(), tt.componentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/component/"+tt.componentID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteComponent(context.Background(), tt.componentID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListProjectComponents(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		responseStatus int
		responseBody   []*Component
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful list",
			projectKey:     "PROJ",
			responseStatus: http.StatusOK,
			responseBody: []*Component{
				{
					ID:          "10000",
					Name:        "Backend",
					Description: "Backend services",
				},
				{
					ID:          "10001",
					Name:        "Frontend",
					Description: "Frontend application",
				},
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
				_, err := service.ListProjectComponents(context.Background(), tt.projectKey)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKey+"/components")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			components, err := service.ListProjectComponents(context.Background(), tt.projectKey)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, components)
				assert.Len(t, components, len(tt.responseBody))
			}
		})
	}
}

func TestCreateVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateVersionInput
		responseStatus int
		responseBody   *Version
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateVersionInput{
				Name:        "v1.0.0",
				Description: "First major release",
				Project:     "PROJ",
				StartDate:   "2024-01-01",
				ReleaseDate: "2024-06-01",
				Released:    false,
				Archived:    false,
			},
			responseStatus: http.StatusCreated,
			responseBody: &Version{
				ID:          "10000",
				Name:        "v1.0.0",
				Description: "First major release",
				StartDate:   "2024-01-01",
				ReleaseDate: "2024-06-01",
				Released:    false,
				Archived:    false,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			input: &CreateVersionInput{
				Project: "PROJ",
			},
			wantErr: true,
			errMsg:  "version name is required",
		},
		{
			name: "missing project",
			input: &CreateVersionInput{
				Name: "v1.0.0",
			},
			wantErr: true,
			errMsg:  "project is required",
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Name == "" || tt.input.Project == "" {
				service := NewService(nil)
				_, err := service.CreateVersion(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/version")

				var received CreateVersionInput
				err := json.NewDecoder(r.Body).Decode(&received)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Name, received.Name)
				assert.Equal(t, tt.input.Project, received.Project)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			version, err := service.CreateVersion(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, version)
				assert.Equal(t, tt.responseBody.ID, version.ID)
				assert.Equal(t, tt.responseBody.Name, version.Name)
			}
		})
	}
}

func TestUpdateVersion(t *testing.T) {
	releasedTrue := true
	releasedFalse := false

	tests := []struct {
		name           string
		versionID      string
		input          *UpdateVersionInput
		responseStatus int
		responseBody   *Version
		wantErr        bool
		errMsg         string
	}{
		{
			name:      "successful update with released flag",
			versionID: "10000",
			input: &UpdateVersionInput{
				Name:     "v1.0.1",
				Released: &releasedTrue,
			},
			responseStatus: http.StatusOK,
			responseBody: &Version{
				ID:       "10000",
				Name:     "v1.0.1",
				Released: true,
			},
			wantErr: false,
		},
		{
			name:      "update without releasing",
			versionID: "10000",
			input: &UpdateVersionInput{
				Description: "Updated description",
				Released:    &releasedFalse,
			},
			responseStatus: http.StatusOK,
			responseBody: &Version{
				ID:          "10000",
				Description: "Updated description",
				Released:    false,
			},
			wantErr: false,
		},
		{
			name:      "empty version ID",
			versionID: "",
			input:     &UpdateVersionInput{Name: "v1.0.0"},
			wantErr:   true,
			errMsg:    "version ID is required",
		},
		{
			name:      "nil input",
			versionID: "10000",
			input:     nil,
			wantErr:   true,
			errMsg:    "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.versionID == "" || tt.input == nil {
				service := NewService(nil)
				_, err := service.UpdateVersion(context.Background(), tt.versionID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/version/"+tt.versionID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			version, err := service.UpdateVersion(context.Background(), tt.versionID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, version)
				if tt.input.Released != nil {
					assert.Equal(t, *tt.input.Released, version.Released)
				}
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name           string
		versionID      string
		responseStatus int
		responseBody   *Version
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			versionID:      "10000",
			responseStatus: http.StatusOK,
			responseBody: &Version{
				ID:          "10000",
				Name:        "v1.0.0",
				Description: "First release",
				Released:    true,
			},
			wantErr: false,
		},
		{
			name:      "empty version ID",
			versionID: "",
			wantErr:   true,
			errMsg:    "version ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.versionID == "" {
				service := NewService(nil)
				_, err := service.GetVersion(context.Background(), tt.versionID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/version/"+tt.versionID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			version, err := service.GetVersion(context.Background(), tt.versionID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, version)
				assert.Equal(t, tt.responseBody.ID, version.ID)
				assert.Equal(t, tt.responseBody.Name, version.Name)
			}
		})
	}
}

func TestDeleteVersion(t *testing.T) {
	tests := []struct {
		name           string
		versionID      string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			versionID:      "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:      "empty version ID",
			versionID: "",
			wantErr:   true,
			errMsg:    "version ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.versionID == "" {
				service := NewService(nil)
				err := service.DeleteVersion(context.Background(), tt.versionID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/version/"+tt.versionID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteVersion(context.Background(), tt.versionID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListProjectVersions(t *testing.T) {
	tests := []struct {
		name           string
		projectKey     string
		responseStatus int
		responseBody   []*Version
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful list",
			projectKey:     "PROJ",
			responseStatus: http.StatusOK,
			responseBody: []*Version{
				{
					ID:       "10000",
					Name:     "v1.0.0",
					Released: true,
				},
				{
					ID:       "10001",
					Name:     "v2.0.0",
					Released: false,
				},
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
				_, err := service.ListProjectVersions(context.Background(), tt.projectKey)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/project/"+tt.projectKey+"/versions")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			versions, err := service.ListProjectVersions(context.Background(), tt.projectKey)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, versions)
				assert.Len(t, versions, len(tt.responseBody))
			}
		})
	}
}
