//go:build live

// Package jirasdk live smoke tests.
//
// These tests talk to a real Jira instance and, for the Cloud flows, to
// Atlassian's real authorization servers. They are the only coverage that
// proves a signature or token this SDK produces is actually accepted — every
// other test in the repo verifies the SDK against a stub we wrote, which cannot
// catch a disagreement with Atlassian.
//
// They are excluded from the normal suite by the "live" build tag. Run them
// explicitly:
//
//	go test -tags live -run TestLive -v .
//
// Each test skips with a note naming the variables it needs, so running with a
// partial set exercises only what you have credentials for. Use -v: the skip
// reasons and the per-probe findings are logged, not asserted.
//
// Every request these tests make is read-only. Nothing here creates, updates,
// transitions or deletes anything, and nothing writes to the instance's
// configuration. They are safe to point at production, though a test instance
// is obviously preferable.
package jirasdk

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/jirasdk/auth"
	"github.com/felixgeelhaar/jirasdk/core/project"
	"github.com/felixgeelhaar/jirasdk/core/search"
	"github.com/felixgeelhaar/jirasdk/core/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// Per-deployment base URL overrides, so Cloud and Server flows can be exercised
// in one run. Both fall back to JIRA_BASE_URL.
const (
	envLiveCloudBaseURL  = "JIRA_LIVE_CLOUD_BASE_URL"
	envLiveServerBaseURL = "JIRA_LIVE_SERVER_BASE_URL"
)

const liveTimeout = 30 * time.Second

// liveContext bounds every live call, so a hanging instance fails the test
// rather than the whole run.
func liveContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), liveTimeout)
	t.Cleanup(cancel)

	return ctx
}

// requireEnv skips the test unless every named variable is set, naming the ones
// that are missing.
func requireEnv(t *testing.T, names ...string) map[string]string {
	t.Helper()

	values := make(map[string]string, len(names))
	var missing []string

	for _, name := range names {
		value := os.Getenv(name)
		if value == "" {
			missing = append(missing, name)
			continue
		}
		values[name] = value
	}

	if len(missing) > 0 {
		t.Skipf("not configured; set %s", strings.Join(missing, ", "))
	}

	return values
}

// baseURLFor resolves the base URL for a deployment, preferring its specific
// override. It skips rather than fails, since an unset base URL means this
// deployment simply was not configured.
func baseURLFor(t *testing.T, specific string) string {
	t.Helper()

	if value := os.Getenv(specific); value != "" {
		return value
	}
	if value := os.Getenv(EnvBaseURL); value != "" {
		return value
	}

	t.Skipf("not configured; set %s or %s", specific, EnvBaseURL)
	return ""
}

// redact shortens a credential so a failure message can identify which one was
// used without disclosing it. Test output routinely lands in CI logs.
func redact(secret string) string {
	if len(secret) <= 8 {
		return "(set, short)"
	}
	return fmt.Sprintf("%s…%s (%d chars)", secret[:4], secret[len(secret)-2:], len(secret))
}

// --- Jira Cloud: API token -------------------------------------------------

func TestLiveAPIToken(t *testing.T) {
	env := requireEnv(t, EnvEmail, EnvAPIToken)
	baseURL := baseURLFor(t, envLiveCloudBaseURL)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithAPIToken(env[EnvEmail], env[EnvAPIToken]),
	)
	require.NoError(t, err)

	user, err := client.Myself.Get(liveContext(t))
	require.NoError(t, err, "API token rejected (token %s)", redact(env[EnvAPIToken]))

	assert.NotEmpty(t, user.AccountID)
	t.Logf("authenticated as %s (%s)", user.DisplayName, user.AccountID)
}

// --- Jira Cloud: OAuth 2.0 (3LO) -------------------------------------------

