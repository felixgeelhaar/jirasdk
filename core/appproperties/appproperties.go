// Package appproperties provides Application Properties management for Jira.
//
// Application properties store system-wide configuration settings.
package appproperties

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for Application Properties.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Application Properties service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// ApplicationProperty represents an application property.
type ApplicationProperty struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	Value         string   `json:"value"`
	Name          string   `json:"name,omitempty"`
	Desc          string   `json:"desc,omitempty"`
	Type          string   `json:"type,omitempty"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
	Example       string   `json:"example,omitempty"`
}

// AdvancedSettings represents Jira advanced settings.
type AdvancedSettings struct {
	Properties []*ApplicationProperty `json:"properties"`
}

// GetAdvancedSettings retrieves all advanced settings.
//
// Example:
//
//	settings, err := client.AppProperties.GetAdvancedSettings(ctx)
func (s *Service) GetAdvancedSettings(ctx context.Context) ([]*ApplicationProperty, error) {
	path := "/rest/api/3/application-properties/advanced-settings"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var properties []*ApplicationProperty
	if err := s.transport.DecodeResponse(resp, &properties); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return properties, nil
}

// GetApplicationProperty retrieves a specific application property.
//
// Example:
//
//	property, err := client.AppProperties.GetApplicationProperty(ctx, "jira.clone.prefix")
func (s *Service) GetApplicationProperty(ctx context.Context, key string) (*ApplicationProperty, error) {
	if key == "" {
		return nil, fmt.Errorf("property key is required")
	}

	path := "/rest/api/3/application-properties"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter
	q := req.URL.Query()
	q.Set("key", key)
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var properties []*ApplicationProperty
	if err := s.transport.DecodeResponse(resp, &properties); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(properties) == 0 {
		return nil, fmt.Errorf("property not found")
	}

	return properties[0], nil
}

// SetApplicationPropertyInput represents input for setting a property.
type SetApplicationPropertyInput struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// SetApplicationProperty sets an application property value.
//
// Example:
//
//	err := client.AppProperties.SetApplicationProperty(ctx, &appproperties.SetApplicationPropertyInput{
//		ID:    "jira.clone.prefix",
//		Value: "CLONE -",
//	})
func (s *Service) SetApplicationProperty(ctx context.Context, input *SetApplicationPropertyInput) error {
	if input == nil || input.ID == "" {
		return fmt.Errorf("property ID is required")
	}

	path := fmt.Sprintf("/rest/api/3/application-properties/%s", input.ID)

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, map[string]string{
		"value": input.Value,
	})
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
