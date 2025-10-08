package issue

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Attachment represents a file attached to an issue.
type Attachment struct {
	ID        string     `json:"id"`
	Self      string     `json:"self,omitempty"`
	Filename  string     `json:"filename"`
	Author    *User      `json:"author,omitempty"`
	Created   *time.Time `json:"created,omitempty"`
	Size      int64      `json:"size"`
	MimeType  string     `json:"mimeType"`
	Content   string     `json:"content,omitempty"`
	Thumbnail string     `json:"thumbnail,omitempty"`
}

// AttachmentMetadata contains metadata for uploading an attachment.
type AttachmentMetadata struct {
	Filename string
	Content  io.Reader
}

// AddAttachment uploads an attachment to an issue.
//
// Example:
//
//	file, _ := os.Open("report.pdf")
//	defer file.Close()
//
//	attachments, err := client.Issue.AddAttachment(ctx, "PROJ-123", &issue.AttachmentMetadata{
//		Filename: "report.pdf",
//		Content:  file,
//	})
func (s *Service) AddAttachment(ctx context.Context, issueKeyOrID string, attachment *AttachmentMetadata) ([]*Attachment, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	if attachment == nil {
		return nil, fmt.Errorf("attachment metadata is required")
	}

	if attachment.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	if attachment.Content == nil {
		return nil, fmt.Errorf("content is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/attachments", issueKeyOrID)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("file", attachment.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content
	if _, err := io.Copy(part, attachment.Content); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request with multipart body
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set multipart content type and body
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "no-check") // Required for attachment uploads
	req.Body = io.NopCloser(body)
	req.ContentLength = int64(body.Len())

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Decode response
	var attachments []*Attachment
	if err := s.transport.DecodeResponse(resp, &attachments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return attachments, nil
}

// GetAttachment retrieves metadata for a specific attachment.
//
// Example:
//
//	attachment, err := client.Issue.GetAttachment(ctx, "10000")
func (s *Service) GetAttachment(ctx context.Context, attachmentID string) (*Attachment, error) {
	if attachmentID == "" {
		return nil, fmt.Errorf("attachment ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/attachment/%s", attachmentID)

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
	var attachment Attachment
	if err := s.transport.DecodeResponse(resp, &attachment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &attachment, nil
}

// DownloadAttachment downloads the content of an attachment.
//
// Example:
//
//	content, err := client.Issue.DownloadAttachment(ctx, "10000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer content.Close()
//
//	// Save to file
//	file, _ := os.Create("downloaded.pdf")
//	defer file.Close()
//	io.Copy(file, content)
func (s *Service) DownloadAttachment(ctx context.Context, attachmentID string) (io.ReadCloser, error) {
	if attachmentID == "" {
		return nil, fmt.Errorf("attachment ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/attachment/content/%s", attachmentID)

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

	// Check status code
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Return response body as ReadCloser
	return resp.Body, nil
}

// DeleteAttachment removes an attachment from Jira.
//
// Example:
//
//	err := client.Issue.DeleteAttachment(ctx, "10000")
func (s *Service) DeleteAttachment(ctx context.Context, attachmentID string) error {
	if attachmentID == "" {
		return fmt.Errorf("attachment ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/attachment/%s", attachmentID)

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
