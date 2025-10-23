package transport

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/felixgeelhaar/jirasdk/auth"
)

// authMiddleware adds authentication to requests.
func authMiddleware(authenticator auth.Authenticator) Middleware {
	return func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			if err := authenticator.Authenticate(req); err != nil {
				return nil, fmt.Errorf("authentication failed: %w", err)
			}
			return next(ctx, req)
		}
	}
}

// userAgentMiddleware sets the User-Agent header.
func userAgentMiddleware(userAgent string) Middleware {
	return func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", userAgent)
			return next(ctx, req)
		}
	}
}

// retryMiddleware implements exponential backoff retry logic.
func retryMiddleware(maxRetries int) Middleware {
	return func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			var lastErr error
			var resp *http.Response

			for attempt := 0; attempt <= maxRetries; attempt++ {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}

				// Execute request
				resp, lastErr = next(ctx, req)

				// If no error and status is not retry-able, return immediately
				if lastErr == nil && !isRetryableStatus(resp.StatusCode) {
					return resp, nil
				}

				// If this was the last attempt, break
				if attempt == maxRetries {
					break
				}

				// Calculate backoff duration with jitter
				backoff := calculateBackoff(attempt)

				// Wait for backoff duration or context cancellation
				select {
				case <-time.After(backoff):
					// Continue to next attempt
				case <-ctx.Done():
					return nil, ctx.Err()
				}

				// Close the response body if present to avoid resource leaks
				if resp != nil && resp.Body != nil {
					_ = resp.Body.Close() // Explicit ignore in cleanup path
				}
			}

			// All retries exhausted
			if lastErr != nil {
				return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
			}

			return resp, nil
		}
	}
}

// rateLimitMiddleware handles rate limiting responses.
func rateLimitMiddleware(buffer time.Duration) Middleware {
	return func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return resp, err
			}

			// Check for rate limiting (429 Too Many Requests)
			if resp.StatusCode == http.StatusTooManyRequests {
				// Parse Retry-After header
				retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))

				// Add buffer to avoid hitting the limit again immediately
				waitDuration := retryAfter + buffer

				// Close the response body
				if resp.Body != nil {
					_ = resp.Body.Close() // Explicit ignore before retry
				}

				// Wait for the specified duration or context cancellation
				select {
				case <-time.After(waitDuration):
					// Retry the request
					return next(ctx, req)
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			// Check for rate limit headers (X-RateLimit-*)
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
				// TODO: Log or emit metric about approaching rate limit
				// In a real implementation, this could trigger adaptive rate limiting
			}

			return resp, nil
		}
	}
}

// isRetryableStatus returns true if the HTTP status code is retry-able.
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// calculateBackoff calculates exponential backoff duration with jitter.
//
// Formula: min(baseDelay * 2^attempt, maxDelay) + jitter
func calculateBackoff(attempt int) time.Duration {
	const (
		baseDelay = 100 * time.Millisecond
		maxDelay  = 30 * time.Second
	)

	// Calculate exponential backoff
	delay := float64(baseDelay) * math.Pow(2, float64(attempt))

	// Cap at max delay
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	// Add jitter (±25%)
	jitter := delay * 0.25 * (rand.Float64()*2 - 1) // #nosec G404 -- Weak random OK for jitter, doesn't need crypto/rand
	finalDelay := time.Duration(delay + jitter)

	return finalDelay
}

// parseRetryAfter parses the Retry-After header value.
//
// The Retry-After header can be either:
// - Number of seconds (e.g., "120")
// - HTTP date (e.g., "Fri, 31 Dec 2025 23:59:59 GMT")
func parseRetryAfter(retryAfter string) time.Duration {
	if retryAfter == "" {
		// Default to 1 second if no Retry-After header
		return time.Second
	}

	// Try to parse as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try to parse as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	// Default to 1 second if parsing fails
	return time.Second
}
