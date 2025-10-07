package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthenticator implements auth.Authenticator for testing
type mockAuthenticator struct{}

func (m *mockAuthenticator) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer mock-token")
	return nil
}

func (m *mockAuthenticator) Type() string {
	return "mock"
}

func TestNew(t *testing.T) {
	baseURL, _ := url.Parse("https://example.atlassian.net")
	client := &http.Client{}

	tests := []struct {
		name string
		opts []TransportOption
		check func(*testing.T, *Transport)
	}{
		{
			name: "default configuration",
			opts: nil,
			check: func(t *testing.T, tr *Transport) {
				assert.Equal(t, 3, tr.maxRetries)
				assert.Equal(t, 5*time.Second, tr.rateLimitBuffer)
				assert.Equal(t, "jira-connect-go/1.0.0", tr.userAgent)
			},
		},
		{
			name: "with custom options",
			opts: []TransportOption{
				WithMaxRetries(5),
				WithRateLimitBuffer(10 * time.Second),
				WithUserAgent("custom-agent/1.0"),
				WithAuthenticator(&mockAuthenticator{}),
			},
			check: func(t *testing.T, tr *Transport) {
				assert.Equal(t, 5, tr.maxRetries)
				assert.Equal(t, 10*time.Second, tr.rateLimitBuffer)
				assert.Equal(t, "custom-agent/1.0", tr.userAgent)
				assert.NotNil(t, tr.authenticator)
			},
		},
		{
			name: "with middleware",
			opts: []TransportOption{
				WithMiddlewares(func(next RoundTripFunc) RoundTripFunc {
					return next
				}),
			},
			check: func(t *testing.T, tr *Transport) {
				assert.Len(t, tr.middlewares, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := New(client, baseURL, tt.opts...)
			require.NotNil(t, transport)
			assert.NotNil(t, transport.roundTripper)
			if tt.check != nil {
				tt.check(t, transport)
			}
		})
	}
}

func TestTransport_NewRequest(t *testing.T) {
	baseURL, _ := url.Parse("https://example.atlassian.net")
	client := &http.Client{}
	transport := New(client, baseURL)

	tests := []struct {
		name    string
		method  string
		path    string
		body    interface{}
		wantErr bool
		check   func(*testing.T, *http.Request)
	}{
		{
			name:    "GET request without body",
			method:  http.MethodGet,
			path:    "/rest/api/3/issue/PROJ-123",
			body:    nil,
			wantErr: false,
			check: func(t *testing.T, req *http.Request) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Contains(t, req.URL.String(), "example.atlassian.net")
				assert.Contains(t, req.URL.Path, "/rest/api/3/issue/PROJ-123")
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Empty(t, req.Header.Get("Content-Type"))
			},
		},
		{
			name:   "POST request with body",
			method: http.MethodPost,
			path:   "/rest/api/3/issue",
			body: map[string]interface{}{
				"fields": map[string]interface{}{
					"summary": "Test issue",
				},
			},
			wantErr: false,
			check: func(t *testing.T, req *http.Request) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := transport.NewRequest(context.Background(), tt.method, tt.path, tt.body)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)
				if tt.check != nil {
					tt.check(t, req)
				}
			}
		})
	}
}

func TestTransport_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication header
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := server.Client()
	transport := New(client, baseURL, WithAuthenticator(&mockAuthenticator{}))

	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := transport.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTransport_DoWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := server.Client()
	transport := New(client, baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	require.NoError(t, err)

	_, err = transport.Do(ctx, req)
	assert.Error(t, err)
}

func TestTransport_DecodeResponse(t *testing.T) {
	baseURL, _ := url.Parse("https://example.atlassian.net")
	client := &http.Client{}
	transport := New(client, baseURL)

	tests := []struct {
		name       string
		statusCode int
		body       string
		target     interface{}
		wantErr    bool
		check      func(*testing.T, interface{}, error)
	}{
		{
			name:       "successful decode",
			statusCode: http.StatusOK,
			body:       `{"key":"PROJ-123","summary":"Test issue"}`,
			target:     &map[string]string{},
			wantErr:    false,
			check: func(t *testing.T, target interface{}, err error) {
				data := target.(*map[string]string)
				assert.Equal(t, "PROJ-123", (*data)["key"])
				assert.Equal(t, "Test issue", (*data)["summary"])
			},
		},
		{
			name:       "nil target",
			statusCode: http.StatusOK,
			body:       `{"key":"value"}`,
			target:     nil,
			wantErr:    false,
		},
		{
			name:       "error response",
			statusCode: http.StatusNotFound,
			body:       `{"errorMessages":["Issue not found"]}`,
			target:     &map[string]string{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
			}

			err := transport.DecodeResponse(resp, tt.target)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, tt.target, err)
				}
			}
		})
	}
}

func TestWithAuthenticator(t *testing.T) {
	cfg := &Config{}
	auth := &mockAuthenticator{}

	opt := WithAuthenticator(auth)
	opt(cfg)

	assert.Equal(t, auth, cfg.authenticator)
}

func TestWithMaxRetries(t *testing.T) {
	cfg := &Config{}

	opt := WithMaxRetries(5)
	opt(cfg)

	assert.Equal(t, 5, cfg.maxRetries)
}

func TestWithRateLimitBuffer(t *testing.T) {
	cfg := &Config{}

	opt := WithRateLimitBuffer(10 * time.Second)
	opt(cfg)

	assert.Equal(t, 10*time.Second, cfg.rateLimitBuffer)
}

func TestWithUserAgent(t *testing.T) {
	cfg := &Config{}

	opt := WithUserAgent("custom-agent/1.0")
	opt(cfg)

	assert.Equal(t, "custom-agent/1.0", cfg.userAgent)
}

func TestWithMiddlewares(t *testing.T) {
	cfg := &Config{}

	middleware1 := func(next RoundTripFunc) RoundTripFunc { return next }
	middleware2 := func(next RoundTripFunc) RoundTripFunc { return next }

	opt := WithMiddlewares(middleware1, middleware2)
	opt(cfg)

	assert.Len(t, cfg.middlewares, 2)
}

func TestBuildMiddlewareChain(t *testing.T) {
	baseURL, _ := url.Parse("https://example.atlassian.net")
	client := &http.Client{}

	customMiddlewareCalled := false
	customMiddleware := func(next RoundTripFunc) RoundTripFunc {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			customMiddlewareCalled = true
			return next(ctx, req)
		}
	}

	transport := New(client, baseURL, WithMiddlewares(customMiddleware))

	assert.NotNil(t, transport.roundTripper)

	// Create a test server to verify middleware is called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	// Update transport to use test server
	serverURL, _ := url.Parse(server.URL)
	transport.baseURL = serverURL
	transport.client = server.Client()

	_, err := transport.Do(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, customMiddlewareCalled)
}
