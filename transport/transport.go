// Package transport provides HTTP transport layer with middleware support.
package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/felixgeelhaar/jirasdk/auth"
)

// Logger is the interface for structured logging in the transport layer.
// This interface matches the jira-connect Logger interface to avoid circular dependencies.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// RoundTripFunc is a function type that performs an HTTP round trip.
type RoundTripFunc func(ctx context.Context, req *http.Request) (*http.Response, error)

// Middleware is a function that wraps a RoundTripFunc to add functionality.
type Middleware func(next RoundTripFunc) RoundTripFunc

// Transport handles HTTP requests with middleware support.
type Transport struct {
	client          *http.Client
	baseURL         *url.URL
	authenticator   auth.Authenticator
	maxRetries      int
	rateLimitBuffer time.Duration
	userAgent       string
	logger          Logger
	middlewares     []Middleware
	roundTripper    RoundTripFunc
}

// Config holds transport configuration.
type Config struct {
	authenticator   auth.Authenticator
	maxRetries      int
	rateLimitBuffer time.Duration
	userAgent       string
	logger          Logger
	middlewares     []Middleware
}

// TransportOption is a functional option for configuring Transport.
type TransportOption func(*Config)

// New creates a new Transport with the given options.
func New(client *http.Client, baseURL *url.URL, opts ...TransportOption) *Transport {
	cfg := &Config{
		maxRetries:      3,
		rateLimitBuffer: 5 * time.Second,
		userAgent:       "jira-connect-go/1.0.0",
		middlewares:     []Middleware{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	t := &Transport{
		client:          client,
		baseURL:         baseURL,
		authenticator:   cfg.authenticator,
		maxRetries:      cfg.maxRetries,
		rateLimitBuffer: cfg.rateLimitBuffer,
		userAgent:       cfg.userAgent,
		logger:          cfg.logger,
		middlewares:     cfg.middlewares,
	}

	// Build middleware chain
	t.buildMiddlewareChain()

	return t
}

// WithAuthenticator sets the authenticator.
func WithAuthenticator(auth auth.Authenticator) TransportOption {
	return func(cfg *Config) {
		cfg.authenticator = auth
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(maxRetries int) TransportOption {
	return func(cfg *Config) {
		cfg.maxRetries = maxRetries
	}
}

// WithRateLimitBuffer sets the rate limit buffer duration.
func WithRateLimitBuffer(buffer time.Duration) TransportOption {
	return func(cfg *Config) {
		cfg.rateLimitBuffer = buffer
	}
}

// WithUserAgent sets the user agent string.
func WithUserAgent(userAgent string) TransportOption {
	return func(cfg *Config) {
		cfg.userAgent = userAgent
	}
}

// WithMiddlewares adds middleware to the transport.
func WithMiddlewares(middlewares ...Middleware) TransportOption {
	return func(cfg *Config) {
		cfg.middlewares = append(cfg.middlewares, middlewares...)
	}
}

// WithLogger sets the logger for the transport.
func WithLogger(logger interface{}) TransportOption {
	return func(cfg *Config) {
		cfg.logger = newLoggerAdapter(logger)
	}
}

// buildMiddlewareChain builds the middleware chain.
func (t *Transport) buildMiddlewareChain() {
	// Start with the base round tripper
	roundTripper := t.baseRoundTrip

	// Apply built-in middleware in order (innermost to outermost):

	// 1. Authentication (closest to the request)
	if t.authenticator != nil {
		roundTripper = authMiddleware(t.authenticator)(roundTripper)
	}

	// 2. User agent
	roundTripper = userAgentMiddleware(t.userAgent)(roundTripper)

	// 3. Rate limiting
	roundTripper = rateLimitMiddleware(t.rateLimitBuffer)(roundTripper)

	// 4. Retry logic
	roundTripper = retryMiddleware(t.maxRetries)(roundTripper)

	// 5. Logging (outermost - logs the final result after all retries)
	if t.logger != nil {
		roundTripper = loggingMiddleware(t.logger)(roundTripper)
	}

	// Apply custom middleware (outermost)
	for i := len(t.middlewares) - 1; i >= 0; i-- {
		roundTripper = t.middlewares[i](roundTripper)
	}

	t.roundTripper = roundTripper
}

// baseRoundTrip performs the actual HTTP request.
func (t *Transport) baseRoundTrip(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Clone the request with context
	req = req.Clone(ctx)

	// Execute the request
	resp, err := t.client.Do(req) //nolint:gosec // G704: URL is from user-configured Jira base URL, not tainted input
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}

// Do executes an HTTP request through the middleware chain.
func (t *Transport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Ensure request has context
	if req.Context() != ctx {
		req = req.Clone(ctx)
	}

	// Execute through middleware chain
	return t.roundTripper(ctx, req)
}

// NewRequest creates a new HTTP request with the base URL.
func (t *Transport) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// Resolve path against base URL
	u, err := t.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Encode request body as JSON
	var bodyReader io.Reader
	if body != nil {
		bodyReader, err = EncodeJSONRequest(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// DecodeResponse decodes a JSON response into the target.
func (t *Transport) DecodeResponse(resp *http.Response, target interface{}) error {
	return DecodeJSONResponse(resp, target)
}
