package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const (
	// DefaultJWTBearerTokenURL is Atlassian's authorization server for the
	// JWT bearer grant used by Connect apps.
	DefaultJWTBearerTokenURL = "https://oauth-2-authorization-server.services.atlassian.com/oauth2/token" //nolint:gosec // G101: public endpoint URL, not a credential

	// JWTBearerGrantType is the RFC 7523 grant type identifier.
	JWTBearerGrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer" //nolint:gosec // G101: public grant type identifier, not a credential

	// jwtBearerAssertionLifetime is the assertion validity window. Atlassian
	// rejects assertions expiring more than 120 seconds in the future.
	jwtBearerAssertionLifetime = 60 * time.Second

	// jwtBearerRefreshLeeway is how long before expiry a cached access token is
	// considered stale, so a request never travels with a token about to lapse.
	jwtBearerRefreshLeeway = 30 * time.Second
)

// JWTBearerConfig configures the OAuth 2.0 JWT bearer grant for Connect apps.
type JWTBearerConfig struct {
	// OAuthClientID is the "oauthClientId" field from the installation
	// lifecycle payload. Note this is distinct from the app key.
	OAuthClientID string

	// SharedSecret is the installation secret, used to sign the assertion.
	SharedSecret string

	// AccountID is the Atlassian account ID of the user to act as.
	AccountID string

	// SiteURL is the tenant URL, for example https://your-domain.atlassian.net.
	// It becomes the "tnt" claim.
	SiteURL string

	// Scopes optionally narrows the access token. Atlassian expects
	// upper-case scope names such as READ or WRITE; leaving this empty
	// requests every scope the app is authorized for.
	Scopes []string

	// TokenURL overrides DefaultJWTBearerTokenURL.
	TokenURL string

	// HTTPClient is used for the token exchange. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// JWTBearerAuth implements Atlassian's OAuth 2.0 JWT bearer token grant, the
// two-legged flow that lets an installed Connect app act as a specific user
// without that user completing an authorization redirect.
//
// The app signs a JWT assertion with its installation shared secret and
// exchanges it for a short-lived bearer token. Tokens are cached and renewed
// automatically.
//
// See https://developer.atlassian.com/cloud/jira/platform/oauth-2-jwt-bearer-token-authorization-grant-type/
type JWTBearerAuth struct {
	clientID     string
	sharedSecret string
	accountID    string
	siteURL      string
	scopes       []string
	tokenURL     string
	httpClient   *http.Client

	mu    sync.Mutex
	token *oauth2.Token
}

// NewJWTBearerAuth creates a JWT bearer authenticator.
//
// Example:
//
//	auth, err := auth.NewJWTBearerAuth(&auth.JWTBearerConfig{
//	    OAuthClientID: install.OAuthClientID,
//	    SharedSecret:  install.SharedSecret,
//	    AccountID:     "5b10ac8d82e05b22cc7d4ef5",
//	    SiteURL:       "https://your-domain.atlassian.net",
//	})
func NewJWTBearerAuth(config *JWTBearerConfig) (*JWTBearerAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("JWT bearer config is required")
	}
	if config.OAuthClientID == "" {
		return nil, fmt.Errorf("OAuth client ID is required")
	}
	if config.SharedSecret == "" {
		return nil, fmt.Errorf("shared secret is required")
	}
	if config.AccountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if config.SiteURL == "" {
		return nil, fmt.Errorf("site URL is required")
	}

	tokenURL := config.TokenURL
	if tokenURL == "" {
		tokenURL = DefaultJWTBearerTokenURL
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &JWTBearerAuth{
		clientID:     config.OAuthClientID,
		sharedSecret: config.SharedSecret,
		accountID:    config.AccountID,
		siteURL:      strings.TrimSuffix(config.SiteURL, "/"),
		scopes:       config.Scopes,
		tokenURL:     tokenURL,
		httpClient:   httpClient,
	}, nil
}

// Authenticate attaches a bearer token, fetching or renewing it as needed.
func (a *JWTBearerAuth) Authenticate(req *http.Request) error {
	token, err := a.Token(req.Context())
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return nil
}

// Token returns a valid access token, exchanging a fresh assertion when the
// cached token is missing or close to expiry.
func (a *JWTBearerAuth) Token(ctx context.Context) (*oauth2.Token, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != nil && tokenUsable(a.token, jwtBearerRefreshLeeway) {
		return a.token, nil
	}

	token, err := a.exchange(ctx)
	if err != nil {
		return nil, err
	}

	a.token = token
	return token, nil
}

// Assertion builds and signs the JWT assertion presented to the authorization
// server. Exported to support debugging and custom exchange flows.
func (a *JWTBearerAuth) Assertion() (string, error) {
	now := time.Now()
	audience := strings.TrimSuffix(a.tokenURL, "/oauth2/token")

	claims := jwt.MapClaims{
		"iss": "urn:atlassian:connect:clientid:" + a.clientID,
		"sub": "urn:atlassian:connect:useraccountid:" + a.accountID,
		"tnt": a.siteURL,
		"aud": audience,
		"iat": now.Unix(),
		"exp": now.Add(jwtBearerAssertionLifetime).Unix(),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.sharedSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT assertion: %w", err)
	}

	return signed, nil
}

// exchange trades the assertion for an access token. The caller holds a.mu.
func (a *JWTBearerAuth) exchange(ctx context.Context) (*oauth2.Token, error) {
	assertion, err := a.Assertion()
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("grant_type", JWTBearerGrantType)
	form.Set("assertion", assertion)
	if len(a.scopes) > 0 {
		form.Set("scope", strings.Join(a.scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req) //nolint:gosec // G704: the URL is the operator-configured token endpoint, not tainted input
	if err != nil {
		return nil, fmt.Errorf("JWT bearer token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWT bearer token request returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"access_token"` //nolint:gosec // G117: field of a token response, not a hardcoded secret
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	if payload.AccessToken == "" {
		return nil, fmt.Errorf("token response contained no access token")
	}

	tokenType := payload.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	token := &oauth2.Token{
		AccessToken: payload.AccessToken,
		TokenType:   tokenType,
	}
	if payload.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}

	return token, nil
}

// Type returns the authentication type.
func (a *JWTBearerAuth) Type() string {
	return "jwt_bearer"
}

// tokenUsable reports whether the token is still valid once leeway is deducted
// from its expiry. A token with no expiry is treated as valid.
func tokenUsable(token *oauth2.Token, leeway time.Duration) bool {
	if token.AccessToken == "" {
		return false
	}
	if token.Expiry.IsZero() {
		return true
	}
	return time.Now().Add(leeway).Before(token.Expiry)
}
