package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewOAuth2Authenticator(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work", "write:jira-work"},
	}

	auth := NewOAuth2Authenticator(config)

	assert.NotNil(t, auth)
	assert.NotNil(t, auth.config)
	assert.Equal(t, config.ClientID, auth.config.ClientID)
	assert.Equal(t, config.ClientSecret, auth.config.ClientSecret)
	assert.Equal(t, config.RedirectURL, auth.config.RedirectURL)
	// offline_access is appended so Atlassian issues a refresh token.
	assert.Equal(t, []string{"read:jira-work", "write:jira-work", ScopeOfflineAccess}, auth.config.Scopes)
	assert.Equal(t, []string{"read:jira-work", "write:jira-work"}, config.Scopes, "caller's slice must not be mutated")
	assert.Equal(t, "https://auth.atlassian.com/authorize", auth.config.Endpoint.AuthURL)
	assert.Equal(t, "https://auth.atlassian.com/oauth/token", auth.config.Endpoint.TokenURL)
}

func TestNewOAuth2Authenticator_CustomEndpoints(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work"},
		AuthURL:      "https://custom.com/auth",
		TokenURL:     "https://custom.com/token",
	}

	auth := NewOAuth2Authenticator(config)

	assert.Equal(t, "https://custom.com/auth", auth.config.Endpoint.AuthURL)
	assert.Equal(t, "https://custom.com/token", auth.config.Endpoint.TokenURL)
}

func TestGetAuthURL(t *testing.T) {
	config := &OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work"},
	}

	auth := NewOAuth2Authenticator(config)
	url := auth.GetAuthURL("test-state")

	assert.Contains(t, url, "https://auth.atlassian.com/authorize")
	assert.Contains(t, url, "client_id=test-client-id")
	assert.Contains(t, url, "redirect_uri=http")
	assert.Contains(t, url, "state=test-state")
	assert.Contains(t, url, "scope=read%3Ajira-work")
}

func TestSetToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	auth.SetToken(token)

	assert.Equal(t, token, auth.GetToken())
	assert.Equal(t, "test-access-token", auth.GetToken().AccessToken)
	assert.Equal(t, "test-refresh-token", auth.GetToken().RefreshToken)
}

func TestGetToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	// Token should be nil initially
	assert.Nil(t, auth.GetToken())

	// Set and retrieve token
	token := &oauth2.Token{AccessToken: "test-token"}
	auth.SetToken(token)
	assert.Equal(t, token, auth.GetToken())
}

func TestAuthenticate_NoToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	req := httptest.NewRequest(http.MethodGet, "https://api.atlassian.com/test", nil)
	err := auth.Authenticate(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no OAuth 2.0 token available")
}

func TestAuthenticate_WithValidToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	// Set a valid token (expires in the future)
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	auth.SetToken(token)

	req := httptest.NewRequest(http.MethodGet, "https://api.atlassian.com/test", nil)
	err := auth.Authenticate(req)

	require.NoError(t, err)
	assert.Equal(t, "Bearer test-access-token", req.Header.Get("Authorization"))
}

func TestType(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	assert.Equal(t, "oauth2", auth.Type())
}

func TestClient_NoToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	client := auth.Client(context.Background())
	assert.Equal(t, http.DefaultClient, client)
}

func TestClient_WithToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	auth.SetToken(token)

	client := auth.Client(context.Background())
	assert.NotNil(t, client)
	assert.NotEqual(t, http.DefaultClient, client)
}

func TestExchange(t *testing.T) {
	// Create a test server that simulates the OAuth 2.0 token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		assert.Equal(t, "test-auth-code", code)

		// Return a token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "test-access-token",
			"refresh_token": "test-refresh-token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer server.Close()

	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		AuthURL:      server.URL + "/authorize",
		TokenURL:     server.URL + "/token",
	})

	token, err := auth.Exchange(context.Background(), "test-auth-code")

	require.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, "test-access-token", token.AccessToken)
	assert.Equal(t, "test-refresh-token", token.RefreshToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, token, auth.GetToken())
}

func TestRefreshToken_NoToken(t *testing.T) {
	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})

	_, err := auth.RefreshToken(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token to refresh")
}

func TestRefreshToken_Success(t *testing.T) {
	// Create a test server that simulates the OAuth 2.0 token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		// Return a new token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "new-access-token",
			"refresh_token": "new-refresh-token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer server.Close()

	auth := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     server.URL,
	})

	// Set an initial token with a refresh token
	auth.SetToken(&oauth2.Token{
		AccessToken:  "old-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-time.Hour), // Expired
	})

	newToken, err := auth.RefreshToken(context.Background())

	require.NoError(t, err)
	require.NotNil(t, newToken)
	assert.Equal(t, "new-access-token", newToken.AccessToken)
	assert.Equal(t, newToken, auth.GetToken())
}

// --- Atlassian 3LO specifics ---

func TestGetAuthURLIncludesAtlassianRequiredParameters(t *testing.T) {
	a := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work"},
	})

	parsed, err := url.Parse(a.GetAuthURL("test-state"))
	require.NoError(t, err)

	q := parsed.Query()
	// Atlassian rejects the authorization request without these two.
	assert.Equal(t, DefaultOAuth2Audience, q.Get("audience"))
	assert.Equal(t, "consent", q.Get("prompt"))

	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "test-client-id", q.Get("client_id"))
	assert.Equal(t, "test-state", q.Get("state"))
	assert.Equal(t, "read:jira-work "+ScopeOfflineAccess, q.Get("scope"))

	// access_type is a Google convention and means nothing to Atlassian.
	assert.Empty(t, q.Get("access_type"))
}

