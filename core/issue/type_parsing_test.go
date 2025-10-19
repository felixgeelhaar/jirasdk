package issue

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNumericTypeFlexibility tests Go's JSON unmarshaler behavior with numeric types.
// This verifies whether Jira returning numbers as strings would cause issues.
func TestNumericTypeFlexibility(t *testing.T) {
	t.Run("int64 from actual number", func(t *testing.T) {
		type TestStruct struct {
			Value int64 `json:"value"`
		}

		jsonData := `{"value": 12000}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.Equal(t, int64(12000), result.Value)
	})

	t.Run("int64 from string number", func(t *testing.T) {
		type TestStruct struct {
			Value int64 `json:"value"`
		}

		// This will FAIL with default Go JSON unmarshaler
		jsonData := `{"value": "12000"}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		// Go's default JSON decoder does NOT accept string numbers for numeric fields
		if err != nil {
			t.Logf("Expected behavior: string number causes error: %v", err)
		} else {
			t.Logf("Unexpected: Go accepted string number, got value: %d", result.Value)
		}
	})

	t.Run("int from actual number", func(t *testing.T) {
		type TestStruct struct {
			Value int `json:"value"`
		}

		jsonData := `{"value": 42}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("int from string number", func(t *testing.T) {
		type TestStruct struct {
			Value int `json:"value"`
		}

		jsonData := `{"value": "42"}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		if err != nil {
			t.Logf("Expected behavior: string number causes error: %v", err)
		} else {
			t.Logf("Unexpected: Go accepted string number, got value: %d", result.Value)
		}
	})
}

// TestBooleanTypeFlexibility tests Go's JSON unmarshaler behavior with boolean types.
func TestBooleanTypeFlexibility(t *testing.T) {
	t.Run("bool from actual boolean true", func(t *testing.T) {
		type TestStruct struct {
			Value bool `json:"value"`
		}

		jsonData := `{"value": true}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.True(t, result.Value)
	})

	t.Run("bool from actual boolean false", func(t *testing.T) {
		type TestStruct struct {
			Value bool `json:"value"`
		}

		jsonData := `{"value": false}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.False(t, result.Value)
	})

	t.Run("bool from string true", func(t *testing.T) {
		type TestStruct struct {
			Value bool `json:"value"`
		}

		// This will FAIL with default Go JSON unmarshaler
		jsonData := `{"value": "true"}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		if err != nil {
			t.Logf("Expected behavior: string boolean causes error: %v", err)
		} else {
			t.Logf("Unexpected: Go accepted string boolean, got value: %v", result.Value)
		}
	})

	t.Run("bool from numeric 1", func(t *testing.T) {
		type TestStruct struct {
			Value bool `json:"value"`
		}

		jsonData := `{"value": 1}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		if err != nil {
			t.Logf("Expected behavior: numeric 1 causes error: %v", err)
		} else {
			t.Logf("Unexpected: Go accepted numeric 1 as bool, got value: %v", result.Value)
		}
	})

	t.Run("bool from numeric 0", func(t *testing.T) {
		type TestStruct struct {
			Value bool `json:"value"`
		}

		jsonData := `{"value": 0}`
		var result TestStruct
		err := json.Unmarshal([]byte(jsonData), &result)

		if err != nil {
			t.Logf("Expected behavior: numeric 0 causes error: %v", err)
		} else {
			t.Logf("Unexpected: Go accepted numeric 0 as bool, got value: %v", result.Value)
		}
	})
}

// TestWorklogTimeSpentSeconds specifically tests the TimeSpentSeconds field
// from the Worklog struct to ensure it handles Jira's format.
func TestWorklogTimeSpentSeconds(t *testing.T) {
	t.Run("TimeSpentSeconds as number", func(t *testing.T) {
		jsonData := `{
			"id": "10000",
			"timeSpentSeconds": 12000,
			"comment": "Test worklog"
		}`

		var worklog Worklog
		err := json.Unmarshal([]byte(jsonData), &worklog)

		require.NoError(t, err)
		assert.Equal(t, int64(12000), worklog.TimeSpentSeconds)
	})

	t.Run("TimeSpentSeconds as string", func(t *testing.T) {
		jsonData := `{
			"id": "10000",
			"timeSpentSeconds": "12000",
			"comment": "Test worklog"
		}`

		var worklog Worklog
		err := json.Unmarshal([]byte(jsonData), &worklog)

		// If Jira returns this as a string, this will fail
		if err != nil {
			t.Logf("String number in TimeSpentSeconds causes error: %v", err)
			t.Logf("This would need custom unmarshaling if Jira returns it this way")
		} else {
			t.Logf("Go accepted string number, got value: %d", worklog.TimeSpentSeconds)
		}
	})
}
