package transport

import (
	"context"
	"net/http"
	"time"
)

// loggingMiddleware adds structured logging to HTTP requests
func loggingMiddleware(logger Logger) Middleware {
	return func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			if logger == nil {
				return next(ctx, req)
			}

			start := time.Now()

			// Log request
			logger.Debug(ctx, "jira_request_started",
				String("method", req.Method),
				String("url", req.URL.String()),
				String("path", req.URL.Path),
			)

			// Execute request
			resp, err := next(ctx, req)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			if err != nil {
				logger.Error(ctx, "jira_request_failed",
					String("method", req.Method),
					String("path", req.URL.Path),
					Duration("duration", duration),
					Err(err),
				)
			} else {
				fields := []Field{
					String("method", req.Method),
					String("path", req.URL.Path),
					Int("status", resp.StatusCode),
					Duration("duration", duration),
				}

				// Add rate limit headers if present
				if rateLimit := resp.Header.Get("X-RateLimit-Limit"); rateLimit != "" {
					fields = append(fields, String("rate_limit", rateLimit))
				}
				if rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining"); rateLimitRemaining != "" {
					fields = append(fields, String("rate_limit_remaining", rateLimitRemaining))
				}

				if resp.StatusCode >= 500 {
					logger.Error(ctx, "jira_request_server_error", fields...)
				} else if resp.StatusCode >= 400 {
					logger.Warn(ctx, "jira_request_client_error", fields...)
				} else {
					logger.Info(ctx, "jira_request_completed", fields...)
				}
			}

			return resp, err
		}
	}
}
