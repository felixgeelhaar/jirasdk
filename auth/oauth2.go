package auth

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"golang.org/x/oauth2"
)

const (
	// DefaultOAuth2AuthURL is Atlassian's authorization endpoint.
	DefaultOAuth2AuthURL = "https://auth.atlassian.com/authorize"

	// DefaultOAuth2TokenURL is Atlassian's token endpoint.
	DefaultOAuth2TokenURL = "https://auth.atlassian.com/oauth/token" //nolint:gosec // G101: public endpoint URL, not a credential

	// DefaultOAuth2Audience is the audience Atlassian requires on the
	// authorization request for Jira and Confluence Cloud APIs.
	DefaultOAuth2Audience = "api.atlassian.com"

	// ScopeOfflineAccess asks Atlassian for a refresh token. Unlike Google's
	// access_type=offline, Atlassian expects this as a scope.
	ScopeOfflineAccess = "offline_access"
)

// OAuth2Authenticator implements OAuth 2.0 (3LO) authentication for Jira.
//
// This is the three-legged flow: a user is redirected to Atlassian, approves
// the requested scopes, and the resulting authorization code is exchanged for
// an access token that acts on that user's behalf.
//
// The authenticator is safe for concurrent use. Expired tokens are refreshed
// transparently and, when a TokenStore is configured, persisted.
type OAuth2Authenticator struct {
	config   *oauth2.Config
	audience string
	store    OAuth2TokenStore

	mu    sync.Mutex
	token *oauth2.Token
}

// OAuth2Config contains configuration for OAuth 2.0 (3LO) authentication.
type OAuth2Config struct {
	// ClientID is the OAuth 2.0 client ID
	ClientID string

	// ClientSecret is the OAuth 2.0 client secret
	ClientSecret string //nolint:gosec // G117: struct field name, not a hardcoded secret

	// RedirectURL is the callback URL for the OAuth 2.0 flow
	RedirectURL string

	// Scopes are the OAuth 2.0 scopes to request. ScopeOfflineAccess is added
	// automatically unless DisableOfflineAccess is set, because without it
	// Atlassian issues no refresh token.
	Scopes []string

	// AuthURL is the authorization endpoint (defaults to Jira Cloud)
	AuthURL string

	// TokenURL is the token endpoint (defaults to Jira Cloud)
	TokenURL string

	// Audience is the "audience" parameter on the authorization request.
	// Defaults to DefaultOAuth2Audience, which Atlassian requires.
	Audience string

	// DisableOfflineAccess stops ScopeOfflineAccess from being added. Set this
	// only if you do not want a refresh token.
	DisableOfflineAccess bool

	// TokenStore optionally persists tokens across process restarts. When set,
	// tokens are saved after every exchange and refresh.
	TokenStore OAuth2TokenStore
}

// NewOAuth2Authenticator creates a new OAuth 2.0 (3LO) authenticator.
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
		config.AuthURL = DefaultOAuth2AuthURL
	}
	if config.TokenURL == "" {
		config.TokenURL = DefaultOAuth2TokenURL
	}
	if config.Audience == "" {
		config.Audience = DefaultOAuth2Audience
	}

	scopes := slices.Clone(config.Scopes)
	if !config.DisableOfflineAccess && !slices.Contains(scopes, ScopeOfflineAccess) {
		scopes = append(scopes, ScopeOfflineAccess)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	return &OAuth2Authenticator{
		config:   oauthConfig,
		audience: config.Audience,
		store:    config.TokenStore,
	}
}

// GetAuthURL returns the authorization URL for the OAuth 2.0 (3LO) flow.
//
// The URL carries the "audience" and "prompt=consent" parameters Atlassian
// requires; without them the authorization request is rejected.
//
// The state value should be unguessable and tied to the user's session, so the
// callback can be verified against it.
//
// Example:
//
//	url := auth.GetAuthURL("state-string")
//	fmt.Println("Visit:", url)
func (a *OAuth2Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("audience", a.audience),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
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

	a.mu.Lock()
	a.token = token
	a.mu.Unlock()

	a.persist(token)
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
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = token
}

// GetToken returns the current OAuth 2.0 token.
func (a *OAuth2Authenticator) GetToken() *oauth2.Token {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.token
}

// LoadToken restores a token from the configured TokenStore.
func (a *OAuth2Authenticator) LoadToken() (*oauth2.Token, error) {
	if a.store == nil {
		return nil, fmt.Errorf("no token store configured")
	}

	token, err := a.store.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	a.SetToken(token)
	return token, nil
}

// RefreshToken refreshes the OAuth 2.0 access token.
//
// Example:
//
//	newToken, err := auth.RefreshToken(ctx)
func (a *OAuth2Authenticator) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.refreshLocked(ctx)
}

// refreshLocked refreshes the token. The caller must hold a.mu.
func (a *OAuth2Authenticator) refreshLocked(ctx context.Context) (*oauth2.Token, error) {
	if a.token == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	newToken, err := a.config.TokenSource(ctx, a.token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	a.token = newToken
	a.persist(newToken)

	return newToken, nil
}

// persist writes the token to the store, if one is configured. A store failure
// must not fail the request the token was obtained for, so it is not surfaced.
func (a *OAuth2Authenticator) persist(token *oauth2.Token) {
	if a.store == nil {
		return
	}
	_ = a.store.SaveToken(token)
}

// Authenticate adds OAuth 2.0 authentication to the request, refreshing the
// token first if it has expired.
func (a *OAuth2Authenticator) Authenticate(req *http.Request) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token == nil {
		return fmt.Errorf("no OAuth 2.0 token available")
	}

	if !a.token.Valid() {
		if _, err := a.refreshLocked(req.Context()); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+a.token.AccessToken)
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
	token := a.GetToken()
	if token == nil {
		return http.DefaultClient
	}
	return a.config.Client(ctx, token)
}

// AccessibleResources lists the Atlassian sites this token can reach. The "id"
// of a site is its cloud ID, which forms the API base URL for 3LO requests:
// https://api.atlassian.com/ex/jira/{cloudID}
func (a *OAuth2Authenticator) AccessibleResources(ctx context.Context) ([]AccessibleResource, error) {
	token := a.GetToken()
	if token == nil {
		return nil, fmt.Errorf("no OAuth 2.0 token available")
	}

	return FetchAccessibleResources(ctx, a.Client(ctx), token.AccessToken)
}

// OAuth2TokenStore defines an interface for storing and retrieving OAuth 2.0 tokens.
//
// Supply an implementation via OAuth2Config.TokenStore to keep refreshed tokens
// across restarts; otherwise a refresh token is lost when the process exits.
type OAuth2TokenStore interface {
	// SaveToken saves an OAuth 2.0 token
	SaveToken(token *oauth2.Token) error

	// LoadToken loads an OAuth 2.0 token
	LoadToken() (*oauth2.Token, error)

	// DeleteToken deletes a stored token
	DeleteToken() error
}

// IsAtlassianCloud reports whether this authenticator targets Atlassian's own
// authorization server. A custom AuthURL implies a self-hosted or proxied
// deployment, where Atlassian's cloud ID routing does not apply.
func (a *OAuth2Authenticator) IsAtlassianCloud() bool {
	return a.config.Endpoint.AuthURL == DefaultOAuth2AuthURL
}
