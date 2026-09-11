package auth

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		rawURL      string
		contextPath string
		want        string
	}{
		{
			name:   "atlassian documentation example",
			method: "GET",
			rawURL: "https://example.atlassian.net/path/to/service?zee_last=param&first=param",
			want:   "GET&/path/to/service&first=param&zee_last=param",
		},
		{
			name:   "no query string",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/myself",
			want:   "GET&/rest/api/3/myself&",
		},
		{
			name:   "method is upper-cased",
			method: "post",
			rawURL: "https://example.atlassian.net/rest/api/3/issue",
			want:   "POST&/rest/api/3/issue&",
		},
		{
			name:   "empty path becomes slash",
			method: "GET",
			rawURL: "https://example.atlassian.net",
			want:   "GET&/&",
		},
		{
			name:   "trailing slash is stripped",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/myself/",
			want:   "GET&/rest/api/3/myself&",
		},
		{
			name:   "root path keeps its slash",
			method: "GET",
			rawURL: "https://example.atlassian.net/",
			want:   "GET&/&",
		},
		{
			name:   "jwt parameter is excluded",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/myself?jwt=abc.def.ghi&expand=groups",
			want:   "GET&/rest/api/3/myself&expand=groups",
		},
		{
			name:   "repeated parameters are comma joined and sorted",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/search?fields=summary&fields=assignee",
			want:   "GET&/rest/api/3/search&fields=assignee,summary",
		},
		{
			name:   "spaces encode as %20 not plus",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/search?jql=project+%3D+TEST",
			want:   "GET&/rest/api/3/search&jql=project%20%3D%20TEST",
		},
		{
			name:   "tilde is left unencoded",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/user?name=~admin",
			want:   "GET&/rest/api/3/user&name=~admin",
		},
		{
			name:   "ampersand in path is encoded",
			method: "GET",
			rawURL: "https://example.atlassian.net/rest/api/3/project/A%26B",
			want:   "GET&/rest/api/3/project/A%26B&",
		},
		{
			name:        "context path is stripped",
			method:      "GET",
			rawURL:      "https://jira.internal/jira/rest/api/2/myself",
			contextPath: "/jira",
			want:        "GET&/rest/api/2/myself&",
		},
		{
			name:        "context path stripped down to root",
			method:      "GET",
			rawURL:      "https://jira.internal/jira",
			contextPath: "/jira",
			want:        "GET&/&",
		},
		{
			name:   "parameters sort by name then value",
			method: "GET",
			rawURL: "https://example.atlassian.net/s?b=2&a=9&a=1",
			want:   "GET&/s&a=1,9&b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.rawURL, nil)
			require.NoError(t, err)

			got, err := CanonicalRequest(req, tt.contextPath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQueryStringHash(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/path/to/service?zee_last=param&first=param", nil)
	require.NoError(t, err)

	qsh, err := QueryStringHash(req, "")
	require.NoError(t, err)

	// SHA-256 hex digest of "GET&/path/to/service&first=param&zee_last=param".
	assert.Equal(t, "b7a226cb55a0de7d42511286707fc07941e6bae1a3692fbac9eabe7942b3e90f", qsh)
	assert.Len(t, qsh, 64)
}

func TestQueryStringHashIsStableAcrossParameterOrder(t *testing.T) {
	a, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/s?b=2&a=1", nil)
	require.NoError(t, err)
	b, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/s?a=1&b=2", nil)
	require.NoError(t, err)

	qshA, err := QueryStringHash(a, "")
	require.NoError(t, err)
	qshB, err := QueryStringHash(b, "")
	require.NoError(t, err)

	assert.Equal(t, qshA, qshB)
}

// TestCanonicalRequestAtlassianPublishedVectors pins the implementation to the
// worked examples and SHA-256 test vectors Atlassian publishes. These are the
// only cases verified against Atlassian's own output rather than our reading of
// the prose, so they are the ones that catch a drifting interpretation.
//
// https://developer.atlassian.com/cloud/bitbucket/query-string-hash/
func TestCanonicalRequestAtlassianPublishedVectors(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		rawURL    string
		canonical string
		qsh       string
	}{
		{
			name:      "GET with one parameter",
			method:    http.MethodGet,
			rawURL:    "https://example.atlassian.net/test?param=value",
			canonical: "GET&/test&param=value",
			qsh:       "be16910858a41fd19ea5c1b4e9decca9a784d1024cb00b2158defe2f29dc86dd",
		},
		{
			name:      "POST with no query string",
			method:    http.MethodPost,
			rawURL:    "https://example.atlassian.net/rest/api/2/issue",
			canonical: "POST&/rest/api/2/issue&",
			qsh:       "43dd1779e33c34fae00c308d62e5dd153a32147d1bcb5d40b3936457fda0ece4",
		},
		{
			name:      "root path",
			method:    http.MethodGet,
			rawURL:    "https://example.atlassian.net/?param=foo",
			canonical: "GET&/&param=foo",
		},
		{
			name:      "POST to a path",
			method:    http.MethodPost,
			rawURL:    "https://example.atlassian.net/user",
			canonical: "POST&/user&",
		},
		{
			// The separator is a literal comma, not %2C. Joining with %2C
			// yields a hash Jira rejects on any request with a repeated
			// parameter, and "fields" and "expand" are routinely repeated.
			name:      "repeated parameters join with a literal comma",
			method:    http.MethodGet,
			rawURL:    "https://example.atlassian.net/rest/api/2/issue?ids=-1&ids=1&ids=10",
			canonical: "GET&/rest/api/2/issue&ids=-1,1,10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.rawURL, nil)
			require.NoError(t, err)

			canonical, err := CanonicalRequest(req, "")
			require.NoError(t, err)
			assert.Equal(t, tt.canonical, canonical)

			if tt.qsh != "" {
				qsh, err := QueryStringHash(req, "")
				require.NoError(t, err)
				assert.Equal(t, tt.qsh, qsh)
			}
		})
	}
}

// TestCanonicalRequestSortsBeforeEncoding guards the sort order against the
// reference implementation, which sorts raw names and values and encodes
// afterwards.
func TestCanonicalRequestSortsBeforeEncoding(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			// Raw: ":" (0x3A) sorts after "0" (0x30). Encoded: "%3A" sorts
			// before "0", because "%" is 0x25. Sorting raw is correct.
			name:   "names needing escapes sort by their raw bytes",
			rawURL: "https://example.atlassian.net/s?a%3A=x&a0=y",
			want:   "GET&/s&a0=y&a%3A=x",
		},
		{
			name:   "values needing escapes sort by their raw bytes",
			rawURL: "https://example.atlassian.net/s?f=a%3Ab&f=a0",
			want:   "GET&/s&f=a0,a%3Ab",
		},
		{
			// A comma inside a value must stay encoded, so it cannot be
			// mistaken for the separator between repeated values.
			name:   "a comma inside a value stays encoded",
			rawURL: "https://example.atlassian.net/s?f=a%2Cb&f=c",
			want:   "GET&/s&f=a%2Cb,c",
		},
		{
			name:   "percent encoding uses upper-case hex",
			rawURL: "https://example.atlassian.net/s?f=%2a",
			want:   "GET&/s&f=%2A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.rawURL, nil)
			require.NoError(t, err)

			got, err := CanonicalRequest(req, "")
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCanonicalRequestIgnoresJSONBody records that a request body never
// contributes to the hash. Atlassian folds a body in only for form-encoded
// POST/PUT requests with an empty query string, and this SDK always sends JSON.
func TestCanonicalRequestIgnoresJSONBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.atlassian.net/rest/api/3/issue",
		strings.NewReader(`{"fields":{"summary":"test"}}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	got, err := CanonicalRequest(req, "")
	require.NoError(t, err)
	assert.Equal(t, "POST&/rest/api/3/issue&", got)
}
