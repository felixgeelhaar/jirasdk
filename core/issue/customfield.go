package issue

import (
	"encoding/json"
	"fmt"
	"time"
)

// CustomFieldType represents the type of a custom field.
type CustomFieldType string

const (
	// CustomFieldTypeString represents a text field
	CustomFieldTypeString CustomFieldType = "string"
	// CustomFieldTypeNumber represents a numeric field
	CustomFieldTypeNumber CustomFieldType = "number"
	// CustomFieldTypeDate represents a date field
	CustomFieldTypeDate CustomFieldType = "date"
	// CustomFieldTypeDateTime represents a datetime field
	CustomFieldTypeDateTime CustomFieldType = "datetime"
	// CustomFieldTypeUser represents a user picker field
	CustomFieldTypeUser CustomFieldType = "user"
	// CustomFieldTypeSelect represents a single select field
	CustomFieldTypeSelect CustomFieldType = "select"
	// CustomFieldTypeMultiSelect represents a multi-select field
	CustomFieldTypeMultiSelect CustomFieldType = "multiselect"
	// CustomFieldTypeURL represents a URL field
	CustomFieldTypeURL CustomFieldType = "url"
	// CustomFieldTypeTextArea represents a text area field
	CustomFieldTypeTextArea CustomFieldType = "textarea"
	// CustomFieldTypeCheckbox represents a checkbox field
	CustomFieldTypeCheckbox CustomFieldType = "checkbox"
	// CustomFieldTypeRadio represents a radio button field
	CustomFieldTypeRadio CustomFieldType = "radio"
	// CustomFieldTypeCascadingSelect represents a cascading select field
	CustomFieldTypeCascadingSelect CustomFieldType = "cascadingselect"
	// CustomFieldTypeVersion represents a version picker field
	CustomFieldTypeVersion CustomFieldType = "version"
	// CustomFieldTypeLabels represents a labels field
	CustomFieldTypeLabels CustomFieldType = "labels"
)

// CustomField represents a custom field value with type information.
type CustomField struct {
	// ID is the custom field ID (e.g., "customfield_10001")
	ID string `json:"-"`

	// Type is the type of the custom field
	Type CustomFieldType `json:"-"`

	// Value is the actual value of the custom field
	Value interface{} `json:"-"`
}

// CustomFields is a type-safe wrapper for custom field operations.
type CustomFields map[string]*CustomField

// NewCustomFields creates a new CustomFields collection.
func NewCustomFields() CustomFields {
	return make(CustomFields)
}

// SetString sets a string custom field value.
//
// Example:
//
//	fields := issue.NewCustomFields()
//	fields.SetString("customfield_10001", "Sprint 1")
func (cf CustomFields) SetString(fieldID, value string) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeString,
		Value: value,
	}
	return cf
}

// GetString retrieves a string custom field value.
func (cf CustomFields) GetString(fieldID string) (string, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return "", false
	}

	if str, ok := field.Value.(string); ok {
		return str, true
	}
	return "", false
}

// SetNumber sets a numeric custom field value.
//
// Example:
//
//	fields.SetNumber("customfield_10002", 42.5)
func (cf CustomFields) SetNumber(fieldID string, value float64) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeNumber,
		Value: value,
	}
	return cf
}

// GetNumber retrieves a numeric custom field value.
func (cf CustomFields) GetNumber(fieldID string) (float64, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return 0, false
	}

	switch v := field.Value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// SetDate sets a date custom field value.
//
// Example:
//
//	fields.SetDate("customfield_10003", time.Now())
func (cf CustomFields) SetDate(fieldID string, value time.Time) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeDate,
		Value: value.Format("2006-01-02"),
	}
	return cf
}

// GetDate retrieves a date custom field value.
// Flexibly parses various date formats from Jira.
func (cf CustomFields) GetDate(fieldID string) (time.Time, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return time.Time{}, false
	}

	if str, ok := field.Value.(string); ok {
		// Use flexible parsing to handle various Jira date formats
		if t, parsed := tryParseDateTime(str); parsed {
			return t, true
		}
	}
	return time.Time{}, false
}

// SetDateTime sets a datetime custom field value.
//
// Example:
//
//	fields.SetDateTime("customfield_10004", time.Now())
func (cf CustomFields) SetDateTime(fieldID string, value time.Time) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeDateTime,
		Value: value.Format(time.RFC3339),
	}
	return cf
}

// GetDateTime retrieves a datetime custom field value.
// Flexibly parses various date/time formats from Jira.
func (cf CustomFields) GetDateTime(fieldID string) (time.Time, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return time.Time{}, false
	}

	if str, ok := field.Value.(string); ok {
		// Use flexible parsing to handle various Jira date/time formats
		if t, parsed := tryParseDateTime(str); parsed {
			return t, true
		}
	}
	return time.Time{}, false
}

// SetUser sets a user custom field value.
//
// Example:
//
//	fields.SetUser("customfield_10005", "accountId123")
func (cf CustomFields) SetUser(fieldID, accountID string) CustomFields {
	cf[fieldID] = &CustomField{
		ID:   fieldID,
		Type: CustomFieldTypeUser,
		Value: map[string]interface{}{
			"accountId": accountID,
		},
	}
	return cf
}

