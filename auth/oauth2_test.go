package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	assert.Equal(t, config.Scopes, auth.config.Scopes)
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
