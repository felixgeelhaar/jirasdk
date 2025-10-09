package filter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper implements RoundTripper for testing.
type mockRoundTripper struct {
	response *http.Response
	err      error
	request  *http.Request
}

func (m *mockRoundTripper) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(data))
	}
	req, err := http.NewRequestWithContext(ctx, method, path, bodyReader)
	if err != nil {
		return nil, err
	}
	m.request = req
	return req, nil
}

func (m *mockRoundTripper) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockRoundTripper) DecodeResponse(resp *http.Response, v interface{}) error {
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name     string
		filterID string
		expand   []string
		response *Filter
		wantErr  bool
	}{
		{
			name:     "success",
			filterID: "12345",
			expand:   []string{"sharePermissions"},
			response: &Filter{
				ID:   "12345",
				Name: "Test Filter",
				JQL:  "project = TEST",
			},
			wantErr: false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var responseBody io.ReadCloser
			if tt.response != nil {
				data, _ := json.Marshal(tt.response)
				responseBody = io.NopCloser(strings.NewReader(string(data)))
			}

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				},
			}
			s := NewService(mockTransport)

			filter, err := s.Get(context.Background(), tt.filterID, tt.expand)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && filter == nil {
				t.Error("Get() returned nil filter")
			}
			if !tt.wantErr && filter.ID != tt.response.ID {
				t.Errorf("Get() filter.ID = %v, want %v", filter.ID, tt.response.ID)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateFilterInput
		wantErr bool
	}{
		{
			name: "success",
			input: &CreateFilterInput{
				Name: "New Filter",
				JQL:  "project = TEST",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "empty name",
			input: &CreateFilterInput{
				JQL: "project = TEST",
			},
			wantErr: true,
		},
		{
			name: "empty JQL",
			input: &CreateFilterInput{
				Name: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Filter{
				ID:   "12345",
				Name: "New Filter",
				JQL:  "project = TEST",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filter, err := s.Create(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && filter == nil {
				t.Error("Create() returned nil filter")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name     string
		filterID string
		input    *UpdateFilterInput
		wantErr  bool
	}{
		{
			name:     "success",
			filterID: "12345",
			input: &UpdateFilterInput{
				Name: "Updated Filter",
				JQL:  "project = UPDATED",
			},
			wantErr: false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			input: &UpdateFilterInput{
				Name: "Test",
			},
			wantErr: true,
		},
		{
			name:     "nil input",
			filterID: "12345",
			input:    nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Filter{
				ID:   "12345",
				Name: "Updated Filter",
				JQL:  "project = UPDATED",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filter, err := s.Update(context.Background(), tt.filterID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && filter == nil {
				t.Error("Update() returned nil filter")
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		filterID string
		wantErr  bool
	}{
		{
			name:     "success",
			filterID: "12345",
			wantErr:  false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			}
			s := NewService(mockTransport)

			err := s.Delete(context.Background(), tt.filterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ListOptions
		wantErr bool
	}{
		{
			name: "success with options",
			opts: &ListOptions{
				StartAt:    0,
				MaxResults: 50,
			},
			wantErr: false,
		},
		{
			name:    "success without options",
			opts:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := struct {
				StartAt    int       `json:"startAt"`
				MaxResults int       `json:"maxResults"`
				Total      int       `json:"total"`
				IsLast     bool      `json:"isLast"`
				Values     []*Filter `json:"values"`
			}{
				StartAt:    0,
				MaxResults: 50,
				Total:      2,
				IsLast:     true,
				Values: []*Filter{
					{ID: "1", Name: "Filter 1"},
					{ID: "2", Name: "Filter 2"},
				},
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filters, err := s.List(context.Background(), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(filters) != 2 {
				t.Errorf("List() returned %d filters, want 2", len(filters))
			}
		})
	}
}

func TestService_GetFavorites(t *testing.T) {
	response := []*Filter{
		{ID: "1", Name: "Favorite 1", Favorite: true},
		{ID: "2", Name: "Favorite 2", Favorite: true},
	}
	data, _ := json.Marshal(response)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(data))),
		},
	}
	s := NewService(mockTransport)

	filters, err := s.GetFavorites(context.Background())
	if err != nil {
		t.Errorf("GetFavorites() error = %v", err)
		return
	}
	if len(filters) != 2 {
		t.Errorf("GetFavorites() returned %d filters, want 2", len(filters))
	}
}

func TestService_GetMyFilters(t *testing.T) {
	tests := []struct {
		name              string
		expand            []string
		includeFavourites bool
		wantErr           bool
	}{
		{
			name:              "success with favorites",
			expand:            []string{"sharePermissions"},
			includeFavourites: true,
			wantErr:           false,
		},
		{
			name:              "success without favorites",
			includeFavourites: false,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := []*Filter{
				{ID: "1", Name: "My Filter 1"},
				{ID: "2", Name: "My Filter 2"},
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filters, err := s.GetMyFilters(context.Background(), tt.expand, tt.includeFavourites)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMyFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(filters) != 2 {
				t.Errorf("GetMyFilters() returned %d filters, want 2", len(filters))
			}
		})
	}
}

func TestService_SetFavorite(t *testing.T) {
	tests := []struct {
		name     string
		filterID string
		wantErr  bool
	}{
		{
			name:     "success",
			filterID: "12345",
			wantErr:  false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Filter{
				ID:       "12345",
				Name:     "Test Filter",
				Favorite: true,
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filter, err := s.SetFavorite(context.Background(), tt.filterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetFavorite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && filter == nil {
				t.Error("SetFavorite() returned nil filter")
			}
		})
	}
}

func TestService_RemoveFavorite(t *testing.T) {
	tests := []struct {
		name     string
		filterID string
		wantErr  bool
	}{
		{
			name:     "success",
			filterID: "12345",
			wantErr:  false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Filter{
				ID:       "12345",
				Name:     "Test Filter",
				Favorite: false,
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			filter, err := s.RemoveFavorite(context.Background(), tt.filterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveFavorite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && filter == nil {
				t.Error("RemoveFavorite() returned nil filter")
			}
		})
	}
}

func TestService_GetDefaultShareScope(t *testing.T) {
	response := struct {
		Scope string `json:"scope"`
	}{
		Scope: "GLOBAL",
	}
	data, _ := json.Marshal(response)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(data))),
		},
	}
	s := NewService(mockTransport)

	scope, err := s.GetDefaultShareScope(context.Background())
	if err != nil {
		t.Errorf("GetDefaultShareScope() error = %v", err)
		return
	}
	if scope == "" {
		t.Error("GetDefaultShareScope() returned empty scope")
	}
	if scope != "GLOBAL" {
		t.Errorf("GetDefaultShareScope() scope = %v, want GLOBAL", scope)
	}
}

func TestService_SetDefaultShareScope(t *testing.T) {
	tests := []struct {
		name    string
		scope   string
		wantErr bool
	}{
		{
			name:    "success GLOBAL",
			scope:   "GLOBAL",
			wantErr: false,
		},
		{
			name:    "success AUTHENTICATED",
			scope:   "AUTHENTICATED",
			wantErr: false,
		},
		{
			name:    "success PRIVATE",
			scope:   "PRIVATE",
			wantErr: false,
		},
		{
			name:    "empty scope",
			scope:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			}
			s := NewService(mockTransport)

			err := s.SetDefaultShareScope(context.Background(), tt.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDefaultShareScope() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetSharePermission(t *testing.T) {
	tests := []struct {
		name         string
		filterID     string
		permissionID int64
		wantErr      bool
	}{
		{
			name:         "success",
			filterID:     "12345",
			permissionID: 67890,
			wantErr:      false,
		},
		{
			name:         "empty filter ID",
			filterID:     "",
			permissionID: 67890,
			wantErr:      true,
		},
		{
			name:         "zero permission ID",
			filterID:     "12345",
			permissionID: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Permission{
				ID:   67890,
				Type: "group",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			perm, err := s.GetSharePermission(context.Background(), tt.filterID, tt.permissionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSharePermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && perm == nil {
				t.Error("GetSharePermission() returned nil permission")
			}
		})
	}
}

func TestService_AddSharePermission(t *testing.T) {
	tests := []struct {
		name       string
		filterID   string
		permission *Permission
		wantErr    bool
	}{
		{
			name:     "success",
			filterID: "12345",
			permission: &Permission{
				Type: "group",
			},
			wantErr: false,
		},
		{
			name:     "empty filter ID",
			filterID: "",
			permission: &Permission{
				Type: "group",
			},
			wantErr: true,
		},
		{
			name:       "nil permission",
			filterID:   "12345",
			permission: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := []*Permission{
				{
					ID:   67890,
					Type: "group",
				},
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			perms, err := s.AddSharePermission(context.Background(), tt.filterID, tt.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddSharePermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && perms == nil {
				t.Error("AddSharePermission() returned nil permissions")
			}
		})
	}
}

func TestService_DeleteSharePermission(t *testing.T) {
	tests := []struct {
		name         string
		filterID     string
		permissionID int64
		wantErr      bool
	}{
		{
			name:         "success",
			filterID:     "12345",
			permissionID: 67890,
			wantErr:      false,
		},
		{
			name:         "empty filter ID",
			filterID:     "",
			permissionID: 67890,
			wantErr:      true,
		},
		{
			name:         "zero permission ID",
			filterID:     "12345",
			permissionID: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			}
			s := NewService(mockTransport)

			err := s.DeleteSharePermission(context.Background(), tt.filterID, tt.permissionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSharePermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
