package jirasdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearAuthEnv removes every authentication variable so each case starts clean.
func clearAuthEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		EnvBaseURL, EnvEmail, EnvAPIToken, EnvPAT, EnvUsername, EnvPassword,
		EnvOAuthClientID, EnvOAuthClientSecret, EnvOAuthRedirectURL, EnvOAuthScopes,
		EnvOAuthAccessToken, EnvOAuthRefreshToken, EnvOAuthTokenURL, EnvOAuthAudience,
		EnvCloudID, EnvConnectAppKey, EnvConnectSharedSecret, EnvConnectOAuthClientID,
		EnvConnectAccountID, EnvOAuth1ConsumerKey, EnvOAuth1ConsumerSecret,
		EnvOAuth1PrivateKey, EnvOAuth1PrivateKeyFile, EnvOAuth1Token,
		EnvOAuth1TokenSecret, EnvOAuth1ImpersonateUser,
	} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}
}

func TestWithEnvAuthMethods(t *testing.T) {
	keyPEM := string(testPrivateKeyPEM(t))

	tests := []struct {
		name     string
		env      map[string]string
		wantType string
	}{
		{
			name:     "API token",
			env:      map[string]string{EnvBaseURL: "https://example.atlassian.net", EnvEmail: "u@e.com", EnvAPIToken: "t"},
			wantType: "api_token",
		},
		{
			name:     "PAT",
			env:      map[string]string{EnvBaseURL: "https://jira.internal", EnvPAT: "pat"},
			wantType: "pat",
		},
		{
			name: "OAuth 2.0 3LO",
			env: map[string]string{
				EnvBaseURL:           "https://example.atlassian.net",
				EnvOAuthClientID:     "cid",
				EnvOAuthClientSecret: "secret",
				EnvOAuthRedirectURL:  "https://app.example.com/callback",
				EnvOAuthAccessToken:  "at",
			},
			wantType: "oauth2",
		},
		{
			name: "Connect JWT",
			env: map[string]string{
				EnvBaseURL:             "https://example.atlassian.net",
				EnvConnectAppKey:       "com.example.app",
				EnvConnectSharedSecret: "shared",
			},
			wantType: "connect_jwt",
		},
		{
			name: "Connect JWT bearer when an account is named",
			env: map[string]string{
				EnvBaseURL:              "https://example.atlassian.net",
				EnvConnectOAuthClientID: "oauth-cid",
				EnvConnectSharedSecret:  "shared",
				EnvConnectAccountID:     "account-id",
			},
			wantType: "jwt_bearer",
		},
		{
			name: "OAuth 1.0a two-legged",
			env: map[string]string{
				EnvBaseURL:           "https://jira.internal",
				EnvOAuth1ConsumerKey: "consumer",
				EnvOAuth1PrivateKey:  keyPEM,
			},
			wantType: "oauth1_2lo",
		},
		{
			name: "OAuth 1.0a three-legged",
			env: map[string]string{
				EnvBaseURL:           "https://jira.internal",
				EnvOAuth1ConsumerKey: "consumer",
				EnvOAuth1PrivateKey:  keyPEM,
				EnvOAuth1Token:       "token",
				EnvOAuth1TokenSecret: "secret",
			},
			wantType: "oauth1_3lo",
		},
		{
			name: "client credentials",
			env: map[string]string{
				EnvBaseURL:           "https://jira.example.com",
				EnvOAuthClientID:     "cid",
				EnvOAuthClientSecret: "secret",
				EnvOAuthTokenURL:     "https://idp.example.com/oauth2/token",
			},
			wantType: "client_credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAuthEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			client, err := NewClient(WithEnv())
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, client.Authenticator.Type())
		})
	}
}

func TestWithEnvOAuth1PrivateKeyFile(t *testing.T) {
	clearAuthEnv(t)

	keyPath := filepath.Join(t.TempDir(), "jira.pem")
	require.NoError(t, os.WriteFile(keyPath, testPrivateKeyPEM(t), 0o600))

	t.Setenv(EnvBaseURL, "https://jira.internal")
	t.Setenv(EnvOAuth1ConsumerKey, "consumer")
	t.Setenv(EnvOAuth1PrivateKeyFile, keyPath)

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.Equal(t, "oauth1_2lo", client.Authenticator.Type())
}

func TestWithEnvOAuth1MissingPrivateKeyFile(t *testing.T) {
	clearAuthEnv(t)

	t.Setenv(EnvBaseURL, "https://jira.internal")
	t.Setenv(EnvOAuth1ConsumerKey, "consumer")
	t.Setenv(EnvOAuth1PrivateKeyFile, filepath.Join(t.TempDir(), "absent.pem"))

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvOAuth1PrivateKeyFile)
}

func TestWithEnvCloudIDMakesBaseURLOptional(t *testing.T) {
	clearAuthEnv(t)

	t.Setenv(EnvCloudID, "cloud-id-one")
	t.Setenv(EnvOAuthClientID, "cid")
	t.Setenv(EnvOAuthClientSecret, "secret")
	t.Setenv(EnvOAuthRedirectURL, "https://app.example.com/callback")
	t.Setenv(EnvOAuthAccessToken, "at")

	client, err := NewClient(WithEnv())
	require.NoError(t, err)

	req, err := client.Transport.NewRequest(t.Context(), "GET", "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id-one/rest/api/3/myself", req.URL.String())
}

func TestWithEnvRequiresBaseURLWithoutCloudID(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv(EnvPAT, "pat")

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvBaseURL)
}

func TestWithEnvScopesOverride(t *testing.T) {
	clearAuthEnv(t)

	t.Setenv(EnvBaseURL, "https://example.atlassian.net")
	t.Setenv(EnvOAuthClientID, "cid")
	t.Setenv(EnvOAuthClientSecret, "secret")
	t.Setenv(EnvOAuthRedirectURL, "https://app.example.com/callback")
	t.Setenv(EnvOAuthScopes, "read:jira-user manage:jira-project")
	t.Setenv(EnvOAuthAccessToken, "at")

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.Equal(t, "oauth2", client.Authenticator.Type())
}

func TestWithEnvNoCredentialsListsEveryMethod(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv(EnvBaseURL, "https://example.atlassian.net")

	_, err := NewClient(WithEnv())
	require.Error(t, err)

	for _, mentioned := range []string{
		EnvAPIToken, EnvPAT, EnvOAuthRedirectURL, EnvConnectAppKey,
		EnvOAuth1ConsumerKey, EnvOAuthTokenURL,
	} {
		assert.Contains(t, err.Error(), mentioned)
	}
}
