// Package field provides Field resource management for Jira.
//
// Fields are the building blocks of Jira issues. This package provides operations
// for managing both system fields and custom fields, including creating custom fields,
// retrieving field configurations, and managing field contexts.
package field

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Field resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Field service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Field represents a Jira field (system or custom).
type Field struct {
	ID                  string              `json:"id"`
	Key                 string              `json:"key,omitempty"`
	Name                string              `json:"name"`
	Custom              bool                `json:"custom"`
	Orderable           bool                `json:"orderable,omitempty"`
	Navigable           bool                `json:"navigable,omitempty"`
	Searchable          bool                `json:"searchable,omitempty"`
	ClauseNames         []string            `json:"clauseNames,omitempty"`
	Schema              *FieldSchema        `json:"schema,omitempty"`
	Scope               *FieldScope         `json:"scope,omitempty"`
	Description         string              `json:"description,omitempty"`
	IsLocked            bool                `json:"isLocked,omitempty"`
	SearcherKey         string              `json:"searcherKey,omitempty"`
	ScreensCount        int                 `json:"screensCount,omitempty"`
	ContextsCount       int                 `json:"contextsCount,omitempty"`
	ProjectsCount       int                 `json:"projectsCount,omitempty"`
	LastUsed            *FieldUsage         `json:"lastUsed,omitempty"`
	FieldConfigScheme   *FieldConfigScheme  `json:"fieldConfigScheme,omitempty"`
}

// FieldSchema represents the schema of a field.
type FieldSchema struct {
	Type     string `json:"type"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Custom   string `json:"custom,omitempty"`
	CustomID int64  `json:"customId,omitempty"`
}

// FieldScope represents the scope of a custom field.
type FieldScope struct {
	Type    string    `json:"type"`
	Project *Project  `json:"project,omitempty"`
}

// Project represents a simplified Jira project.
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// FieldUsage represents field usage information.
type FieldUsage struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// FieldConfigScheme represents a field configuration scheme.
type FieldConfigScheme struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateFieldInput represents input for creating a custom field.
type CreateFieldInput struct {
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Type         string       `json:"type"`
	SearcherKey  string       `json:"searcherKey"`
}

// UpdateFieldInput represents input for updating a custom field.
type UpdateFieldInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	SearcherKey string `json:"searcherKey,omitempty"`
}

// FieldContext represents a custom field context.
type FieldContext struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	IsGlobalContext bool     `json:"isGlobalContext"`
	IsAnyIssueType  bool     `json:"isAnyIssueType"`
}

// CreateContextInput represents input for creating a field context.
type CreateContextInput struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	ProjectIDs     []string `json:"projectIds,omitempty"`
	IssueTypeIDs   []string `json:"issueTypeIds,omitempty"`
}

// UpdateContextInput represents input for updating a field context.
type UpdateContextInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// FieldOption represents an option for a select/multi-select field.
type FieldOption struct {
	ID       int64  `json:"id,omitempty"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled,omitempty"`
	Position int    `json:"position,omitempty"`
}

// CreateOptionInput represents input for creating a field option.
type CreateOptionInput struct {
	Value    string `json:"value"`
	Disabled bool   `json:"disabled,omitempty"`
}

// List retrieves all fields (system and custom).
//
// Example:
//
//	fields, err := client.Field.List(ctx)
func (s *Service) List(ctx context.Context) ([]*Field, error) {
	path := "/rest/api/3/field"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var fields []*Field
	if err := s.transport.DecodeResponse(resp, &fields); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return fields, nil
}

// Get retrieves a specific field by ID or key.
//
// Example:
//
//	field, err := client.Field.Get(ctx, "customfield_10000")
func (s *Service) Get(ctx context.Context, fieldID string) (*Field, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s", fieldID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var field Field
	if err := s.transport.DecodeResponse(resp, &field); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &field, nil
}

