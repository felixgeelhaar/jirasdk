package agile

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

func TestGetBoards(t *testing.T) {
	tests := []struct {
		name           string
		opts           *BoardsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		checkResult    func(*testing.T, []*Board)
	}{
		{
			name:           "successful get boards",
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"maxResults": 50,
				"startAt":    0,
				"total":      2,
				"isLast":     true,
				"values": []map[string]interface{}{
					{
						"id":   int64(123),
						"name": "Sprint Board",
						"type": "scrum",
					},
					{
						"id":   int64(124),
						"name": "Kanban Board",
						"type": "kanban",
					},
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, boards []*Board) {
				assert.Len(t, boards, 2)
				assert.Equal(t, int64(123), boards[0].ID)
				assert.Equal(t, "Sprint Board", boards[0].Name)
				assert.Equal(t, "scrum", boards[0].Type)
			},
		},
		{
			name: "with filters",
			opts: &BoardsOptions{
				Type:       "scrum",
				MaxResults: 10,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":   int64(123),
						"name": "Sprint Board",
						"type": "scrum",
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
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board")

				if tt.opts != nil {
					if tt.opts.Type != "" {
						assert.Equal(t, tt.opts.Type, r.URL.Query().Get("type"))
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
			boards, err := service.GetBoards(context.Background(), tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, boards)
				if tt.checkResult != nil {
					tt.checkResult(t, boards)
				}
			}
		})
	}
}

func TestGetBoard(t *testing.T) {
	tests := []struct {
		name           string
		boardID        int64
		responseStatus int
		responseBody   *Board
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			boardID:        123,
			responseStatus: http.StatusOK,
			responseBody: &Board{
				ID:   123,
				Name: "Sprint Board",
				Type: "scrum",
			},
			wantErr: false,
		},
		{
			name:    "invalid board ID",
			boardID: 0,
			wantErr: true,
			errMsg:  "board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.boardID <= 0 {
				service := NewService(nil)
				_, err := service.GetBoard(context.Background(), tt.boardID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board/123")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			board, err := service.GetBoard(context.Background(), tt.boardID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, board)
				assert.Equal(t, tt.responseBody.ID, board.ID)
				assert.Equal(t, tt.responseBody.Name, board.Name)
			}
		})
	}
}

func TestCreateBoard(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateBoardInput
		responseStatus int
		responseBody   *Board
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateBoardInput{
				Name:     "New Sprint Board",
				Type:     "scrum",
				FilterID: 10000,
			},
			responseStatus: http.StatusCreated,
			responseBody: &Board{
				ID:   125,
				Name: "New Sprint Board",
				Type: "scrum",
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
			input: &CreateBoardInput{
				Type:     "scrum",
				FilterID: 10000,
			},
			wantErr: true,
			errMsg:  "board name is required",
		},
		{
			name: "missing type",
			input: &CreateBoardInput{
				Name:     "Board",
				FilterID: 10000,
			},
			wantErr: true,
			errMsg:  "board type is required",
		},
		{
			name: "missing filter ID",
			input: &CreateBoardInput{
				Name: "Board",
				Type: "scrum",
			},
			wantErr: true,
			errMsg:  "filter ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Name == "" || tt.input.Type == "" || tt.input.FilterID <= 0 {
				service := NewService(nil)
				_, err := service.CreateBoard(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			board, err := service.CreateBoard(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, board)
				assert.Equal(t, tt.responseBody.ID, board.ID)
			}
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	tests := []struct {
		name           string
		boardID        int64
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			boardID:        123,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:    "invalid board ID",
			boardID: 0,
			wantErr: true,
			errMsg:  "board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.boardID <= 0 {
				service := NewService(nil)
				err := service.DeleteBoard(context.Background(), tt.boardID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board/123")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteBoard(context.Background(), tt.boardID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetBoardSprints(t *testing.T) {
	tests := []struct {
		name           string
		boardID        int64
		opts           *SprintsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get sprints",
			boardID:        123,
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":    int64(456),
						"name":  "Sprint 25",
						"state": "active",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "with options",
			boardID: 123,
			opts: &SprintsOptions{
				State:      "active,future",
				StartAt:    10,
				MaxResults: 50,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":    int64(456),
						"name":  "Sprint 25",
						"state": "active",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid board ID",
			boardID: 0,
			wantErr: true,
			errMsg:  "board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.boardID <= 0 {
				service := NewService(nil)
				_, err := service.GetBoardSprints(context.Background(), tt.boardID, tt.opts)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board/123/sprint")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			sprints, err := service.GetBoardSprints(context.Background(), tt.boardID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sprints)
			}
		})
	}
}

func TestGetSprint(t *testing.T) {
	tests := []struct {
		name           string
		sprintID       int64
		responseStatus int
		responseBody   *Sprint
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			sprintID:       456,
			responseStatus: http.StatusOK,
			responseBody: &Sprint{
				ID:    456,
				Name:  "Sprint 25",
				State: "active",
			},
			wantErr: false,
		},
		{
			name:     "invalid sprint ID",
			sprintID: 0,
			wantErr:  true,
			errMsg:   "sprint ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sprintID <= 0 {
				service := NewService(nil)
				_, err := service.GetSprint(context.Background(), tt.sprintID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/sprint/456")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			sprint, err := service.GetSprint(context.Background(), tt.sprintID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sprint)
				assert.Equal(t, tt.responseBody.ID, sprint.ID)
			}
		})
	}
}

