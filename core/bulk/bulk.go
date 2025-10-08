// Package bulk provides bulk operations for Jira resources.
//
// This package implements Jira's bulk operation APIs for efficient batch processing
// of issues and other resources. Bulk operations allow you to perform actions on
// multiple items in a single API call, reducing network overhead and improving performance.
//
// Note: Jira has a strict limit of 1000 issues per bulk operation request.
package bulk

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Service provides bulk operations for Jira resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper is the interface for executing HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, target interface{}) error
}

// NewService creates a new bulk operations service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

const (
	// MaxBulkIssues is the maximum number of issues that can be processed in a single bulk operation
	MaxBulkIssues = 1000

	// BulkOperationStatusRunning indicates the operation is in progress
	BulkOperationStatusRunning = "RUNNING"
	// BulkOperationStatusComplete indicates the operation completed successfully
	BulkOperationStatusComplete = "COMPLETE"
	// BulkOperationStatusFailed indicates the operation failed
	BulkOperationStatusFailed = "FAILED"
	// BulkOperationStatusCancelled indicates the operation was cancelled
	BulkOperationStatusCancelled = "CANCELLED"
)

// IssueUpdate represents a single issue update in a bulk operation.
type IssueUpdate struct {
	// Fields contains the field values to update
	Fields map[string]interface{} `json:"fields"`

	// Update contains field update operations (add, set, remove)
	Update map[string][]FieldOperation `json:"update,omitempty"`

	// HistoryMetadata contains metadata for the change
	HistoryMetadata *HistoryMetadata `json:"historyMetadata,omitempty"`

	// Properties contains issue properties to set
	Properties []EntityProperty `json:"properties,omitempty"`
}

// FieldOperation represents an operation on a field (add, set, remove).
type FieldOperation struct {
	// Add adds a value to a multi-value field
	Add interface{} `json:"add,omitempty"`

	// Set sets the value of a field
	Set interface{} `json:"set,omitempty"`

	// Remove removes a value from a multi-value field
	Remove interface{} `json:"remove,omitempty"`
}

// HistoryMetadata provides metadata about the change.
type HistoryMetadata struct {
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Actor       *HistoryMetadataActor  `json:"actor,omitempty"`
	Cause       *HistoryMetadataCause  `json:"cause,omitempty"`
	ExtraData   map[string]interface{} `json:"extraData,omitempty"`
}

// HistoryMetadataActor represents the actor in history metadata.
type HistoryMetadataActor struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Type        string `json:"type,omitempty"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	URL         string `json:"url,omitempty"`
}

// HistoryMetadataCause represents the cause in history metadata.
type HistoryMetadataCause struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// EntityProperty represents an entity property.
type EntityProperty struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// CreateIssuesInput contains the input for creating multiple issues.
type CreateIssuesInput struct {
	// IssueUpdates contains the issues to create
	IssueUpdates []IssueUpdate `json:"issueUpdates"`
}

// CreateIssuesResult contains the result of a bulk create operation.
type CreateIssuesResult struct {
	// Issues contains the created issues
	Issues []CreatedIssue `json:"issues,omitempty"`

	// Errors contains any errors that occurred
	Errors []BulkOperationError `json:"errors,omitempty"`
}

// CreatedIssue represents a successfully created issue.
type CreatedIssue struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// BulkOperationError represents an error in a bulk operation.
type BulkOperationError struct {
	// Status is the HTTP status code
	Status int `json:"status,omitempty"`

	// ElementErrors contains field-specific errors
	ElementErrors *ElementErrors `json:"elementErrors,omitempty"`

	// FailedElementNumber indicates which element in the request failed
	FailedElementNumber int `json:"failedElementNumber,omitempty"`
}

// ElementErrors contains field-specific validation errors.
type ElementErrors struct {
	ErrorMessages []string          `json:"errorMessages,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
}

// EditIssuesInput contains the input for bulk editing issues.
type EditIssuesInput struct {
	// IssueUpdates maps issue IDs/keys to their updates
	IssueUpdates map[string]IssueUpdate `json:"issueUpdates"`
}

// BulkOperationProgress represents the progress of a long-running bulk operation.
type BulkOperationProgress struct {
	// TaskID is the unique identifier for this operation
	TaskID string `json:"taskId"`

	// Status indicates the current status
	Status string `json:"status"`

	// ProgressPercent indicates completion percentage (0-100)
	ProgressPercent int `json:"progressPercent,omitempty"`

	// Message provides additional information
	Message string `json:"message,omitempty"`

	// Result contains the operation result (when complete)
	Result *BulkOperationResult `json:"result,omitempty"`

	// SubmittedBy indicates who submitted the operation
	SubmittedBy *User `json:"submittedBy,omitempty"`

	// Created is when the operation was created
	Created int64 `json:"created,omitempty"`

	// Started is when the operation started processing
	Started int64 `json:"started,omitempty"`

	// Updated is when the operation was last updated
	Updated int64 `json:"updated,omitempty"`

	// Completed is when the operation finished
	Completed int64 `json:"completed,omitempty"`
}