// TestLiveOAuth2ThreeLegged is the test that matters most for 3LO, because it
// is the only way to confirm the cloud ID routing is right: a 3LO token is
// rejected at a site URL, so a wrong base URL fails with 401 rather than
// anything more descriptive.
func TestLiveOAuth2ThreeLegged(t *testing.T) {
	env := requireEnv(t, EnvOAuthAccessToken)

	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:     os.Getenv(EnvOAuthClientID),
		ClientSecret: os.Getenv(EnvOAuthClientSecret),
	})
	oauth.SetToken(&oauth2.Token{
		AccessToken:  env[EnvOAuthAccessToken],
		RefreshToken: os.Getenv(EnvOAuthRefreshToken),
		TokenType:    "Bearer",
	})

	ctx := liveContext(t)

	// Step 1: the accessible-resources lookup the resolver depends on.
	resources, err := oauth.AccessibleResources(ctx)
	require.NoError(t, err, "accessible-resources rejected the token (%s)", redact(env[EnvOAuthAccessToken]))
	require.NotEmpty(t, resources, "token grants access to no sites; check its scopes")

	for _, r := range resources {
		t.Logf("accessible site: %s %s (cloud ID %s, %d scopes)", r.Name, r.URL, r.ID, len(r.Scopes))
	}

	// Step 2: a request through the resolved gateway URL.
	opts := []Option{WithOAuth2(oauth)}
	if len(resources) > 1 {
		// Ambiguous otherwise, by design.
		opts = append(opts, WithCloudID(resources[0].ID))
		t.Logf("several sites accessible; pinning %s", resources[0].URL)
	}

	client, err := NewClient(opts...)
	require.NoError(t, err)

	req, err := client.Transport.NewRequest(ctx, http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)

	// Confirm the host and the cloud ID prefix survived, rather than inferring
	// it from the call succeeding.
	assert.Equal(t, "api.atlassian.com", req.URL.Host, "3LO requests must go through the API gateway")
	assert.Contains(t, req.URL.Path, "/ex/jira/"+resources[0].ID+"/rest/api/3/myself")
	t.Logf("resolved request URL: %s", req.URL)

	user, err := client.Myself.Get(ctx)
	require.NoError(t, err, "gateway request rejected")
	t.Logf("authenticated as %s (%s)", user.DisplayName, user.AccountID)

	// Step 3: the same token against the site URL, which should fail. This is
	// logged rather than asserted — it documents why routing exists, and a
	// future Atlassian change making it work should not fail the build.
	t.Run("site URL rejects a 3LO token", func(t *testing.T) {
		siteClient, err := NewClient(
			WithBaseURL(resources[0].URL),
			WithOAuth2(oauth),
			WithoutCloudIDResolution(),
		)
		require.NoError(t, err)

		if _, err := siteClient.Myself.Get(liveContext(t)); err != nil {
			t.Logf("as expected, the site URL rejected the token: %v", err)
		} else {
			t.Logf("NOTE: the site URL accepted the token; Atlassian may have changed this")
		}
	})

	// Step 4: refresh, when a refresh token is available. This is what proves
	// offline_access was actually granted.
	if os.Getenv(EnvOAuthRefreshToken) == "" {
		t.Log("no refresh token set; skipping the refresh probe")
		return
	}

	t.Run("refresh", func(t *testing.T) {
		requireEnv(t, EnvOAuthClientID, EnvOAuthClientSecret)

		before := oauth.GetToken().AccessToken
		refreshed, err := oauth.RefreshToken(liveContext(t))
		require.NoError(t, err, "refresh failed; was offline_access among the granted scopes?")

		assert.NotEmpty(t, refreshed.AccessToken)
		assert.NotEqual(t, before, refreshed.AccessToken, "refresh returned the same access token")
		t.Logf("refreshed; new token expires %s", refreshed.Expiry.Format(time.RFC3339))
	})
}

