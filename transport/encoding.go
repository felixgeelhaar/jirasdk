package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EncodeJSONRequest encodes a request body as JSON.
func EncodeJSONRequest(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	return buf, nil
}

// DecodeJSONResponse decodes a JSON response into the target struct.
func DecodeJSONResponse(resp *http.Response, target interface{}) error {
	if target == nil {
		return nil
	}

	defer resp.Body.Close()

	// Read response body for error handling
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, body)
	}

	// Decode successful response
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// ErrorResponse represents a Jira API error response.
type ErrorResponse struct {
	StatusCode int                    `json:"-"`
	ErrorMessages []string            `json:"errorMessages,omitempty"`
	Errors        map[string]string   `json:"errors,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Jira API error (HTTP %d): %s", e.StatusCode, e.Message)
	}

	if len(e.ErrorMessages) > 0 {
		return fmt.Sprintf("Jira API error (HTTP %d): %s", e.StatusCode, e.ErrorMessages[0])
	}

	if len(e.Errors) > 0 {
		for field, msg := range e.Errors {
			return fmt.Sprintf("Jira API error (HTTP %d): %s: %s", e.StatusCode, field, msg)
		}
	}

	return fmt.Sprintf("Jira API error (HTTP %d)", e.StatusCode)
}

// parseErrorResponse attempts to parse a Jira error response.
func parseErrorResponse(statusCode int, body []byte) error {
	errResp := &ErrorResponse{
		StatusCode: statusCode,
	}

	// Try to unmarshal as Jira error format
	if err := json.Unmarshal(body, errResp); err != nil {
		// If parsing fails, use raw body as message
		errResp.Message = string(body)
	}

	return errResp
}

// IsNotFound returns true if the error is a 404 Not Found.
func IsNotFound(err error) bool {
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized.
func IsUnauthorized(err error) bool {
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden returns true if the error is a 403 Forbidden.
func IsForbidden(err error) bool {
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp.StatusCode == http.StatusForbidden
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp.StatusCode == http.StatusTooManyRequests
	}
	return false
}
