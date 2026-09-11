package jirasdk

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/felixgeelhaar/jirasdk/auth"
	"golang.org/x/oauth2"
)

// Environment variable names for Jira SDK configuration
const (
	// Base URL configuration
	EnvBaseURL = "JIRA_BASE_URL" // Required: Jira instance URL

	// Authentication - API Token (Jira Cloud)
	EnvEmail    = "JIRA_EMAIL"     // Required for API token auth
	EnvAPIToken = "JIRA_API_TOKEN" //nolint:gosec // G101: Not a credential, just env var name

	// Authentication - Personal Access Token (Jira Server/Data Center)
	EnvPAT = "JIRA_PAT" //nolint:gosec // G101: Not a hardcoded credential

	// Authentication - Basic Auth (Legacy, not recommended)
	EnvUsername = "JIRA_USERNAME" // Legacy basic auth username
	EnvPassword = "JIRA_PASSWORD" //nolint:gosec // G101: Not a hardcoded credential

	// OAuth 2.0 (3LO) configuration
	EnvOAuthClientID     = "JIRA_OAUTH_CLIENT_ID"
	EnvOAuthClientSecret = "JIRA_OAUTH_CLIENT_SECRET" //nolint:gosec // G101: Not a credential, just env var name
	EnvOAuthRedirectURL  = "JIRA_OAUTH_REDIRECT_URL"
	EnvOAuthScopes       = "JIRA_OAUTH_SCOPES"        // Space-separated scope list
	EnvOAuthAccessToken  = "JIRA_OAUTH_ACCESS_TOKEN"  //nolint:gosec // G101: Not a hardcoded credential
	EnvOAuthRefreshToken = "JIRA_OAUTH_REFRESH_TOKEN" //nolint:gosec // G101: Not a hardcoded credential

	// OAuth 2.0 client credentials (two-legged) configuration
	EnvOAuthTokenURL = "JIRA_OAUTH_TOKEN_URL" //nolint:gosec // G101: Not a credential, just env var name
	EnvOAuthAudience = "JIRA_OAUTH_AUDIENCE"

	// Jira Cloud site selection for OAuth 2.0 (3LO). Setting this skips the
	// accessible-resources lookup.
	EnvCloudID = "JIRA_CLOUD_ID"

	// Atlassian Connect (two-legged) configuration
	EnvConnectAppKey        = "JIRA_CONNECT_APP_KEY"
	EnvConnectSharedSecret  = "JIRA_CONNECT_SHARED_SECRET" //nolint:gosec // G101: Not a hardcoded credential
	EnvConnectOAuthClientID = "JIRA_CONNECT_OAUTH_CLIENT_ID"
	EnvConnectAccountID     = "JIRA_CONNECT_ACCOUNT_ID" // Account ID to act as

	// OAuth 1.0a (Jira Server/Data Center) configuration
	EnvOAuth1ConsumerKey     = "JIRA_OAUTH1_CONSUMER_KEY"
	EnvOAuth1ConsumerSecret  = "JIRA_OAUTH1_CONSUMER_SECRET" //nolint:gosec // G101: Not a hardcoded credential
	EnvOAuth1PrivateKey      = "JIRA_OAUTH1_PRIVATE_KEY"     //nolint:gosec // G101: Not a hardcoded credential
	EnvOAuth1PrivateKeyFile  = "JIRA_OAUTH1_PRIVATE_KEY_FILE"
	EnvOAuth1Token           = "JIRA_OAUTH1_TOKEN"        //nolint:gosec // G101: Not a hardcoded credential
	EnvOAuth1TokenSecret     = "JIRA_OAUTH1_TOKEN_SECRET" //nolint:gosec // G101: Not a hardcoded credential
	EnvOAuth1ImpersonateUser = "JIRA_OAUTH1_IMPERSONATE_USER"

	// Client configuration
	EnvTimeout      = "JIRA_TIMEOUT"           // HTTP timeout in seconds (default: 30)
	EnvMaxRetries   = "JIRA_MAX_RETRIES"       // Max retry attempts (default: 3)
	EnvRateLimitBuf = "JIRA_RATE_LIMIT_BUFFER" // Rate limit buffer in seconds (default: 5)
	EnvUserAgent    = "JIRA_USER_AGENT"        // Custom user agent string
)