// GetUser retrieves a user custom field value.
func (cf CustomFields) GetUser(fieldID string) (string, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return "", false
	}

	if userMap, ok := field.Value.(map[string]interface{}); ok {
		if accountID, ok := userMap["accountId"].(string); ok {
			return accountID, true
		}
	}
	return "", false
}

// SetSelect sets a single select custom field value.
//
// Example:
//
//	fields.SetSelect("customfield_10006", "option1")
func (cf CustomFields) SetSelect(fieldID, value string) CustomFields {
	cf[fieldID] = &CustomField{
		ID:   fieldID,
		Type: CustomFieldTypeSelect,
		Value: map[string]interface{}{
			"value": value,
		},
	}
	return cf
}

// GetSelect retrieves a single select custom field value.
func (cf CustomFields) GetSelect(fieldID string) (string, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return "", false
	}

	if selectMap, ok := field.Value.(map[string]interface{}); ok {
		if value, ok := selectMap["value"].(string); ok {
			return value, true
		}
	}
	return "", false
}

// SetMultiSelect sets a multi-select custom field value.
//
// Example:
//
//	fields.SetMultiSelect("customfield_10007", []string{"option1", "option2"})
func (cf CustomFields) SetMultiSelect(fieldID string, values []string) CustomFields {
	options := make([]map[string]interface{}, len(values))
	for i, v := range values {
		options[i] = map[string]interface{}{
			"value": v,
		}
	}

	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeMultiSelect,
		Value: options,
	}
	return cf
}

// GetMultiSelect retrieves a multi-select custom field value.
func (cf CustomFields) GetMultiSelect(fieldID string) ([]string, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return nil, false
	}

	// Try []map[string]interface{} first (what SetMultiSelect creates)
	if options, ok := field.Value.([]map[string]interface{}); ok {
		values := make([]string, 0, len(options))
		for _, optMap := range options {
			if value, ok := optMap["value"].(string); ok {
				values = append(values, value)
			}
		}
		return values, len(values) > 0
	}

	// Try []interface{} (from JSON unmarshaling)
	if options, ok := field.Value.([]interface{}); ok {
		values := make([]string, 0, len(options))
		for _, opt := range options {
			if optMap, ok := opt.(map[string]interface{}); ok {
				if value, ok := optMap["value"].(string); ok {
					values = append(values, value)
				}
			}
		}
		return values, len(values) > 0
	}

	return nil, false
}

// SetLabels sets a labels custom field value.
//
// Example:
//
//	fields.SetLabels("customfield_10008", []string{"bug", "urgent"})
func (cf CustomFields) SetLabels(fieldID string, labels []string) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Type:  CustomFieldTypeLabels,
		Value: labels,
	}
	return cf
}

// GetLabels retrieves a labels custom field value.
func (cf CustomFields) GetLabels(fieldID string) ([]string, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return nil, false
	}

	if labels, ok := field.Value.([]string); ok {
		return labels, true
	}

	// Handle []interface{} case from JSON unmarshaling
	if labelsInterface, ok := field.Value.([]interface{}); ok {
		labels := make([]string, 0, len(labelsInterface))
		for _, l := range labelsInterface {
			if str, ok := l.(string); ok {
				labels = append(labels, str)
			}
		}
		return labels, len(labels) > 0
	}

	return nil, false
}

// SetRaw sets a custom field with a raw value (for advanced use cases).
//
// Example:
//
//	fields.SetRaw("customfield_10009", map[string]interface{}{
//	    "complex": "structure",
//	})
func (cf CustomFields) SetRaw(fieldID string, value interface{}) CustomFields {
	cf[fieldID] = &CustomField{
		ID:    fieldID,
		Value: value,
	}
	return cf
}

// GetRaw retrieves a raw custom field value.
func (cf CustomFields) GetRaw(fieldID string) (interface{}, bool) {
	field, ok := cf[fieldID]
	if !ok {
		return nil, false
	}
	return field.Value, true
}

// MarshalJSON converts CustomFields to the JSON format expected by Jira API.
func (cf CustomFields) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	for fieldID, field := range cf {
		result[fieldID] = field.Value
	}
	return json.Marshal(result)
}

// UnmarshalJSON parses custom fields from Jira API response.
func (cf *CustomFields) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if *cf == nil {
		*cf = make(CustomFields)
	}

	for fieldID, value := range raw {
		(*cf)[fieldID] = &CustomField{
			ID:    fieldID,
			Value: value,
		}
	}

	return nil
}

// ToMap converts CustomFields to a map for use in issue updates.
func (cf CustomFields) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for fieldID, field := range cf {
		result[fieldID] = field.Value
	}
	return result
}

// FromMap creates CustomFields from a map.
func FromMap(m map[string]interface{}) CustomFields {
	cf := NewCustomFields()
	for fieldID, value := range m {
		cf[fieldID] = &CustomField{
			ID:    fieldID,
			Value: value,
		}
	}
	return cf
}

// Merge merges another CustomFields into this one, overwriting conflicts.
func (cf CustomFields) Merge(other CustomFields) CustomFields {
	for fieldID, field := range other {
		cf[fieldID] = field
	}
	return cf
}

// CustomFieldError represents an error with custom field operations.
type CustomFieldError struct {
	FieldID string
	Message string
}

// Error implements the error interface.
func (e *CustomFieldError) Error() string {
	return fmt.Sprintf("custom field %s: %s", e.FieldID, e.Message)
}
