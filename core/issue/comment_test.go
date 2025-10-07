package issue

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListComments(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		responseBody   *CommentsResult
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful list",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusOK,
			responseBody: &CommentsResult{
				Comments: []*Comment{
					{
						ID:      "10000",
						Body:    "First comment",
						Created: &now,
					},
					{
						ID:      "10001",
						Body:    "Second comment",
						Created: &now,
					},
				},
				Total: 2,
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				_, err := service.ListComments(context.Background(), tt.issueKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/comment")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			comments, err := service.ListComments(context.Background(), tt.issueKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comments)
				assert.Equal(t, len(tt.responseBody.Comments), len(comments))
			}
		})
	}
}

func TestAddComment(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		issueKeyOrID   string
		input          *AddCommentInput
		responseStatus int
		responseBody   *Comment
		wantErr        bool
		errMsg         string
	}{
		{
			name:         "successful add",
			issueKeyOrID: "PROJ-123",
			input: &AddCommentInput{
				Body: "This is a test comment",
			},
			responseStatus: http.StatusCreated,
			responseBody: &Comment{
				ID:      "10000",
				Body:    "This is a test comment",
				Created: &now,
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			input: &AddCommentInput{
				Body: "Comment",
			},
			wantErr: true,
			errMsg:  "issue key or ID is required",
		},
		{
			name:         "nil input",
			issueKeyOrID: "PROJ-123",
			input:        nil,
			wantErr:      true,
			errMsg:       "comment body is required",
		},
		{
			name:         "empty body",
			issueKeyOrID: "PROJ-123",
			input: &AddCommentInput{
				Body: "",
			},
			wantErr: true,
			errMsg:  "comment body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.input == nil || tt.input.Body == "" {
				service := NewService(nil)
				_, err := service.AddComment(context.Background(), tt.issueKeyOrID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/comment")

				var body AddCommentInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Body, body.Body)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			comment, err := service.AddComment(context.Background(), tt.issueKeyOrID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comment)
				assert.Equal(t, tt.responseBody.ID, comment.ID)
				assert.Equal(t, tt.responseBody.Body, comment.Body)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		issueKeyOrID   string
		commentID      string
		input          *UpdateCommentInput
		responseStatus int
		responseBody   *Comment
		wantErr        bool
		errMsg         string
	}{
		{
			name:         "successful update",
			issueKeyOrID: "PROJ-123",
			commentID:    "10000",
			input: &UpdateCommentInput{
				Body: "Updated comment",
			},
			responseStatus: http.StatusOK,
			responseBody: &Comment{
				ID:      "10000",
				Body:    "Updated comment",
				Updated: &now,
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			commentID:    "10000",
			input: &UpdateCommentInput{
				Body: "Updated",
			},
			wantErr: true,
			errMsg:  "issue key or ID is required",
		},
		{
			name:         "empty comment ID",
			issueKeyOrID: "PROJ-123",
			commentID:    "",
			input: &UpdateCommentInput{
				Body: "Updated",
			},
			wantErr: true,
			errMsg:  "comment ID is required",
		},
		{
			name:         "nil input",
			issueKeyOrID: "PROJ-123",
			commentID:    "10000",
			input:        nil,
			wantErr:      true,
			errMsg:       "comment body is required",
		},
		{
			name:         "empty body",
			issueKeyOrID: "PROJ-123",
			commentID:    "10000",
			input: &UpdateCommentInput{
				Body: "",
			},
			wantErr: true,
			errMsg:  "comment body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.commentID == "" || tt.input == nil || tt.input.Body == "" {
				service := NewService(nil)
				_, err := service.UpdateComment(context.Background(), tt.issueKeyOrID, tt.commentID, tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/comment/"+tt.commentID)

				var body UpdateCommentInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Body, body.Body)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			comment, err := service.UpdateComment(context.Background(), tt.issueKeyOrID, tt.commentID, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, comment)
				assert.Equal(t, tt.responseBody.ID, comment.ID)
				assert.Equal(t, tt.responseBody.Body, comment.Body)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		commentID      string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			issueKeyOrID:   "PROJ-123",
			commentID:      "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			commentID:    "10000",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty comment ID",
			issueKeyOrID: "PROJ-123",
			commentID:    "",
			wantErr:      true,
			errMsg:       "comment ID is required",
		},
		{
			name:           "unexpected status code",
			issueKeyOrID:   "PROJ-123",
			commentID:      "10000",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.commentID == "" {
				service := NewService(nil)
				err := service.DeleteComment(context.Background(), tt.issueKeyOrID, tt.commentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/comment/"+tt.commentID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteComment(context.Background(), tt.issueKeyOrID, tt.commentID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
