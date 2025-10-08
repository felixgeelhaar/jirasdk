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
		"description": "Test description",
		"customfield_10001": "Sprint 1",
		"customfield_10002": 42.5,
		"customfield_10003": {"accountId": "123"}
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", fields.Summary)
	assert.Equal(t, "Test description", fields.Description)

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
		"description": "Test description"
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	require.NoError(t, err)

	assert.Equal(t, "Test issue", fields.Summary)
	assert.Equal(t, "Test description", fields.Description)

	// Custom fields should be initialized but empty
	assert.NotNil(t, fields.Custom)
	assert.Len(t, fields.Custom, 0)
}

func TestIssueFields_RoundTrip(t *testing.T) {
	original := &IssueFields{
		Summary:     "Test issue",
		Description: "Test description",
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