// WithEnv configures the client from environment variables.
//
// This option loads configuration from standard environment variables,
// following the pattern used by AWS SDK, Azure SDK, and other enterprise SDKs.
//
// Required environment variables:
//   - JIRA_BASE_URL: Your Jira instance URL (e.g., https://your-domain.atlassian.net)
//
// Authentication, tried in this order:
//   - JIRA_EMAIL + JIRA_API_TOKEN: API token auth for Jira Cloud (recommended)
//   - JIRA_PAT: Personal Access Token for Jira Server/Data Center
//   - JIRA_USERNAME + JIRA_PASSWORD: Basic auth (legacy, not recommended)
//   - JIRA_OAUTH_CLIENT_ID + JIRA_OAUTH_CLIENT_SECRET + JIRA_OAUTH_REDIRECT_URL:
//     OAuth 2.0 (3LO) for Jira Cloud. Add JIRA_OAUTH_ACCESS_TOKEN and
//     JIRA_OAUTH_REFRESH_TOKEN to restore a previously granted token, and
//     JIRA_OAUTH_SCOPES to override the requested scopes.
//   - JIRA_CONNECT_APP_KEY + JIRA_CONNECT_SHARED_SECRET: Connect JWT (two-legged)
//     for Jira Cloud. Add JIRA_CONNECT_OAUTH_CLIENT_ID + JIRA_CONNECT_ACCOUNT_ID
//     to use the JWT bearer grant and act as that user instead.
//   - JIRA_OAUTH1_CONSUMER_KEY + JIRA_OAUTH1_PRIVATE_KEY (or
//     JIRA_OAUTH1_PRIVATE_KEY_FILE): OAuth 1.0a for Jira Server/Data Center.
//     Two-legged by default; add JIRA_OAUTH1_TOKEN + JIRA_OAUTH1_TOKEN_SECRET
//     for three-legged, or JIRA_OAUTH1_IMPERSONATE_USER to act as a user.
//   - JIRA_OAUTH_CLIENT_ID + JIRA_OAUTH_CLIENT_SECRET + JIRA_OAUTH_TOKEN_URL:
//     OAuth 2.0 client credentials, for Jira behind a gateway or custom IdP.
//
// Jira Cloud site selection:
//   - JIRA_CLOUD_ID: pins the site for OAuth 2.0 (3LO), skipping the
//     accessible-resources lookup. When set, JIRA_BASE_URL is optional.
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
		// Load base URL. It is required unless a cloud ID pins the Jira Cloud
		// site, which determines the API base URL on its own.
		cloudID := os.Getenv(EnvCloudID)
		baseURL := os.Getenv(EnvBaseURL)
		if baseURL == "" && cloudID == "" {
			return fmt.Errorf("environment variable %s is required", EnvBaseURL)
		}

		if baseURL != "" {
			if err := WithBaseURL(baseURL)(cfg); err != nil {
				return fmt.Errorf("invalid %s: %w", EnvBaseURL, err)
			}
		}

		if cloudID != "" {
			if err := WithCloudID(cloudID)(cfg); err != nil {
				return fmt.Errorf("invalid %s: %w", EnvCloudID, err)
			}
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

	// Priority 4: OAuth 2.0 (3LO)
	clientID := os.Getenv(EnvOAuthClientID)
	clientSecret := os.Getenv(EnvOAuthClientSecret)
	redirectURL := os.Getenv(EnvOAuthRedirectURL)
	if clientID != "" && clientSecret != "" && redirectURL != "" {
		return configureOAuth2FromEnv(cfg, clientID, clientSecret, redirectURL)
	}

	// Priority 5: Atlassian Connect (two-legged)
	if authenticator, ok, err := connectAuthFromEnv(); ok || err != nil {
		if err != nil {
			return err
		}
		return WithAuthenticator(authenticator)(cfg)
	}

	// Priority 6: OAuth 1.0a (Jira Server/Data Center)
	if authenticator, ok, err := oauth1AuthFromEnv(); ok || err != nil {
		if err != nil {
			return err
		}
		return WithAuthenticator(authenticator)(cfg)
	}

	// Priority 7: OAuth 2.0 client credentials (two-legged)
	tokenURL := os.Getenv(EnvOAuthTokenURL)
	if clientID != "" && clientSecret != "" && tokenURL != "" {
		return WithClientCredentials(&auth.ClientCredentialsConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
			Scopes:       strings.Fields(os.Getenv(EnvOAuthScopes)),
			Audience:     os.Getenv(EnvOAuthAudience),
		})(cfg)
	}

	// No valid authentication found
	return fmt.Errorf("no valid authentication credentials found in environment variables; set one of: "+
		"(%s + %s) for API token, %s for PAT, (%s + %s) for basic auth, "+
		"(%s + %s + %s) for OAuth 2.0 (3LO), (%s + %s) for Connect JWT, "+
		"(%s + %s) for OAuth 1.0a, or (%s + %s + %s) for client credentials",
		EnvEmail, EnvAPIToken, EnvPAT, EnvUsername, EnvPassword,
		EnvOAuthClientID, EnvOAuthClientSecret, EnvOAuthRedirectURL,
		EnvConnectAppKey, EnvConnectSharedSecret,
		EnvOAuth1ConsumerKey, EnvOAuth1PrivateKey,
		EnvOAuthClientID, EnvOAuthClientSecret, EnvOAuthTokenURL)
}

