// Package jiraconnect provides an idiomatic Go client for Jira Cloud and Server/Data Center REST APIs.
//
// This library follows enterprise-grade patterns including:
//   - Functional options for flexible configuration
//   - Context propagation for cancellation and timeouts
//   - Automatic retry with exponential backoff
//   - Rate limiting and quota management
//   - Type-safe domain models
//
// Example usage:
//
//	client, err := jiraconnect.NewClient(
//		jiraconnect.WithBaseURL("https://your-domain.atlassian.net"),
//		jiraconnect.WithAPIToken("your-email@example.com", "your-api-token"),
//		jiraconnect.WithTimeout(30*time.Second),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	issue, err := client.Issue.Get(ctx, "PROJ-123")
//	if err != nil {
//		log.Fatal(err)
//	}
package jiraconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/felixgeelhaar/jira-connect/auth"
	"github.com/felixgeelhaar/jira-connect/core/agile"
	"github.com/felixgeelhaar/jira-connect/core/bulk"
	"github.com/felixgeelhaar/jira-connect/core/issue"
	"github.com/felixgeelhaar/jira-connect/core/permission"
	"github.com/felixgeelhaar/jira-connect/core/project"
	"github.com/felixgeelhaar/jira-connect/core/search"
	"github.com/felixgeelhaar/jira-connect/core/user"
	"github.com/felixgeelhaar/jira-connect/core/workflow"
	"github.com/felixgeelhaar/jira-connect/transport"
)

const (
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3

	// DefaultRateLimitBuffer is the buffer time before rate limit reset
	DefaultRateLimitBuffer = 5 * time.Second
)

// Client is the main Jira API client.
type Client struct {
	// BaseURL is the Jira instance URL (e.g., https://your-domain.atlassian.net)
	BaseURL *url.URL

	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client

	// Authenticator handles authentication for requests
	Authenticator auth.Authenticator

	// Transport provides HTTP transport with middleware support
	Transport *transport.Transport

	// Domain service clients
	Issue      *issue.Service
	Project    *project.Service
	User       *user.Service
	Workflow   *workflow.Service
	Search     *search.Service
	Agile      *agile.Service
	Permission *permission.Service
	Bulk       *bulk.Service
}

// Config holds the client configuration.
type Config struct {
	baseURL           *url.URL
	authenticator     auth.Authenticator
	httpClient        *http.Client
	timeout           time.Duration
	maxRetries        int
	rateLimitBuffer   time.Duration
	middlewares       []transport.Middleware
	userAgent         string
	enableCompression bool
	logger            Logger
	resilience        Resilience
}

// Option is a functional option for configuring the Client.
type Option func(*Config) error

// NewClient creates a new Jira API client with the provided options.
//
// At minimum, you must provide a base URL and authentication method.
//
// Example:
//
//	client, err := NewClient(
//		WithBaseURL("https://example.atlassian.net"),
//		WithAPIToken("user@example.com", "token"),
//	)
func NewClient(opts ...Option) (*Client, error) {
	// Initialize with defaults
	cfg := &Config{
		timeout:           DefaultTimeout,
		maxRetries:        DefaultMaxRetries,
		rateLimitBuffer:   DefaultRateLimitBuffer,
		userAgent:         "jira-connect-go/1.0.0",
		enableCompression: true,
		middlewares:       []transport.Middleware{},
		logger:            NewNoopLogger(),
		resilience:        NewNoopResilience(),
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate required configuration
	if cfg.baseURL == nil {
		return nil, fmt.Errorf("base URL is required")
	}

	if cfg.authenticator == nil {
		return nil, fmt.Errorf("authentication method is required")
	}

	// Create HTTP client if not provided
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{
			Timeout: cfg.timeout,
		}
	}

	// Create transport with middleware
	tr := transport.New(
		cfg.httpClient,
		cfg.baseURL,
		transport.WithAuthenticator(cfg.authenticator),
		transport.WithMaxRetries(cfg.maxRetries),
		transport.WithRateLimitBuffer(cfg.rateLimitBuffer),
		transport.WithUserAgent(cfg.userAgent),
		transport.WithLogger(cfg.logger),
		transport.WithMiddlewares(cfg.middlewares...),
	)

	client := &Client{
		BaseURL:       cfg.baseURL,
		HTTPClient:    cfg.httpClient,
		Authenticator: cfg.authenticator,
		Transport:     tr,
	}

	// Initialize domain services
	client.Issue = issue.NewService(tr)
	client.Project = project.NewService(tr)
	client.User = user.NewService(tr)
	client.Workflow = workflow.NewService(tr)
	client.Search = search.NewService(tr)
	client.Agile = agile.NewService(tr)
	client.Permission = permission.NewService(tr)
	client.Bulk = bulk.NewService(tr)

	return client, nil
}

// WithBaseURL sets the Jira instance base URL.
//
// Example:
//
//	WithBaseURL("https://your-domain.atlassian.net")
func WithBaseURL(baseURL string) Option {
	return func(cfg *Config) error {
		u, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("invalid base URL: %w", err)
		}

		if u.Scheme != "https" && u.Scheme != "http" {
			return fmt.Errorf("base URL must use http or https scheme")
		}

		cfg.baseURL = u
		return nil
	}
}

