package issue

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddWorklog(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		issueKeyOrID   string
		input          *AddWorklogInput
		responseStatus int
		responseBody   *Worklog
		wantErr        bool
		errMsg         string
	}{
		{
			name:         "successful worklog with time spent string",
			issueKeyOrID: "PROJ-123",
			input: &AddWorklogInput{
				TimeSpent: "3h 20m",
				Started:   &now,
				Comment:   "Implemented feature",
			},
			responseStatus: http.StatusCreated,
			responseBody: &Worklog{
				ID:               "10000",
				TimeSpent:        "3h 20m",
				TimeSpentSeconds: 12000,
				Comment:          "Implemented feature",
			},
			wantErr: false,
		},
		{
			name:         "successful worklog with seconds",
			issueKeyOrID: "PROJ-123",
			input: &AddWorklogInput{
				TimeSpentSeconds: 7200, // 2 hours
				Started:          &now,
			},
			responseStatus: http.StatusCreated,
			responseBody: &Worklog{
				ID:               "10001",
				TimeSpentSeconds: 7200,
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			input: &AddWorklogInput{
				TimeSpent: "1h",
			},
			wantErr: true,
			errMsg:  "issue key or ID is required",
		},
		{
			name:         "nil input",
			issueKeyOrID: "PROJ-123",
			input:        nil,
			wantErr:      true,
			errMsg:       "worklog input is required",
		},
		{
			name:         "missing time",
			issueKeyOrID: "PROJ-123",
			input: &AddWorklogInput{
				Comment: "No time specified",
			},
			wantErr: true,
			errMsg:  "either timeSpent or timeSpentSeconds is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.input == nil || (tt.input.TimeSpent == "" && tt.input.TimeSpentSeconds == 0) {
				service := NewService(nil)
				_, err := service.AddWorklog(context.Background(), tt.issueKeyOrID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/worklog")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			worklog, err := service.AddWorklog(context.Background(), tt.issueKeyOrID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, worklog)
				assert.Equal(t, tt.responseBody.ID, worklog.ID)
			}
		})
	}
}

func TestGetWorklog(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		worklogID      string
		responseStatus int
		responseBody   *Worklog
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			issueKeyOrID:   "PROJ-123",
			worklogID:      "10000",
			responseStatus: http.StatusOK,
			responseBody: &Worklog{
				ID:               "10000",
				TimeSpent:        "2h",
				TimeSpentSeconds: 7200,
				Comment:          "Development work",
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			worklogID:    "10000",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty worklog ID",
			issueKeyOrID: "PROJ-123",
			worklogID:    "",
			wantErr:      true,
			errMsg:       "worklog ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.worklogID == "" {
				service := NewService(nil)
				_, err := service.GetWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/worklog/"+tt.worklogID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			worklog, err := service.GetWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, worklog)
				assert.Equal(t, tt.responseBody.ID, worklog.ID)
				assert.Equal(t, tt.responseBody.TimeSpent, worklog.TimeSpent)
			}
		})
	}
}

func TestListWorklogs(t *testing.T) {
	worklogs := []*Worklog{
		{
			ID:               "10000",
			TimeSpent:        "2h",
			TimeSpentSeconds: 7200,
		},
		{
			ID:               "10001",
			TimeSpent:        "1h 30m",
			TimeSpentSeconds: 5400,
		},
	}

	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful list",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusOK,
			wantErr:        false,
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
				_, err := service.ListWorklogs(context.Background(), tt.issueKeyOrID, nil)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/worklog")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"worklogs": worklogs,
				})
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.ListWorklogs(context.Background(), tt.issueKeyOrID, nil)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, 2)
				assert.Equal(t, "10000", result[0].ID)
				assert.Equal(t, "10001", result[1].ID)
			}
		})
	}
}

func TestUpdateWorklog(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		worklogID      string
		input          *UpdateWorklogInput
		responseStatus int
		responseBody   *Worklog
		wantErr        bool
		errMsg         string
	}{
		{
			name:         "successful update",
			issueKeyOrID: "PROJ-123",
			worklogID:    "10000",
			input: &UpdateWorklogInput{
				TimeSpent: "4h",
				Comment:   "Updated estimate",
			},
			responseStatus: http.StatusOK,
			responseBody: &Worklog{
				ID:        "10000",
				TimeSpent: "4h",
				Comment:   "Updated estimate",
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			worklogID:    "10000",
			input:        &UpdateWorklogInput{TimeSpent: "1h"},
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty worklog ID",
			issueKeyOrID: "PROJ-123",
			worklogID:    "",
			input:        &UpdateWorklogInput{TimeSpent: "1h"},
			wantErr:      true,
			errMsg:       "worklog ID is required",
		},
		{
			name:         "nil input",
			issueKeyOrID: "PROJ-123",
			worklogID:    "10000",
			input:        nil,
			wantErr:      true,
			errMsg:       "update worklog input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.worklogID == "" || tt.input == nil {
				service := NewService(nil)
				_, err := service.UpdateWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/worklog/"+tt.worklogID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			worklog, err := service.UpdateWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, worklog)
				assert.Equal(t, tt.responseBody.ID, worklog.ID)
			}
		})
	}
}

func TestDeleteWorklog(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		worklogID      string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			issueKeyOrID:   "PROJ-123",
			worklogID:      "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			worklogID:    "10000",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty worklog ID",
			issueKeyOrID: "PROJ-123",
			worklogID:    "",
			wantErr:      true,
			errMsg:       "worklog ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.worklogID == "" {
				service := NewService(nil)
				err := service.DeleteWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/worklog/"+tt.worklogID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteWorklog(context.Background(), tt.issueKeyOrID, tt.worklogID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{0, "0m"},
		{60, "1m"},
		{3600, "1h "},
		{7200, "2h "},
		{12000, "3h 20m"},
		{86400, "1d "},
		{604800, "1w "},
		{694800, "1w 1d 1h "},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}