func TestCreateSprint(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateSprintInput
		responseStatus int
		responseBody   *Sprint
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful create",
			input: &CreateSprintInput{
				Name:          "Sprint 26",
				OriginBoardID: 123,
				StartDate:     "2024-06-01T09:00:00.000Z",
				EndDate:       "2024-06-14T17:00:00.000Z",
			},
			responseStatus: http.StatusCreated,
			responseBody: &Sprint{
				ID:    457,
				Name:  "Sprint 26",
				State: "future",
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
			input: &CreateSprintInput{
				OriginBoardID: 123,
			},
			wantErr: true,
			errMsg:  "sprint name is required",
		},
		{
			name: "missing board ID",
			input: &CreateSprintInput{
				Name: "Sprint",
			},
			wantErr: true,
			errMsg:  "origin board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Name == "" || tt.input.OriginBoardID <= 0 {
				service := NewService(nil)
				_, err := service.CreateSprint(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/sprint")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			sprint, err := service.CreateSprint(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sprint)
			}
		})
	}
}

func TestUpdateSprint(t *testing.T) {
	tests := []struct {
		name           string
		sprintID       int64
		input          *UpdateSprintInput
		responseStatus int
		responseBody   *Sprint
		wantErr        bool
		errMsg         string
	}{
		{
			name:     "successful update",
			sprintID: 456,
			input: &UpdateSprintInput{
				State: "active",
				Goal:  "Complete authentication",
			},
			responseStatus: http.StatusOK,
			responseBody: &Sprint{
				ID:    456,
				State: "active",
				Goal:  "Complete authentication",
			},
			wantErr: false,
		},
		{
			name:     "invalid sprint ID",
			sprintID: 0,
			input:    &UpdateSprintInput{},
			wantErr:  true,
			errMsg:   "sprint ID is required",
		},
		{
			name:     "nil input",
			sprintID: 456,
			input:    nil,
			wantErr:  true,
			errMsg:   "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sprintID <= 0 || tt.input == nil {
				service := NewService(nil)
				_, err := service.UpdateSprint(context.Background(), tt.sprintID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/sprint/456")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			sprint, err := service.UpdateSprint(context.Background(), tt.sprintID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sprint)
			}
		})
	}
}