func TestGetAuthURLCustomAudience(t *testing.T) {
	a := NewOAuth2Authenticator(&OAuth2Config{
		ClientID: "test-client-id",
		Audience: "custom.example.com",
	})

	parsed, err := url.Parse(a.GetAuthURL("state"))
	require.NoError(t, err)
	assert.Equal(t, "custom.example.com", parsed.Query().Get("audience"))
}

func TestOfflineAccessScopeHandling(t *testing.T) {
	tests := []struct {
		name   string
		config *OAuth2Config
		want   []string
	}{
		{
			name:   "appended when absent",
			config: &OAuth2Config{Scopes: []string{"read:jira-work"}},
			want:   []string{"read:jira-work", ScopeOfflineAccess},
		},
		{
			name:   "not duplicated when already present",
			config: &OAuth2Config{Scopes: []string{"read:jira-work", ScopeOfflineAccess}},
			want:   []string{"read:jira-work", ScopeOfflineAccess},
		},
		{
			name:   "omitted when disabled",
			config: &OAuth2Config{Scopes: []string{"read:jira-work"}, DisableOfflineAccess: true},
			want:   []string{"read:jira-work"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewOAuth2Authenticator(tt.config)
			assert.Equal(t, tt.want, a.config.Scopes)
		})
	}
}

// memoryTokenStore is an in-memory OAuth2TokenStore for tests.
type memoryTokenStore struct {
	mu      sync.Mutex
	token   *oauth2.Token
	saves   int
	loadErr error
}

func (s *memoryTokenStore) SaveToken(token *oauth2.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = token
	s.saves++
	return nil
}

func (s *memoryTokenStore) LoadToken() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return s.token, nil
}

func (s *memoryTokenStore) DeleteToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = nil
	return nil
}

func (s *memoryTokenStore) saved() (*oauth2.Token, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.token, s.saves
}

func TestTokenStorePersistsExchange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"at","refresh_token":"rt","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	store := &memoryTokenStore{}
	a := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:   "id",
		TokenURL:   server.URL,
		TokenStore: store,
	})

	_, err := a.Exchange(context.Background(), "code")
	require.NoError(t, err)

	token, saves := store.saved()
	require.NotNil(t, token)
	assert.Equal(t, "at", token.AccessToken)
	assert.Equal(t, "rt", token.RefreshToken)
	assert.Equal(t, 1, saves)
}

func TestTokenStorePersistsRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed","refresh_token":"rt2","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	store := &memoryTokenStore{}
	a := NewOAuth2Authenticator(&OAuth2Config{
		ClientID:   "id",
		TokenURL:   server.URL,
		TokenStore: store,
	})
	a.SetToken(&oauth2.Token{
		AccessToken:  "expired",
		RefreshToken: "rt",
		Expiry:       time.Now().Add(-time.Hour),
	})

	refreshed, err := a.RefreshToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "refreshed", refreshed.AccessToken)

	token, saves := store.saved()
	require.NotNil(t, token)
	assert.Equal(t, "refreshed", token.AccessToken)
	assert.Equal(t, 1, saves, "a refresh must be written back, or it is lost on restart")
}

func TestLoadTokenFromStore(t *testing.T) {
	store := &memoryTokenStore{token: &oauth2.Token{AccessToken: "stored"}}

	a := NewOAuth2Authenticator(&OAuth2Config{ClientID: "id", TokenStore: store})

	loaded, err := a.LoadToken()
	require.NoError(t, err)
	assert.Equal(t, "stored", loaded.AccessToken)
	assert.Equal(t, "stored", a.GetToken().AccessToken)
}

func TestLoadTokenWithoutStore(t *testing.T) {
	a := NewOAuth2Authenticator(&OAuth2Config{ClientID: "id"})

	_, err := a.LoadToken()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token store configured")
}

func TestAuthenticateRefreshesExpiredToken(t *testing.T) {
	var refreshes int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&refreshes, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	a := NewOAuth2Authenticator(&OAuth2Config{ClientID: "id", TokenURL: server.URL})
	a.SetToken(&oauth2.Token{
		AccessToken:  "expired",
		RefreshToken: "rt",
		Expiry:       time.Now().Add(-time.Hour),
	})

	req, err := http.NewRequest(http.MethodGet, "https://api.atlassian.com/test", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	assert.Equal(t, "Bearer refreshed", req.Header.Get("Authorization"))
	assert.Equal(t, int32(1), atomic.LoadInt32(&refreshes))
}

func TestConcurrentAuthenticateIsRaceFree(t *testing.T) {
	a := NewOAuth2Authenticator(&OAuth2Config{ClientID: "id"})
	a.SetToken(&oauth2.Token{AccessToken: "at", Expiry: time.Now().Add(time.Hour)})

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodGet, "https://api.atlassian.com/test", nil)
			assert.NoError(t, err)
			assert.NoError(t, a.Authenticate(req))

			if i%8 == 0 {
				a.SetToken(&oauth2.Token{AccessToken: "at", Expiry: time.Now().Add(time.Hour)})
			}
			_ = a.GetToken()
		}(i)
	}
	wg.Wait()
}
