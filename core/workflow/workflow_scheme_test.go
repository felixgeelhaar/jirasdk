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
