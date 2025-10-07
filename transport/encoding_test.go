package transport

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeJSONRequest(t *testing.T) {
	tests := []struct {
		name    string
		body    interface{}
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil body",
			body:    nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "valid struct",
			body: map[string]string{
				"key": "value",
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "complex nested struct",
			body: map[string]interface{}{
				"fields": map[string]interface{}{
					"summary": "Test",
					"labels":  []string{"bug", "urgent"},
				},
			},
			wantNil: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := EncodeJSONRequest(tt.body)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, reader)
				} else {
					assert.NotNil(t, reader)
					// Verify we can read the encoded data
					data, err := io.ReadAll(reader)
					require.NoError(t, err)
					assert.NotEmpty(t, data)
				}
			}
		})
	}
}

func TestDecodeJSONResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		target     interface{}
		wantErr    bool
		check      func(*testing.T, interface{})
	}{
		{
			name:       "successful decode",
			statusCode: http.StatusOK,
			body:       `{"key":"PROJ-123","summary":"Test"}`,
			target:     &map[string]string{},
			wantErr:    false,
			check: func(t *testing.T, target interface{}) {
				data := target.(*map[string]string)
				assert.Equal(t, "PROJ-123", (*data)["key"])
				assert.Equal(t, "Test", (*data)["summary"])
			},
		},
		{
			name:       "nil target",
			statusCode: http.StatusOK,
			body:       `{"key":"value"}`,
			target:     nil,
			wantErr:    false,
		},
		{
			name:       "400 error with message",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Bad request"}`,
			target:     &map[string]string{},
			wantErr:    true,
		},
		{
			name:       "404 error with errorMessages",
			statusCode: http.StatusNotFound,
			body:       `{"errorMessages":["Issue not found"]}`,
			target:     &map[string]string{},
			wantErr:    true,
		},
		{
			name:       "500 error with errors map",
			statusCode: http.StatusInternalServerError,
			body:       `{"errors":{"field":"Invalid value"}}`,
			target:     &map[string]string{},
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			statusCode: http.StatusOK,
			body:       `{invalid json}`,
			target:     &map[string]string{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
			}

			err := DecodeJSONResponse(resp, tt.target)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, tt.target)
				}
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name     string
		errResp  *ErrorResponse
		expected string
	}{
		{
			name: "with message",
			errResp: &ErrorResponse{
				StatusCode: http.StatusBadRequest,
				Message:    "Bad request",
			},
			expected: "Jira API error (HTTP 400): Bad request",
		},
		{
			name: "with errorMessages",
			errResp: &ErrorResponse{
				StatusCode:    http.StatusNotFound,
				ErrorMessages: []string{"Issue not found", "Other error"},
			},
			expected: "Jira API error (HTTP 404): Issue not found",
		},
		{
			name: "with errors map",
			errResp: &ErrorResponse{
				StatusCode: http.StatusBadRequest,
				Errors: map[string]string{
					"field": "Invalid value",
				},
			},
			expected: "Jira API error (HTTP 400): field: Invalid value",
		},
		{
			name: "no details",
			errResp: &ErrorResponse{
				StatusCode: http.StatusInternalServerError,
			},
			expected: "Jira API error (HTTP 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errResp.Error()
			assert.Contains(t, result, "Jira API error")
			assert.Contains(t, result, "HTTP")
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		check      func(*testing.T, error)
	}{
		{
			name:       "valid error JSON",
			statusCode: http.StatusBadRequest,
			body:       `{"errorMessages":["Error 1"],"errors":{"field":"Error"}}`,
			check: func(t *testing.T, err error) {
				errResp, ok := err.(*ErrorResponse)
				require.True(t, ok)
				assert.Equal(t, http.StatusBadRequest, errResp.StatusCode)
				assert.Len(t, errResp.ErrorMessages, 1)
				assert.Len(t, errResp.Errors, 1)
			},
		},
		{
			name:       "invalid JSON falls back to raw message",
			statusCode: http.StatusInternalServerError,
			body:       `Invalid JSON`,
			check: func(t *testing.T, err error) {
				errResp, ok := err.(*ErrorResponse)
				require.True(t, ok)
				assert.Equal(t, http.StatusInternalServerError, errResp.StatusCode)
				assert.Equal(t, "Invalid JSON", errResp.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseErrorResponse(tt.statusCode, []byte(tt.body))
			require.Error(t, err)
			if tt.check != nil {
				tt.check(t, err)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "404 error",
			err:      &ErrorResponse{StatusCode: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "400 error",
			err:      &ErrorResponse{StatusCode: http.StatusBadRequest},
			expected: false,
		},
		{
			name:     "non-ErrorResponse error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "401 error",
			err:      &ErrorResponse{StatusCode: http.StatusUnauthorized},
			expected: true,
		},
		{
			name:     "403 error",
			err:      &ErrorResponse{StatusCode: http.StatusForbidden},
			expected: false,
		},
		{
			name:     "non-ErrorResponse error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorized(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "403 error",
			err:      &ErrorResponse{StatusCode: http.StatusForbidden},
			expected: true,
		},
		{
			name:     "401 error",
			err:      &ErrorResponse{StatusCode: http.StatusUnauthorized},
			expected: false,
		},
		{
			name:     "non-ErrorResponse error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForbidden(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "429 error",
			err:      &ErrorResponse{StatusCode: http.StatusTooManyRequests},
			expected: true,
		},
		{
			name:     "500 error",
			err:      &ErrorResponse{StatusCode: http.StatusInternalServerError},
			expected: false,
		},
		{
			name:     "non-ErrorResponse error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRateLimited(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
