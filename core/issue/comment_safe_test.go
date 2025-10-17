package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommentSafeHelperMethods tests all safe helper methods for Comment
func TestCommentSafeHelperMethods(t *testing.T) {
	// Test time values
	testCreated := timePtr(2024, 1, 15, 10, 30, 0)
	testUpdated := timePtr(2024, 2, 20, 14, 45, 0)
	testAuthor := &User{AccountID: "user1", DisplayName: "John Doe"}

	t.Run("GetAuthor with nil Author", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		assert.Nil(t, comment.GetAuthor())
	})

	t.Run("GetAuthor with populated Author", func(t *testing.T) {
		comment := &Comment{
			ID:     "10001",
			Body:   "Test comment",
			Author: testAuthor,
		}
		author := comment.GetAuthor()
		assert.NotNil(t, author)
		assert.Equal(t, testAuthor.AccountID, author.AccountID)
		assert.Equal(t, testAuthor.DisplayName, author.DisplayName)
	})

	t.Run("GetAuthorName with nil Author", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		assert.Equal(t, "", comment.GetAuthorName())
	})

	t.Run("GetAuthorName with populated Author", func(t *testing.T) {
		comment := &Comment{
			ID:     "10001",
			Body:   "Test comment",
			Author: testAuthor,
		}
		assert.Equal(t, "John Doe", comment.GetAuthorName())
	})

	t.Run("GetCreated with nil Created", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		assert.Nil(t, comment.GetCreated())
	})

	t.Run("GetCreated with populated Created", func(t *testing.T) {
		comment := &Comment{
			ID:      "10001",
			Body:    "Test comment",
			Created: testCreated,
		}
		created := comment.GetCreated()
		assert.NotNil(t, created)
		assert.Equal(t, *testCreated, *created)
	})

	t.Run("GetCreatedTime with nil Created", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		created := comment.GetCreatedTime()
		assert.True(t, created.IsZero())
	})

	t.Run("GetCreatedTime with populated Created", func(t *testing.T) {
		comment := &Comment{
			ID:      "10001",
			Body:    "Test comment",
			Created: testCreated,
		}
		created := comment.GetCreatedTime()
		assert.False(t, created.IsZero())
		assert.Equal(t, *testCreated, created)
	})

	t.Run("GetUpdated with nil Updated", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		assert.Nil(t, comment.GetUpdated())
	})

	t.Run("GetUpdated with populated Updated", func(t *testing.T) {
		comment := &Comment{
			ID:      "10001",
			Body:    "Test comment",
			Updated: testUpdated,
		}
		updated := comment.GetUpdated()
		assert.NotNil(t, updated)
		assert.Equal(t, *testUpdated, *updated)
	})

	t.Run("GetUpdatedTime with nil Updated", func(t *testing.T) {
		comment := &Comment{ID: "10001", Body: "Test comment"}
		updated := comment.GetUpdatedTime()
		assert.True(t, updated.IsZero())
	})

	t.Run("GetUpdatedTime with populated Updated", func(t *testing.T) {
		comment := &Comment{
			ID:      "10001",
			Body:    "Test comment",
			Updated: testUpdated,
		}
		updated := comment.GetUpdatedTime()
		assert.False(t, updated.IsZero())
		assert.Equal(t, *testUpdated, updated)
	})

	t.Run("All fields with complete comment", func(t *testing.T) {
		comment := &Comment{
			ID:      "10001",
			Body:    "Complete test comment",
			Author:  testAuthor,
			Created: testCreated,
			Updated: testUpdated,
		}

		// Test pointer methods
		assert.NotNil(t, comment.GetAuthor())
		assert.Equal(t, "John Doe", comment.GetAuthorName())
		assert.NotNil(t, comment.GetCreated())
		assert.NotNil(t, comment.GetUpdated())

		// Test value methods
		assert.False(t, comment.GetCreatedTime().IsZero())
		assert.False(t, comment.GetUpdatedTime().IsZero())

		// Test values match
		assert.Equal(t, *testCreated, comment.GetCreatedTime())
		assert.Equal(t, *testUpdated, comment.GetUpdatedTime())
	})
}
