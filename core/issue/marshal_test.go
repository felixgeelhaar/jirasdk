package issue

import (
	"encoding/json"
	"testing"
)

// TestProjectMarshaling verifies that empty fields are omitted when marshaling Project
func TestProjectMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		project  *Project
		expected string
	}{
		{
			name:     "Project with only Key",
			project:  &Project{Key: "PROJ"},
			expected: `{"key":"PROJ"}`,
		},
		{
			name:     "Project with only ID",
			project:  &Project{ID: "10000"},
			expected: `{"id":"10000"}`,
		},
		{
			name:     "Project with ID and Key",
			project:  &Project{ID: "10000", Key: "PROJ"},
			expected: `{"id":"10000","key":"PROJ"}`,
		},
		{
			name:     "Project with all fields",
			project:  &Project{ID: "10000", Key: "PROJ", Name: "My Project", Self: "https://example.com/proj"},
			expected: `{"id":"10000","key":"PROJ","name":"My Project","self":"https://example.com/proj"}`,
		},
		{
			name:     "Empty Project",
			project:  &Project{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.project)
			if err != nil {
				t.Fatalf("Failed to marshal project: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// TestIssueTypeMarshaling verifies that empty fields are omitted when marshaling IssueType
func TestIssueTypeMarshaling(t *testing.T) {
	tests := []struct {
		name      string
		issueType *IssueType
		expected  string
	}{
		{
			name:      "IssueType with only ID",
			issueType: &IssueType{ID: "10001"},
			expected:  `{"id":"10001"}`,
		},
		{
			name:      "IssueType with only Name",
			issueType: &IssueType{Name: "Bug"},
			expected:  `{"name":"Bug"}`,
		},
		{
			name:      "IssueType with ID and Name",
			issueType: &IssueType{ID: "10001", Name: "Bug"},
			expected:  `{"id":"10001","name":"Bug"}`,
		},
		{
			name:      "Empty IssueType",
			issueType: &IssueType{},
			expected:  `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.issueType)
			if err != nil {
				t.Fatalf("Failed to marshal issue type: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// TestPriorityMarshaling verifies that empty fields are omitted when marshaling Priority
func TestPriorityMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		priority *Priority
		expected string
	}{
		{
			name:     "Priority with only ID",
			priority: &Priority{ID: "3"},
			expected: `{"id":"3"}`,
		},
		{
			name:     "Priority with only Name",
			priority: &Priority{Name: "High"},
			expected: `{"name":"High"}`,
		},
		{
			name:     "Empty Priority",
			priority: &Priority{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.priority)
			if err != nil {
				t.Fatalf("Failed to marshal priority: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// TestComponentMarshaling verifies that empty fields are omitted when marshaling Component
func TestComponentMarshaling(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		expected  string
	}{
		{
			name:      "Component with only ID",
			component: &Component{ID: "10050"},
			expected:  `{"id":"10050"}`,
		},
		{
			name:      "Component with only Name",
			component: &Component{Name: "Backend"},
			expected:  `{"name":"Backend"}`,
		},
		{
			name:      "Empty Component",
			component: &Component{},
			expected:  `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.component)
			if err != nil {
				t.Fatalf("Failed to marshal component: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// TestIssueFieldsMarshaling verifies complete IssueFields marshaling behavior
func TestIssueFieldsMarshaling(t *testing.T) {
	fields := &IssueFields{
		Summary:   "Test Issue",
		Project:   &Project{Key: "PROJ"},   // Only Key, no ID
		IssueType: &IssueType{Name: "Bug"}, // Only Name, no ID
	}

	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("Failed to marshal fields: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify project has only key, not id
	project := result["project"].(map[string]interface{})
	if _, hasID := project["id"]; hasID {
		t.Error("Project should not have 'id' field when empty")
	}
	if key, hasKey := project["key"]; !hasKey || key != "PROJ" {
		t.Error("Project should have 'key' field with value 'PROJ'")
	}

	// Verify issuetype has only name, not id
	issueType := result["issuetype"].(map[string]interface{})
	if _, hasID := issueType["id"]; hasID {
		t.Error("IssueType should not have 'id' field when empty")
	}
	if name, hasName := issueType["name"]; !hasName || name != "Bug" {
		t.Error("IssueType should have 'name' field with value 'Bug'")
	}
}
