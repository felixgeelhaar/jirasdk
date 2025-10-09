// Package webhook provides Webhook resource management for Jira.
//
// Webhooks allow external systems to receive notifications when events occur in Jira.
// This package provides operations for managing webhooks and their configurations.
package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Service provides operations for Webhook resources.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Webhook service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// Webhook represents a Jira webhook.
type Webhook struct {
	ID              int64    `json:"id,omitempty"`
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Events          []string `json:"events"`
	JQLFilter       string   `json:"jqlFilter,omitempty"`
	ExcludeBody     bool     `json:"excludeBody,omitempty"`
	Enabled         bool     `json:"enabled,omitempty"`
	ExpirationDate  int64    `json:"expirationDate,omitempty"`
	LastUpdatedUser string   `json:"lastUpdatedUser,omitempty"`
	LastUpdatedDate int64    `json:"lastUpdatedDate,omitempty"`
}

// CreateWebhookInput represents input for creating a webhook.
type CreateWebhookInput struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	JQLFilter   string   `json:"jqlFilter,omitempty"`
	ExcludeBody bool     `json:"excludeBody,omitempty"`
}

// UpdateWebhookInput represents input for updating a webhook.
type UpdateWebhookInput struct {
	Name        string   `json:"name,omitempty"`
	URL         string   `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	JQLFilter   string   `json:"jqlFilter,omitempty"`
	ExcludeBody bool     `json:"excludeBody,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// WebhookRegistrationResult represents the result of webhook registration.
type WebhookRegistrationResult struct {
	CreatedWebhookID int64          `json:"createdWebhookId,omitempty"`
	Errors           []WebhookError `json:"errors,omitempty"`
}

// WebhookError represents an error in webhook operations.
type WebhookError struct {
	Message string `json:"message"`
}

// ListOptions represents options for listing webhooks.
type ListOptions struct {
	StartAt    int `json:"startAt,omitempty"`
	MaxResults int `json:"maxResults,omitempty"`
}

// FailedWebhook represents a failed webhook execution.
type FailedWebhook struct {
	ID          int64  `json:"id"`
	Body        string `json:"body,omitempty"`
	URL         string `json:"url"`
	FailureTime int64  `json:"failureTime"`
}

// List retrieves all webhooks with pagination.
//
// Example:
//
//	webhooks, err := client.Webhook.List(ctx, &webhook.ListOptions{MaxResults: 50})
func (s *Service) List(ctx context.Context, opts *ListOptions) ([]*Webhook, error) {
	path := "/rest/api/3/webhook"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.StartAt > 0 {
			q.Set("startAt", strconv.Itoa(opts.StartAt))
		}
		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*Webhook `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}

// Get retrieves a specific webhook by ID.
//
// Example:
//
//	webhook, err := client.Webhook.Get(ctx, 10000)
func (s *Service) Get(ctx context.Context, webhookID int64) (*Webhook, error) {
	if webhookID <= 0 {
		return nil, fmt.Errorf("webhook ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/webhook/%d", webhookID)

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var webhook Webhook
	if err := s.transport.DecodeResponse(resp, &webhook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhook, nil
}

// Create creates new webhooks.
//
// Example:
//
//	webhooks := []*webhook.CreateWebhookInput{
//	    {
//	        Name:   "Issue Created Webhook",
//	        URL:    "https://example.com/webhooks/jira",
//	        Events: []string{"jira:issue_created"},
//	    },
//	}
//	results, err := client.Webhook.Create(ctx, webhooks)
func (s *Service) Create(ctx context.Context, webhooks []*CreateWebhookInput) ([]*WebhookRegistrationResult, error) {
	if len(webhooks) == 0 {
		return nil, fmt.Errorf("at least one webhook is required")
	}

	for i, webhook := range webhooks {
		if webhook.Name == "" {
			return nil, fmt.Errorf("webhook %d: name is required", i)
		}
		if webhook.URL == "" {
			return nil, fmt.Errorf("webhook %d: URL is required", i)
		}
		if len(webhook.Events) == 0 {
			return nil, fmt.Errorf("webhook %d: at least one event is required", i)
		}
	}

	path := "/rest/api/3/webhook"

	input := struct {
		Webhooks []*CreateWebhookInput `json:"webhooks"`
	}{
		Webhooks: webhooks,
	}

	req, err := s.transport.NewRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		WebhookRegistrationResult []*WebhookRegistrationResult `json:"webhookRegistrationResult"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.WebhookRegistrationResult, nil
}

// Update updates a webhook.
//
// Example:
//
//	enabled := false
//	webhook, err := client.Webhook.Update(ctx, 10000, &webhook.UpdateWebhookInput{
//	    Name:    "Updated Webhook",
//	    Enabled: &enabled,
//	})
func (s *Service) Update(ctx context.Context, webhookID int64, input *UpdateWebhookInput) (*Webhook, error) {
	if webhookID <= 0 {
		return nil, fmt.Errorf("webhook ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	path := fmt.Sprintf("/rest/api/3/webhook/%d", webhookID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var webhook Webhook
	if err := s.transport.DecodeResponse(resp, &webhook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhook, nil
}

// Delete deletes webhooks by ID.
//
// Example:
//
//	err := client.Webhook.Delete(ctx, []int64{10000, 10001})
func (s *Service) Delete(ctx context.Context, webhookIDs []int64) error {
	if len(webhookIDs) == 0 {
		return fmt.Errorf("at least one webhook ID is required")
	}

	path := "/rest/api/3/webhook"

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add webhook IDs as query parameters
	q := req.URL.Query()
	for _, id := range webhookIDs {
		q.Add("webhookId", strconv.FormatInt(id, 10))
	}
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// Refresh extends the expiration date of webhooks.
//
// Example:
//
//	expiration, err := client.Webhook.Refresh(ctx, []int64{10000, 10001})
func (s *Service) Refresh(ctx context.Context, webhookIDs []int64) (int64, error) {
	if len(webhookIDs) == 0 {
		return 0, fmt.Errorf("at least one webhook ID is required")
	}

	path := "/rest/api/3/webhook/refresh"

	input := struct {
		WebhookIDs []int64 `json:"webhookIds"`
	}{
		WebhookIDs: webhookIDs,
	}

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		ExpirationDate int64 `json:"expirationDate"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ExpirationDate, nil
}

// GetDynamicModules retrieves dynamic modules for webhooks.
//
// Example:
//
//	modules, err := client.Webhook.GetDynamicModules(ctx)
func (s *Service) GetDynamicModules(ctx context.Context) (map[string]interface{}, error) {
	path := "/rest/api/3/app/module/dynamic"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var modules map[string]interface{}
	if err := s.transport.DecodeResponse(resp, &modules); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return modules, nil
}

// GetFailedWebhooks retrieves webhooks that have recently failed.
//
// Example:
//
//	failed, err := client.Webhook.GetFailedWebhooks(ctx, &webhook.ListOptions{MaxResults: 10})
func (s *Service) GetFailedWebhooks(ctx context.Context, opts *ListOptions) ([]*FailedWebhook, error) {
	path := "/rest/api/3/webhook/failed"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(opts.MaxResults))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var result struct {
		Values []*FailedWebhook `json:"values"`
	}

	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Values, nil
}
