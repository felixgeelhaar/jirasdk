package auth

import (
	"net/http"
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
			want:   "GET&/rest/api/3/search&fields=assignee%2Csummary",
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
			want:   "GET&/s&a=1%2C9&b=2",
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
