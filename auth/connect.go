package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultConnectJWTLifetime is the token lifetime recommended by Atlassian for
// Connect JWTs. Tokens are minted per request, so it only needs to cover clock
// skew and network latency.
const DefaultConnectJWTLifetime = 180 * time.Second

// ConnectJWTConfig configures Connect JWT authentication.
type ConnectJWTConfig struct {
	// AppKey is the app key from your atlassian-connect.json descriptor. It
	// becomes the "iss" claim.
	AppKey string

	// SharedSecret is the installation secret received in the "installed"
	// lifecycle callback. It signs the JWT with HMAC-SHA256.
	SharedSecret string

	// Subject optionally sets the "sub" claim to an Atlassian account ID,
	// associating the request with a user.
	Subject string

	// ContextPath is the product's context path, stripped from the URI before
	// computing the query string hash. Empty for Jira Cloud; typically "/jira"
	// for a Server instance deployed under a path.
	ContextPath string

	// TokenLifetime overrides DefaultConnectJWTLifetime.
	TokenLifetime time.Duration
}

// ConnectJWTAuth implements Atlassian Connect JWT authentication.
//
// This is the two-legged flow for Connect apps: the app authenticates as
// itself using the shared secret exchanged during installation, with no user
// authorization step. Each request carries a freshly minted JWT whose "qsh"
// claim binds it to that specific method, path and query string.
//
// See https://developer.atlassian.com/cloud/jira/platform/understanding-jwt-for-connect-apps/
type ConnectJWTAuth struct {
	appKey       string
	sharedSecret string
	subject      string
	contextPath  string
	lifetime     time.Duration
}

// NewConnectJWTAuth creates a Connect JWT authenticator.
//
// Example:
//
//	auth, err := auth.NewConnectJWTAuth(&auth.ConnectJWTConfig{
//	    AppKey:       "com.example.my-app",
//	    SharedSecret: installPayload.SharedSecret,
//	})
func NewConnectJWTAuth(config *ConnectJWTConfig) (*ConnectJWTAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("connect JWT config is required")
	}
	if config.AppKey == "" {
		return nil, fmt.Errorf("app key is required")
	}
	if config.SharedSecret == "" {
		return nil, fmt.Errorf("shared secret is required")
	}

	lifetime := config.TokenLifetime
	if lifetime <= 0 {
		lifetime = DefaultConnectJWTLifetime
	}

	return &ConnectJWTAuth{
		appKey:       config.AppKey,
		sharedSecret: config.SharedSecret,
		subject:      config.Subject,
		contextPath:  config.ContextPath,
		lifetime:     lifetime,
	}, nil
}

// Authenticate mints a request-scoped JWT and attaches it to the request.
func (a *ConnectJWTAuth) Authenticate(req *http.Request) error {
	token, err := a.Sign(req)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "JWT "+token)
	return nil
}

// Sign builds and signs the Connect JWT for the given request without mutating
// it. Exported so callers can reuse the token for requests this SDK does not
// model, such as a raw attachment upload.
func (a *ConnectJWTAuth) Sign(req *http.Request) (string, error) {
	qsh, err := QueryStringHash(req, a.contextPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute query string hash: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       a.appKey,
		"iat":       now.Unix(),
		"exp":       now.Add(a.lifetime).Unix(),
		JWTClaimQSH: qsh,
	}
	if a.subject != "" {
		claims["sub"] = a.subject
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.sharedSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign connect JWT: %w", err)
	}

	return signed, nil
}

// Type returns the authentication type.
func (a *ConnectJWTAuth) Type() string {
	return "connect_jwt"
}
