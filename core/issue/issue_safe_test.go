package issue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestIssueSafeHelperMethods tests all safe helper methods on the Issue struct
func TestIssueSafeHelperMethods(t *testing.T) {
	t.Run("SafeFields with nil Fields", func(t *testing.T) {
		issue := &Issue{
			ID:  "10001",
			Key: "PROJ-123",
			// Fields is nil
		}

		// Should not panic
		fields := issue.SafeFields()
		assert.NotNil(t, fields)
		assert.Equal(t, "", fields.Summary)
	})

	t.Run("SafeFields with populated Fields", func(t *testing.T) {
		issue := &Issue{
			ID:  "10001",
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Test summary",
				Description: "Test description",
			},
		}

		fields := issue.SafeFields()
		assert.NotNil(t, fields)
		assert.Equal(t, "Test summary", fields.Summary)
		assert.Equal(t, "Test description", fields.Description)
	})

	t.Run("GetSummary with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetSummary())
	})

	t.Run("GetSummary with populated Fields", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test summary"},
		}
		assert.Equal(t, "Test summary", issue.GetSummary())
	})

	t.Run("GetDescription with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetDescription())
	})

	t.Run("GetDescription with populated Fields", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Description: "Test description"},
		}
		assert.Equal(t, "Test description", issue.GetDescription())
	})

	t.Run("GetStatus with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetStatus())
	})

	t.Run("GetStatus with nil Status", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Status is nil
		}
		assert.Nil(t, issue.GetStatus())
	})

	t.Run("GetStatus with populated Status", func(t *testing.T) {
		status := &Status{ID: "1", Name: "To Do"}
		issue := &Issue{
			Fields: &IssueFields{Status: status},
		}
		assert.Equal(t, status, issue.GetStatus())
		assert.Equal(t, "To Do", issue.GetStatus().Name)
	})

	t.Run("GetStatusName with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetStatusName())

		issue.Fields = &IssueFields{}
		assert.Equal(t, "", issue.GetStatusName())
	})

	t.Run("GetStatusName with populated Status", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Status: &Status{Name: "In Progress"},
			},
		}
		assert.Equal(t, "In Progress", issue.GetStatusName())
	})

	t.Run("GetPriority with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetPriority())
	})

	t.Run("GetPriority with populated Priority", func(t *testing.T) {
		priority := &Priority{ID: "1", Name: "High"}
		issue := &Issue{
			Fields: &IssueFields{Priority: priority},
		}
		assert.Equal(t, priority, issue.GetPriority())
	})

	t.Run("GetPriorityName with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetPriorityName())
	})

	t.Run("GetPriorityName with populated Priority", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Priority: &Priority{Name: "Critical"},
			},
		}
		assert.Equal(t, "Critical", issue.GetPriorityName())
	})

	t.Run("GetAssignee with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetAssignee())
	})

	t.Run("GetAssignee with populated Assignee", func(t *testing.T) {
		assignee := &User{AccountID: "123", DisplayName: "John Doe"}
		issue := &Issue{
			Fields: &IssueFields{Assignee: assignee},
		}
		assert.Equal(t, assignee, issue.GetAssignee())
	})

	t.Run("GetAssigneeName with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetAssigneeName())
	})

	t.Run("GetAssigneeName with populated Assignee", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Assignee: &User{DisplayName: "Jane Smith"},
			},
		}
		assert.Equal(t, "Jane Smith", issue.GetAssigneeName())
	})

	t.Run("GetReporter with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetReporter())
	})

	t.Run("GetReporter with populated Reporter", func(t *testing.T) {
		reporter := &User{AccountID: "456", DisplayName: "Bob Reporter"}
		issue := &Issue{
			Fields: &IssueFields{Reporter: reporter},
		}
		assert.Equal(t, reporter, issue.GetReporter())
	})

	t.Run("GetReporterName with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetReporterName())
	})

	t.Run("GetReporterName with populated Reporter", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Reporter: &User{DisplayName: "Alice Reporter"},
			},
		}
		assert.Equal(t, "Alice Reporter", issue.GetReporterName())
	})

	t.Run("GetProject with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetProject())
	})

	t.Run("GetProject with populated Project", func(t *testing.T) {
		project := &Project{ID: "10000", Key: "PROJ"}
		issue := &Issue{
			Fields: &IssueFields{Project: project},
		}
		assert.Equal(t, project, issue.GetProject())
	})

	t.Run("GetProjectKey with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetProjectKey())
	})

	t.Run("GetProjectKey with populated Project", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Project: &Project{Key: "TEST"},
			},
		}
		assert.Equal(t, "TEST", issue.GetProjectKey())
	})

	t.Run("GetIssueType with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetIssueType())
	})

	t.Run("GetIssueType with populated IssueType", func(t *testing.T) {
		issueType := &IssueType{ID: "1", Name: "Task"}
		issue := &Issue{
			Fields: &IssueFields{IssueType: issueType},
		}
		assert.Equal(t, issueType, issue.GetIssueType())
	})

	t.Run("GetIssueTypeName with nil chain", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetIssueTypeName())
	})

	t.Run("GetIssueTypeName with populated IssueType", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				IssueType: &IssueType{Name: "Bug"},
			},
		}
		assert.Equal(t, "Bug", issue.GetIssueTypeName())
	})

	t.Run("GetLabels with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		labels := issue.GetLabels()
		assert.NotNil(t, labels)
		assert.Equal(t, 0, len(labels))
	})

	t.Run("GetLabels with nil Labels", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Labels is nil
		}
		labels := issue.GetLabels()
		assert.NotNil(t, labels)
		assert.Equal(t, 0, len(labels))
	})

	t.Run("GetLabels with populated Labels", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Labels: []string{"backend", "urgent"},
			},
		}
		labels := issue.GetLabels()
		assert.Equal(t, 2, len(labels))
		assert.Equal(t, "backend", labels[0])
		assert.Equal(t, "urgent", labels[1])
	})

	t.Run("GetComponents with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		components := issue.GetComponents()
		assert.NotNil(t, components)
		assert.Equal(t, 0, len(components))
	})

	t.Run("GetComponents with nil Components", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Components is nil
		}
		components := issue.GetComponents()
		assert.NotNil(t, components)
		assert.Equal(t, 0, len(components))
	})

	t.Run("GetComponents with populated Components", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{
				Components: []*Component{
					{ID: "1", Name: "API"},
					{ID: "2", Name: "Frontend"},
				},
			},
		}
		components := issue.GetComponents()
		assert.Equal(t, 2, len(components))
		assert.Equal(t, "API", components[0].Name)
		assert.Equal(t, "Frontend", components[1].Name)
	})

	t.Run("All safe methods on fully populated issue", func(t *testing.T) {
		issue := &Issue{
			ID:  "10001",
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Full test issue",
				Description: "Complete description",
				Status:      &Status{ID: "1", Name: "Open"},
				Priority:    &Priority{ID: "2", Name: "High"},
				Assignee:    &User{DisplayName: "John Assignee"},
				Reporter:    &User{DisplayName: "Jane Reporter"},
				Project:     &Project{Key: "PROJ", Name: "Test Project"},
				IssueType:   &IssueType{Name: "Bug"},
				Labels:      []string{"critical", "production"},
				Components:  []*Component{{Name: "Backend"}},
			},
		}

		// All methods should work without panicking
		assert.NotNil(t, issue.SafeFields())
		assert.Equal(t, "Full test issue", issue.GetSummary())
		assert.Equal(t, "Complete description", issue.GetDescription())
		assert.Equal(t, "Open", issue.GetStatusName())
		assert.Equal(t, "High", issue.GetPriorityName())
		assert.Equal(t, "John Assignee", issue.GetAssigneeName())
		assert.Equal(t, "Jane Reporter", issue.GetReporterName())
		assert.Equal(t, "PROJ", issue.GetProjectKey())
		assert.Equal(t, "Bug", issue.GetIssueTypeName())
		assert.Equal(t, 2, len(issue.GetLabels()))
		assert.Equal(t, 1, len(issue.GetComponents()))
	})
}