// WithAPIToken configures API token authentication for Jira Cloud.
//
// This is the recommended authentication method for Jira Cloud.
// Generate an API token at: https://id.atlassian.com/manage-profile/security/api-tokens
//
// Example:
//
//	WithAPIToken("user@example.com", "your-api-token")
func WithAPIToken(email, token string) Option {
	return func(cfg *Config) error {
		if email == "" || token == "" {
			return fmt.Errorf("email and token are required")
		}
		cfg.authenticator = auth.NewAPITokenAuth(email, token)
		return nil
	}
}

// WithPAT configures Personal Access Token authentication for Jira Server/Data Center.
//
// Example:
//
//	WithPAT("your-personal-access-token")
func WithPAT(token string) Option {
	return func(cfg *Config) error {
		if token == "" {
			return fmt.Errorf("PAT token is required")
		}
		cfg.authenticator = auth.NewPATAuth(token)
		return nil
	}
}

// WithOAuth2 configures OAuth 2.0 authentication.
//
// Example:
//
//	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURL:  "http://localhost:8080/callback",
//	    Scopes:       []string{"read:jira-work", "write:jira-work"},
//	})
//	WithOAuth2(oauth)
func WithOAuth2(oauth *auth.OAuth2Authenticator) Option {
	return func(cfg *Config) error {
		if oauth == nil {
			return fmt.Errorf("OAuth 2.0 authenticator is required")
		}
		cfg.authenticator = oauth
		return nil
	}
}

// WithBasicAuth configures basic authentication (legacy, not recommended).
//
// Example:
//
//	WithBasicAuth("username", "password")
func WithBasicAuth(username, password string) Option {
	return func(cfg *Config) error {
		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}
		cfg.authenticator = auth.NewBasicAuth(username, password)
		return nil
	}
}

// WithTimeout sets the HTTP client timeout.
//
// Example:
//
//	WithTimeout(60 * time.Second)
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout <= 0 {
			return fmt.Errorf("timeout must be positive")
		}
		cfg.timeout = timeout
		return nil
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
//
// Example:
//
//	WithMaxRetries(5)
func WithMaxRetries(maxRetries int) Option {
	return func(cfg *Config) error {
		if maxRetries < 0 {
			return fmt.Errorf("max retries must be non-negative")
		}
		cfg.maxRetries = maxRetries
		return nil
	}
}

// WithRateLimitBuffer sets the buffer time before rate limit reset.
//
// Example:
//
//	WithRateLimitBuffer(10 * time.Second)
func WithRateLimitBuffer(buffer time.Duration) Option {
	return func(cfg *Config) error {
		if buffer < 0 {
			return fmt.Errorf("rate limit buffer must be non-negative")
		}
		cfg.rateLimitBuffer = buffer
		return nil
	}
}

// WithHTTPClient provides a custom HTTP client.
//
// Example:
//
//	WithHTTPClient(&http.Client{Timeout: 60 * time.Second})
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *Config) error {
		if client == nil {
			return fmt.Errorf("HTTP client cannot be nil")
		}
		cfg.httpClient = client
		return nil
	}
}

// WithMiddleware adds a custom middleware to the transport chain.
//
// Example:
//
//	WithMiddleware(func(next transport.RoundTripFunc) transport.RoundTripFunc {
//		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
//			// Custom logic before request
//			resp, err := next(ctx, req)
//			// Custom logic after response
//			return resp, err
//		}
//	})
func WithMiddleware(middleware transport.Middleware) Option {
	return func(cfg *Config) error {
		cfg.middlewares = append(cfg.middlewares, middleware)
		return nil
	}
}

// WithUserAgent sets a custom user agent string.
//
// Example:
//
//	WithUserAgent("MyApp/1.0.0")
func WithUserAgent(userAgent string) Option {
	return func(cfg *Config) error {
		if userAgent == "" {
			return fmt.Errorf("user agent cannot be empty")
		}
		cfg.userAgent = userAgent
		return nil
	}
}

// WithLogger sets a custom logger for structured logging.
//
// By default, a no-op logger is used. Use the bolt adapter for
// zero-allocation structured logging with OpenTelemetry integration.
//
// Example:
//
//	import "github.com/felixgeelhaar/jira-connect/logger/bolt"
//	import "github.com/felixgeelhaar/bolt"
//
//	logger := bolt.New(bolt.NewJSONHandler(os.Stdout))
//	WithLogger(boltadapter.NewAdapter(logger))
func WithLogger(logger Logger) Option {
	return func(cfg *Config) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}
		cfg.logger = logger
		return nil
	}
}

// WithResilience sets custom resilience patterns for the client.
//
// By default, basic retry and rate limiting are used. Use the fortify adapter for
// production-grade resilience patterns including circuit breakers, advanced retries,
// rate limiting, timeouts, and bulkheads.
//
// Example:
//
//	import "github.com/felixgeelhaar/jira-connect/resilience/fortify"
//
//	resilience := fortify.NewAdapter(jira.DefaultResilienceConfig())
//	WithResilience(resilience)
func WithResilience(resilience Resilience) Option {
	return func(cfg *Config) error {
		if resilience == nil {
			return fmt.Errorf("resilience cannot be nil")
		}
		cfg.resilience = resilience
		return nil
	}
}

// Do executes an HTTP request with context.
//
// This is a low-level method for advanced use cases. Most users should
// use the domain-specific service methods instead.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.Transport.Do(ctx, req)
}
