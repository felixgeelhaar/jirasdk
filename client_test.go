package jirasdk

import (
	"net/http"
	"testing"
	"time"

	"github.com/felixgeelhaar/jirasdk/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration with API token",
			opts: []Option{
				WithBaseURL("https://example.atlassian.net"),
				WithAPIToken("user@example.com", "token123"),
			},
			wantErr: false,
		},
		{
			name: "valid configuration with PAT",
			opts: []Option{
				WithBaseURL("https://jira.company.com"),
				WithPAT("pat-token-123"),
			},
			wantErr: false,
		},
		{
			name: "valid configuration with basic auth",
			opts: []Option{
				WithBaseURL("https://jira.company.com"),
				WithBasicAuth("username", "password"),
			},
			wantErr: false,
		},
		{
			name:    "missing base URL",
			opts:    []Option{WithAPIToken("user@example.com", "token")},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name:    "missing authentication",
			opts:    []Option{WithBaseURL("https://example.atlassian.net")},
			wantErr: true,
			errMsg:  "authentication method is required",
		},
		{
			name: "invalid base URL",
			opts: []Option{
				WithBaseURL("not-a-url"),
				WithAPIToken("user@example.com", "token"),
			},
			wantErr: true,
		},
		{
			name: "base URL without scheme",
			opts: []Option{
				WithBaseURL("example.atlassian.net"),
				WithAPIToken("user@example.com", "token"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				assert.NotNil(t, client.BaseURL)
				assert.NotNil(t, client.HTTPClient)
				assert.NotNil(t, client.Authenticator)
				assert.NotNil(t, client.Transport)
				assert.NotNil(t, client.Issue)
				assert.NotNil(t, client.Project)
				assert.NotNil(t, client.User)
				assert.NotNil(t, client.Workflow)
			}
		})
	}
}

func TestWithBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://example.atlassian.net",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "not a url",
			wantErr: true,
		},
		{
			name:    "URL without scheme",
			url:     "example.atlassian.net",
			wantErr: true,
		},
		{
			name:    "FTP scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithBaseURL(tt.url)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.baseURL)
			}
		})
	}
}

func TestWithAPIToken(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		token   string
		wantErr bool
	}{
		{
			name:    "valid credentials",
			email:   "user@example.com",
			token:   "token123",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			token:   "token123",
			wantErr: true,
		},
		{
			name:    "empty token",
			email:   "user@example.com",
			token:   "",
			wantErr: true,
		},
		{
			name:    "both empty",
			email:   "",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithAPIToken(tt.email, tt.token)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.authenticator)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid timeout",
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0,
			wantErr: true,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithTimeout(tt.timeout)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.timeout, cfg.timeout)
			}
		})
	}
}

func TestWithMaxRetries(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
		wantErr    bool
	}{
		{
			name:       "valid retries",
			maxRetries: 3,
			wantErr:    false,
		},
		{
			name:       "zero retries",
			maxRetries: 0,
			wantErr:    false,
		},
		{
			name:       "negative retries",
			maxRetries: -1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithMaxRetries(tt.maxRetries)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.maxRetries, cfg.maxRetries)
			}
		})
	}
}

func TestWithUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		wantErr   bool
	}{
		{
			name:      "valid user agent",
			userAgent: "MyApp/1.0.0",
			wantErr:   false,
		},
		{
			name:      "empty user agent",
			userAgent: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithUserAgent(tt.userAgent)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userAgent, cfg.userAgent)
			}
		})
	}
}

func TestWithPAT(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid PAT",
			token:   "pat-token-123",
			wantErr: false,
		},
		{
			name:    "empty PAT",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithPAT(tt.token)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.authenticator)
			}
		})
	}
}

func TestWithBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "valid credentials",
			username: "admin",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty password",
			username: "admin",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithBasicAuth(tt.username, tt.password)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg.authenticator)
			}
		})
	}
}

func TestWithRateLimitBuffer(t *testing.T) {
	tests := []struct {
		name    string
		buffer  time.Duration
		wantErr bool
	}{
		{
			name:    "valid buffer",
			buffer:  10 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero buffer",
			buffer:  0,
			wantErr: false,
		},
		{
			name:    "negative buffer",
			buffer:  -5 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithRateLimitBuffer(tt.buffer)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.buffer, cfg.rateLimitBuffer)
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		client  *http.Client
		wantErr bool
	}{
		{
			name:    "valid HTTP client",
			client:  &http.Client{Timeout: 60 * time.Second},
			wantErr: false,
		},
		{
			name:    "nil HTTP client",
			client:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithHTTPClient(tt.client)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.client, cfg.httpClient)
			}
		})
	}
}

func TestWithMiddleware(t *testing.T) {
	testMiddleware := func(next transport.RoundTripFunc) transport.RoundTripFunc {
		return next
	}

	tests := []struct {
		name       string
		middleware transport.Middleware
		wantErr    bool
	}{
		{
			name:       "valid middleware",
			middleware: testMiddleware,
			wantErr:    false,
		},
		{
			name:       "nil middleware",
			middleware: nil,
			wantErr:    false, // WithMiddleware doesn't validate nil, just appends
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			opt := WithMiddleware(tt.middleware)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cfg.middlewares, 1)
			}
		})
	}
}
