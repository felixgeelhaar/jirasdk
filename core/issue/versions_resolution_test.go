package issue

import (
	"encoding/json"
	"testing"

	"github.com/felixgeelhaar/jirasdk/core/project"
	"github.com/felixgeelhaar/jirasdk/core/resolution"
)

// TestResolutionField_MarshalJSON tests that the Resolution field is correctly marshaled to JSON.
func TestResolutionField_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		fields   *IssueFields
		expected string
	}{
		{
			name: "resolution with name",
			fields: &IssueFields{
				Summary:    "Bug fix",
				IssueType:  &IssueType{Name: "Bug"},
				Project:    &Project{Key: "PROJ"},
				Resolution: &resolution.Resolution{Name: "Done"},
			},
			expected: `{"summary":"Bug fix","issuetype":{"name":"Bug"},"project":{"key":"PROJ"},"resolution":{"name":"Done"}}`,
		},
		{
			name: "resolution with ID",
			fields: &IssueFields{
				Summary:    "Bug fix",
				IssueType:  &IssueType{Name: "Bug"},
				Project:    &Project{Key: "PROJ"},
				Resolution: &resolution.Resolution{ID: "10001"},
			},
			expected: `{"summary":"Bug fix","issuetype":{"name":"Bug"},"project":{"key":"PROJ"},"resolution":{"id":"10001"}}`,
		},
		{
			name: "resolution with both name and ID",
			fields: &IssueFields{
				Summary:    "Bug fix",
				IssueType:  &IssueType{Name: "Bug"},
				Project:    &Project{Key: "PROJ"},
				Resolution: &resolution.Resolution{ID: "10001", Name: "Done"},
			},
			expected: `{"summary":"Bug fix","issuetype":{"name":"Bug"},"project":{"key":"PROJ"},"resolution":{"id":"10001","name":"Done"}}`,
		},
		{
			name: "no resolution field",
			fields: &IssueFields{
				Summary:   "Bug fix",
				IssueType: &IssueType{Name: "Bug"},
				Project:   &Project{Key: "PROJ"},
			},
			expected: `{"summary":"Bug fix","issuetype":{"name":"Bug"},"project":{"key":"PROJ"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.fields)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			got := string(data)
			if got != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestResolutionField_UnmarshalJSON tests that the Resolution field is correctly unmarshaled from JSON.
func TestResolutionField_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, fields *IssueFields)
	}{
		{
			name: "unmarshal resolution with name",
			json: `{"summary":"Bug fix","resolution":{"name":"Done"}}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Resolution == nil {
					t.Fatal("Resolution is nil")
				}
				if fields.Resolution.Name != "Done" {
					t.Errorf("Resolution.Name = %s, want Done", fields.Resolution.Name)
				}
			},
		},
		{
			name: "unmarshal resolution with ID",
			json: `{"summary":"Bug fix","resolution":{"id":"10001"}}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Resolution == nil {
					t.Fatal("Resolution is nil")
				}
				if fields.Resolution.ID != "10001" {
					t.Errorf("Resolution.ID = %s, want 10001", fields.Resolution.ID)
				}
			},
		},
		{
			name: "unmarshal without resolution",
			json: `{"summary":"Bug fix"}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.Resolution != nil {
					t.Errorf("Resolution = %v, want nil", fields.Resolution)
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

// TestFixVersionsField_MarshalJSON tests that the FixVersions field is correctly marshaled to JSON.
func TestFixVersionsField_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		fields   *IssueFields
		expected string
	}{
		{
			name: "single fix version",
			fields: &IssueFields{
				Summary:   "Feature",
				IssueType: &IssueType{Name: "Story"},
				Project:   &Project{Key: "PROJ"},
				FixVersions: []*project.Version{
					{Name: "v1.0.0"},
				},
			},
			expected: `{"summary":"Feature","issuetype":{"name":"Story"},"project":{"key":"PROJ"},"fixVersions":[{"name":"v1.0.0"}]}`,
		},
		{
			name: "multiple fix versions",
			fields: &IssueFields{
				Summary:   "Feature",
				IssueType: &IssueType{Name: "Story"},
				Project:   &Project{Key: "PROJ"},
				FixVersions: []*project.Version{
					{ID: "10001", Name: "v1.0.0"},
					{ID: "10002", Name: "v2.0.0"},
				},
			},
			expected: `{"summary":"Feature","issuetype":{"name":"Story"},"project":{"key":"PROJ"},"fixVersions":[{"id":"10001","name":"v1.0.0"},{"id":"10002","name":"v2.0.0"}]}`,
		},
		{
			name: "no fix versions",
			fields: &IssueFields{
				Summary:   "Feature",
				IssueType: &IssueType{Name: "Story"},
				Project:   &Project{Key: "PROJ"},
			},
			expected: `{"summary":"Feature","issuetype":{"name":"Story"},"project":{"key":"PROJ"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.fields)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			got := string(data)
			if got != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestAffectsVersionsField_MarshalJSON tests that the AffectsVersions field is correctly marshaled to JSON.
func TestAffectsVersionsField_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		fields   *IssueFields
		expected string
	}{
		{
			name: "single affects version",
			fields: &IssueFields{
				Summary:   "Bug",
				IssueType: &IssueType{Name: "Bug"},
				Project:   &Project{Key: "PROJ"},
				AffectsVersions: []*project.Version{
					{Name: "v1.0.0"},
				},
			},
			expected: `{"summary":"Bug","issuetype":{"name":"Bug"},"project":{"key":"PROJ"},"versions":[{"name":"v1.0.0"}]}`,
		},
		{
			name: "multiple affects versions",
			fields: &IssueFields{
				Summary:   "Bug",
				IssueType: &IssueType{Name: "Bug"},
				Project:   &Project{Key: "PROJ"},
				AffectsVersions: []*project.Version{
					{ID: "10001", Name: "v1.0.0"},
					{ID: "10002", Name: "v1.1.0"},
				},
			},
			expected: `{"summary":"Bug","issuetype":{"name":"Bug"},"project":{"key":"PROJ"},"versions":[{"id":"10001","name":"v1.0.0"},{"id":"10002","name":"v1.1.0"}]}`,
		},
		{
			name: "no affects versions",
			fields: &IssueFields{
				Summary:   "Bug",
				IssueType: &IssueType{Name: "Bug"},
				Project:   &Project{Key: "PROJ"},
			},
			expected: `{"summary":"Bug","issuetype":{"name":"Bug"},"project":{"key":"PROJ"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.fields)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			got := string(data)
			if got != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestVersionsField_UnmarshalJSON tests that version fields are correctly unmarshaled from JSON.
func TestVersionsField_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, fields *IssueFields)
	}{
		{
			name: "unmarshal fix versions",
			json: `{"summary":"Feature","fixVersions":[{"id":"10001","name":"v1.0.0"},{"id":"10002","name":"v2.0.0"}]}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if len(fields.FixVersions) != 2 {
					t.Fatalf("FixVersions length = %d, want 2", len(fields.FixVersions))
				}
				if fields.FixVersions[0].Name != "v1.0.0" {
					t.Errorf("FixVersions[0].Name = %s, want v1.0.0", fields.FixVersions[0].Name)
				}
				if fields.FixVersions[1].Name != "v2.0.0" {
					t.Errorf("FixVersions[1].Name = %s, want v2.0.0", fields.FixVersions[1].Name)
				}
			},
		},
		{
			name: "unmarshal affects versions",
			json: `{"summary":"Bug","versions":[{"id":"10001","name":"v1.0.0"}]}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if len(fields.AffectsVersions) != 1 {
					t.Fatalf("AffectsVersions length = %d, want 1", len(fields.AffectsVersions))
				}
				if fields.AffectsVersions[0].Name != "v1.0.0" {
					t.Errorf("AffectsVersions[0].Name = %s, want v1.0.0", fields.AffectsVersions[0].Name)
				}
			},
		},
		{
			name: "unmarshal without versions",
			json: `{"summary":"Feature"}`,
			validate: func(t *testing.T, fields *IssueFields) {
				if fields.FixVersions != nil {
					t.Errorf("FixVersions = %v, want nil", fields.FixVersions)
				}
				if fields.AffectsVersions != nil {
					t.Errorf("AffectsVersions = %v, want nil", fields.AffectsVersions)
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

// TestIssue_GetResolution tests the safe Resolution accessor method.
func TestIssue_GetResolution(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected *resolution.Resolution
	}{
		{
			name: "resolution exists",
			issue: &Issue{
				Fields: &IssueFields{
					Resolution: &resolution.Resolution{Name: "Done"},
				},
			},
			expected: &resolution.Resolution{Name: "Done"},
		},
		{
			name: "resolution is nil",
			issue: &Issue{
				Fields: &IssueFields{
					Resolution: nil,
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
			got := tt.issue.GetResolution()
			if got == nil && tt.expected == nil {
				return
			}
			if got == nil || tt.expected == nil {
				t.Errorf("GetResolution() = %v, want %v", got, tt.expected)
				return
			}
			if got.Name != tt.expected.Name {
				t.Errorf("GetResolution().Name = %s, want %s", got.Name, tt.expected.Name)
			}
		})
	}
}

// TestIssue_GetResolutionName tests the safe ResolutionName accessor method.
func TestIssue_GetResolutionName(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected string
	}{
		{
			name: "resolution name exists",
			issue: &Issue{
				Fields: &IssueFields{
					Resolution: &resolution.Resolution{Name: "Done"},
				},
			},
			expected: "Done",
		},
		{
			name: "resolution exists but name is empty",
			issue: &Issue{
				Fields: &IssueFields{
					Resolution: &resolution.Resolution{ID: "10001"},
				},
			},
			expected: "",
		},
		{
			name: "resolution is nil",
			issue: &Issue{
				Fields: &IssueFields{
					Resolution: nil,
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
			got := tt.issue.GetResolutionName()
			if got != tt.expected {
				t.Errorf("GetResolutionName() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestIssue_GetFixVersions tests the safe FixVersions accessor method.
func TestIssue_GetFixVersions(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected int // number of versions expected
	}{
		{
			name: "fix versions exist",
			issue: &Issue{
				Fields: &IssueFields{
					FixVersions: []*project.Version{
						{Name: "v1.0.0"},
						{Name: "v2.0.0"},
					},
				},
			},
			expected: 2,
		},
		{
			name: "fix versions is nil",
			issue: &Issue{
				Fields: &IssueFields{
					FixVersions: nil,
				},
			},
			expected: 0,
		},
		{
			name: "fields is nil",
			issue: &Issue{
				Fields: nil,
			},
			expected: 0,
		},
		{
			name:     "issue is empty",
			issue:    &Issue{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.GetFixVersions()
			if got == nil && tt.expected == 0 {
				return
			}
			if len(got) != tt.expected {
				t.Errorf("GetFixVersions() length = %d, want %d", len(got), tt.expected)
			}
		})
	}
}

// TestIssue_GetAffectsVersions tests the safe AffectsVersions accessor method.
func TestIssue_GetAffectsVersions(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected int // number of versions expected
	}{
		{
			name: "affects versions exist",
			issue: &Issue{
				Fields: &IssueFields{
					AffectsVersions: []*project.Version{
						{Name: "v1.0.0"},
						{Name: "v1.1.0"},
					},
				},
			},
			expected: 2,
		},
		{
			name: "affects versions is nil",
			issue: &Issue{
				Fields: &IssueFields{
					AffectsVersions: nil,
				},
			},
			expected: 0,
		},
		{
			name: "fields is nil",
			issue: &Issue{
				Fields: nil,
			},
			expected: 0,
		},
		{
			name:     "issue is empty",
			issue:    &Issue{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.GetAffectsVersions()
			if got == nil && tt.expected == 0 {
				return
			}
			if len(got) != tt.expected {
				t.Errorf("GetAffectsVersions() length = %d, want %d", len(got), tt.expected)
			}
		})
	}
}

// TestCreateIssueWithVersionsAndResolution tests creating an issue with version and resolution fields.
func TestCreateIssueWithVersionsAndResolution(t *testing.T) {
	input := &CreateInput{
		Fields: &IssueFields{
			Project:   &Project{Key: "PROJ"},
			Summary:   "Bug fix for version 1.0",
			IssueType: &IssueType{Name: "Bug"},
			FixVersions: []*project.Version{
				{Name: "v1.0.0"},
				{Name: "v2.0.0"},
			},
			AffectsVersions: []*project.Version{
				{Name: "v1.0.0"},
			},
			Resolution: &resolution.Resolution{Name: "Done"},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify the JSON contains all the fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("fields is not a map")
	}

	// Check fix versions
	fixVersions, ok := fields["fixVersions"].([]interface{})
	if !ok {
		t.Fatal("fixVersions field is missing or not an array")
	}
	if len(fixVersions) != 2 {
		t.Errorf("fixVersions length = %d, want 2", len(fixVersions))
	}

	// Check affects versions (JSON field name is "versions")
	affectsVersions, ok := fields["versions"].([]interface{})
	if !ok {
		t.Fatal("versions field is missing or not an array")
	}
	if len(affectsVersions) != 1 {
		t.Errorf("versions length = %d, want 1", len(affectsVersions))
	}

	// Check resolution
	resolutionField, ok := fields["resolution"].(map[string]interface{})
	if !ok {
		t.Fatal("resolution field is missing or not a map")
	}
	if resolutionField["name"] != "Done" {
		t.Errorf("resolution.name = %v, want Done", resolutionField["name"])
	}
}

// TestVersionsResolution_NilSafety tests that accessing version and resolution fields doesn't cause panics.
func TestVersionsResolution_NilSafety(t *testing.T) {
	// These should not panic
	var issue *Issue

	// Test with nil issue
	t.Run("nil issue", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Methods panicked with nil issue: %v", r)
			}
		}()
		if issue != nil {
			_ = issue.GetResolution()
			_ = issue.GetResolutionName()
			_ = issue.GetFixVersions()
			_ = issue.GetAffectsVersions()
		}
	})

	// Test with nil fields
	issue = &Issue{}
	t.Run("nil fields", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Methods panicked with nil fields: %v", r)
			}
		}()
		resolution := issue.GetResolution()
		if resolution != nil {
			t.Errorf("GetResolution() = %v, want nil", resolution)
		}
		resolutionName := issue.GetResolutionName()
		if resolutionName != "" {
			t.Errorf("GetResolutionName() = %s, want empty", resolutionName)
		}
		fixVersions := issue.GetFixVersions()
		if fixVersions == nil {
			t.Errorf("GetFixVersions() = nil, want empty slice")
		} else if len(fixVersions) != 0 {
			t.Errorf("GetFixVersions() = %v, want empty slice", fixVersions)
		}
		affectsVersions := issue.GetAffectsVersions()
		if affectsVersions == nil {
			t.Errorf("GetAffectsVersions() = nil, want empty slice")
		} else if len(affectsVersions) != 0 {
			t.Errorf("GetAffectsVersions() = %v, want empty slice", affectsVersions)
		}
	})
}
