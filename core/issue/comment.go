package issue

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Comment represents an issue comment.
type Comment struct {
	ID      string     `json:"id"`
	Self    string     `json:"self,omitempty"`
	Author  *User      `json:"author,omitempty"`
	Body    string     `json:"body"`
	Created *time.Time `json:"created,omitempty"`
	Updated *time.Time `json:"updated,omitempty"`
}

// GetAuthor safely retrieves the comment author.
// Returns nil if Author is nil.
func (c *Comment) GetAuthor() *User {
	return c.Author
}

// GetAuthorName safely retrieves the author display name.
// Returns an empty string if Author or Author.DisplayName is not available.
func (c *Comment) GetAuthorName() string {
	if c.Author == nil {
		return ""
	}
	return c.Author.DisplayName
}

// GetCreated safely retrieves the comment creation timestamp.
// Returns nil if Created is nil.
func (c *Comment) GetCreated() *time.Time {
	return c.Created
}

// GetCreatedTime safely retrieves the comment creation timestamp as a value.
// Returns zero time (time.Time{}) if Created is nil.
// Use this method when you need a time.Time value instead of a pointer.
func (c *Comment) GetCreatedTime() time.Time {
	if c.Created == nil {
		return time.Time{}
	}
	return *c.Created
}

// GetUpdated safely retrieves the comment last update timestamp.
// Returns nil if Updated is nil.
func (c *Comment) GetUpdated() *time.Time {
	return c.Updated
}

// GetUpdatedTime safely retrieves the comment last update timestamp as a value.
// Returns zero time (time.Time{}) if Updated is nil.
// Use this method when you need a time.Time value instead of a pointer.
func (c *Comment) GetUpdatedTime() time.Time {
	if c.Updated == nil {
		return time.Time{}
	}
	return *c.Updated
}

// CommentsResult contains a list of comments with pagination.
type CommentsResult struct {
	Comments   []*Comment `json:"comments"`
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Total      int        `json:"total"`
}

// ListComments retrieves all comments for an issue.
//
// Example:
//
//	comments, err := client.Issue.ListComments(ctx, "PROJ-123")
func (s *Service) ListComments(ctx context.Context, issueKeyOrID string) ([]*Comment, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/comment", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var result CommentsResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Comments, nil
}

// AddCommentInput contains the data for adding a comment.
type AddCommentInput struct {
	Body string `json:"body"`
}

// AddComment adds a comment to an issue.
//
// Example:
//
//	comment, err := client.Issue.AddComment(ctx, "PROJ-123", &issue.AddCommentInput{
//		Body: "This is a comment",
//	})
func (s *Service) AddComment(ctx context.Context, issueKeyOrID string, input *AddCommentInput) (*Comment, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if input == nil || input.Body == "" {
		return nil, fmt.Errorf("comment body is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/comment", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var comment Comment
	if err := s.transport.DecodeResponse(resp, &comment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &comment, nil
}

// UpdateCommentInput contains the data for updating a comment.
type UpdateCommentInput struct {
	Body string `json:"body"`
}

// UpdateComment updates an existing comment.
//
// Example:
//
//	comment, err := client.Issue.UpdateComment(ctx, "PROJ-123", "10000", &issue.UpdateCommentInput{
//		Body: "Updated comment",
//	})
func (s *Service) UpdateComment(ctx context.Context, issueKeyOrID, commentID string, input *UpdateCommentInput) (*Comment, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if commentID == "" {
		return nil, fmt.Errorf("comment ID is required")
	}

	if input == nil || input.Body == "" {
		return nil, fmt.Errorf("comment body is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/comment/%s", issueKeyOrID, commentID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var comment Comment
	if err := s.transport.DecodeResponse(resp, &comment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &comment, nil
}

// DeleteComment deletes a comment from an issue.
//
// Example:
//
//	err := client.Issue.DeleteComment(ctx, "PROJ-123", "10000")
func (s *Service) DeleteComment(ctx context.Context, issueKeyOrID, commentID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if commentID == "" {
		return fmt.Errorf("comment ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/comment/%s", issueKeyOrID, commentID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Close response body
	defer resp.Body.Close()

	// Delete returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
