package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnvironmentSafeHelperMethods tests all safe helper methods for Environment field
func TestEnvironmentSafeHelperMethods(t *testing.T) {
	t.Run("GetEnvironment with nil Fields", func(t *testing.T) {
		issue := &Issue{Key: "PROJ-123"}
		assert.Nil(t, issue.GetEnvironment())
	})

	t.Run("GetEnvironment with nil Environment", func(t *testing.T) {
		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary: "Test issue",
			},
		}
		assert.Nil(t, issue.GetEnvironment())
	})

	t.Run("GetEnvironment with populated Environment", func(t *testing.T) {
		envADF := ADFFromText("Production environment")
		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Test issue",
				Environment: envADF,
			},
		}
		env := issue.GetEnvironment()
		assert.NotNil(t, env)
		assert.Equal(t, "Production environment", env.ToText())
	})

	t.Run("GetEnvironmentText with nil Fields", func(t *testing.T) {
		issue := &Issue{Key: "PROJ-123"}
		assert.Equal(t, "", issue.GetEnvironmentText())
	})

	t.Run("GetEnvironmentText with nil Environment", func(t *testing.T) {
		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary: "Test issue",
			},
		}
		assert.Equal(t, "", issue.GetEnvironmentText())
	})

	t.Run("GetEnvironmentText with populated Environment", func(t *testing.T) {
		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Test issue",
				Environment: ADFFromText("Staging environment"),
			},
		}
		assert.Equal(t, "Staging environment", issue.GetEnvironmentText())
	})

	t.Run("GetEnvironmentText with rich ADF Environment", func(t *testing.T) {
		richEnv := NewADF().
			AddHeading("Environment Details", 3).
			AddParagraph("OS: Ubuntu 22.04").
			AddParagraph("Browser: Chrome 120")

		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Test issue",
				Environment: richEnv,
			},
		}
		text := issue.GetEnvironmentText()
		assert.Contains(t, text, "Environment Details")
		assert.Contains(t, text, "Ubuntu 22.04")
		assert.Contains(t, text, "Chrome 120")
	})
}

// TestIssueFieldsEnvironmentSetters tests the setter methods for Environment
func TestIssueFieldsEnvironmentSetters(t *testing.T) {
	t.Run("SetEnvironmentText creates ADF from plain text", func(t *testing.T) {
		fields := &IssueFields{
			Summary: "Test issue",
		}
		fields.SetEnvironmentText("Development environment")

		assert.NotNil(t, fields.Environment)
		assert.Equal(t, "Development environment", fields.Environment.ToText())
	})

	t.Run("SetEnvironment sets ADF directly", func(t *testing.T) {
		fields := &IssueFields{
			Summary: "Test issue",
		}
		envADF := NewADF().
			AddParagraph("Test environment").
			AddParagraph("Node.js 18.x")

		fields.SetEnvironment(envADF)

		assert.NotNil(t, fields.Environment)
		assert.Equal(t, envADF, fields.Environment)
		text := fields.Environment.ToText()
		assert.Contains(t, text, "Test environment")
		assert.Contains(t, text, "Node.js 18.x")
	})

	t.Run("SetEnvironmentText with empty string", func(t *testing.T) {
		fields := &IssueFields{
			Summary: "Test issue",
		}
		fields.SetEnvironmentText("")

		assert.NotNil(t, fields.Environment)
		// Empty text should create empty ADF
		assert.True(t, fields.Environment.IsEmpty())
	})

	t.Run("SetEnvironment with nil", func(t *testing.T) {
		fields := &IssueFields{
			Summary: "Test issue",
		}
		fields.SetEnvironment(nil)

		assert.Nil(t, fields.Environment)
	})

	t.Run("SetEnvironmentText overwrites existing Environment", func(t *testing.T) {
		fields := &IssueFields{
			Summary:     "Test issue",
			Environment: ADFFromText("Old environment"),
		}

		fields.SetEnvironmentText("New environment")

		assert.NotNil(t, fields.Environment)
		assert.Equal(t, "New environment", fields.Environment.ToText())
	})

	t.Run("SetEnvironment overwrites existing Environment", func(t *testing.T) {
		fields := &IssueFields{
			Summary:     "Test issue",
			Environment: ADFFromText("Old environment"),
		}

		newEnv := NewADF().AddParagraph("Completely new environment")
		fields.SetEnvironment(newEnv)

		assert.NotNil(t, fields.Environment)
		assert.Equal(t, newEnv, fields.Environment)
		assert.Equal(t, "Completely new environment", fields.Environment.ToText())
	})
}

// TestEnvironmentIntegration tests Environment field in complete issue workflows
func TestEnvironmentIntegration(t *testing.T) {
	t.Run("Complete issue with all Environment methods", func(t *testing.T) {
		// Create issue with environment
		issue := &Issue{
			Key: "PROJ-123",
			Fields: &IssueFields{
				Summary:     "Bug in production",
				Environment: ADFFromText("Production - AWS us-east-1"),
			},
		}

		// Test getter methods
		assert.NotNil(t, issue.GetEnvironment())
		assert.Equal(t, "Production - AWS us-east-1", issue.GetEnvironmentText())

		// Update environment using setter
		issue.Fields.SetEnvironmentText("Staging - AWS us-west-2")
		assert.Equal(t, "Staging - AWS us-west-2", issue.GetEnvironmentText())

		// Update with rich ADF
		richEnv := NewADF().
			AddHeading("Environment", 3).
			AddParagraph("Region: us-west-2").
			AddParagraph("Instance: i-1234567890abcdef0")

		issue.Fields.SetEnvironment(richEnv)
		text := issue.GetEnvironmentText()
		assert.Contains(t, text, "Environment")
		assert.Contains(t, text, "us-west-2")
		assert.Contains(t, text, "i-1234567890abcdef0")

		// Clear environment
		issue.Fields.SetEnvironment(nil)
		assert.Nil(t, issue.GetEnvironment())
		assert.Equal(t, "", issue.GetEnvironmentText())
	})
}
