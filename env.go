package jirasdk

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/felixgeelhaar/jirasdk/auth"
)

// Environment variable names for Jira SDK configuration
const (
	// Base URL configuration
	EnvBaseURL = "JIRA_BASE_URL" // Required: Jira instance URL

	// Authentication - API Token (Jira Cloud)
	EnvEmail    = "JIRA_EMAIL"     // Required for API token auth
	EnvAPIToken = "JIRA_API_TOKEN" // Required for API token auth

	// Authentication - Personal Access Token (Jira Server/Data Center)
	EnvPAT = "JIRA_PAT" // Alternative to API token

	// Authentication - Basic Auth (Legacy, not recommended)
	EnvUsername = "JIRA_USERNAME" // Legacy basic auth username
	EnvPassword = "JIRA_PASSWORD" // Legacy basic auth password

	// OAuth 2.0 configuration
	EnvOAuthClientID     = "JIRA_OAUTH_CLIENT_ID"
	EnvOAuthClientSecret = "JIRA_OAUTH_CLIENT_SECRET"
	EnvOAuthRedirectURL  = "JIRA_OAUTH_REDIRECT_URL"

	// Client configuration
	EnvTimeout      = "JIRA_TIMEOUT"       // HTTP timeout in seconds (default: 30)
	EnvMaxRetries   = "JIRA_MAX_RETRIES"   // Max retry attempts (default: 3)
	EnvRateLimitBuf = "JIRA_RATE_LIMIT_BUFFER" // Rate limit buffer in seconds (default: 5)
	EnvUserAgent    = "JIRA_USER_AGENT"    // Custom user agent string
)

// WithEnv configures the client from environment variables.
//
// This option loads configuration from standard environment variables,
// following the pattern used by AWS SDK, Azure SDK, and other enterprise SDKs.
//
// Required environment variables:
//   - JIRA_BASE_URL: Your Jira instance URL (e.g., https://your-domain.atlassian.net)
//
// Authentication (one of the following):
//   - JIRA_EMAIL + JIRA_API_TOKEN: API token auth for Jira Cloud (recommended)
//   - JIRA_PAT: Personal Access Token for Jira Server/Data Center
//   - JIRA_USERNAME + JIRA_PASSWORD: Basic auth (legacy, not recommended)
//
// Optional configuration:
//   - JIRA_TIMEOUT: HTTP timeout in seconds (default: 30)
//   - JIRA_MAX_RETRIES: Maximum retry attempts (default: 3)
//   - JIRA_RATE_LIMIT_BUFFER: Rate limit buffer in seconds (default: 5)
//   - JIRA_USER_AGENT: Custom user agent string
//
// Example:
//
//	export JIRA_BASE_URL="https://your-domain.atlassian.net"
//	export JIRA_EMAIL="user@example.com"
//	export JIRA_API_TOKEN="your-api-token"
//
//	client, err := jirasdk.NewClient(
//	    jirasdk.WithEnv(),
//	)
//
// For advanced use cases, combine WithEnv() with other options:
//
//	client, err := jirasdk.NewClient(
//	    jirasdk.WithEnv(),              // Load from environment
//	    jirasdk.WithTimeout(60*time.Second), // Override timeout
//	)
func WithEnv() Option {
	return func(cfg *Config) error {
		// Load base URL (required)
		baseURL := os.Getenv(EnvBaseURL)
		if baseURL == "" {
			return fmt.Errorf("environment variable %s is required", EnvBaseURL)
		}

		// Apply base URL
		if err := WithBaseURL(baseURL)(cfg); err != nil {
			return fmt.Errorf("invalid %s: %w", EnvBaseURL, err)
		}

		// Determine authentication method based on available environment variables
		if err := configureAuthFromEnv(cfg); err != nil {
			return err
		}

		// Load optional configuration
		if err := configureOptionalFromEnv(cfg); err != nil {
			return err
		}

		return nil
	}
}

