package expression

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

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name           string
		input          *EvaluationInput
		responseStatus int
		responseBody   *EvaluationResult
		wantErr        bool
		errMsg         string
		checkRequest   func(*testing.T, *http.Request)
	}{
		{
			name: "successful evaluation",
			input: &EvaluationInput{
				Expression: "issue.summary",
				Context: map[string]interface{}{
					"issue": map[string]interface{}{
						"key": "PROJ-123",
					},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: "Test Issue Summary",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "expression is required",
		},
		{
			name: "empty expression",
			input: &EvaluationInput{
				Expression: "",
			},
			wantErr: true,
			errMsg:  "expression is required",
		},
		{
			name: "with complexity metadata",
			input: &EvaluationInput{
				Expression: "user.displayName",
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: "John Doe",
				Meta: &EvaluationMeta{
					Complexity: &Complexity{
						Steps:               5,
						ExpensiveOperations: 1,
						Beans:               2,
						PrimitiveValues:     3,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with errors",
			input: &EvaluationInput{
				Expression: "invalid.expression",
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: nil,
				Errors: []*EvaluationError{
					{
						Type:    "syntax",
						Message: "Invalid expression syntax",
						Line:    1,
						Column:  8,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Expression == "" {
				service := NewService(nil)
				_, err := service.Evaluate(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/expression/eval")

				// Verify request body
				var body EvaluationInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Expression, body.Expression)

				if tt.checkRequest != nil {
					tt.checkRequest(t, r)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.Evaluate(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.responseBody.Value, result.Value)
			}
		})
	}
}

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		name           string
		input          *EvaluationInput
		responseStatus int
		responseBody   *EvaluationResult
		wantErr        bool
		errMsg         string
		checkRequest   func(*testing.T, *http.Request)
	}{
		{
			name: "successful evaluation with new endpoint",
			input: &EvaluationInput{
				Expression: "issue.summary",
				Context: map[string]interface{}{
					"issue": map[string]interface{}{
						"key": "PROJ-123",
					},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: "Test Issue Summary",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "expression is required",
		},
		{
			name: "empty expression",
			input: &EvaluationInput{
				Expression: "",
			},
			wantErr: true,
			errMsg:  "expression is required",
		},
		{
			name: "complex expression with metadata",
			input: &EvaluationInput{
				Expression: "issue.fields.status.name + ' - ' + issue.fields.priority.name",
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: "Open - High",
				Meta: &EvaluationMeta{
					Complexity: &Complexity{
						Steps:               10,
						ExpensiveOperations: 2,
						Beans:               4,
						PrimitiveValues:     6,
					},
					Issues: []string{"PROJ-123"},
				},
			},
			wantErr: false,
		},
		{
			name: "evaluation with errors",
			input: &EvaluationInput{
				Expression: "invalid.field.access",
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: nil,
				Errors: []*EvaluationError{
					{
						Type:    "type",
						Message: "Property 'field' not found",
						Line:    1,
						Column:  8,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "verify new endpoint path",
			input: &EvaluationInput{
				Expression: "user.accountId",
			},
			responseStatus: http.StatusOK,
			responseBody: &EvaluationResult{
				Value: "abc123",
			},
			wantErr: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				// Verify it's using the NEW endpoint (exactly "/evaluate", not "/eval")
				assert.Equal(t, "/rest/api/3/expression/evaluate", r.URL.Path)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Expression == "" {
				service := NewService(nil)
				_, err := service.EvaluateExpression(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/expression/evaluate")

				// Verify request body
				var body EvaluationInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Expression, body.Expression)

				if tt.checkRequest != nil {
					tt.checkRequest(t, r)
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.EvaluateExpression(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.responseBody.Value, result.Value)
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name           string
		input          *AnalysisInput
		responseStatus int
		responseBody   *AnalysisResult
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful analysis",
			input: &AnalysisInput{
				Expressions: []string{"issue.summary", "user.displayName"},
			},
			responseStatus: http.StatusOK,
			responseBody: &AnalysisResult{
				Results: []*ExpressionAnalysis{
					{
						Expression: "issue.summary",
						Valid:      true,
						Type:       "String",
					},
					{
						Expression: "user.displayName",
						Valid:      true,
						Type:       "String",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "at least one expression is required",
		},
		{
			name: "empty expressions",
			input: &AnalysisInput{
				Expressions: []string{},
			},
			wantErr: true,
			errMsg:  "at least one expression is required",
		},
		{
			name: "analysis with errors",
			input: &AnalysisInput{
				Expressions: []string{"invalid.expression", "valid.expression"},
			},
			responseStatus: http.StatusOK,
			responseBody: &AnalysisResult{
				Results: []*ExpressionAnalysis{
					{
						Expression: "invalid.expression",
						Valid:      false,
						Errors: []*EvaluationError{
							{
								Type:    "syntax",
								Message: "Invalid syntax",
							},
						},
					},
					{
						Expression: "valid.expression",
						Valid:      true,
						Type:       "Any",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "analysis with complexity",
			input: &AnalysisInput{
				Expressions: []string{"issue.fields.status.name"},
			},
			responseStatus: http.StatusOK,
			responseBody: &AnalysisResult{
				Results: []*ExpressionAnalysis{
					{
						Expression: "issue.fields.status.name",
						Valid:      true,
						Type:       "String",
						Complexity: &Complexity{
							Steps:               8,
							ExpensiveOperations: 1,
							Beans:               3,
							PrimitiveValues:     5,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || len(tt.input.Expressions) == 0 {
				service := NewService(nil)
				_, err := service.Analyze(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/expression/analyse")

				// Verify request body
				var body AnalysisInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Expressions, body.Expressions)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			result, err := service.Analyze(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, len(tt.responseBody.Results), len(result.Results))
			}
		})
	}
}
