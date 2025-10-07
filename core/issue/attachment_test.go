package issue

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAttachment(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		issueKeyOrID   string
		attachment     *AttachmentMetadata
		responseStatus int
		responseBody   []*Attachment
		wantErr        bool
		errMsg         string
	}{
		{
			name:         "successful upload",
			issueKeyOrID: "PROJ-123",
			attachment: &AttachmentMetadata{
				Filename: "test.pdf",
				Content:  strings.NewReader("file content"),
			},
			responseStatus: http.StatusOK,
			responseBody: []*Attachment{
				{
					ID:       "10000",
					Filename: "test.pdf",
					Size:     12,
					MimeType: "application/pdf",
					Created:  &now,
				},
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			attachment: &AttachmentMetadata{
				Filename: "test.pdf",
				Content:  strings.NewReader("content"),
			},
			wantErr: true,
			errMsg:  "issue key or ID is required",
		},
		{
			name:         "nil attachment",
			issueKeyOrID: "PROJ-123",
			attachment:   nil,
			wantErr:      true,
			errMsg:       "attachment metadata is required",
		},
		{
			name:         "empty filename",
			issueKeyOrID: "PROJ-123",
			attachment: &AttachmentMetadata{
				Filename: "",
				Content:  strings.NewReader("content"),
			},
			wantErr: true,
			errMsg:  "filename is required",
		},
		{
			name:         "nil content",
			issueKeyOrID: "PROJ-123",
			attachment: &AttachmentMetadata{
				Filename: "test.pdf",
				Content:  nil,
			},
			wantErr: true,
			errMsg:  "content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.attachment == nil || tt.attachment.Filename == "" || tt.attachment.Content == nil {
				service := NewService(nil)
				_, err := service.AddAttachment(context.Background(), tt.issueKeyOrID, tt.attachment)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/attachments")
				assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
				assert.Equal(t, "no-check", r.Header.Get("X-Atlassian-Token"))

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			attachments, err := service.AddAttachment(context.Background(), tt.issueKeyOrID, tt.attachment)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, attachments)
				assert.Equal(t, len(tt.responseBody), len(attachments))
				if len(attachments) > 0 {
					assert.Equal(t, tt.responseBody[0].ID, attachments[0].ID)
					assert.Equal(t, tt.responseBody[0].Filename, attachments[0].Filename)
				}
			}
		})
	}
}

func TestGetAttachment(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		attachmentID   string
		responseStatus int
		responseBody   *Attachment
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			attachmentID:   "10000",
			responseStatus: http.StatusOK,
			responseBody: &Attachment{
				ID:       "10000",
				Filename: "report.pdf",
				Size:     1024,
				MimeType: "application/pdf",
				Created:  &now,
			},
			wantErr: false,
		},
		{
			name:         "empty attachment ID",
			attachmentID: "",
			wantErr:      true,
			errMsg:       "attachment ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attachmentID == "" {
				service := NewService(nil)
				_, err := service.GetAttachment(context.Background(), tt.attachmentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/attachment/"+tt.attachmentID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			attachment, err := service.GetAttachment(context.Background(), tt.attachmentID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, attachment)
				assert.Equal(t, tt.responseBody.ID, attachment.ID)
				assert.Equal(t, tt.responseBody.Filename, attachment.Filename)
			}
		})
	}
}

func TestDownloadAttachment(t *testing.T) {
	tests := []struct {
		name           string
		attachmentID   string
		responseStatus int
		responseBody   string
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful download",
			attachmentID:   "10000",
			responseStatus: http.StatusOK,
			responseBody:   "file content here",
			wantErr:        false,
		},
		{
			name:         "empty attachment ID",
			attachmentID: "",
			wantErr:      true,
			errMsg:       "attachment ID is required",
		},
		{
			name:           "not found",
			attachmentID:   "10000",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attachmentID == "" {
				service := NewService(nil)
				_, err := service.DownloadAttachment(context.Background(), tt.attachmentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/attachment/content/"+tt.attachmentID)

				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus == http.StatusOK {
					w.Write([]byte(tt.responseBody))
				}
			})
			defer transport.Close()

			service := NewService(transport)
			content, err := service.DownloadAttachment(context.Background(), tt.attachmentID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, content)
				defer content.Close()

				// Read and verify content
				body := &bytes.Buffer{}
				_, err := io.Copy(body, content)
				require.NoError(t, err)
				assert.Equal(t, tt.responseBody, body.String())
			}
		})
	}
}

func TestDeleteAttachment(t *testing.T) {
	tests := []struct {
		name           string
		attachmentID   string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			attachmentID:   "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty attachment ID",
			attachmentID: "",
			wantErr:      true,
			errMsg:       "attachment ID is required",
		},
		{
			name:           "unexpected status code",
			attachmentID:   "10000",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attachmentID == "" {
				service := NewService(nil)
				err := service.DeleteAttachment(context.Background(), tt.attachmentID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/attachment/"+tt.attachmentID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteAttachment(context.Background(), tt.attachmentID)

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
