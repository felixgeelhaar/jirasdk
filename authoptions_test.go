package jirasdk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/felixgeelhaar/jirasdk/auth"
	"github.com/felixgeelhaar/jirasdk/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func testPrivateKeyPEM(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func TestWithConnectJWT(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithConnectJWT(&auth.ConnectJWTConfig{
			AppKey:       "com.example.app",
			SharedSecret: "shared-secret",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "connect_jwt", client.Authenticator.Type())
}

func TestWithConnectJWTValidationError(t *testing.T) {
	_, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithConnectJWT(&auth.ConnectJWTConfig{AppKey: "com.example.app"}),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shared secret is required")
}

func TestWithJWTBearer(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithJWTBearer(&auth.JWTBearerConfig{
			OAuthClientID: "oauth-client-id",
			SharedSecret:  "shared-secret",
			AccountID:     "account-id",
			SiteURL:       "https://example.atlassian.net",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "jwt_bearer", client.Authenticator.Type())
}

func TestWithOAuth1TwoLegged(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://jira.internal"),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   "my-consumer",
			PrivateKeyPEM: testPrivateKeyPEM(t),
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "oauth1_2lo", client.Authenticator.Type())
}

func TestWithOAuth1ThreeLegged(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://jira.internal"),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   "my-consumer",
			PrivateKeyPEM: testPrivateKeyPEM(t),
			Token:         "access-token",
			TokenSecret:   "access-secret",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "oauth1_3lo", client.Authenticator.Type())
}

func TestWithOAuth1ValidationError(t *testing.T) {
	_, err := NewClient(
		WithBaseURL("https://jira.internal"),
		WithOAuth1(&auth.OAuth1Config{ConsumerKey: "my-consumer"}),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "private key is required")
}

func TestWithClientCredentials(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://jira.example.com"),
		WithClientCredentials(&auth.ClientCredentialsConfig{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			TokenURL:     "https://idp.example.com/oauth2/token",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "client_credentials", client.Authenticator.Type())
}

func TestWithAuthenticator(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithAuthenticator(auth.NewPATAuth("token")),
	)
	require.NoError(t, err)
	assert.Equal(t, "pat", client.Authenticator.Type())

	_, err = NewClient(WithBaseURL("https://example.atlassian.net"), WithAuthenticator(nil))
	require.Error(t, err)
}

// --- cloud ID resolution ---

func newOAuth2Client(t *testing.T, opts ...Option) *Client {
	t.Helper()

	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	oauth.SetToken(&oauth2.Token{AccessToken: "access-token"})

	client, err := NewClient(append([]Option{WithOAuth2(oauth)}, opts...)...)
	require.NoError(t, err)

	return client
}

func TestOAuth2ClientRoutesThroughGatewayWithPinnedCloudID(t *testing.T) {
	client := newOAuth2Client(t,
		WithBaseURL("https://example.atlassian.net"),
		WithCloudID("pinned-cloud-id"),
	)

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)

	// A 3LO token is rejected at the site URL, so requests must go to the
	// gateway with the cloud ID prefix intact.
	assert.Equal(t, "https://api.atlassian.com/ex/jira/pinned-cloud-id/rest/api/3/myself", req.URL.String())
}

func TestOAuth2ClientNeedsNoBaseURL(t *testing.T) {
	client := newOAuth2Client(t, WithCloudID("pinned-cloud-id"))

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://api.atlassian.com/ex/jira/pinned-cloud-id/rest/api/3/myself", req.URL.String())
}

func TestWithoutCloudIDResolutionKeepsBaseURL(t *testing.T) {
	client := newOAuth2Client(t,
		WithBaseURL("https://proxy.example.com"),
		WithoutCloudIDResolution(),
	)

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://proxy.example.com/rest/api/3/myself", req.URL.String())
}

func TestCustomOAuth2EndpointSkipsCloudResolution(t *testing.T) {
	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID: "client-id",
		AuthURL:  "https://keycloak.internal/auth",
		TokenURL: "https://keycloak.internal/token",
	})
	oauth.SetToken(&oauth2.Token{AccessToken: "access-token"})

	client, err := NewClient(
		WithBaseURL("https://jira.internal"),
		WithOAuth2(oauth),
	)
	require.NoError(t, err)

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/2/myself", nil)
	require.NoError(t, err)

	// A non-Atlassian authorization server implies a deployment where cloud ID
	// routing does not apply.
	assert.Equal(t, "https://jira.internal/rest/api/2/myself", req.URL.String())
}

func TestNonOAuth2AuthSkipsCloudResolution(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithAPIToken("user@example.com", "token"),
	)
	require.NoError(t, err)

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.atlassian.net/rest/api/3/myself", req.URL.String())
}

func TestWithCloudIDRequiresValue(t *testing.T) {
	_, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithAPIToken("user@example.com", "token"),
		WithCloudID(""),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cloud ID is required")
}

// fixedResolver returns a constant base URL.
type fixedResolver struct{ raw string }

func (r fixedResolver) ResolveBaseURL(_ context.Context) (*url.URL, error) {
	return url.Parse(r.raw)
}

func TestWithBaseURLResolverOverridesBaseURL(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithAPIToken("user@example.com", "token"),
		WithBaseURLResolver(fixedResolver{raw: "https://gateway.example.com/jira"}),
	)
	require.NoError(t, err)

	req, err := client.Transport.NewRequest(context.Background(), http.MethodGet, "/rest/api/3/myself", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://gateway.example.com/jira/rest/api/3/myself", req.URL.String())

	_, err = NewClient(
		WithBaseURL("https://example.atlassian.net"),
		WithAPIToken("user@example.com", "token"),
		WithBaseURLResolver(nil),
	)
	require.Error(t, err)
}

func TestBaseURLResolverInterfaceIsSatisfied(t *testing.T) {
	resolver, err := auth.NewCloudIDResolver(&auth.CloudIDResolverConfig{CloudID: "cloud-id"})
	require.NoError(t, err)

	var _ transport.BaseURLResolver = resolver
}

// --- end-to-end request signing ---

func TestConnectJWTClientSignsRealRequest(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accountId":"123","displayName":"Test User"}`))
	}))
	defer server.Close()

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithConnectJWT(&auth.ConnectJWTConfig{
			AppKey:       "com.example.app",
			SharedSecret: "shared-secret",
		}),
	)
	require.NoError(t, err)

	_, err = client.Myself.Get(context.Background())
	require.NoError(t, err)
	assert.Contains(t, gotAuth, "JWT ")
}

func TestOAuth1ClientSignsRealRequest(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accountId":"123","displayName":"Test User"}`))
	}))
	defer server.Close()

	client, err := NewClient(
		WithBaseURL(server.URL),
		WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   "my-consumer",
			PrivateKeyPEM: testPrivateKeyPEM(t),
		}),
	)
	require.NoError(t, err)

	_, err = client.Myself.Get(context.Background())
	require.NoError(t, err)
	assert.Contains(t, gotAuth, "OAuth ")
	assert.Contains(t, gotAuth, `oauth_signature_method="RSA-SHA1"`)
}
