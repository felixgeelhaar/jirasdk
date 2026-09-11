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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clientCredentialsServer struct {
	*httptest.Server
	calls     atomic.Int32
	mu        sync.Mutex
	lastForm  url.Values
	lastAuth  string
	expiresIn int64
}

func newClientCredentialsServer(t *testing.T) *clientCredentialsServer {
	t.Helper()

	s := &clientCredentialsServer{expiresIn: 3600}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())

		s.calls.Add(1)
		s.mu.Lock()
		s.lastForm = r.PostForm
		s.lastAuth = r.Header.Get("Authorization")
		expiresIn := s.expiresIn
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "cc-access-token",
			"token_type":   "Bearer",
			"expires_in":   expiresIn,
		})
	}))
	t.Cleanup(s.Close)

	return s
}

func (s *clientCredentialsServer) form() url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastForm
}

func TestNewClientCredentialsAuthValidation(t *testing.T) {
	valid := ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     "https://idp.example.com/oauth2/token",
	}

	tests := []struct {
		name    string
		mutate  func(*ClientCredentialsConfig)
		wantErr string
	}{
		{"missing client ID", func(c *ClientCredentialsConfig) { c.ClientID = "" }, "client ID is required"},
		{"missing client secret", func(c *ClientCredentialsConfig) { c.ClientSecret = "" }, "client secret is required"},
		{"missing token URL", func(c *ClientCredentialsConfig) { c.TokenURL = "" }, "token URL is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)

			a, err := NewClientCredentialsAuth(&cfg)
			require.Error(t, err)
			assert.Nil(t, a)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}

	t.Run("nil config", func(t *testing.T) {
		_, err := NewClientCredentialsAuth(nil)
		require.Error(t, err)
	})
}

func TestClientCredentialsAuthenticate(t *testing.T) {
	server := newClientCredentialsServer(t)

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
		Scopes:       []string{"jira:read", "jira:write"},
	})
	require.NoError(t, err)
	assert.Equal(t, "client_credentials", a.Type())

	req, err := http.NewRequest(http.MethodGet, "https://jira.example.com/rest/api/2/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	assert.Equal(t, "Bearer cc-access-token", req.Header.Get("Authorization"))

	form := server.form()
	assert.Equal(t, "client_credentials", form.Get("grant_type"))
	assert.Equal(t, "jira:read jira:write", form.Get("scope"))
}

func TestClientCredentialsSendsAudience(t *testing.T) {
	server := newClientCredentialsServer(t)

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
		Audience:     "https://jira.example.com",
	})
	require.NoError(t, err)

	_, err = a.Token(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://jira.example.com", server.form().Get("audience"))
}

func TestClientCredentialsSendsEndpointParams(t *testing.T) {
	server := newClientCredentialsServer(t)

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:       "client-id",
		ClientSecret:   "client-secret",
		TokenURL:       server.URL,
		EndpointParams: map[string][]string{"resource": {"jira"}},
	})
	require.NoError(t, err)

	_, err = a.Token(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "jira", server.form().Get("resource"))
}

func TestClientCredentialsCachesToken(t *testing.T) {
	server := newClientCredentialsServer(t)

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err := a.Token(context.Background())
		require.NoError(t, err)
	}

	assert.Equal(t, int32(1), server.calls.Load())
}

func TestClientCredentialsRenewsNearExpiry(t *testing.T) {
	server := newClientCredentialsServer(t)
	server.expiresIn = 5 // shorter than clientCredentialsRefreshLeeway

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	})
	require.NoError(t, err)

	_, err = a.Token(context.Background())
	require.NoError(t, err)
	_, err = a.Token(context.Background())
	require.NoError(t, err)

	assert.Equal(t, int32(2), server.calls.Load())
}

func TestClientCredentialsPropagatesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer server.Close()

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	})
	require.NoError(t, err)

	_, err = a.Token(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_client")
}

func TestClientCredentialsConcurrentTokenFetch(t *testing.T) {
	server := newClientCredentialsServer(t)

	a, err := NewClientCredentialsAuth(&ClientCredentialsConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	})
	require.NoError(t, err)

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

	assert.Equal(t, int32(1), server.calls.Load())
}
