package jiraconnect

import (
	"context"
	"net/http"
	"time"
)

// Resilience defines the interface for resilience patterns.
// Implementations can use fortify or custom resilience strategies.
type Resilience interface {
	// ExecuteRequest wraps an HTTP request with resilience patterns
	ExecuteRequest(ctx context.Context, req *http.Request, do func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error)
}

// ResilienceConfig holds configuration for resilience patterns
type ResilienceConfig struct {
	// Circuit Breaker settings
	CircuitBreakerEnabled   bool
	CircuitBreakerThreshold int           // Consecutive failures before opening
	CircuitBreakerInterval  time.Duration // Time window for counting failures
	CircuitBreakerTimeout   time.Duration // Time to wait in open state before half-open

	// Retry settings
	RetryEnabled      bool
	RetryMaxAttempts  int
	RetryInitialDelay time.Duration
	RetryMaxDelay     time.Duration
	RetryMultiplier   float64
	RetryJitter       bool

	// Rate Limiting settings
	RateLimitEnabled bool
	RateLimitRate    int           // Requests per interval
	RateLimitBurst   int           // Maximum burst size
	RateLimitWindow  time.Duration // Time window for rate limiting

	// Timeout settings
	TimeoutEnabled bool
	TimeoutDuration time.Duration

	// Bulkhead settings
	BulkheadEnabled      bool
	BulkheadMaxConcurrent int           // Maximum concurrent requests
	BulkheadMaxQueue      int           // Maximum queued requests
	BulkheadQueueTimeout  time.Duration // Max time to wait in queue
}

// DefaultResilienceConfig returns a reasonable default configuration
func DefaultResilienceConfig() ResilienceConfig {
	return ResilienceConfig{
		// Circuit Breaker
		CircuitBreakerEnabled:   true,
		CircuitBreakerThreshold: 5,
		CircuitBreakerInterval:  time.Minute,
		CircuitBreakerTimeout:   30 * time.Second,

		// Retry
		RetryEnabled:      true,
		RetryMaxAttempts:  3,
		RetryInitialDelay: 100 * time.Millisecond,
		RetryMaxDelay:     10 * time.Second,
		RetryMultiplier:   2.0,
		RetryJitter:       true,

		// Rate Limiting
		RateLimitEnabled: true,
		RateLimitRate:    100, // 100 req/min for Jira Cloud
		RateLimitBurst:   10,
		RateLimitWindow:  time.Minute,

		// Timeout
		TimeoutEnabled:  true,
		TimeoutDuration: 30 * time.Second,

		// Bulkhead
		BulkheadEnabled:      true,
		BulkheadMaxConcurrent: 10,
		BulkheadMaxQueue:      20,
		BulkheadQueueTimeout:  5 * time.Second,
	}
}

// noopResilience is a resilience implementation that does nothing
type noopResilience struct{}

func (n *noopResilience) ExecuteRequest(ctx context.Context, req *http.Request, do func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error) {
	return do(ctx, req)
}

// NewNoopResilience creates a resilience implementation that does nothing.
// This is the default when no resilience is configured.
func NewNoopResilience() Resilience {
	return &noopResilience{}
}
