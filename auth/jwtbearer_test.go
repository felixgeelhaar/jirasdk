package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jwtBearerServer stands in for Atlassian's authorization server, recording the
// form it received and counting exchanges.
type jwtBearerServer struct {
	*httptest.Server
	calls     atomic.Int32
	mu        sync.Mutex
	lastForm  url.Values
	expiresIn int64
}

func newJWTBearerServer(t *testing.T) *jwtBearerServer {
	t.Helper()

	s := &jwtBearerServer{expiresIn: 900}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())

		s.calls.Add(1)
		s.mu.Lock()
		s.lastForm = r.PostForm
		expiresIn := s.expiresIn
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "issued-access-token",
			"token_type":   "Bearer",
			"expires_in":   expiresIn,
		})
	}))
	t.Cleanup(s.Close)

	return s
}

func (s *jwtBearerServer) form() url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastForm
}

func newTestJWTBearerAuth(t *testing.T, tokenURL string) *JWTBearerAuth {
	t.Helper()

	a, err := NewJWTBearerAuth(&JWTBearerConfig{
		OAuthClientID: "oauth-client-id",
		SharedSecret:  testSharedSecret,
		AccountID:     "5b10ac8d82e05b22cc7d4ef5",
		SiteURL:       "https://example.atlassian.net",
		TokenURL:      tokenURL,
	})
	require.NoError(t, err)
	return a
}

func TestNewJWTBearerAuthValidation(t *testing.T) {
	valid := JWTBearerConfig{
		OAuthClientID: "oauth-client-id",
		SharedSecret:  testSharedSecret,
		AccountID:     "account-id",
		SiteURL:       "https://example.atlassian.net",
	}

	tests := []struct {
		name    string
		mutate  func(*JWTBearerConfig)
		wantErr string
	}{
		{"missing client ID", func(c *JWTBearerConfig) { c.OAuthClientID = "" }, "OAuth client ID is required"},
		{"missing shared secret", func(c *JWTBearerConfig) { c.SharedSecret = "" }, "shared secret is required"},
		{"missing account ID", func(c *JWTBearerConfig) { c.AccountID = "" }, "account ID is required"},
		{"missing site URL", func(c *JWTBearerConfig) { c.SiteURL = "" }, "site URL is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)

			a, err := NewJWTBearerAuth(&cfg)
			require.Error(t, err)
			assert.Nil(t, a)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}

	t.Run("nil config", func(t *testing.T) {
		_, err := NewJWTBearerAuth(nil)
		require.Error(t, err)
	})

	t.Run("defaults to Atlassian token URL", func(t *testing.T) {
		a, err := NewJWTBearerAuth(&valid)
		require.NoError(t, err)
		assert.Equal(t, DefaultJWTBearerTokenURL, a.tokenURL)
		assert.Equal(t, "jwt_bearer", a.Type())
	})
}

func TestJWTBearerAssertionClaims(t *testing.T) {
	a := newTestJWTBearerAuth(t, DefaultJWTBearerTokenURL)

	assertion, err := a.Assertion()
	require.NoError(t, err)

	claims := parseConnectToken(t, assertion)
	assert.Equal(t, "urn:atlassian:connect:clientid:oauth-client-id", claims["iss"])
	assert.Equal(t, "urn:atlassian:connect:useraccountid:5b10ac8d82e05b22cc7d4ef5", claims["sub"])
	assert.Equal(t, "https://example.atlassian.net", claims["tnt"])
	assert.Equal(t, "https://oauth-2-authorization-server.services.atlassian.com", claims["aud"])

	// Atlassian rejects assertions valid for more than 120 seconds.
	iat, err := claims.GetIssuedAt()
	require.NoError(t, err)
	exp, err := claims.GetExpirationTime()
	require.NoError(t, err)
	assert.LessOrEqual(t, exp.Sub(iat.Time), 120*time.Second)

	// An assertion is not a Connect request JWT and carries no qsh.
	assert.NotContains(t, claims, JWTClaimQSH)
}

func TestJWTBearerAuthenticate(t *testing.T) {
	server := newJWTBearerServer(t)
	a := newTestJWTBearerAuth(t, server.URL)

	req, err := http.NewRequest(http.MethodGet, "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself", nil)
	require.NoError(t, err)

	require.NoError(t, a.Authenticate(req))
	assert.Equal(t, "Bearer issued-access-token", req.Header.Get("Authorization"))

	form := server.form()
	assert.Equal(t, JWTBearerGrantType, form.Get("grant_type"))
	assert.NotEmpty(t, form.Get("assertion"))
	assert.Empty(t, form.Get("scope"), "no scope should be sent when none configured")

	// The assertion on the wire must verify against the shared secret.
	_, err = jwt.Parse(form.Get("assertion"), func(*jwt.Token) (interface{}, error) {
		return []byte(testSharedSecret), nil
	})
	require.NoError(t, err)
}

func TestJWTBearerSendsScopes(t *testing.T) {
	server := newJWTBearerServer(t)

	a, err := NewJWTBearerAuth(&JWTBearerConfig{
		OAuthClientID: "oauth-client-id",
		SharedSecret:  testSharedSecret,
		AccountID:     "account-id",
		SiteURL:       "https://example.atlassian.net",
		TokenURL:      server.URL,
		Scopes:        []string{"READ", "WRITE"},
	})
	require.NoError(t, err)

	_, err = a.Token(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "READ WRITE", server.form().Get("scope"))
}

func TestJWTBearerCachesToken(t *testing.T) {
	server := newJWTBearerServer(t)
	a := newTestJWTBearerAuth(t, server.URL)

	for i := 0; i < 5; i++ {
		_, err := a.Token(context.Background())
		require.NoError(t, err)
	}

	assert.Equal(t, int32(1), server.calls.Load(), "valid token should be reused")
}

func TestJWTBearerRenewsNearExpiry(t *testing.T) {
	server := newJWTBearerServer(t)
	// Shorter than jwtBearerRefreshLeeway, so the token is never reusable.
	server.expiresIn = 5

	a := newTestJWTBearerAuth(t, server.URL)

	_, err := a.Token(context.Background())
	require.NoError(t, err)
	_, err = a.Token(context.Background())
	require.NoError(t, err)

	assert.Equal(t, int32(2), server.calls.Load(), "token inside the leeway window must be renewed")
}

func TestJWTBearerPropagatesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	a := newTestJWTBearerAuth(t, server.URL)

	_, err := a.Token(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "invalid_grant")
}

func TestJWTBearerRejectsEmptyAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer server.Close()

	a := newTestJWTBearerAuth(t, server.URL)

	_, err := a.Token(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no access token")
}

func TestJWTBearerConcurrentTokenFetch(t *testing.T) {
	server := newJWTBearerServer(t)
	a := newTestJWTBearerAuth(t, server.URL)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := a.Token(context.Background())
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), server.calls.Load(), "concurrent callers should share one exchange")
}