// TestIssueDateSafeHelperMethods tests all date-related safe helper methods
func TestIssueDateSafeHelperMethods(t *testing.T) {
	// Test time values for assertions
	testCreated := timePtr(2024, 1, 15, 10, 30, 0)
	testUpdated := timePtr(2024, 2, 20, 14, 45, 0)
	testDueDate := timePtr(2024, 3, 31, 23, 59, 59)

	t.Run("GetCreated with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetCreated())
	})

	t.Run("GetCreated with nil Created", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Created is nil
		}
		assert.Nil(t, issue.GetCreated())
	})

	t.Run("GetCreated with populated Created", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Created: testCreated},
		}
		created := issue.GetCreated()
		assert.NotNil(t, created)
		assert.Equal(t, *testCreated, *created)
	})

	t.Run("GetCreatedTime with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		created := issue.GetCreatedTime()
		assert.True(t, created.IsZero())
	})

	t.Run("GetCreatedTime with nil Created", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Created is nil
		}
		created := issue.GetCreatedTime()
		assert.True(t, created.IsZero())
	})

	t.Run("GetCreatedTime with populated Created", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Created: testCreated},
		}
		created := issue.GetCreatedTime()
		assert.False(t, created.IsZero())
		assert.Equal(t, *testCreated, created)
	})

	t.Run("GetUpdated with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetUpdated())
	})

	t.Run("GetUpdated with nil Updated", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Updated is nil
		}
		assert.Nil(t, issue.GetUpdated())
	})

	t.Run("GetUpdated with populated Updated", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Updated: testUpdated},
		}
		updated := issue.GetUpdated()
		assert.NotNil(t, updated)
		assert.Equal(t, *testUpdated, *updated)
	})

	t.Run("GetUpdatedTime with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		updated := issue.GetUpdatedTime()
		assert.True(t, updated.IsZero())
	})

	t.Run("GetUpdatedTime with nil Updated", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// Updated is nil
		}
		updated := issue.GetUpdatedTime()
		assert.True(t, updated.IsZero())
	})

	t.Run("GetUpdatedTime with populated Updated", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Updated: testUpdated},
		}
		updated := issue.GetUpdatedTime()
		assert.False(t, updated.IsZero())
		assert.Equal(t, *testUpdated, updated)
	})

	t.Run("GetDueDate with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		assert.Nil(t, issue.GetDueDate())
	})

	t.Run("GetDueDate with nil DueDate", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// DueDate is nil
		}
		assert.Nil(t, issue.GetDueDate())
	})

	t.Run("GetDueDate with populated DueDate", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{DueDate: testDueDate},
		}
		dueDate := issue.GetDueDate()
		assert.NotNil(t, dueDate)
		assert.Equal(t, *testDueDate, *dueDate)
	})

	t.Run("GetDueDateValue with nil Fields", func(t *testing.T) {
		issue := &Issue{ID: "10001", Key: "PROJ-123"}
		dueDate := issue.GetDueDateValue()
		assert.True(t, dueDate.IsZero())
	})

	t.Run("GetDueDateValue with nil DueDate", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{Summary: "Test"},
			// DueDate is nil
		}
		dueDate := issue.GetDueDateValue()
		assert.True(t, dueDate.IsZero())
	})

	t.Run("GetDueDateValue with populated DueDate", func(t *testing.T) {
		issue := &Issue{
			Fields: &IssueFields{DueDate: testDueDate},
		}
		dueDate := issue.GetDueDateValue()
		assert.False(t, dueDate.IsZero())
		assert.Equal(t, *testDueDate, dueDate)
	})

	t.Run("All date fields with complete issue", func(t *testing.T) {
		issue := &Issue{
			ID:  "10001",
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary: "Complete issue",
				Created: testCreated,
				Updated: testUpdated,
				DueDate: testDueDate,
			},
		}

		// Test pointer methods
		assert.NotNil(t, issue.GetCreated())
		assert.NotNil(t, issue.GetUpdated())
		assert.NotNil(t, issue.GetDueDate())

		// Test value methods
		assert.False(t, issue.GetCreatedTime().IsZero())
		assert.False(t, issue.GetUpdatedTime().IsZero())
		assert.False(t, issue.GetDueDateValue().IsZero())

		// Test values match
		assert.Equal(t, *testCreated, issue.GetCreatedTime())
		assert.Equal(t, *testUpdated, issue.GetUpdatedTime())
		assert.Equal(t, *testDueDate, issue.GetDueDateValue())
	})
}

// timePtr is a helper function to create time.Time pointers for tests
func timePtr(year, month, day, hour, min, sec int) *time.Time {
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
	return &t
}
