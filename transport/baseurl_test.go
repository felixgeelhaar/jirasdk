package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{
			name:    "site base URL with absolute path",
			baseURL: "https://example.atlassian.net",
			path:    "/rest/api/3/myself",
			want:    "https://example.atlassian.net/rest/api/3/myself",
		},
		{
			name:    "site base URL with trailing slash",
			baseURL: "https://example.atlassian.net/",
			path:    "/rest/api/3/myself",
			want:    "https://example.atlassian.net/rest/api/3/myself",
		},
		{
			name:    "gateway base URL keeps its cloud ID prefix",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id",
			path:    "/rest/api/3/myself",
			want:    "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself",
		},
		{
			name:    "gateway base URL with trailing slash",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id/",
			path:    "/rest/api/3/myself",
			want:    "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself",
		},
		{
			name:    "server context path is preserved",
			baseURL: "https://jira.internal/jira",
			path:    "/rest/api/2/myself",
			want:    "https://jira.internal/jira/rest/api/2/myself",
		},
		{
			name:    "relative path",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id",
			path:    "rest/api/3/myself",
			want:    "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself",
		},
		{
			name:    "query string is carried over",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id",
			path:    "/rest/api/3/search?jql=project%3DTEST&maxResults=50",
			want:    "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/search?jql=project%3DTEST&maxResults=50",
		},
		{
			name:    "absolute URL overrides the base",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id",
			path:    "https://other.example.com/rest/api/3/myself",
			want:    "https://other.example.com/rest/api/3/myself",
		},
		{
			name:    "empty path yields the base",
			baseURL: "https://api.atlassian.com/ex/jira/cloud-id",
			path:    "",
			want:    "https://api.atlassian.com/ex/jira/cloud-id",
		},
		{
			name:    "base with no path",
			baseURL: "https://example.atlassian.net",
			path:    "",
			want:    "https://example.atlassian.net/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, err := url.Parse(tt.baseURL)
			require.NoError(t, err)

			got, err := resolveURL(base, tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestResolveURLErrors(t *testing.T) {
	_, err := resolveURL(nil, "/rest/api/3/myself")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")

	base, err := url.Parse("https://example.atlassian.net")
	require.NoError(t, err)

	_, err = resolveURL(base, "://not a url")
	require.Error(t, err)
}

// stubResolver records how often it was consulted.
type stubResolver struct {
	baseURL string
	err     error
	calls   atomic.Int32
}

func (r *stubResolver) ResolveBaseURL(_ context.Context) (*url.URL, error) {
	r.calls.Add(1)
	if r.err != nil {
		return nil, r.err
	}
	return url.Parse(r.baseURL)
}

func TestTransportUsesBaseURLResolver(t *testing.T) {
	resolver := &stubResolver{baseURL: "https://api.atlassian.com/ex/jira/cloud-id"}

	configured, err := url.Parse("https://example.atlassian.net")
	require.NoError(t, err)

	tr := New(http.DefaultClient, configured, WithBaseURLResolver(resolver))

	req, err := tr.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)

	// The resolver must win over the base URL passed to New.
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself", req.URL.String())
	assert.Equal(t, int32(1), resolver.calls.Load())
}

func TestTransportWithoutResolverUsesConfiguredBaseURL(t *testing.T) {
	configured, err := url.Parse("https://example.atlassian.net")
	require.NoError(t, err)

	tr := New(http.DefaultClient, configured)

	req, err := tr.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.atlassian.net/rest/api/3/myself", req.URL.String())
}

func TestTransportSurfacesResolverError(t *testing.T) {
	resolver := &stubResolver{err: fmt.Errorf("token grants access to no Atlassian sites")}

	tr := New(http.DefaultClient, nil, WithBaseURLResolver(resolver))

	_, err := tr.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Atlassian sites")
}

func TestTransportRequiresABaseURL(t *testing.T) {
	tr := New(http.DefaultClient, nil)

	_, err := tr.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}

func TestTransportConcurrentRequestBuilding(t *testing.T) {
	resolver := &stubResolver{baseURL: "https://api.atlassian.com/ex/jira/cloud-id"}
	tr := New(http.DefaultClient, nil, WithBaseURLResolver(resolver))

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := tr.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
			assert.NoError(t, err)
			assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id/rest/api/3/myself", req.URL.String())
		}()
	}
	wg.Wait()
}