// configureOAuth2FromEnv builds the 3LO authenticator, restoring a stored token
// when one is present so a configured client is usable straight away.
func configureOAuth2FromEnv(cfg *Config, clientID, clientSecret, redirectURL string) error {
	scopes := strings.Fields(os.Getenv(EnvOAuthScopes))
	if len(scopes) == 0 {
		scopes = []string{"read:jira-work", "write:jira-work"}
	}

	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	})

	if accessToken := os.Getenv(EnvOAuthAccessToken); accessToken != "" {
		oauth.SetToken(&oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: os.Getenv(EnvOAuthRefreshToken),
			TokenType:    "Bearer",
		})
	}

	return WithOAuth2(oauth)(cfg)
}

// connectAuthFromEnv builds a Connect authenticator when the environment
// describes one. An account ID selects the JWT bearer grant, which acts as that
// user; otherwise the app authenticates as itself with a Connect JWT.
func connectAuthFromEnv() (auth.Authenticator, bool, error) {
	appKey := os.Getenv(EnvConnectAppKey)
	sharedSecret := os.Getenv(EnvConnectSharedSecret)
	oauthClientID := os.Getenv(EnvConnectOAuthClientID)
	accountID := os.Getenv(EnvConnectAccountID)

	if sharedSecret == "" || (appKey == "" && oauthClientID == "") {
		return nil, false, nil
	}

	if accountID != "" && oauthClientID != "" {
		authenticator, err := auth.NewJWTBearerAuth(&auth.JWTBearerConfig{
			OAuthClientID: oauthClientID,
			SharedSecret:  sharedSecret,
			AccountID:     accountID,
			SiteURL:       os.Getenv(EnvBaseURL),
		})
		if err != nil {
			return nil, true, fmt.Errorf("invalid Connect JWT bearer configuration: %w", err)
		}
		return authenticator, true, nil
	}

	authenticator, err := auth.NewConnectJWTAuth(&auth.ConnectJWTConfig{
		AppKey:       appKey,
		SharedSecret: sharedSecret,
		Subject:      accountID,
	})
	if err != nil {
		return nil, true, fmt.Errorf("invalid Connect JWT configuration: %w", err)
	}

	return authenticator, true, nil
}

// oauth1AuthFromEnv builds an OAuth 1.0a authenticator when the environment
// describes one. Supplying a token selects the three-legged flow; otherwise
// requests are signed two-legged.
func oauth1AuthFromEnv() (auth.Authenticator, bool, error) {
	consumerKey := os.Getenv(EnvOAuth1ConsumerKey)
	if consumerKey == "" {
		return nil, false, nil
	}

	privateKey := []byte(os.Getenv(EnvOAuth1PrivateKey))
	if keyFile := os.Getenv(EnvOAuth1PrivateKeyFile); len(privateKey) == 0 && keyFile != "" {
		contents, err := os.ReadFile(keyFile) //nolint:gosec // G304: path is operator-supplied configuration
		if err != nil {
			return nil, true, fmt.Errorf("failed to read %s: %w", EnvOAuth1PrivateKeyFile, err)
		}
		privateKey = contents
	}

	config := &auth.OAuth1Config{
		ConsumerKey:     consumerKey,
		ConsumerSecret:  os.Getenv(EnvOAuth1ConsumerSecret),
		PrivateKeyPEM:   privateKey,
		Token:           os.Getenv(EnvOAuth1Token),
		TokenSecret:     os.Getenv(EnvOAuth1TokenSecret),
		ImpersonateUser: os.Getenv(EnvOAuth1ImpersonateUser),
		BaseURL:         os.Getenv(EnvBaseURL),
	}
	if len(privateKey) == 0 {
		config.SignatureMethod = auth.SignatureMethodHMACSHA1
	}

	authenticator, err := auth.NewOAuth1Auth(config)
	if err != nil {
		return nil, true, fmt.Errorf("invalid OAuth 1.0a configuration: %w", err)
	}

	return authenticator, true, nil
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