// Create creates a new custom field.
//
// Example:
//
//	field, err := client.Field.Create(ctx, &field.CreateFieldInput{
//	    Name:        "Story Points",
//	    Description: "Estimation in story points",
//	    Type:        "com.atlassian.jira.plugin.system.customfieldtypes:float",
//	    SearcherKey: "com.atlassian.jira.plugin.system.customfieldtypes:exactnumber",
//	})
func (s *Service) Create(ctx context.Context, input *CreateFieldInput) (*Field, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}

	if input.Type == "" {
		return nil, fmt.Errorf("field type is required")
	}

	if input.SearcherKey == "" {
		return nil, fmt.Errorf("searcher key is required")
	}

	path := "/rest/api/3/field"

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var field Field
	if err := s.transport.DecodeResponse(resp, &field); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &field, nil
}

// Update updates a custom field.
//
// Example:
//
//	field, err := client.Field.Update(ctx, "customfield_10000", &field.UpdateFieldInput{
//	    Name:        "Updated Story Points",
//	    Description: "Updated estimation field",
//	})
func (s *Service) Update(ctx context.Context, fieldID string, input *UpdateFieldInput) (*Field, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s", fieldID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var field Field
	if err := s.transport.DecodeResponse(resp, &field); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &field, nil
}

// Delete deletes a custom field.
//
// Example:
//
//	err := client.Field.Delete(ctx, "customfield_10000")
func (s *Service) Delete(ctx context.Context, fieldID string) error {
	if fieldID == "" {
		return fmt.Errorf("field ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s", fieldID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// ListContexts retrieves contexts for a custom field.
//
// Example:
//
//	contexts, err := client.Field.ListContexts(ctx, "customfield_10000")
func (s *Service) ListContexts(ctx context.Context, fieldID string) ([]*FieldContext, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context", fieldID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*FieldContext `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// CreateContext creates a custom field context.
//
// Example:
//
//	context, err := client.Field.CreateContext(ctx, "customfield_10000", &field.CreateContextInput{
//	    Name:        "Software Projects Context",
//	    Description: "Context for software development projects",
//	    ProjectIDs:  []string{"10000", "10001"},
//	})
func (s *Service) CreateContext(ctx context.Context, fieldID string, input *CreateContextInput) (*FieldContext, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("context name is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context", fieldID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var context FieldContext
	if err := s.transport.DecodeResponse(resp, &context); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &context, nil
}

// UpdateContext updates a custom field context.
//
// Example:
//
//	context, err := client.Field.UpdateContext(ctx, "customfield_10000", "10100", &field.UpdateContextInput{
//	    Name:        "Updated Context Name",
//	    Description: "Updated description",
//	})
func (s *Service) UpdateContext(ctx context.Context, fieldID, contextID string, input *UpdateContextInput) (*FieldContext, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	if contextID == "" {
		return nil, fmt.Errorf("context ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context/%s", fieldID, contextID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var context FieldContext
	if err := s.transport.DecodeResponse(resp, &context); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &context, nil
}

// DeleteContext deletes a custom field context.
//
// Example:
//
//	err := client.Field.DeleteContext(ctx, "customfield_10000", "10100")
func (s *Service) DeleteContext(ctx context.Context, fieldID, contextID string) error {
	if fieldID == "" {
		return fmt.Errorf("field ID is required")
	}

	if contextID == "" {
		return fmt.Errorf("context ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context/%s", fieldID, contextID)

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// ListOptions retrieves options for a select/multi-select custom field context.
//
// Example:
//
//	options, err := client.Field.ListOptions(ctx, "customfield_10000", "10100")
func (s *Service) ListOptions(ctx context.Context, fieldID, contextID string) ([]*FieldOption, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	if contextID == "" {
		return nil, fmt.Errorf("context ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context/%s/option", fieldID, contextID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*FieldOption `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// CreateOption creates an option for a select/multi-select custom field context.
//
// Example:
//
//	option, err := client.Field.CreateOption(ctx, "customfield_10000", "10100", &field.CreateOptionInput{
//	    Value: "High Priority",
//	})
func (s *Service) CreateOption(ctx context.Context, fieldID, contextID string, input *CreateOptionInput) (*FieldOption, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("field ID is required")
	}

	if contextID == "" {
		return nil, fmt.Errorf("context ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	if input.Value == "" {
		return nil, fmt.Errorf("option value is required")
	}

	path := fmt.Sprintf("/rest/api/3/field/%s/context/%s/option", fieldID, contextID)

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var option FieldOption
	if err := s.transport.DecodeResponse(resp, &option); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &option, nil
}