func TestDeleteSprint(t *testing.T) {
	tests := []struct {
		name           string
		sprintID       int64
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			sprintID:       456,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "invalid sprint ID",
			sprintID: 0,
			wantErr:  true,
			errMsg:   "sprint ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sprintID <= 0 {
				service := NewService(nil)
				err := service.DeleteSprint(context.Background(), tt.sprintID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/sprint/456")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteSprint(context.Background(), tt.sprintID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetBoardEpics(t *testing.T) {
	doneTrue := true

	tests := []struct {
		name           string
		boardID        int64
		opts           *EpicsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get epics",
			boardID:        123,
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":      int64(789),
						"key":     "PROJ-100",
						"name":    "User Management",
						"summary": "Epic for user management features",
						"done":    false,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "with done filter",
			boardID: 123,
			opts: &EpicsOptions{
				Done:       &doneTrue,
				StartAt:    0,
				MaxResults: 25,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"values": []map[string]interface{}{
					{
						"id":   int64(789),
						"done": true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid board ID",
			boardID: 0,
			wantErr: true,
			errMsg:  "board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.boardID <= 0 {
				service := NewService(nil)
				_, err := service.GetBoardEpics(context.Background(), tt.boardID, nil)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board/123/epic")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			epics, err := service.GetBoardEpics(context.Background(), tt.boardID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, epics)
			}
		})
	}
}

func TestGetEpic(t *testing.T) {
	tests := []struct {
		name           string
		epicID         int64
		responseStatus int
		responseBody   *Epic
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			epicID:         789,
			responseStatus: http.StatusOK,
			responseBody: &Epic{
				ID:      789,
				Key:     "PROJ-100",
				Name:    "User Management",
				Summary: "Epic for user management features",
				Done:    false,
			},
			wantErr: false,
		},
		{
			name:    "invalid epic ID",
			epicID:  0,
			wantErr: true,
			errMsg:  "epic ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.epicID <= 0 {
				service := NewService(nil)
				_, err := service.GetEpic(context.Background(), tt.epicID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/epic/789")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			epic, err := service.GetEpic(context.Background(), tt.epicID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, epic)
				assert.Equal(t, tt.responseBody.ID, epic.ID)
			}
		})
	}
}

func TestMoveIssuesToSprint(t *testing.T) {
	tests := []struct {
		name           string
		sprintID       int64
		input          *MoveIssuesToSprintInput
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:     "successful move",
			sprintID: 456,
			input: &MoveIssuesToSprintInput{
				Issues: []string{"PROJ-123", "PROJ-124"},
			},
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:     "invalid sprint ID",
			sprintID: 0,
			input: &MoveIssuesToSprintInput{
				Issues: []string{"PROJ-123"},
			},
			wantErr: true,
			errMsg:  "sprint ID is required",
		},
		{
			name:     "no issues",
			sprintID: 456,
			input:    &MoveIssuesToSprintInput{},
			wantErr:  true,
			errMsg:   "at least one issue is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sprintID <= 0 || tt.input == nil || len(tt.input.Issues) == 0 {
				service := NewService(nil)
				err := service.MoveIssuesToSprint(context.Background(), tt.sprintID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/sprint/456/issue")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.MoveIssuesToSprint(context.Background(), tt.sprintID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetBacklog(t *testing.T) {
	tests := []struct {
		name           string
		boardID        int64
		opts           *BoardsOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get backlog",
			boardID:        123,
			opts:           nil,
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"maxResults": 50,
				"startAt":    0,
				"total":      2,
				"issues": []interface{}{
					map[string]interface{}{"key": "PROJ-123"},
					map[string]interface{}{"key": "PROJ-124"},
				},
			},
			wantErr: false,
		},
		{
			name:    "with pagination",
			boardID: 123,
			opts: &BoardsOptions{
				StartAt:    10,
				MaxResults: 25,
			},
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"issues": []interface{}{
					map[string]interface{}{"key": "PROJ-125"},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid board ID",
			boardID: 0,
			wantErr: true,
			errMsg:  "board ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.boardID <= 0 {
				service := NewService(nil)
				_, err := service.GetBacklog(context.Background(), tt.boardID, nil)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/agile/1.0/board/123/backlog")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			issues, err := service.GetBacklog(context.Background(), tt.boardID, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, issues)
			}
		})
	}
}
