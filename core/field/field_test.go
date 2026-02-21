package field

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

func TestService_List(t *testing.T) {
	response := []*Field{
		{ID: "customfield_10000", Name: "Story Points", Custom: true},
		{ID: "summary", Name: "Summary", Custom: false},
	}
	data, _ := json.Marshal(response)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(data))),
		},
	}
	s := NewService(mockTransport)

	fields, err := s.List(context.Background())
	if err != nil {
		t.Errorf("List() error = %v", err)
		return
	}
	if len(fields) != 2 {
		t.Errorf("List() returned %d fields, want 2", len(fields))
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Field{
				ID:     "customfield_10000",
				Name:   "Story Points",
				Custom: true,
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			field, err := s.Get(context.Background(), tt.fieldID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && field == nil {
				t.Error("Get() returned nil field")
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateFieldInput
		wantErr bool
	}{
		{
			name: "success",
			input: &CreateFieldInput{
				Name:        "Story Points",
				Type:        "com.atlassian.jira.plugin.system.customfieldtypes:float",
				SearcherKey: "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber",
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
			input: &CreateFieldInput{
				Type:        "com.atlassian.jira.plugin.system.customfieldtypes:float",
				SearcherKey: "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			input: &CreateFieldInput{
				Name:        "Story Points",
				SearcherKey: "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber",
			},
			wantErr: true,
		},
		{
			name: "empty searcher key",
			input: &CreateFieldInput{
				Name: "Story Points",
				Type: "com.atlassian.jira.plugin.system.customfieldtypes:float",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Field{
				ID:     "customfield_10000",
				Name:   "Story Points",
				Custom: true,
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			field, err := s.Create(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && field == nil {
				t.Error("Create() returned nil field")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		input   *UpdateFieldInput
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			input: &UpdateFieldInput{
				Name: "Updated Story Points",
			},
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			input: &UpdateFieldInput{
				Name: "Updated",
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			fieldID: "customfield_10000",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Field{
				ID:     "customfield_10000",
				Name:   "Updated Story Points",
				Custom: true,
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			field, err := s.Update(context.Background(), tt.fieldID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && field == nil {
				t.Error("Update() returned nil field")
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			wantErr: true,
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

			err := s.Delete(context.Background(), tt.fieldID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ListContexts(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := struct {
				Values []*FieldContext `json:"values"`
			}{
				Values: []*FieldContext{
					{ID: "10100", Name: "Default Context"},
					{ID: "10101", Name: "Custom Context"},
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

			contexts, err := s.ListContexts(context.Background(), tt.fieldID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListContexts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(contexts) != 2 {
				t.Errorf("ListContexts() returned %d contexts, want 2", len(contexts))
			}
		})
	}
}

func TestService_CreateContext(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		input   *CreateContextInput
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			input: &CreateContextInput{
				Name: "New Context",
			},
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			input: &CreateContextInput{
				Name: "New Context",
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			fieldID: "customfield_10000",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "empty context name",
			fieldID: "customfield_10000",
			input:   &CreateContextInput{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &FieldContext{
				ID:   "10100",
				Name: "New Context",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			context, err := s.CreateContext(context.Background(), tt.fieldID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && context == nil {
				t.Error("CreateContext() returned nil context")
			}
		})
	}
}

func TestService_UpdateContext(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		input     *UpdateContextInput
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input: &UpdateContextInput{
				Name: "Updated Context",
			},
			wantErr: false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			input: &UpdateContextInput{
				Name: "Updated",
			},
			wantErr: true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			input: &UpdateContextInput{
				Name: "Updated",
			},
			wantErr: true,
		},
		{
			name:      "nil input",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &FieldContext{
				ID:   "10100",
				Name: "Updated Context",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			context, err := s.UpdateContext(context.Background(), tt.fieldID, tt.contextID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && context == nil {
				t.Error("UpdateContext() returned nil context")
			}
		})
	}
}

func TestService_DeleteContext(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			wantErr:   false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			wantErr:   true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			wantErr:   true,
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

			err := s.DeleteContext(context.Background(), tt.fieldID, tt.contextID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ListOptions(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			wantErr:   false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			wantErr:   true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := struct {
				Values []*FieldOption `json:"values"`
			}{
				Values: []*FieldOption{
					{ID: 1, Value: "Option 1"},
					{ID: 2, Value: "Option 2"},
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

			options, err := s.ListOptions(context.Background(), tt.fieldID, tt.contextID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(options) != 2 {
				t.Errorf("ListOptions() returned %d options, want 2", len(options))
			}
		})
	}
}

func TestService_CreateOption(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		input     *CreateOptionInput
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input: &CreateOptionInput{
				Value: "New Option",
			},
			wantErr: false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			input: &CreateOptionInput{
				Value: "New Option",
			},
			wantErr: true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			input: &CreateOptionInput{
				Value: "New Option",
			},
			wantErr: true,
		},
		{
			name:      "nil input",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     nil,
			wantErr:   true,
		},
		{
			name:      "empty value",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     &CreateOptionInput{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &FieldOption{
				ID:    1,
				Value: "New Option",
			}
			data, _ := json.Marshal(response)

			mockTransport := &mockRoundTripper{
				response: &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader(string(data))),
				},
			}
			s := NewService(mockTransport)

			option, err := s.CreateOption(context.Background(), tt.fieldID, tt.contextID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && option == nil {
				t.Error("CreateOption() returned nil option")
			}
		})
	}
}

func TestService_AssociateContextProjects(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		input     *AssociateContextProjectsInput
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input: &AssociateContextProjectsInput{
				ProjectIDs: []string{"10000", "10001"},
			},
			wantErr: false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			input: &AssociateContextProjectsInput{
				ProjectIDs: []string{"10000"},
			},
			wantErr: true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			input: &AssociateContextProjectsInput{
				ProjectIDs: []string{"10000"},
			},
			wantErr: true,
		},
		{
			name:      "nil input",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     nil,
			wantErr:   true,
		},
		{
			name:      "empty project IDs",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     &AssociateContextProjectsInput{},
			wantErr:   true,
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

			err := s.AssociateContextProjects(context.Background(), tt.fieldID, tt.contextID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssociateContextProjects() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_RemoveContextProjects(t *testing.T) {
	tests := []struct {
		name      string
		fieldID   string
		contextID string
		input     *RemoveContextProjectsInput
		wantErr   bool
	}{
		{
			name:      "success",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input: &RemoveContextProjectsInput{
				ProjectIDs: []string{"10000"},
			},
			wantErr: false,
		},
		{
			name:      "empty field ID",
			fieldID:   "",
			contextID: "10100",
			input: &RemoveContextProjectsInput{
				ProjectIDs: []string{"10000"},
			},
			wantErr: true,
		},
		{
			name:      "empty context ID",
			fieldID:   "customfield_10000",
			contextID: "",
			input: &RemoveContextProjectsInput{
				ProjectIDs: []string{"10000"},
			},
			wantErr: true,
		},
		{
			name:      "nil input",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     nil,
			wantErr:   true,
		},
		{
			name:      "empty project IDs",
			fieldID:   "customfield_10000",
			contextID: "10100",
			input:     &RemoveContextProjectsInput{},
			wantErr:   true,
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

			err := s.RemoveContextProjects(context.Background(), tt.fieldID, tt.contextID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveContextProjects() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetContextProjectMappings(t *testing.T) {
	tests := []struct {
		name    string
		fieldID string
		opts    *GetContextProjectMappingsOptions
		wantErr bool
	}{
		{
			name:    "success",
			fieldID: "customfield_10000",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "with context IDs filter",
			fieldID: "customfield_10000",
			opts: &GetContextProjectMappingsOptions{
				ContextIDs: []string{"10100", "10101"},
				MaxResults: 50,
			},
			wantErr: false,
		},
		{
			name:    "empty field ID",
			fieldID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := struct {
				Values []*ContextProjectMapping `json:"values"`
			}{
				Values: []*ContextProjectMapping{
					{ContextID: "10100", ProjectID: "10000", IsGlobal: false},
					{ContextID: "10100", ProjectID: "10001", IsGlobal: false},
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

			mappings, err := s.GetContextProjectMappings(context.Background(), tt.fieldID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContextProjectMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(mappings) != 2 {
				t.Errorf("GetContextProjectMappings() returned %d mappings, want 2", len(mappings))
			}
		})
	}
}
