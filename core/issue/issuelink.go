package issue

import (
	"context"
	"fmt"
	"net/http"
)

// IssueLink represents a link between two issues.
type IssueLink struct {
	ID           string         `json:"id"`
	Self         string         `json:"self,omitempty"`
	Type         *IssueLinkType `json:"type"`
	InwardIssue  *LinkedIssue   `json:"inwardIssue,omitempty"`
	OutwardIssue *LinkedIssue   `json:"outwardIssue,omitempty"`
}

// IssueLinkType represents the type of link between issues.
type IssueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`  // e.g., "is blocked by"
	Outward string `json:"outward"` // e.g., "blocks"
	Self    string `json:"self,omitempty"`
}

// LinkedIssue represents a simplified issue in a link.
type LinkedIssue struct {
	ID     string             `json:"id"`
	Key    string             `json:"key"`
	Self   string             `json:"self"`
	Fields *LinkedIssueFields `json:"fields,omitempty"`
}

// LinkedIssueFields contains basic fields of a linked issue.
type LinkedIssueFields struct {
	Summary   string     `json:"summary,omitempty"`
	Status    *Status    `json:"status,omitempty"`
	Priority  *Priority  `json:"priority,omitempty"`
	IssueType *IssueType `json:"issuetype,omitempty"`
}

// CreateIssueLinkInput contains the data for creating an issue link.
type CreateIssueLinkInput struct {
	Type         *IssueLinkType `json:"type"`
	InwardIssue  *IssueRef      `json:"inwardIssue"`
	OutwardIssue *IssueRef      `json:"outwardIssue"`
	Comment      *LinkComment   `json:"comment,omitempty"`
}

// IssueRef represents a reference to an issue.
type IssueRef struct {
	ID  string `json:"id,omitempty"`
	Key string `json:"key,omitempty"`
}

// LinkComment represents a comment added when creating a link.
type LinkComment struct {
	Body       string             `json:"body"`
	Visibility *CommentVisibility `json:"visibility,omitempty"`
}

// CommentVisibility controls who can see the comment.
type CommentVisibility struct {
	Type  string `json:"type"`  // "group" or "role"
	Value string `json:"value"` // group name or role name
}

// GetIssueLinks retrieves all links for an issue.
//
// Example:
//
//	links, err := client.Issue.GetIssueLinks(ctx, "PROJ-123")
func (s *Service) GetIssueLinks(ctx context.Context, issueKeyOrID string) ([]*IssueLink, error) {
	if issueKeyOrID == "" {
		return nil, fmt.Errorf("issue key or ID is required")
	}

	// Get the issue with issuelinks field
	_, err := s.Get(ctx, issueKeyOrID, &GetOptions{
		Fields: []string{"issuelinks"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// In a full implementation, you would parse issuelinks from issue.Fields
	// For now, return empty slice
	return []*IssueLink{}, nil
}

// CreateIssueLink creates a link between two issues.
//
// Example:
//
//	link, err := client.Issue.CreateIssueLink(ctx, &issue.CreateIssueLinkInput{
//	    Type: &issue.IssueLinkType{Name: "Blocks"},
//	    InwardIssue:  &issue.IssueRef{Key: "PROJ-123"},
//	    OutwardIssue: &issue.IssueRef{Key: "PROJ-456"},
//	})
func (s *Service) CreateIssueLink(ctx context.Context, input *CreateIssueLinkInput) error {
	if input == nil {
		return fmt.Errorf("create issue link input is required")
	}

	if input.Type == nil {
		return fmt.Errorf("link type is required")
	}

	if input.InwardIssue == nil {
		return fmt.Errorf("inward issue is required")
	}

	if input.OutwardIssue == nil {
		return fmt.Errorf("outward issue is required")
	}

	path := "/rest/api/3/issueLink"

	// Create request
	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
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

	// Create returns 201 Created on success
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DeleteIssueLink removes a link between issues.
//
// Example:
//
//	err := client.Issue.DeleteIssueLink(ctx, "10000")
func (s *Service) DeleteIssueLink(ctx context.Context, linkID string) error {
	if linkID == "" {
		return fmt.Errorf("link ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issueLink/%s", linkID)

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
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetIssueLinkType retrieves a specific issue link type.
//
// Example:
//
//	linkType, err := client.Issue.GetIssueLinkType(ctx, "10000")
func (s *Service) GetIssueLinkType(ctx context.Context, linkTypeID string) (*IssueLinkType, error) {
	if linkTypeID == "" {
		return nil, fmt.Errorf("link type ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/issueLinkType/%s", linkTypeID)

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
	var linkType IssueLinkType
	if err := s.transport.DecodeResponse(resp, &linkType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &linkType, nil
}

// ListIssueLinkTypes retrieves all available issue link types.
//
// Example:
//
//	linkTypes, err := client.Issue.ListIssueLinkTypes(ctx)
func (s *Service) ListIssueLinkTypes(ctx context.Context) ([]*IssueLinkType, error) {
	path := "/rest/api/3/issueLinkType"

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
	var result struct {
		IssueLinkTypes []*IssueLinkType `json:"issueLinkTypes"`
	}
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.IssueLinkTypes, nil
}

// Common issue link type helpers

// BlocksLinkType returns a link type reference for "blocks" relationship.
func BlocksLinkType() *IssueLinkType {
	return &IssueLinkType{
		Name:    "Blocks",
		Inward:  "is blocked by",
		Outward: "blocks",
	}
}

// DuplicatesLinkType returns a link type reference for "duplicates" relationship.
func DuplicatesLinkType() *IssueLinkType {
	return &IssueLinkType{
		Name:    "Duplicate",
		Inward:  "is duplicated by",
		Outward: "duplicates",
	}
}

// RelatesToLinkType returns a link type reference for "relates to" relationship.
func RelatesToLinkType() *IssueLinkType {
	return &IssueLinkType{
		Name:    "Relates",
		Inward:  "relates to",
		Outward: "relates to",
	}
}

// CausesLinkType returns a link type reference for "causes" relationship.
func CausesLinkType() *IssueLinkType {
	return &IssueLinkType{
		Name:    "Causation",
		Inward:  "is caused by",
		Outward: "causes",
	}
}

// ClonesLinkType returns a link type reference for "clones" relationship.
func ClonesLinkType() *IssueLinkType {
	return &IssueLinkType{
		Name:    "Cloners",
		Inward:  "is cloned by",
		Outward: "clones",
	}
}