// TestLiveOAuth2AuthURL checks that Atlassian accepts the authorization URL we
// build. A missing audience or prompt parameter is rejected at this step, which
// is the failure the stub tests cannot see.
func TestLiveOAuth2AuthURL(t *testing.T) {
	// The redirect URI must be one registered for the app: Atlassian rejects a
	// mismatch, which would look identical to a malformed URL.
	env := requireEnv(t, EnvOAuthClientID, EnvOAuthRedirectURL)

	scopes := strings.Fields(os.Getenv(EnvOAuthScopes))
	if len(scopes) == 0 {
		scopes = []string{"read:jira-work"}
	}

	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:    env[EnvOAuthClientID],
		RedirectURL: env[EnvOAuthRedirectURL],
		Scopes:      scopes,
	})

	authURL := oauth.GetAuthURL("live-smoke-state")

	// Do not follow the redirect to the login page; we only care that the
	// request itself is not rejected as malformed.
	client := &http.Client{
		Timeout:       liveTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	req, err := http.NewRequestWithContext(liveContext(t), http.MethodGet, authURL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	t.Logf("authorize endpoint returned %s", resp.Status)

	// A rejected authorization request comes back as 400, or redirects to an
	// error page. Anything else means the parameters were accepted.
	assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode,
		"Atlassian rejected the authorization URL as malformed: %s", authURL)

	if location := resp.Header.Get("Location"); strings.Contains(location, "error") {
		t.Errorf("authorize endpoint redirected to an error: %s\n"+
			"if the parameters look right, check that %s matches a redirect URI "+
			"registered for this app and that %s names scopes it is granted",
			location, EnvOAuthRedirectURL, EnvOAuthScopes)
	}
}

// --- Jira Cloud: Connect JWT (2LO) -----------------------------------------

func TestLiveConnectJWT(t *testing.T) {
	env := requireEnv(t, EnvConnectAppKey, EnvConnectSharedSecret)
	baseURL := baseURLFor(t, envLiveCloudBaseURL)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithConnectJWT(&auth.ConnectJWTConfig{
			AppKey:       env[EnvConnectAppKey],
			SharedSecret: env[EnvConnectSharedSecret],
			ContextPath:  os.Getenv("JIRA_CONTEXT_PATH"),
		}),
	)
	require.NoError(t, err)

	ctx := liveContext(t)

	// A request with no query string: the simplest qsh there is.
	t.Run("no query string", func(t *testing.T) {
		info, err := client.ServerInfo.Get(liveContext(t))
		require.NoError(t, err, "Connect JWT rejected (secret %s)", redact(env[EnvConnectSharedSecret]))
		t.Logf("Jira %s (%s)", info.Version, info.DeploymentType)
	})

	// A request with a single query parameter.
	t.Run("single query parameter", func(t *testing.T) {
		projects, err := client.Project.List(liveContext(t), &project.ListOptions{
			Expand: []string{"description"},
		})
		require.NoError(t, err, "qsh rejected for a single query parameter")
		t.Logf("%d project(s) visible to the app", len(projects))
	})

	// A request with REPEATED query parameters. This is the case that was
	// broken: joining the values with %2C instead of a literal comma produces
	// a hash Jira rejects, and this SDK emits repeated "expand" parameters on
	// many endpoints. No stub test can catch a wrong-but-self-consistent hash.
	t.Run("repeated query parameters", func(t *testing.T) {
		projects, err := client.Project.List(liveContext(t), &project.ListOptions{
			Expand: []string{"description", "lead", "issueTypes"},
		})
		require.NoError(t, err,
			"qsh rejected for repeated query parameters; check that canonicalQuery "+
				"joins repeated values with a literal comma, not %%2C")
		t.Logf("%d project(s) with three expand values", len(projects))
	})

	// A GET whose parameter value needs percent-encoding, exercising the
	// encoding and sort rules together. A space must encode as %20, not "+".
	t.Run("value needing percent encoding", func(t *testing.T) {
		users, err := client.User.Search(liveContext(t), &user.SearchOptions{
			Query:      "a b",
			MaxResults: 1,
		})
		require.NoError(t, err,
			"qsh rejected for a value needing encoding; check that spaces encode "+
				"as %%20 rather than \"+\"")
		t.Logf("user search returned %d result(s)", len(users))
	})

	// A POST carrying a JSON body and no query string. Atlassian folds a body
	// into the hash only for form-encoded requests, so this confirms the JSON
	// body is correctly left out: including it would produce a rejected hash.
	t.Run("POST with a JSON body", func(t *testing.T) {
		results, err := client.Search.Search(liveContext(t), &search.SearchOptions{
			JQL:        "order by created DESC",
			MaxResults: 1,
		})
		require.NoError(t, err, "qsh rejected for a POST with a JSON body")
		t.Logf("search returned %d issue(s)", len(results.Issues))
	})

	_ = ctx
}

