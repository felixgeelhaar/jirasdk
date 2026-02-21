package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWorkflowScheme(t *testing.T) {
	tests := []struct {
		name           string
		schemeID       int64
		responseStatus int
		responseBody   *WorkflowScheme
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get workflow scheme",
			schemeID:       10000,
			responseStatus: http.StatusOK,
			responseBody: &WorkflowScheme{
				ID:              10000,
				Name:            "Default Workflow Scheme",
				Description:     "The default workflow scheme",
				DefaultWorkflow: "classic-default-workflow",
				IssueTypeMappings: map[string]string{
					"10001": "software-development-workflow",
					"10002": "bug-workflow",
				},
				Draft: false,
			},
			wantErr: false,
		},
		{
			name:     "invalid scheme ID",
			schemeID: 0,
			wantErr:  true,
			errMsg:   "workflow scheme ID is required",
		},
		{
			name:     "negative scheme ID",
			schemeID: -1,
			wantErr:  true,
			errMsg:   "workflow scheme ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schemeID <= 0 {
				service := NewService(nil)
				_, err := service.GetWorkflowScheme(context.Background(), tt.schemeID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme/")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.GetWorkflowScheme(context.Background(), tt.schemeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scheme)
				assert.Equal(t, tt.responseBody.ID, scheme.ID)
				assert.Equal(t, tt.responseBody.Name, scheme.Name)
				assert.Equal(t, tt.responseBody.DefaultWorkflow, scheme.DefaultWorkflow)
			}
		})
	}
}

func TestListWorkflowSchemes(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListWorkflowSchemesOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		checkResult    func(*testing.T, []*WorkflowScheme)
	}{
		{
			name:           "successful list",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":              10000,
						"name":            "Default Workflow Scheme",
						"defaultWorkflow": "classic-default-workflow",
						"draft":           false,
					},
					{
						"id":              10001,
						"name":            "Custom Workflow Scheme",
						"defaultWorkflow": "software-workflow",
						"draft":           true,
					},
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, schemes []*WorkflowScheme) {
				assert.Len(t, schemes, 2)
				assert.Equal(t, int64(10000), schemes[0].ID)
				assert.Equal(t, "Default Workflow Scheme", schemes[0].Name)
				assert.False(t, schemes[0].Draft)
				assert.Equal(t, int64(10001), schemes[1].ID)
				assert.True(t, schemes[1].Draft)
			},
		},
		{
			name: "with pagination",
			opts: &ListWorkflowSchemesOptions{
				StartAt:    10,
				MaxResults: 50,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":   10000,
						"name": "Default Workflow Scheme",
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
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme")

				if tt.opts != nil {
					if tt.opts.StartAt > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("startAt"))
					}
					if tt.opts.MaxResults > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("maxResults"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			schemes, err := service.ListWorkflowSchemes(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, schemes)
				if tt.checkResult != nil {
					tt.checkResult(t, schemes)
				}
			}
		})
	}
}

func TestCreateWorkflowScheme(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateWorkflowSchemeInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			input: &CreateWorkflowSchemeInput{
				Name:            "My Workflow Scheme",
				Description:     "Custom workflow scheme",
				DefaultWorkflow: "classic-default-workflow",
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
			input: &CreateWorkflowSchemeInput{
				DefaultWorkflow: "classic-default-workflow",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				_, err := service.CreateWorkflowScheme(context.Background(), tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/rest/api/3/workflowscheme", r.URL.Path)

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(&WorkflowScheme{
					ID:   10010,
					Name: tt.input.Name,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.CreateWorkflowScheme(context.Background(), tt.input)

			require.NoError(t, err)
			require.NotNil(t, scheme)
			assert.Equal(t, tt.input.Name, scheme.Name)
		})
	}
}

func TestUpdateWorkflowScheme(t *testing.T) {
	tests := []struct {
		name     string
		schemeID int64
		input    *UpdateWorkflowSchemeInput
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: 10000,
			input: &UpdateWorkflowSchemeInput{
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:     "invalid ID",
			schemeID: 0,
			input:    &UpdateWorkflowSchemeInput{Description: "Updated"},
			wantErr:  true,
			errMsg:   "workflow scheme ID is required",
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
			if tt.wantErr {
				service := NewService(nil)
				_, err := service.UpdateWorkflowScheme(context.Background(), tt.schemeID, tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme/")

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&WorkflowScheme{
					ID:          tt.schemeID,
					Name:        "Updated Scheme",
					Description: tt.input.Description,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			scheme, err := service.UpdateWorkflowScheme(context.Background(), tt.schemeID, tt.input)

			require.NoError(t, err)
			require.NotNil(t, scheme)
		})
	}
}

func TestDeleteWorkflowScheme(t *testing.T) {
	tests := []struct {
		name     string
		schemeID int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: 10000,
			wantErr:  false,
		},
		{
			name:     "invalid ID",
			schemeID: 0,
			wantErr:  true,
			errMsg:   "workflow scheme ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.DeleteWorkflowScheme(context.Background(), tt.schemeID)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme/")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteWorkflowScheme(context.Background(), tt.schemeID)

			require.NoError(t, err)
		})
	}
}

func TestSetWorkflowSchemeIssueType(t *testing.T) {
	tests := []struct {
		name     string
		schemeID int64
		input    *WorkflowSchemeIssueType
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success",
			schemeID: 10000,
			input: &WorkflowSchemeIssueType{
				IssueType: "10001",
				Workflow:   "software-simplified-workflow",
			},
			wantErr: false,
		},
		{
			name:     "invalid ID",
			schemeID: 0,
			input: &WorkflowSchemeIssueType{
				IssueType: "10001",
				Workflow:   "software-simplified-workflow",
			},
			wantErr: true,
			errMsg:  "workflow scheme ID is required",
		},
		{
			name:     "nil input",
			schemeID: 10000,
			input:    nil,
			wantErr:  true,
			errMsg:   "issue type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.SetWorkflowSchemeIssueType(context.Background(), tt.schemeID, tt.input)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme/")
				assert.Contains(t, r.URL.Path, "/issuetype/")

				w.WriteHeader(http.StatusOK)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.SetWorkflowSchemeIssueType(context.Background(), tt.schemeID, tt.input)

			require.NoError(t, err)
		})
	}
}

func TestDeleteWorkflowSchemeIssueType(t *testing.T) {
	tests := []struct {
		name      string
		schemeID  int64
		issueType string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success",
			schemeID:  10000,
			issueType: "10001",
			wantErr:   false,
		},
		{
			name:      "invalid ID",
			schemeID:  0,
			issueType: "10001",
			wantErr:   true,
			errMsg:    "workflow scheme ID is required",
		},
		{
			name:      "empty issue type",
			schemeID:  10000,
			issueType: "",
			wantErr:   true,
			errMsg:    "issue type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				service := NewService(nil)
				err := service.DeleteWorkflowSchemeIssueType(context.Background(), tt.schemeID, tt.issueType)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/workflowscheme/")
				assert.Contains(t, r.URL.Path, "/issuetype/")

				w.WriteHeader(http.StatusNoContent)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteWorkflowSchemeIssueType(context.Background(), tt.schemeID, tt.issueType)

			require.NoError(t, err)
		})
	}
}
