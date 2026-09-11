package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const twoSitesJSON = `[
  {"id":"cloud-id-one","url":"https://one.atlassian.net","name":"One","scopes":["read:jira-work"]},
  {"id":"cloud-id-two","url":"https://two.atlassian.net","name":"Two","scopes":["read:jira-work"]}
]`

const oneSiteJSON = `[{"id":"cloud-id-one","url":"https://one.atlassian.net","name":"One","scopes":["read:jira-work"]}]`

// resourcesServer serves a canned accessible-resources payload and counts hits.
type resourcesServer struct {
	*httptest.Server
	calls atomic.Int32
}

func newResourcesServer(t *testing.T, status int, payload string) *resourcesServer {
	t.Helper()

	s := &resourcesServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.calls.Add(1)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(s.Close)

	return s
}

// tokenAuth is a minimal Authenticator for exercising the resolver.
type tokenAuth struct{ token string }

func (a *tokenAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.token)
	return nil
}

func (a *tokenAuth) Type() string { return "test" }

// newResolverAgainst points a resolver at a test server instead of Atlassian.
func newResolverAgainst(t *testing.T, server *resourcesServer, siteURL string) *CloudIDResolver {
	t.Helper()

	r, err := NewCloudIDResolver(&CloudIDResolverConfig{
		Authenticator: &tokenAuth{token: "test-token"},
		SiteURL:       siteURL,
		HTTPClient:    server.Client(),
	})
	require.NoError(t, err)

	// Redirect the lookup at the test server.
	r.httpClient = &http.Client{Transport: rewriteHost{to: server.URL, base: server.Client().Transport}}
	return r
}

// rewriteHost sends every request to the test server regardless of its URL.
type rewriteHost struct {
	to   string
	base http.RoundTripper
}

func (t rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := http.NewRequestWithContext(req.Context(), req.Method, t.to, req.Body)
	if err != nil {
		return nil, err
	}
	target.Header = req.Header

	rt := t.base
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(target)
}

func TestJiraCloudBaseURL(t *testing.T) {
	u, err := JiraCloudBaseURL("cloud-id-one")
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id-one", u.String())

	_, err = JiraCloudBaseURL("")
	require.Error(t, err)
}

func TestNewCloudIDResolverValidation(t *testing.T) {
	_, err := NewCloudIDResolver(nil)
	require.Error(t, err)

	_, err = NewCloudIDResolver(&CloudIDResolverConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authenticator is required")

	// An explicit cloud ID removes the need for an authenticator.
	r, err := NewCloudIDResolver(&CloudIDResolverConfig{CloudID: "cloud-id-one"})
	require.NoError(t, err)
	assert.NotNil(t, r)
}

func TestResolveBaseURLWithExplicitCloudID(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, oneSiteJSON)

	r, err := NewCloudIDResolver(&CloudIDResolverConfig{
		Authenticator: &tokenAuth{token: "test-token"},
		CloudID:       "pinned-cloud-id",
		HTTPClient:    server.Client(),
	})
	require.NoError(t, err)

	base, err := r.ResolveBaseURL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/pinned-cloud-id", base.String())
	assert.Equal(t, int32(0), server.calls.Load(), "an explicit cloud ID must skip the lookup")
}

func TestResolveBaseURLFromSingleSite(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, oneSiteJSON)
	r := newResolverAgainst(t, server, "")

	base, err := r.ResolveBaseURL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id-one", base.String())

	cloudID, err := r.CloudID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "cloud-id-one", cloudID)
}

func TestResolveBaseURLSelectsBySiteURL(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, twoSitesJSON)
	r := newResolverAgainst(t, server, "https://two.atlassian.net")

	base, err := r.ResolveBaseURL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id-two", base.String())
}

func TestResolveBaseURLSiteURLIgnoresTrailingSlashAndCase(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, twoSitesJSON)
	r := newResolverAgainst(t, server, "https://TWO.atlassian.net/")

	base, err := r.ResolveBaseURL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/cloud-id-two", base.String())
}

func TestResolveBaseURLAmbiguousSites(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, twoSitesJSON)
	r := newResolverAgainst(t, server, "")

	_, err := r.ResolveBaseURL(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2 sites")
	// The error should name the candidates so the fix is obvious.
	assert.Contains(t, err.Error(), "https://one.atlassian.net")
	assert.Contains(t, err.Error(), "https://two.atlassian.net")
}

func TestResolveBaseURLUnknownSite(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, twoSitesJSON)
	r := newResolverAgainst(t, server, "https://three.atlassian.net")

	_, err := r.ResolveBaseURL(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no access to https://three.atlassian.net")
}

func TestResolveBaseURLNoSites(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, `[]`)
	r := newResolverAgainst(t, server, "")

	_, err := r.ResolveBaseURL(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Atlassian sites")
}

func TestResolveBaseURLPropagatesHTTPError(t *testing.T) {
	server := newResourcesServer(t, http.StatusUnauthorized, `{"message":"Unauthorized"}`)
	r := newResolverAgainst(t, server, "")

	_, err := r.ResolveBaseURL(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestResolveBaseURLCachesResult(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, oneSiteJSON)
	r := newResolverAgainst(t, server, "")

	for i := 0; i < 5; i++ {
		_, err := r.ResolveBaseURL(context.Background())
		require.NoError(t, err)
	}

	assert.Equal(t, int32(1), server.calls.Load(), "the lookup must happen once per client")
}

func TestResolveBaseURLConcurrent(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, oneSiteJSON)
	r := newResolverAgainst(t, server, "")

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.ResolveBaseURL(context.Background())
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), server.calls.Load())
}

func TestFetchAccessibleResources(t *testing.T) {
	server := newResourcesServer(t, http.StatusOK, twoSitesJSON)

	// The canonical Atlassian URL is not reachable from a test, so route the
	// request to the stub server instead.
	client := &http.Client{Transport: rewriteHost{to: server.URL}}

	resources, err := FetchAccessibleResources(context.Background(), client, "test-token")
	require.NoError(t, err)

	require.Len(t, resources, 2)
	assert.Equal(t, "cloud-id-one", resources[0].ID)
	assert.Equal(t, "https://one.atlassian.net", resources[0].URL)
	assert.Equal(t, "One", resources[0].Name)
	assert.Equal(t, []string{"read:jira-work"}, resources[0].Scopes)
}

func TestOAuth2AuthenticatorAccessibleResourcesRequiresToken(t *testing.T) {
	a := NewOAuth2Authenticator(&OAuth2Config{ClientID: "id"})

	_, err := a.AccessibleResources(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no OAuth 2.0 token available")

	a.SetToken(&oauth2.Token{AccessToken: "at"})
	assert.NotNil(t, a.GetToken())
}
