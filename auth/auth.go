// Package auth provides authentication mechanisms for Jira API requests.
package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// Authenticator is the interface that all authentication methods must implement.
type Authenticator interface {
	// Authenticate modifies the HTTP request to include authentication credentials
	Authenticate(req *http.Request) error

	// Type returns the authentication type for logging and debugging
	Type() string
}

// APITokenAuth implements API token authentication for Jira Cloud.
//
// API tokens are the recommended authentication method for Jira Cloud.
// They use HTTP Basic Auth with email as username and token as password.
type APITokenAuth struct {
	email string
	token string
}

// NewAPITokenAuth creates a new API token authenticator.
func NewAPITokenAuth(email, token string) *APITokenAuth {
	return &APITokenAuth{
		email: email,
		token: token,
	}
}

// Authenticate adds API token authentication to the request.
func (a *APITokenAuth) Authenticate(req *http.Request) error {
	credentials := fmt.Sprintf("%s:%s", a.email, a.token)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	return nil
}

// Type returns the authentication type.
func (a *APITokenAuth) Type() string {
	return "api_token"
}

// PATAuth implements Personal Access Token authentication for Jira Server/Data Center.
//
// PATs use Bearer token authentication and are the recommended method for
// Jira Server and Data Center instances.
type PATAuth struct {
	token string
}

// NewPATAuth creates a new PAT authenticator.
func NewPATAuth(token string) *PATAuth {
	return &PATAuth{
		token: token,
	}
}

// Authenticate adds PAT authentication to the request.
func (a *PATAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
	return nil
}

// Type returns the authentication type.
func (a *PATAuth) Type() string {
	return "pat"
}

// BasicAuth implements HTTP Basic authentication.
//
// This is a legacy authentication method and should only be used when
// other methods are not available. It's not recommended for production use.
type BasicAuth struct {
	username string
	password string
}

// NewBasicAuth creates a new basic authenticator.
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}

// Authenticate adds basic authentication to the request.
func (a *BasicAuth) Authenticate(req *http.Request) error {
	credentials := fmt.Sprintf("%s:%s", a.username, a.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	return nil
}

// Type returns the authentication type.
func (a *BasicAuth) Type() string {
	return "basic"
}
