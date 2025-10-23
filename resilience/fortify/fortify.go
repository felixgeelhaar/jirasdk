// Package fortify provides a fortify-based resilience adapter for jira-connect.
//
// This adapter integrates fortify's resilience patterns (circuit breakers, retries,
// rate limiting, timeouts, and bulkheads) with the jira-connect HTTP client.
//
// Example usage:
//
//	resilience := fortify.NewAdapter(jira.DefaultResilienceConfig())
//	client, err := jira.NewClient(
//		jira.WithBaseURL("https://your-domain.atlassian.net"),
//		jira.WithAPIToken("email", "token"),
//		jira.WithResilience(resilience),
//	)
package fortify

import (
	"context"
	"net/http"

	"github.com/felixgeelhaar/fortify/bulkhead"
	"github.com/felixgeelhaar/fortify/circuitbreaker"
	"github.com/felixgeelhaar/fortify/ratelimit"
	"github.com/felixgeelhaar/fortify/retry"
	"github.com/felixgeelhaar/fortify/timeout"
	jira "github.com/felixgeelhaar/jirasdk"
)

// Adapter adapts fortify resilience patterns to the jira-connect Resilience interface
type Adapter struct {
	circuitBreaker circuitbreaker.CircuitBreaker[*http.Response]
	retrier        retry.Retry[*http.Response]
	rateLimiter    ratelimit.RateLimiter
	timer          timeout.Timeout[*http.Response]
	bulkheadPool   bulkhead.Bulkhead[*http.Response]
	config         jira.ResilienceConfig
}

// NewAdapter creates a new fortify adapter with the given configuration
func NewAdapter(config jira.ResilienceConfig) *Adapter {
	adapter := &Adapter{
		config: config,
	}

	// Initialize circuit breaker if enabled
	if config.CircuitBreakerEnabled {
		adapter.circuitBreaker = circuitbreaker.New[*http.Response](circuitbreaker.Config{
			MaxRequests: 100,
			Interval:    config.CircuitBreakerInterval,
			Timeout:     config.CircuitBreakerTimeout,
			ReadyToTrip: func(counts circuitbreaker.Counts) bool {
				return counts.ConsecutiveFailures >= uint32(config.CircuitBreakerThreshold) // #nosec G115 - Safe conversion, config value is validated
			},
		})
	}

	// Initialize retry if enabled
	if config.RetryEnabled {
		adapter.retrier = retry.New[*http.Response](retry.Config{
			MaxAttempts:   config.RetryMaxAttempts,
			InitialDelay:  config.RetryInitialDelay,
			MaxDelay:      config.RetryMaxDelay,
			BackoffPolicy: retry.BackoffExponential,
			Multiplier:    config.RetryMultiplier,
			Jitter:        config.RetryJitter,
			IsRetryable: func(err error) bool {
				// Retry on network errors and 5xx status codes
				return isRetryableError(err)
			},
		})
	}

	// Initialize rate limiter if enabled
	if config.RateLimitEnabled {
		adapter.rateLimiter = ratelimit.New(ratelimit.Config{
			Rate:     config.RateLimitRate,
			Burst:    config.RateLimitBurst,
			Interval: config.RateLimitWindow,
		})
	}

	// Initialize timeout if enabled
	if config.TimeoutEnabled {
		adapter.timer = timeout.New[*http.Response](timeout.Config{
			DefaultTimeout: config.TimeoutDuration,
		})
	}

	// Initialize bulkhead if enabled
	if config.BulkheadEnabled {
		adapter.bulkheadPool = bulkhead.New[*http.Response](bulkhead.Config{
			MaxConcurrent: config.BulkheadMaxConcurrent,
			MaxQueue:      config.BulkheadMaxQueue,
			QueueTimeout:  config.BulkheadQueueTimeout,
		})
	}

	return adapter
}

// ExecuteRequest wraps an HTTP request with all enabled resilience patterns
func (a *Adapter) ExecuteRequest(ctx context.Context, req *http.Request, do func(context.Context, *http.Request) (*http.Response, error)) (*http.Response, error) {
	// Build the execution function with all enabled patterns
	execute := func(ctx context.Context) (*http.Response, error) {
		return do(ctx, req)
	}

	// Apply patterns in order: Bulkhead -> RateLimit -> Timeout -> CircuitBreaker -> Retry

	// 1. Bulkhead - Limit concurrency
	if a.bulkheadPool != nil {
		bulkheadExecute := execute
		execute = func(ctx context.Context) (*http.Response, error) {
			return a.bulkheadPool.Execute(ctx, bulkheadExecute)
		}
	}

	// 2. Rate Limit - Check quotas
	if a.rateLimiter != nil {
		rateLimitExecute := execute
		execute = func(ctx context.Context) (*http.Response, error) {
			// Wait for rate limit token
			if err := a.rateLimiter.Wait(ctx, "jira-api"); err != nil {
				return nil, err
			}
			return rateLimitExecute(ctx)
		}
	}

	// 3. Timeout - Enforce time limits
	if a.timer != nil {
		timeoutExecute := execute
		timeout := a.config.TimeoutDuration
		execute = func(ctx context.Context) (*http.Response, error) {
			return a.timer.Execute(ctx, timeout, timeoutExecute)
		}
	}

	// 4. Circuit Breaker - Check service health
	if a.circuitBreaker != nil {
		circuitBreakerExecute := execute
		execute = func(ctx context.Context) (*http.Response, error) {
			return a.circuitBreaker.Execute(ctx, circuitBreakerExecute)
		}
	}

	// 5. Retry - Handle transient failures
	if a.retrier != nil {
		retryExecute := execute
		execute = func(ctx context.Context) (*http.Response, error) {
			return a.retrier.Do(ctx, retryExecute)
		}
	}

	return execute(ctx)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Retry on specific HTTP status codes if we can determine them
	// For now, retry on all errors - can be made more sophisticated
	return true
}

// GetCircuitBreakerState returns the current circuit breaker state
func (a *Adapter) GetCircuitBreakerState() string {
	if a.circuitBreaker == nil {
		return "disabled"
	}
	state := a.circuitBreaker.State()
	return state.String()
}

// Close closes all resilience resources
func (a *Adapter) Close() error {
	if a.bulkheadPool != nil {
		return a.bulkheadPool.Close()
	}
	return nil
}
