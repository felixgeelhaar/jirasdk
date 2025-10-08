package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// OAuth2Authenticator implements OAuth 2.0 authentication for Jira.
type OAuth2Authenticator struct {
	config *oauth2.Config
	token  *oauth2.Token
}

// OAuth2Config contains configuration for OAuth 2.0 authentication.
type OAuth2Config struct {
	// ClientID is the OAuth 2.0 client ID
	ClientID string

	// ClientSecret is the OAuth 2.0 client secret
	ClientSecret string

	// RedirectURL is the callback URL for the OAuth 2.0 flow
	RedirectURL string

	// Scopes are the OAuth 2.0 scopes to request
	Scopes []string

	// AuthURL is the authorization endpoint (defaults to Jira Cloud)
	AuthURL string

	// TokenURL is the token endpoint (defaults to Jira Cloud)
	TokenURL string
}

// NewOAuth2Authenticator creates a new OAuth 2.0 authenticator.
//
// Example:
//
//	auth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURL:  "http://localhost:8080/callback",
//	    Scopes:       []string{"read:jira-work", "write:jira-work"},
//	})
func NewOAuth2Authenticator(config *OAuth2Config) *OAuth2Authenticator {
	if config.AuthURL == "" {
		config.AuthURL = "https://auth.atlassian.com/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://auth.atlassian.com/oauth/token"
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	return &OAuth2Authenticator{
		config: oauthConfig,
	}
}

// GetAuthURL returns the authorization URL for the OAuth 2.0 flow.
//
// Example:
//
//	url := auth.GetAuthURL("state-string")
//	fmt.Println("Visit:", url)
func (a *OAuth2Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange exchanges an authorization code for an access token.
//
// Example:
//
//	token, err := auth.Exchange(ctx, "authorization-code")
func (a *OAuth2Authenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	a.token = token
	return token, nil
}

// SetToken sets the OAuth 2.0 token for subsequent requests.
//
// Example:
//
//	auth.SetToken(&oauth2.Token{
//	    AccessToken:  "access-token",
//	    RefreshToken: "refresh-token",
//	})
func (a *OAuth2Authenticator) SetToken(token *oauth2.Token) {
	a.token = token
}

// GetToken returns the current OAuth 2.0 token.
func (a *OAuth2Authenticator) GetToken() *oauth2.Token {
	return a.token
}

// RefreshToken refreshes the OAuth 2.0 access token.
//
// Example:
//
//	newToken, err := auth.RefreshToken(ctx)
func (a *OAuth2Authenticator) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	if a.token == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	tokenSource := a.config.TokenSource(ctx, a.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	a.token = newToken
	return newToken, nil
}

// Authenticate adds OAuth 2.0 authentication to the request.
func (a *OAuth2Authenticator) Authenticate(req *http.Request) error {
	if a.token == nil {
		return fmt.Errorf("no OAuth 2.0 token available")
	}

	// Check if token needs refresh
	if !a.token.Valid() {
		ctx := req.Context()
		if _, err := a.RefreshToken(ctx); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Add Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token.AccessToken))
	return nil
}

// Type returns the authentication type.
func (a *OAuth2Authenticator) Type() string {
	return "oauth2"
}

// Client returns an HTTP client that automatically handles OAuth 2.0 authentication.
//
// Example:
//
//	httpClient := auth.Client(ctx)
//	resp, err := httpClient.Get("https://api.atlassian.com/...")
func (a *OAuth2Authenticator) Client(ctx context.Context) *http.Client {
	if a.token == nil {
		return http.DefaultClient
	}
	return a.config.Client(ctx, a.token)
}

// OAuth2TokenStore defines an interface for storing and retrieving OAuth 2.0 tokens.
type OAuth2TokenStore interface {
	// SaveToken saves an OAuth 2.0 token
	SaveToken(token *oauth2.Token) error

	// LoadToken loads an OAuth 2.0 token
	LoadToken() (*oauth2.Token, error)

	// DeleteToken deletes a stored token
	DeleteToken() error
}
