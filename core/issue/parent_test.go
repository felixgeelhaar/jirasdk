package issue

import (
	"encoding/json"
	"testing"
)

// TestParentField_MarshalJSON tests that the Parent field is correctly marshaled to JSON.
func TestParentField_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		fields   *IssueFields
		expected string
	}{
		{
			name: "parent with key",
			fields: &IssueFields{
				Summary:   "Subtask summary",
				IssueType: &IssueType{Name: "Sub-task"},
				Project:   &Project{Key: "PROJ"},
				Parent:    &IssueRef{Key: "PROJ-123"},
			},
			expected: `{"summary":"Subtask summary","issuetype":{"name":"Sub-task"},"project":{"key":"PROJ"},"parent":{"key":"PROJ-123"}}`,
		},
		{
			name: "parent with ID",
			fields: &IssueFields{
				Summary:   "Subtask summary",
				IssueType: &IssueType{Name: "Sub-task"},
				Project:   &Project{Key: "PROJ"},
				Parent:    &IssueRef{ID: "10001"},
			},
			expected: `{"summary":"Subtask summary","issuetype":{"name":"Sub-task"},"project":{"key":"PROJ"},"parent":{"id":"10001"}}`,
		},
		{
			name: "parent with both key and ID",
			fields: &IssueFields{
				Summary:   "Subtask summary",
				IssueType: &IssueType{Name: "Sub-task"},
				Project:   &Project{Key: "PROJ"},
				Parent:    &IssueRef{Key: "PROJ-123", ID: "10001"},
			},
			expected: `{"summary":"Subtask summary","issuetype":{"name":"Sub-task"},"project":{"key":"PROJ"},"parent":{"id":"10001","key":"PROJ-123"}}`,
		},
		{
			name: "no parent field",
			fields: &IssueFields{
				Summary:   "Regular task",
				IssueType: &IssueType{Name: "Task"},
				Project:   &Project{Key: "PROJ"},
			},
			expected: `{"summary":"Regular task","issuetype":{"name":"Task"},"project":{"key":"PROJ"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.fields)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			// Compare JSON strings
			got := string(data)
			if got != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestParentField_UnmarshalJSON tests that the Parent field is correctly unmarshaled from JSON.
func TestParentField_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, fields *IssueFields)
	}{
		{
			name: "unmarshal parent with key",
			json: `{"summary":"Subtask","parent":{"key":"PROJ-123"}}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Parent == nil {
					t.Fatal("Parent is nil")
				}
				if fields.Parent.Key != "PROJ-123" {
					t.Errorf("Parent.Key = %s, want PROJ-123", fields.Parent.Key)
				}
				if fields.Parent.ID != "" {
					t.Errorf("Parent.ID = %s, want empty", fields.Parent.ID)
				}
			},
		},
		{
			name: "unmarshal parent with ID",
			json: `{"summary":"Subtask","parent":{"id":"10001"}}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Parent == nil {
					t.Fatal("Parent is nil")
				}
				if fields.Parent.ID != "10001" {
					t.Errorf("Parent.ID = %s, want 10001", fields.Parent.ID)
				}
				if fields.Parent.Key != "" {
					t.Errorf("Parent.Key = %s, want empty", fields.Parent.Key)
				}
			},
		},
		{
			name: "unmarshal parent with both",
			json: `{"summary":"Subtask","parent":{"key":"PROJ-123","id":"10001"}}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Parent == nil {
					t.Fatal("Parent is nil")
				}
				if fields.Parent.Key != "PROJ-123" {
					t.Errorf("Parent.Key = %s, want PROJ-123", fields.Parent.Key)
				}
				if fields.Parent.ID != "10001" {
					t.Errorf("Parent.ID = %s, want 10001", fields.Parent.ID)
				}
			},
		},
		{
			name: "unmarshal without parent",
			json: `{"summary":"Regular task"}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Parent != nil {
					t.Errorf("Parent = %v, want nil", fields.Parent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fields IssueFields
			err := json.Unmarshal([]byte(tt.json), &fields)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			tt.validate(t, &fields)
		})
	}
}

// TestCreateIssueWithParent tests creating an issue with a parent field.
func TestCreateIssueWithParent(t *testing.T) {
	input := &CreateInput{
		Fields: &IssueFields{
			Project:   &Project{Key: "PROJ"},
			Summary:   "Implement API endpoint",
			IssueType: &IssueType{Name: "Sub-task"},
			Parent:    &IssueRef{Key: "PROJ-123"},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify the JSON contains the parent field
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("fields is not a map")
	}

	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatal("parent field is missing or not a map")
	}

	if parent["key"] != "PROJ-123" {
		t.Errorf("parent.key = %v, want PROJ-123", parent["key"])
	}
}

// TestIssue_GetParent tests the safe Parent accessor method.
func TestIssue_GetParent(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected *IssueRef
	}{
		{
			name: "parent exists",
			issue: &Issue{
				Fields: &IssueFields{
					Parent: &IssueRef{Key: "PROJ-123"},
				},
			},
			expected: &IssueRef{Key: "PROJ-123"},
		},
		{
			name: "parent is nil",
			issue: &Issue{
				Fields: &IssueFields{
					Parent: nil,
				},
			},
			expected: nil,
		},
		{
			name: "fields is nil",
			issue: &Issue{
				Fields: nil,
			},
			expected: nil,
		},
		{
			name:     "issue is empty",
			issue:    &Issue{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.GetParent()
			if got == nil && tt.expected == nil {
				return
			}
			if got == nil || tt.expected == nil {
				t.Errorf("GetParent() = %v, want %v", got, tt.expected)
				return
			}
			if got.Key != tt.expected.Key {
				t.Errorf("GetParent().Key = %s, want %s", got.Key, tt.expected.Key)
			}
		})
	}
}

// TestIssue_GetParentKey tests the safe ParentKey accessor method.
func TestIssue_GetParentKey(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected string
	}{
		{
			name: "parent key exists",
			issue: &Issue{
				Fields: &IssueFields{
					Parent: &IssueRef{Key: "PROJ-123"},
				},
			},
			expected: "PROJ-123",
		},
		{
			name: "parent exists but key is empty",
			issue: &Issue{
				Fields: &IssueFields{
					Parent: &IssueRef{ID: "10001"},
				},
			},
			expected: "",
		},
		{
			name: "parent is nil",
			issue: &Issue{
				Fields: &IssueFields{
					Parent: nil,
				},
			},
			expected: "",
		},
		{
			name: "fields is nil",
			issue: &Issue{
				Fields: nil,
			},
			expected: "",
		},
		{
			name:     "issue is empty",
			issue:    &Issue{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.GetParentKey()
			if got != tt.expected {
				t.Errorf("GetParentKey() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestParentField_NilSafety tests that accessing parent fields doesn't cause panics.
func TestParentField_NilSafety(t *testing.T) {
	// These should not panic
	var issue *Issue

	// Test with nil issue
	t.Run("nil issue", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetParent() panicked with nil issue: %v", r)
			}
		}()
		if issue != nil {
			_ = issue.GetParent()
		}
	})

	// Test with nil fields
	issue = &Issue{}
	t.Run("nil fields", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetParent() panicked with nil fields: %v", r)
			}
		}()
		parent := issue.GetParent()
		if parent != nil {
			t.Errorf("GetParent() = %v, want nil", parent)
		}
	})

	// Test GetParentKey with nil parent
	t.Run("nil parent key", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetParentKey() panicked: %v", r)
			}
		}()
		key := issue.GetParentKey()
		if key != "" {
			t.Errorf("GetParentKey() = %s, want empty", key)
		}
	})
}