// BulkOperationResult contains the result of a bulk operation.
type BulkOperationResult struct {
	// SuccessCount is the number of successfully processed items
	SuccessCount int `json:"successCount,omitempty"`

	// ErrorCount is the number of failed items
	ErrorCount int `json:"errorCount,omitempty"`

	// Errors contains details of any failures
	Errors []BulkOperationError `json:"errors,omitempty"`
}

// User represents a minimal user reference.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// CreateIssues creates multiple issues in a single request.
//
// Note: Maximum 1000 issues per request.
//
// Example:
//
//	result, err := client.Bulk.CreateIssues(ctx, &bulk.CreateIssuesInput{
//	    IssueUpdates: []bulk.IssueUpdate{
//	        {
//	            Fields: map[string]interface{}{
//	                "project":   map[string]string{"key": "PROJ"},
//	                "summary":   "Bulk created issue 1",
//	                "issuetype": map[string]string{"name": "Task"},
//	            },
//	        },
//	        {
//	            Fields: map[string]interface{}{
//	                "project":   map[string]string{"key": "PROJ"},
//	                "summary":   "Bulk created issue 2",
//	                "issuetype": map[string]string{"name": "Task"},
//	            },
//	        },
//	    },
//	})
func (s *Service) CreateIssues(ctx context.Context, input *CreateIssuesInput) (*CreateIssuesResult, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if len(input.IssueUpdates) == 0 {
		return nil, fmt.Errorf("at least one issue update is required")
	}

	if len(input.IssueUpdates) > MaxBulkIssues {
		return nil, fmt.Errorf("cannot create more than %d issues in a single request", MaxBulkIssues)
	}

	path := "/rest/api/3/issue/bulk"

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
	var result CreateIssuesResult
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteIssuesInput contains the input for deleting multiple issues.
type DeleteIssuesInput struct {
	// IssueIDs contains the issue IDs or keys to delete
	IssueIDs []string `json:"issueIdsOrKeys"`
}

// DeleteIssues deletes multiple issues in a single request.
//
// Note: Maximum 1000 issues per request.
// Warning: This operation cannot be undone.
//
// Example:
//
//	err := client.Bulk.DeleteIssues(ctx, &bulk.DeleteIssuesInput{
//	    IssueIDs: []string{"PROJ-123", "PROJ-124", "PROJ-125"},
//	})
func (s *Service) DeleteIssues(ctx context.Context, input *DeleteIssuesInput) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}

	if len(input.IssueIDs) == 0 {
		return fmt.Errorf("at least one issue ID is required")
	}

	if len(input.IssueIDs) > MaxBulkIssues {
		return fmt.Errorf("cannot delete more than %d issues in a single request", MaxBulkIssues)
	}

	path := "/rest/api/3/issue/bulk"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, input)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Check for successful deletion (204 No Content)
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// WaitForCompletion polls for the completion of a long-running bulk operation.
//
// This function blocks until the operation completes, fails, or the context is cancelled.
// It polls every pollInterval until the operation reaches a terminal state.
//
// Example:
//
//	progress, err := client.Bulk.WaitForCompletion(ctx, taskID, 5*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Operation completed with %d successes and %d errors\n",
//	    progress.Result.SuccessCount, progress.Result.ErrorCount)
func (s *Service) WaitForCompletion(ctx context.Context, taskID string, pollInterval time.Duration) (*BulkOperationProgress, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			progress, err := s.GetProgress(ctx, taskID)
			if err != nil {
				return nil, err
			}

			// Check if operation has reached a terminal state
			switch progress.Status {
			case BulkOperationStatusComplete, BulkOperationStatusFailed, BulkOperationStatusCancelled:
				return progress, nil
			}
		}
	}
}

// GetProgress retrieves the current progress of a long-running bulk operation.
//
// Example:
//
//	progress, err := client.Bulk.GetProgress(ctx, taskID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Operation is %d%% complete\n", progress.ProgressPercent)
func (s *Service) GetProgress(ctx context.Context, taskID string) (*BulkOperationProgress, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/task/%s", taskID)

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
	var progress BulkOperationProgress
	if err := s.transport.DecodeResponse(resp, &progress); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &progress, nil
}
