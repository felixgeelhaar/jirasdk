package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWorklogSafeHelperMethods tests all safe helper methods for Worklog
func TestWorklogSafeHelperMethods(t *testing.T) {
	// Test time values
	testCreated := timePtr(2024, 1, 15, 10, 30, 0)
	testUpdated := timePtr(2024, 2, 20, 14, 45, 0)
	testStarted := timePtr(2024, 1, 15, 9, 0, 0)
	testAuthor := &User{AccountID: "user1", DisplayName: "John Doe"}
	testUpdateAuthor := &User{AccountID: "user2", DisplayName: "Jane Smith"}

	t.Run("GetAuthor with nil Author", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Nil(t, worklog.GetAuthor())
	})

	t.Run("GetAuthor with populated Author", func(t *testing.T) {
		worklog := &Worklog{
			ID:     "10001",
			Author: testAuthor,
		}
		author := worklog.GetAuthor()
		assert.NotNil(t, author)
		assert.Equal(t, testAuthor.AccountID, author.AccountID)
		assert.Equal(t, testAuthor.DisplayName, author.DisplayName)
	})

	t.Run("GetAuthorName with nil Author", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Equal(t, "", worklog.GetAuthorName())
	})

	t.Run("GetAuthorName with populated Author", func(t *testing.T) {
		worklog := &Worklog{
			ID:     "10001",
			Author: testAuthor,
		}
		assert.Equal(t, "John Doe", worklog.GetAuthorName())
	})

	t.Run("GetUpdateAuthor with nil UpdateAuthor", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Nil(t, worklog.GetUpdateAuthor())
	})

	t.Run("GetUpdateAuthor with populated UpdateAuthor", func(t *testing.T) {
		worklog := &Worklog{
			ID:           "10001",
			UpdateAuthor: testUpdateAuthor,
		}
		updateAuthor := worklog.GetUpdateAuthor()
		assert.NotNil(t, updateAuthor)
		assert.Equal(t, testUpdateAuthor.AccountID, updateAuthor.AccountID)
		assert.Equal(t, testUpdateAuthor.DisplayName, updateAuthor.DisplayName)
	})

	t.Run("GetUpdateAuthorName with nil UpdateAuthor", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Equal(t, "", worklog.GetUpdateAuthorName())
	})

	t.Run("GetUpdateAuthorName with populated UpdateAuthor", func(t *testing.T) {
		worklog := &Worklog{
			ID:           "10001",
			UpdateAuthor: testUpdateAuthor,
		}
		assert.Equal(t, "Jane Smith", worklog.GetUpdateAuthorName())
	})

	t.Run("GetCreated with nil Created", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Nil(t, worklog.GetCreated())
	})

	t.Run("GetCreated with populated Created", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Created: testCreated,
		}
		created := worklog.GetCreated()
		assert.NotNil(t, created)
		assert.Equal(t, *testCreated, *created)
	})

	t.Run("GetCreatedTime with nil Created", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		created := worklog.GetCreatedTime()
		assert.True(t, created.IsZero())
	})

	t.Run("GetCreatedTime with populated Created", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Created: testCreated,
		}
		created := worklog.GetCreatedTime()
		assert.False(t, created.IsZero())
		assert.Equal(t, *testCreated, created)
	})

	t.Run("GetUpdated with nil Updated", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Nil(t, worklog.GetUpdated())
	})

	t.Run("GetUpdated with populated Updated", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Updated: testUpdated,
		}
		updated := worklog.GetUpdated()
		assert.NotNil(t, updated)
		assert.Equal(t, *testUpdated, *updated)
	})

	t.Run("GetUpdatedTime with nil Updated", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		updated := worklog.GetUpdatedTime()
		assert.True(t, updated.IsZero())
	})

	t.Run("GetUpdatedTime with populated Updated", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Updated: testUpdated,
		}
		updated := worklog.GetUpdatedTime()
		assert.False(t, updated.IsZero())
		assert.Equal(t, *testUpdated, updated)
	})

	t.Run("GetStarted with nil Started", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		assert.Nil(t, worklog.GetStarted())
	})

	t.Run("GetStarted with populated Started", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Started: testStarted,
		}
		started := worklog.GetStarted()
		assert.NotNil(t, started)
		assert.Equal(t, *testStarted, *started)
	})

	t.Run("GetStartedTime with nil Started", func(t *testing.T) {
		worklog := &Worklog{ID: "10001"}
		started := worklog.GetStartedTime()
		assert.True(t, started.IsZero())
	})

	t.Run("GetStartedTime with populated Started", func(t *testing.T) {
		worklog := &Worklog{
			ID:      "10001",
			Started: testStarted,
		}
		started := worklog.GetStartedTime()
		assert.False(t, started.IsZero())
		assert.Equal(t, *testStarted, started)
	})

	t.Run("All fields with complete worklog", func(t *testing.T) {
		worklog := &Worklog{
			ID:           "10001",
			Author:       testAuthor,
			UpdateAuthor: testUpdateAuthor,
			Created:      testCreated,
			Updated:      testUpdated,
			Started:      testStarted,
			TimeSpent:    "3h 20m",
		}

		// Test pointer methods
		assert.NotNil(t, worklog.GetAuthor())
		assert.Equal(t, "John Doe", worklog.GetAuthorName())
		assert.NotNil(t, worklog.GetUpdateAuthor())
		assert.Equal(t, "Jane Smith", worklog.GetUpdateAuthorName())
		assert.NotNil(t, worklog.GetCreated())
		assert.NotNil(t, worklog.GetUpdated())
		assert.NotNil(t, worklog.GetStarted())

		// Test value methods
		assert.False(t, worklog.GetCreatedTime().IsZero())
		assert.False(t, worklog.GetUpdatedTime().IsZero())
		assert.False(t, worklog.GetStartedTime().IsZero())

		// Test values match
		assert.Equal(t, *testCreated, worklog.GetCreatedTime())
		assert.Equal(t, *testUpdated, worklog.GetUpdatedTime())
		assert.Equal(t, *testStarted, worklog.GetStartedTime())
	})
}
