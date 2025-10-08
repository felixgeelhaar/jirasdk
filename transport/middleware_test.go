package transport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name        string
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{
			name:        "first attempt",
			attempt:     0,
			minExpected: 75 * time.Millisecond,  // 100ms - 25% jitter
			maxExpected: 125 * time.Millisecond, // 100ms + 25% jitter
		},
		{
			name:        "second attempt",
			attempt:     1,
			minExpected: 150 * time.Millisecond, // 200ms - 25% jitter
			maxExpected: 250 * time.Millisecond, // 200ms + 25% jitter
		},
		{
			name:        "third attempt",
			attempt:     2,
			minExpected: 300 * time.Millisecond, // 400ms - 25% jitter
			maxExpected: 500 * time.Millisecond, // 400ms + 25% jitter
		},
		{
			name:        "max backoff",
			attempt:     10,
			minExpected: 22500 * time.Millisecond, // 30s - 25% jitter
			maxExpected: 37500 * time.Millisecond, // 30s + 25% jitter (capped)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt)
			assert.GreaterOrEqual(t, backoff, tt.minExpected)
			assert.LessOrEqual(t, backoff, tt.maxExpected)
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		expected   time.Duration
	}{
		{
			name:       "empty header",
			retryAfter: "",
			expected:   time.Second,
		},
		{
			name:       "seconds format",
			retryAfter: "120",
			expected:   120 * time.Second,
		},
		{
			name:       "zero seconds",
			retryAfter: "0",
			expected:   0,
		},
		{
			name:       "invalid format",
			retryAfter: "invalid",
			expected:   time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.retryAfter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "429 Too Many Requests",
			statusCode: 429,
			expected:   true,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			expected:   true,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: 502,
			expected:   true,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: 503,
			expected:   true,
		},
		{
			name:       "504 Gateway Timeout",
			statusCode: 504,
			expected:   true,
		},
		{
			name:       "200 OK",
			statusCode: 200,
			expected:   false,
		},
		{
			name:       "400 Bad Request",
			statusCode: 400,
			expected:   false,
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			expected:   false,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}
