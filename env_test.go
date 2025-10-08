package jirasdk

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithEnv_APIToken(t *testing.T) {
	// Setup environment
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://test.atlassian.net", client.BaseURL.String())
	assert.NotNil(t, client.Authenticator)
}

func TestWithEnv_PAT(t *testing.T) {
	// Setup environment
	os.Setenv(EnvBaseURL, "https://jira.company.com")
	os.Setenv(EnvPAT, "test-pat-token")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://jira.company.com", client.BaseURL.String())
	assert.NotNil(t, client.Authenticator)
}

func TestWithEnv_BasicAuth(t *testing.T) {
	// Setup environment
	os.Setenv(EnvBaseURL, "https://jira.company.com")
	os.Setenv(EnvUsername, "testuser")
	os.Setenv(EnvPassword, "testpass")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://jira.company.com", client.BaseURL.String())
	assert.NotNil(t, client.Authenticator)
}

func TestWithEnv_OAuth2(t *testing.T) {
	// Setup environment
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvOAuthClientID, "test-client-id")
	os.Setenv(EnvOAuthClientSecret, "test-client-secret")
	os.Setenv(EnvOAuthRedirectURL, "http://localhost:8080/callback")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://test.atlassian.net", client.BaseURL.String())
	assert.NotNil(t, client.Authenticator)
}

func TestWithEnv_MissingBaseURL(t *testing.T) {
	// Setup environment - missing base URL
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvBaseURL)
}

func TestWithEnv_MissingAuth(t *testing.T) {
	// Setup environment - missing authentication
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid authentication credentials")
}

func TestWithEnv_PartialAPIToken(t *testing.T) {
	// Setup environment - email without token
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	// Missing JIRA_API_TOKEN
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid authentication credentials")
}

func TestWithEnv_OptionalTimeout(t *testing.T) {
	// Setup environment with custom timeout
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvTimeout, "60")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 60*time.Second, client.HTTPClient.Timeout)
}

func TestWithEnv_OptionalMaxRetries(t *testing.T) {
	// Setup environment with custom max retries
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvMaxRetries, "5")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	// Max retries is internal to transport, can't directly assert
}

func TestWithEnv_OptionalUserAgent(t *testing.T) {
	// Setup environment with custom user agent
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvUserAgent, "MyApp/1.0.0")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	// User agent is internal to transport, can't directly assert
}

func TestWithEnv_InvalidTimeout(t *testing.T) {
	// Setup environment with invalid timeout
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvTimeout, "invalid")
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvTimeout)
}

func TestWithEnv_NegativeTimeout(t *testing.T) {
	// Setup environment with negative timeout
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvTimeout, "-10")
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvTimeout)
}

func TestWithEnv_InvalidMaxRetries(t *testing.T) {
	// Setup environment with invalid max retries
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvMaxRetries, "invalid")
	defer cleanupEnv()

	_, err := NewClient(WithEnv())
	require.Error(t, err)
	assert.Contains(t, err.Error(), EnvMaxRetries)
}

func TestWithEnv_CombinedWithOptions(t *testing.T) {
	// Test combining WithEnv with other options
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	defer cleanupEnv()

	// Override timeout with explicit option
	client, err := NewClient(
		WithEnv(),
		WithTimeout(90*time.Second),
	)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 90*time.Second, client.HTTPClient.Timeout)
}

func TestWithEnv_AuthPriority(t *testing.T) {
	// Test that API token takes priority when multiple auth methods are set
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	os.Setenv(EnvPAT, "test-pat")
	os.Setenv(EnvUsername, "testuser")
	os.Setenv(EnvPassword, "testpass")
	defer cleanupEnv()

	client, err := NewClient(WithEnv())
	require.NoError(t, err)
	assert.NotNil(t, client)
	// API token should be used (priority 1)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Setup environment
	os.Setenv(EnvBaseURL, "https://test.atlassian.net")
	os.Setenv(EnvEmail, "test@example.com")
	os.Setenv(EnvAPIToken, "test-token")
	defer cleanupEnv()

	client, err := LoadConfigFromEnv()
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://test.atlassian.net", client.BaseURL.String())
}

func TestLoadConfigFromEnv_MissingEnv(t *testing.T) {
	// Clean environment
	cleanupEnv()

	_, err := LoadConfigFromEnv()
	require.Error(t, err)
}

// cleanupEnv removes all jirasdk environment variables
func cleanupEnv() {
	envVars := []string{
		EnvBaseURL,
		EnvEmail,
		EnvAPIToken,
		EnvPAT,
		EnvUsername,
		EnvPassword,
		EnvOAuthClientID,
		EnvOAuthClientSecret,
		EnvOAuthRedirectURL,
		EnvTimeout,
		EnvMaxRetries,
		EnvRateLimitBuf,
		EnvUserAgent,
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