// configureAuthFromEnv determines and configures authentication from environment variables
func configureAuthFromEnv(cfg *Config) error {
	// Priority 1: API Token (Jira Cloud - recommended)
	email := os.Getenv(EnvEmail)
	apiToken := os.Getenv(EnvAPIToken)
	if email != "" && apiToken != "" {
		return WithAPIToken(email, apiToken)(cfg)
	}

	// Priority 2: Personal Access Token (Jira Server/Data Center)
	pat := os.Getenv(EnvPAT)
	if pat != "" {
		return WithPAT(pat)(cfg)
	}

	// Priority 3: Basic Auth (Legacy)
	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)
	if username != "" && password != "" {
		return WithBasicAuth(username, password)(cfg)
	}

	// Priority 4: OAuth 2.0 (if configured)
	clientID := os.Getenv(EnvOAuthClientID)
	clientSecret := os.Getenv(EnvOAuthClientSecret)
	redirectURL := os.Getenv(EnvOAuthRedirectURL)
	if clientID != "" && clientSecret != "" && redirectURL != "" {
		oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"read:jira-work", "write:jira-work"},
		})
		return WithOAuth2(oauth)(cfg)
	}

	// No valid authentication found
	return fmt.Errorf("no valid authentication credentials found in environment variables; " +
		"set either (%s + %s) for API token, %s for PAT, or (%s + %s) for basic auth",
		EnvEmail, EnvAPIToken, EnvPAT, EnvUsername, EnvPassword)
}

// configureOptionalFromEnv loads optional configuration from environment variables
func configureOptionalFromEnv(cfg *Config) error {
	// Timeout
	if timeoutStr := os.Getenv(EnvTimeout); timeoutStr != "" {
		timeoutSec, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid %s: must be an integer (seconds)", EnvTimeout)
		}
		if timeoutSec <= 0 {
			return fmt.Errorf("invalid %s: must be positive", EnvTimeout)
		}
		if err := WithTimeout(time.Duration(timeoutSec) * time.Second)(cfg); err != nil {
			return err
		}
	}

	// Max retries
	if maxRetriesStr := os.Getenv(EnvMaxRetries); maxRetriesStr != "" {
		maxRetries, err := strconv.Atoi(maxRetriesStr)
		if err != nil {
			return fmt.Errorf("invalid %s: must be an integer", EnvMaxRetries)
		}
		if maxRetries < 0 {
			return fmt.Errorf("invalid %s: must be non-negative", EnvMaxRetries)
		}
		if err := WithMaxRetries(maxRetries)(cfg); err != nil {
			return err
		}
	}

	// Rate limit buffer
	if bufferStr := os.Getenv(EnvRateLimitBuf); bufferStr != "" {
		bufferSec, err := strconv.Atoi(bufferStr)
		if err != nil {
			return fmt.Errorf("invalid %s: must be an integer (seconds)", EnvRateLimitBuf)
		}
		if bufferSec < 0 {
			return fmt.Errorf("invalid %s: must be non-negative", EnvRateLimitBuf)
		}
		if err := WithRateLimitBuffer(time.Duration(bufferSec) * time.Second)(cfg); err != nil {
			return err
		}
	}

	// User agent
	if userAgent := os.Getenv(EnvUserAgent); userAgent != "" {
		if err := WithUserAgent(userAgent)(cfg); err != nil {
			return err
		}
	}

	return nil
}

// LoadConfigFromEnv is a convenience function that creates a new client
// configured entirely from environment variables.
//
// This is equivalent to:
//
//	jirasdk.NewClient(jirasdk.WithEnv())
//
// Example:
//
//	export JIRA_BASE_URL="https://your-domain.atlassian.net"
//	export JIRA_EMAIL="user@example.com"
//	export JIRA_API_TOKEN="your-api-token"
//
//	client, err := jirasdk.LoadConfigFromEnv()
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadConfigFromEnv() (*Client, error) {
	return NewClient(WithEnv())
}
