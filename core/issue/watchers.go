package issue

import (
	"context"
	"fmt"
	"net/http"
)

// Watchers represents the watchers of an issue.
type Watchers struct {
	Self       string  `json:"self,omitempty"`
	IsWatching bool    `json:"isWatching"`
	WatchCount int     `json:"watchCount"`
	Watchers   []*User `json:"watchers,omitempty"`
}

// GetWatchers retrieves the watchers for an issue.
//
// Example:
//
//	watchers, err := client.Issue.GetWatchers(ctx, "PROJ-123")
func (s *Service) GetWatchers(ctx context.Context, issueKeyOrID string) (*Watchers, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/watchers", issueKeyOrID)

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
	var watchers Watchers
	if err := s.transport.DecodeResponse(resp, &watchers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &watchers, nil
}

// AddWatcher adds a user to the watchers list.
//
// Example:
//
//	err := client.Issue.AddWatcher(ctx, "PROJ-123", "account-id")
func (s *Service) AddWatcher(ctx context.Context, issueKeyOrID, accountID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/watchers", issueKeyOrID)

	// Request body is just the account ID as a JSON string
	body := fmt.Sprintf(`"%s"`, accountID)

	// Create request manually with string body
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Override body with raw JSON string
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody
	req.ContentLength = int64(len(body))

	// Execute request
	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Close response body
	defer resp.Body.Close()

	// Add watcher returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// RemoveWatcher removes a user from the watchers list.
//
// Example:
//
//	err := client.Issue.RemoveWatcher(ctx, "PROJ-123", "account-id")
func (s *Service) RemoveWatcher(ctx context.Context, issueKeyOrID, accountID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/watchers?accountId=%s", issueKeyOrID, accountID)

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

	// Remove watcher returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Votes represents the votes for an issue.
type Votes struct {
	Self     string  `json:"self,omitempty"`
	Votes    int     `json:"votes"`
	HasVoted bool    `json:"hasVoted"`
	Voters   []*User `json:"voters,omitempty"`
}

// GetVotes retrieves the votes for an issue.
//
// Example:
//
//	votes, err := client.Issue.GetVotes(ctx, "PROJ-123")
func (s *Service) GetVotes(ctx context.Context, issueKeyOrID string) (*Votes, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/votes", issueKeyOrID)

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
	var votes Votes
	if err := s.transport.DecodeResponse(resp, &votes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &votes, nil
}

// AddVote adds a vote from the current user.
//
// Example:
//
//	err := client.Issue.AddVote(ctx, "PROJ-123")
func (s *Service) AddVote(ctx context.Context, issueKeyOrID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/votes", issueKeyOrID)

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, nil)
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

	// Add vote returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// RemoveVote removes the vote from the current user.
//
// Example:
//
//	err := client.Issue.RemoveVote(ctx, "PROJ-123")
func (s *Service) RemoveVote(ctx context.Context, issueKeyOrID string) error {
	if issueKeyOrID == "" {
		return fmt.Errorf("issue key or ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issue/%s/votes", issueKeyOrID)

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

	// Remove vote returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
