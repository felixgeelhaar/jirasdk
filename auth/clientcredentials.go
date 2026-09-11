package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// clientCredentialsRefreshLeeway is how long before expiry a cached token is
// treated as stale, so a request never travels with a token about to lapse.
const clientCredentialsRefreshLeeway = 30 * time.Second

// ClientCredentialsConfig configures the OAuth 2.0 client credentials grant.
type ClientCredentialsConfig struct {
	// ClientID identifies the client to the authorization server.
	ClientID string

	// ClientSecret authenticates the client.
	ClientSecret string //nolint:gosec // G117: struct field name, not a hardcoded secret

	// TokenURL is the authorization server's token endpoint.
	TokenURL string

	// Scopes optionally narrows the access token.
	Scopes []string

	// Audience is sent as an extra "audience" parameter when non-empty, which
	// several authorization servers require to target a specific API.
	Audience string

	// EndpointParams carries any further parameters the authorization server
	// expects on the token request.
	EndpointParams map[string][]string

	// AuthStyle controls whether credentials travel in the Authorization header
	// or the request body. The default probes both.
	AuthStyle oauth2.AuthStyle

	// HTTPClient is used for the token exchange.
	HTTPClient *http.Client
}

// ClientCredentialsAuth implements the OAuth 2.0 client credentials grant, a
// two-legged flow in which the client authenticates as itself with no user
// involved.
//
// Atlassian's own Jira REST API does not offer this grant — for Jira Cloud use
// ConnectJWTAuth or JWTBearerAuth, and for Server/Data Center use OAuth1Auth.
// This authenticator exists for deployments that front Jira with an API gateway
// or an in-house identity provider issuing tokens Jira accepts.
type ClientCredentialsAuth struct {
	config     *clientcredentials.Config
	httpClient *http.Client

	mu    sync.Mutex
	token *oauth2.Token
}

// NewClientCredentialsAuth creates a client credentials authenticator.
//
// Example:
//
//	auth, err := auth.NewClientCredentialsAuth(&auth.ClientCredentialsConfig{
//	    ClientID:     "service-account",
//	    ClientSecret: os.Getenv("CLIENT_SECRET"),
//	    TokenURL:     "https://idp.example.com/oauth2/token",
//	    Scopes:       []string{"jira:read", "jira:write"},
//	})
func NewClientCredentialsAuth(config *ClientCredentialsConfig) (*ClientCredentialsAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("client credentials config is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}
	if config.TokenURL == "" {
		return nil, fmt.Errorf("token URL is required")
	}

	params := map[string][]string{}
	for k, v := range config.EndpointParams {
		params[k] = v
	}
	if config.Audience != "" {
		params["audience"] = []string{config.Audience}
	}

	return &ClientCredentialsAuth{
		config: &clientcredentials.Config{
			ClientID:       config.ClientID,
			ClientSecret:   config.ClientSecret,
			TokenURL:       config.TokenURL,
			Scopes:         config.Scopes,
			EndpointParams: params,
			AuthStyle:      config.AuthStyle,
		},
		httpClient: config.HTTPClient,
	}, nil
}

// Authenticate attaches a bearer token, fetching or renewing it as needed.
// The underlying token source caches tokens and refreshes them before expiry.
func (a *ClientCredentialsAuth) Authenticate(req *http.Request) error {
	token, err := a.Token(req.Context())
	if err != nil {
		return err
	}

	tokenType := token.Type()
	if tokenType == "" {
		tokenType = "Bearer"
	}

	req.Header.Set("Authorization", tokenType+" "+token.AccessToken)
	return nil
}

// Token returns a valid access token, requesting a new one when the cached
// token is missing or close to expiry.
func (a *ClientCredentialsAuth) Token(ctx context.Context) (*oauth2.Token, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != nil && tokenUsable(a.token, clientCredentialsRefreshLeeway) {
		return a.token, nil
	}

	token, err := a.config.Token(a.withHTTPClient(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to obtain client credentials token: %w", err)
	}

	a.token = token
	return token, nil
}

// Client returns an HTTP client that injects the access token automatically.
func (a *ClientCredentialsAuth) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, a.config.TokenSource(a.withHTTPClient(ctx)))
}

// withHTTPClient binds the configured HTTP client to the context, which is how
// golang.org/x/oauth2 accepts a custom transport.
func (a *ClientCredentialsAuth) withHTTPClient(ctx context.Context) context.Context {
	if a.httpClient == nil {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, a.httpClient)
}

// Type returns the authentication type.
func (a *ClientCredentialsAuth) Type() string {
	return "client_credentials"
}
