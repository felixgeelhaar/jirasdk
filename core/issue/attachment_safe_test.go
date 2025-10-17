package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAttachmentSafeHelperMethods tests all safe helper methods for Attachment
func TestAttachmentSafeHelperMethods(t *testing.T) {
	// Test time values
	testCreated := timePtr(2024, 1, 15, 10, 30, 0)
	testAuthor := &User{AccountID: "user1", DisplayName: "John Doe"}

	t.Run("GetAuthor with nil Author", func(t *testing.T) {
		attachment := &Attachment{ID: "10001", Filename: "test.pdf"}
		assert.Nil(t, attachment.GetAuthor())
	})

	t.Run("GetAuthor with populated Author", func(t *testing.T) {
		attachment := &Attachment{
			ID:       "10001",
			Filename: "test.pdf",
			Author:   testAuthor,
		}
		author := attachment.GetAuthor()
		assert.NotNil(t, author)
		assert.Equal(t, testAuthor.AccountID, author.AccountID)
		assert.Equal(t, testAuthor.DisplayName, author.DisplayName)
	})

	t.Run("GetAuthorName with nil Author", func(t *testing.T) {
		attachment := &Attachment{ID: "10001", Filename: "test.pdf"}
		assert.Equal(t, "", attachment.GetAuthorName())
	})

	t.Run("GetAuthorName with populated Author", func(t *testing.T) {
		attachment := &Attachment{
			ID:       "10001",
			Filename: "test.pdf",
			Author:   testAuthor,
		}
		assert.Equal(t, "John Doe", attachment.GetAuthorName())
	})

	t.Run("GetCreated with nil Created", func(t *testing.T) {
		attachment := &Attachment{ID: "10001", Filename: "test.pdf"}
		assert.Nil(t, attachment.GetCreated())
	})

	t.Run("GetCreated with populated Created", func(t *testing.T) {
		attachment := &Attachment{
			ID:       "10001",
			Filename: "test.pdf",
			Created:  testCreated,
		}
		created := attachment.GetCreated()
		assert.NotNil(t, created)
		assert.Equal(t, *testCreated, *created)
	})

	t.Run("GetCreatedTime with nil Created", func(t *testing.T) {
		attachment := &Attachment{ID: "10001", Filename: "test.pdf"}
		created := attachment.GetCreatedTime()
		assert.True(t, created.IsZero())
	})

	t.Run("GetCreatedTime with populated Created", func(t *testing.T) {
		attachment := &Attachment{
			ID:       "10001",
			Filename: "test.pdf",
			Created:  testCreated,
		}
		created := attachment.GetCreatedTime()
		assert.False(t, created.IsZero())
		assert.Equal(t, *testCreated, created)
	})

	t.Run("All fields with complete attachment", func(t *testing.T) {
		attachment := &Attachment{
			ID:       "10001",
			Filename: "complete-test.pdf",
			Author:   testAuthor,
			Created:  testCreated,
			Size:     1024,
			MimeType: "application/pdf",
		}

		// Test pointer methods
		assert.NotNil(t, attachment.GetAuthor())
		assert.Equal(t, "John Doe", attachment.GetAuthorName())
		assert.NotNil(t, attachment.GetCreated())

		// Test value methods
		assert.False(t, attachment.GetCreatedTime().IsZero())

		// Test values match
		assert.Equal(t, *testCreated, attachment.GetCreatedTime())
	})
}
