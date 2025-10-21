package issue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomFields_SetAndGetString(t *testing.T) {
	cf := NewCustomFields()

	cf.SetString("customfield_10001", "Sprint 1")

	value, ok := cf.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", value)

	// Test non-existent field
	_, ok = cf.GetString("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetNumber(t *testing.T) {
	cf := NewCustomFields()

	cf.SetNumber("customfield_10002", 42.5)

	value, ok := cf.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, value)

	// Test non-existent field
	_, ok = cf.GetNumber("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetDate(t *testing.T) {
	cf := NewCustomFields()
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	cf.SetDate("customfield_10003", expectedDate)

	value, ok := cf.GetDate("customfield_10003")
	assert.True(t, ok)
	assert.Equal(t, expectedDate.Format("2006-01-02"), value.Format("2006-01-02"))

	// Test non-existent field
	_, ok = cf.GetDate("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetDateTime(t *testing.T) {
	cf := NewCustomFields()
	expectedTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	cf.SetDateTime("customfield_10004", expectedTime)

	value, ok := cf.GetDateTime("customfield_10004")
	assert.True(t, ok)
	assert.Equal(t, expectedTime.Unix(), value.Unix())

	// Test non-existent field
	_, ok = cf.GetDateTime("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetUser(t *testing.T) {
	cf := NewCustomFields()

	cf.SetUser("customfield_10005", "accountId123")

	accountID, ok := cf.GetUser("customfield_10005")
	assert.True(t, ok)
	assert.Equal(t, "accountId123", accountID)

	// Test non-existent field
	_, ok = cf.GetUser("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetSelect(t *testing.T) {
	cf := NewCustomFields()

	cf.SetSelect("customfield_10006", "High")

	value, ok := cf.GetSelect("customfield_10006")
	assert.True(t, ok)
	assert.Equal(t, "High", value)

	// Test non-existent field
	_, ok = cf.GetSelect("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetMultiSelect(t *testing.T) {
	cf := NewCustomFields()
	expected := []string{"option1", "option2", "option3"}

	cf.SetMultiSelect("customfield_10007", expected)

	values, ok := cf.GetMultiSelect("customfield_10007")
	assert.True(t, ok)
	assert.Equal(t, expected, values)

	// Test non-existent field
	_, ok = cf.GetMultiSelect("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetLabels(t *testing.T) {
	cf := NewCustomFields()
	expected := []string{"bug", "urgent", "frontend"}

	cf.SetLabels("customfield_10008", expected)

	labels, ok := cf.GetLabels("customfield_10008")
	assert.True(t, ok)
	assert.Equal(t, expected, labels)

	// Test non-existent field
	_, ok = cf.GetLabels("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_SetAndGetRaw(t *testing.T) {
	cf := NewCustomFields()
	expected := map[string]interface{}{
		"nested": "structure",
		"count":  42,
	}

	cf.SetRaw("customfield_10009", expected)

	value, ok := cf.GetRaw("customfield_10009")
	assert.True(t, ok)
	assert.Equal(t, expected, value)

	// Test non-existent field
	_, ok = cf.GetRaw("customfield_99999")
	assert.False(t, ok)
}

func TestCustomFields_FluentAPI(t *testing.T) {
	cf := NewCustomFields().
		SetString("customfield_10001", "Sprint 1").
		SetNumber("customfield_10002", 42.5).
		SetLabels("customfield_10008", []string{"bug", "urgent"})

	assert.Len(t, cf, 3)

	str, ok := cf.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", str)

	num, ok := cf.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)

	labels, ok := cf.GetLabels("customfield_10008")
	assert.True(t, ok)
	assert.Equal(t, []string{"bug", "urgent"}, labels)
}

func TestCustomFields_MarshalJSON(t *testing.T) {
	cf := NewCustomFields().
		SetString("customfield_10001", "Sprint 1").
		SetNumber("customfield_10002", 42.5)

	data, err := json.Marshal(cf)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "Sprint 1", result["customfield_10001"])
	assert.Equal(t, 42.5, result["customfield_10002"])
}

func TestCustomFields_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"customfield_10001": "Sprint 1",
		"customfield_10002": 42.5,
		"customfield_10003": {"accountId": "123"}
	}`

	var cf CustomFields
	err := json.Unmarshal([]byte(jsonData), &cf)
	require.NoError(t, err)

	assert.Len(t, cf, 3)

	str, ok := cf.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", str)

	num, ok := cf.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)

	_, ok = cf.GetRaw("customfield_10003")
	assert.True(t, ok)
}

func TestCustomFields_ToMap(t *testing.T) {
	cf := NewCustomFields().
		SetString("customfield_10001", "Sprint 1").
		SetNumber("customfield_10002", 42.5)

	m := cf.ToMap()

	assert.Len(t, m, 2)
	assert.Equal(t, "Sprint 1", m["customfield_10001"])
	assert.Equal(t, 42.5, m["customfield_10002"])
}

func TestCustomFields_FromMap(t *testing.T) {
	m := map[string]interface{}{
		"customfield_10001": "Sprint 1",
		"customfield_10002": 42.5,
	}

	cf := FromMap(m)

	assert.Len(t, cf, 2)

	str, ok := cf.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", str)

	num, ok := cf.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)
}

func TestCustomFields_Merge(t *testing.T) {
	cf1 := NewCustomFields().
		SetString("customfield_10001", "Sprint 1").
		SetNumber("customfield_10002", 42.5)

	cf2 := NewCustomFields().
		SetString("customfield_10001", "Sprint 2"). // Overwrites
		SetLabels("customfield_10008", []string{"bug"})

	cf1.Merge(cf2)

	assert.Len(t, cf1, 3)

	// Check overwritten value
	str, ok := cf1.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 2", str)

	// Check preserved value
	num, ok := cf1.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)

	// Check merged value
	labels, ok := cf1.GetLabels("customfield_10008")
	assert.True(t, ok)
	assert.Equal(t, []string{"bug"}, labels)
}

func TestCustomFieldError(t *testing.T) {
	err := &CustomFieldError{
		FieldID: "customfield_10001",
		Message: "invalid format",
	}

	assert.Equal(t, "custom field customfield_10001: invalid format", err.Error())
}

func TestIssueFields_MarshalJSON_WithCustomFields(t *testing.T) {
	fields := &IssueFields{
		Summary: "Test issue",
		Custom: NewCustomFields().
			SetString("customfield_10001", "Sprint 1").
			SetNumber("customfield_10002", 42.5),
	}

	data, err := json.Marshal(fields)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", result["summary"])
	assert.Equal(t, "Sprint 1", result["customfield_10001"])
	assert.Equal(t, 42.5, result["customfield_10002"])
}

func TestIssueFields_MarshalJSON_WithoutCustomFields(t *testing.T) {
	fields := &IssueFields{
		Summary: "Test issue",
	}

	data, err := json.Marshal(fields)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", result["summary"])
	_, hasCustom := result["customfield_10001"]
	assert.False(t, hasCustom)
}

func TestIssueFields_UnmarshalJSON_WithCustomFields(t *testing.T) {
	jsonData := `{
		"summary": "Test issue",
		"description": {
			"type": "doc",
			"version": 1,
			"content": [
				{
					"type": "paragraph",
					"content": [
						{
							"type": "text",
							"text": "Test description"
						}
					]
				}
			]
		},
		"customfield_10001": "Sprint 1",
		"customfield_10002": 42.5,
		"customfield_10003": {"accountId": "123"}
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", fields.Summary)
	require.NotNil(t, fields.Description)
	assert.Equal(t, "Test description", fields.Description.ToText())

	// Check custom fields were extracted
	assert.Len(t, fields.Custom, 3)

	str, ok := fields.Custom.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", str)

	num, ok := fields.Custom.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)

	_, ok = fields.Custom.GetRaw("customfield_10003")
	assert.True(t, ok)
}

func TestIssueFields_UnmarshalJSON_WithoutCustomFields(t *testing.T) {
	jsonData := `{
		"summary": "Test issue",
		"description": {
			"type": "doc",
			"version": 1,
			"content": [
				{
					"type": "paragraph",
					"content": [
						{
							"type": "text",
							"text": "Test description"
						}
					]
				}
			]
		}
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", fields.Summary)
	require.NotNil(t, fields.Description)
	assert.Equal(t, "Test description", fields.Description.ToText())

	// Custom fields should be initialized but empty
	assert.NotNil(t, fields.Custom)
	assert.Len(t, fields.Custom, 0)
}

func TestIssueFields_RoundTrip(t *testing.T) {
	original := &IssueFields{
		Summary:     "Test issue",
		Description: ADFFromText("Test description"),
		Custom: NewCustomFields().
			SetString("customfield_10001", "Sprint 1").
			SetNumber("customfield_10002", 42.5).
			SetLabels("customfield_10008", []string{"bug", "urgent"}),
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded IssueFields
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify standard fields
	assert.Equal(t, original.Summary, decoded.Summary)
	assert.Equal(t, original.Description, decoded.Description)

	// Verify custom fields
	str, ok := decoded.Custom.GetString("customfield_10001")
	assert.True(t, ok)
	assert.Equal(t, "Sprint 1", str)

	num, ok := decoded.Custom.GetNumber("customfield_10002")
	assert.True(t, ok)
	assert.Equal(t, 42.5, num)

	labels, ok := decoded.Custom.GetLabels("customfield_10008")
	assert.True(t, ok)
	assert.Equal(t, []string{"bug", "urgent"}, labels)
}

func TestIssueFields_UnmarshalJSON_FlexibleDateFormats(t *testing.T) {
	t.Run("DueDate with date-only format (YYYY-MM-DD)", func(t *testing.T) {
		jsonData := `{
			"summary": "Test issue",
			"duedate": "2025-10-30"
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Test issue", fields.Summary)
		require.NotNil(t, fields.DueDate)

		// Verify the date was parsed correctly
		expected := time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, *fields.DueDate)
	})

	t.Run("DueDate with RFC3339 format (fallback)", func(t *testing.T) {
		jsonData := `{
			"summary": "Test issue",
			"duedate": "2025-10-30T15:04:05Z"
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Test issue", fields.Summary)
		require.NotNil(t, fields.DueDate)

		// Verify the date was parsed correctly with time component
		expected := time.Date(2025, 10, 30, 15, 4, 5, 0, time.UTC)
		assert.Equal(t, expected, *fields.DueDate)
	})

	t.Run("DueDate with null value", func(t *testing.T) {
		jsonData := `{
			"summary": "Test issue",
			"duedate": null
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Test issue", fields.Summary)
		assert.Nil(t, fields.DueDate)
	})

	t.Run("DueDate missing from JSON", func(t *testing.T) {
		jsonData := `{
			"summary": "Test issue"
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Test issue", fields.Summary)
		assert.Nil(t, fields.DueDate)
	})

	t.Run("Created and Updated with RFC3339 format", func(t *testing.T) {
		jsonData := `{
			"summary": "Test issue",
			"created": "2024-01-15T10:30:00.000Z",
			"updated": "2024-01-20T14:45:30.000Z",
			"duedate": "2025-10-30"
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Test issue", fields.Summary)

		// Verify Created
		require.NotNil(t, fields.Created)
		expectedCreated := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		assert.Equal(t, expectedCreated, *fields.Created)

		// Verify Updated
		require.NotNil(t, fields.Updated)
		expectedUpdated := time.Date(2024, 1, 20, 14, 45, 30, 0, time.UTC)
		assert.Equal(t, expectedUpdated, *fields.Updated)

		// Verify DueDate (date-only format)
		require.NotNil(t, fields.DueDate)
		expectedDue := time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedDue, *fields.DueDate)
	})

	t.Run("All date fields with realistic Jira response", func(t *testing.T) {
		// Realistic Jira API response format
		jsonData := `{
			"summary": "Implement user authentication",
			"description": {
				"type": "doc",
				"version": 1,
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "Add OAuth 2.0 support"
							}
						]
					}
				]
			},
			"created": "2024-01-01T10:30:00.000+0000",
			"updated": "2024-01-05T15:20:30.000+0000",
			"duedate": "2024-02-01"
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		assert.Equal(t, "Implement user authentication", fields.Summary)
		require.NotNil(t, fields.Description)
		assert.Equal(t, "Add OAuth 2.0 support", fields.Description.ToText())

		// All date fields should be parsed successfully
		require.NotNil(t, fields.Created)
		require.NotNil(t, fields.Updated)
		require.NotNil(t, fields.DueDate)
	})

	t.Run("Custom date fields with various formats", func(t *testing.T) {
		// Test that custom fields also benefit from flexible date parsing
		jsonData := `{
			"summary": "Test with custom dates",
			"customfield_10001": "2025-12-25",
			"customfield_10002": "2024-06-15T09:30:00.000+0000",
			"customfield_10003": "2024-07-20T14:00:00Z",
			"customfield_10004": "not-a-date",
			"customfield_10005": 12345
		}`

		var fields IssueFields
		err := json.Unmarshal([]byte(jsonData), &fields)
		require.NoError(t, err)

		// Custom date-only field should be parsed
		date1, ok := fields.Custom.GetDate("customfield_10001")
		assert.True(t, ok)
		expected1 := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected1, date1)

		// Custom datetime with Jira format should be parsed
		date2, ok := fields.Custom.GetDateTime("customfield_10002")
		assert.True(t, ok)
		assert.False(t, date2.IsZero())

		// Custom datetime with RFC3339 should be parsed
		date3, ok := fields.Custom.GetDateTime("customfield_10003")
		assert.True(t, ok)
		expected3 := time.Date(2024, 7, 20, 14, 0, 0, 0, time.UTC)
		assert.Equal(t, expected3, date3)

		// Non-date string should remain as string
		str, ok := fields.Custom.GetString("customfield_10004")
		assert.True(t, ok)
		assert.Equal(t, "not-a-date", str)

		// Number should remain as number
		num, ok := fields.Custom.GetNumber("customfield_10005")
		assert.True(t, ok)
		assert.Equal(t, 12345.0, num)
	})
}