// --- Jira Cloud: JWT bearer grant (2LO impersonation) ----------------------

func TestLiveJWTBearer(t *testing.T) {
	env := requireEnv(t, EnvConnectOAuthClientID, EnvConnectSharedSecret, EnvConnectAccountID)
	baseURL := baseURLFor(t, envLiveCloudBaseURL)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithJWTBearer(&auth.JWTBearerConfig{
			OAuthClientID: env[EnvConnectOAuthClientID],
			SharedSecret:  env[EnvConnectSharedSecret],
			AccountID:     env[EnvConnectAccountID],
			SiteURL:       baseURL,
		}),
	)
	require.NoError(t, err)

	user, err := client.Myself.Get(liveContext(t))
	require.NoError(t, err,
		"JWT bearer assertion rejected; confirm %s is the oauthClientId from the "+
			"install payload and not the app key", EnvConnectOAuthClientID)

	// The token should act as the requested user, not as the app.
	assert.Equal(t, env[EnvConnectAccountID], user.AccountID,
		"impersonation returned a different account than requested")
	t.Logf("acting as %s (%s)", user.DisplayName, user.AccountID)
}

// --- Jira Server / Data Center: PAT ----------------------------------------

func TestLivePAT(t *testing.T) {
	env := requireEnv(t, EnvPAT)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithPAT(env[EnvPAT]),
	)
	require.NoError(t, err)

	info, err := client.ServerInfo.Get(liveContext(t))
	require.NoError(t, err, "PAT rejected (%s)", redact(env[EnvPAT]))
	t.Logf("Jira %s (%s)", info.Version, info.DeploymentType)
}

// --- Jira Server / Data Center: OAuth 1.0a ---------------------------------

func TestLiveOAuth1TwoLegged(t *testing.T) {
	env := requireEnv(t, EnvOAuth1ConsumerKey, EnvOAuth1PrivateKeyFile)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	privateKey, err := os.ReadFile(env[EnvOAuth1PrivateKeyFile])
	require.NoError(t, err, "cannot read the private key named by %s", EnvOAuth1PrivateKeyFile)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   env[EnvOAuth1ConsumerKey],
			PrivateKeyPEM: privateKey,
		}),
	)
	require.NoError(t, err)

	t.Run("no query string", func(t *testing.T) {
		info, err := client.ServerInfo.Get(liveContext(t))
		require.NoError(t, err,
			"OAuth 1.0a signature rejected; confirm the public key is registered "+
				"on the application link for consumer %q", env[EnvOAuth1ConsumerKey])
		t.Logf("Jira %s (%s)", info.Version, info.DeploymentType)
	})

	// Query parameters are part of the OAuth 1.0a signature base string, so a
	// signature that works without them can still fail with them.
	t.Run("query parameters needing encoding", func(t *testing.T) {
		users, err := client.User.Search(liveContext(t), &user.SearchOptions{
			Query:      "a b",
			MaxResults: 1,
		})
		require.NoError(t, err, "signature rejected once query parameters were present")
		t.Logf("user search returned %d result(s)", len(users))
	})
}

