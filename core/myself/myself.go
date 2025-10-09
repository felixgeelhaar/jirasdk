// Package myself provides current user (myself) operations for Jira.
//
// The myself resource provides operations for the currently authenticated user.
package myself

import (
	"context"
	"fmt"
	"net/http"
)

// Service provides operations for the current user.
type Service struct {
	transport RoundTripper
}

// RoundTripper defines the interface for making HTTP requests.
type RoundTripper interface {
	NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DecodeResponse(resp *http.Response, v interface{}) error
}

// NewService creates a new Myself service.
func NewService(transport RoundTripper) *Service {
	return &Service{
		transport: transport,
	}
}

// User represents the current user.
type User struct {
	Self         string      `json:"self,omitempty"`
	AccountID    string      `json:"accountId,omitempty"`
	AccountType  string      `json:"accountType,omitempty"`
	EmailAddress string      `json:"emailAddress,omitempty"`
	DisplayName  string      `json:"displayName,omitempty"`
	Active       bool        `json:"active,omitempty"`
	TimeZone     string      `json:"timeZone,omitempty"`
	Locale       string      `json:"locale,omitempty"`
	AvatarURLs   *AvatarURLs `json:"avatarUrls,omitempty"`
	Groups       *Groups     `json:"groups,omitempty"`
	Expand       string      `json:"expand,omitempty"`
}

// AvatarURLs contains URLs for different avatar sizes.
type AvatarURLs struct {
	Size16 string `json:"16x16,omitempty"`
	Size24 string `json:"24x24,omitempty"`
	Size32 string `json:"32x32,omitempty"`
	Size48 string `json:"48x48,omitempty"`
}

// Groups represents user group membership.
type Groups struct {
	Size  int          `json:"size,omitempty"`
	Items []*GroupItem `json:"items,omitempty"`
}

// GroupItem represents a single group.
type GroupItem struct {
	Name string `json:"name,omitempty"`
	Self string `json:"self,omitempty"`
}

// Preferences represents user preferences.
type Preferences struct {
	Locale   string `json:"locale,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

// Get retrieves the current user.
//
// Example:
//
//	user, err := client.Myself.Get(ctx)
func (s *Service) Get(ctx context.Context) (*User, error) {
	path := "/rest/api/3/myself"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var user User
	if err := s.transport.DecodeResponse(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// GetPreferences retrieves current user preferences.
//
// Example:
//
//	prefs, err := client.Myself.GetPreferences(ctx)
func (s *Service) GetPreferences(ctx context.Context) (*Preferences, error) {
	path := "/rest/api/3/mypreferences"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	var prefs Preferences
	if err := s.transport.DecodeResponse(resp, &prefs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &prefs, nil
}

// SetPreferences sets current user preferences.
//
// Example:
//
//	err := client.Myself.SetPreferences(ctx, &myself.Preferences{
//		Locale:   "en_US",
//		TimeZone: "America/New_York",
//	})
func (s *Service) SetPreferences(ctx context.Context, prefs *Preferences) error {
	if prefs == nil {
		return fmt.Errorf("preferences are required")
	}

	path := "/rest/api/3/mypreferences"

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, prefs)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// GetPreference retrieves a specific preference.
//
// Example:
//
//	value, err := client.Myself.GetPreference(ctx, "locale")
func (s *Service) GetPreference(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("preference key is required")
	}

	path := "/rest/api/3/mypreferences"

	req, err := s.transport.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter
	q := req.URL.Query()
	q.Set("key", key)
	req.URL.RawQuery = q.Encode()

	resp, err := s.transport.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	var result map[string]string
	if err := s.transport.DecodeResponse(resp, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	value, ok := result[key]
	if !ok {
		return "", fmt.Errorf("preference not found")
	}

	return value, nil
}

// SetPreference sets a specific preference.
//
// Example:
//
//	err := client.Myself.SetPreference(ctx, "locale", "en_US")
func (s *Service) SetPreference(ctx context.Context, key, value string) error {
	if key == "" {
		return fmt.Errorf("preference key is required")
	}

	path := "/rest/api/3/mypreferences"

	req, err := s.transport.NewRequest(ctx, http.MethodPut, path, map[string]string{
		key: value,
	})
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter
	q := req.URL.Query()
	q.Set("key", key)
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}

// DeletePreference deletes a specific preference.
//
// Example:
//
//	err := client.Myself.DeletePreference(ctx, "customSetting")
func (s *Service) DeletePreference(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("preference key is required")
	}

	path := "/rest/api/3/mypreferences"

	req, err := s.transport.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter
	q := req.URL.Query()
	q.Set("key", key)
	req.URL.RawQuery = q.Encode()

	_, err = s.transport.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