func TestLiveOAuth1Impersonation(t *testing.T) {
	env := requireEnv(t, EnvOAuth1ConsumerKey, EnvOAuth1PrivateKeyFile, EnvOAuth1ImpersonateUser)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	privateKey, err := os.ReadFile(env[EnvOAuth1PrivateKeyFile])
	require.NoError(t, err)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:     env[EnvOAuth1ConsumerKey],
			PrivateKeyPEM:   privateKey,
			ImpersonateUser: env[EnvOAuth1ImpersonateUser],
		}),
	)
	require.NoError(t, err)

	user, err := client.Myself.Get(liveContext(t))
	require.NoError(t, err,
		"two-legged impersonation rejected; the application link must allow "+
			"2-legged OAuth with user impersonation, and user_id must be signed")
	t.Logf("acting as %s (%s)", user.DisplayName, user.EmailAddress)
}

func TestLiveOAuth1ThreeLegged(t *testing.T) {
	env := requireEnv(t, EnvOAuth1ConsumerKey, EnvOAuth1PrivateKeyFile, EnvOAuth1Token, EnvOAuth1TokenSecret)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	privateKey, err := os.ReadFile(env[EnvOAuth1PrivateKeyFile])
	require.NoError(t, err)

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   env[EnvOAuth1ConsumerKey],
			PrivateKeyPEM: privateKey,
			Token:         env[EnvOAuth1Token],
			TokenSecret:   env[EnvOAuth1TokenSecret],
		}),
	)
	require.NoError(t, err)

	user, err := client.Myself.Get(liveContext(t))
	require.NoError(t, err, "three-legged signature rejected (token %s)", redact(env[EnvOAuth1Token]))
	t.Logf("authenticated as %s (%s)", user.DisplayName, user.EmailAddress)
}

// TestLiveOAuth1Handshake runs the first leg of the three-legged handshake,
// which is the only part that needs no human. A temporary credential coming
// back proves the consumer key and signature are accepted.
func TestLiveOAuth1Handshake(t *testing.T) {
	env := requireEnv(t, EnvOAuth1ConsumerKey, EnvOAuth1PrivateKeyFile)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	privateKey, err := os.ReadFile(env[EnvOAuth1PrivateKeyFile])
	require.NoError(t, err)

	authenticator, err := auth.NewOAuth1Auth(&auth.OAuth1Config{
		ConsumerKey:   env[EnvOAuth1ConsumerKey],
		PrivateKeyPEM: privateKey,
		BaseURL:       baseURL,
	})
	require.NoError(t, err)

	requestToken, err := authenticator.RequestToken(liveContext(t), "oob")
	require.NoError(t, err, "request-token endpoint rejected the signature")

	assert.NotEmpty(t, requestToken.Token)
	t.Logf("request token obtained: %s", redact(requestToken.Token))

	authURL, err := authenticator.AuthorizationURL(requestToken.Token)
	require.NoError(t, err)
	t.Logf("a user would approve at: %s", authURL)
	t.Log("the remaining legs need a human; see examples/oauth1")
}

// --- Gateway / custom IdP: client credentials ------------------------------

func TestLiveClientCredentials(t *testing.T) {
	env := requireEnv(t, EnvOAuthClientID, EnvOAuthClientSecret, EnvOAuthTokenURL)
	baseURL := baseURLFor(t, envLiveServerBaseURL)

	authenticator, err := auth.NewClientCredentialsAuth(&auth.ClientCredentialsConfig{
		ClientID:     env[EnvOAuthClientID],
		ClientSecret: env[EnvOAuthClientSecret],
		TokenURL:     env[EnvOAuthTokenURL],
		Scopes:       strings.Fields(os.Getenv(EnvOAuthScopes)),
		Audience:     os.Getenv(EnvOAuthAudience),
	})
	require.NoError(t, err)

	ctx := liveContext(t)

	token, err := authenticator.Token(ctx)
	require.NoError(t, err, "token endpoint rejected the client credentials")
	assert.NotEmpty(t, token.AccessToken)
	t.Logf("token obtained, type %s", token.Type())

	client, err := NewClient(
		WithBaseURL(baseURL),
		WithAuthenticator(authenticator),
	)
	require.NoError(t, err)

	info, err := client.ServerInfo.Get(ctx)
	require.NoError(t, err, "Jira rejected the issued token")
	t.Logf("Jira %s (%s)", info.Version, info.DeploymentType)
}
