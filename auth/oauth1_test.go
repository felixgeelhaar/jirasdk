package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" //nolint:gosec // G505: verifying RSA-SHA1 signatures requires SHA-1
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRSAKey(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return key, pemBytes
}

// parseAuthorizationHeader turns an OAuth header back into its parameters,
// percent-decoding each value.
func parseAuthorizationHeader(t *testing.T, header string) map[string]string {
	t.Helper()

	require.True(t, strings.HasPrefix(header, "OAuth "), "expected OAuth scheme, got %q", header)

	params := map[string]string{}
	for _, part := range regexp.MustCompile(`,\s*`).Split(strings.TrimPrefix(header, "OAuth "), -1) {
		name, value, ok := strings.Cut(part, "=")
		require.True(t, ok, "malformed parameter %q", part)

		decoded, err := url.QueryUnescape(strings.Trim(value, `"`))
		require.NoError(t, err)
		params[name] = decoded
	}

	return params
}

// verifyRSASignature checks the header's signature against the base string the
// server would reconstruct.
func verifyRSASignature(t *testing.T, pub *rsa.PublicKey, req *http.Request, params map[string]string) {
	t.Helper()

	signature, err := base64.StdEncoding.DecodeString(params["oauth_signature"])
	require.NoError(t, err)

	oauthParams := map[string]string{}
	for k, v := range params {
		if k != "oauth_signature" {
			oauthParams[k] = v
		}
	}

	base, err := SignatureBaseString(req, oauthParams)
	require.NoError(t, err)

	digest := sha1.Sum([]byte(base)) //nolint:gosec // G401: RSA-SHA1 verification
	require.NoError(t, rsa.VerifyPKCS1v15(pub, crypto.SHA1, digest[:], signature))
}

func TestNewOAuth1AuthValidation(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	tests := []struct {
		name    string
		config  *OAuth1Config
		wantErr string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: "config is required",
		},
		{
			name:    "missing consumer key",
			config:  &OAuth1Config{PrivateKeyPEM: pemBytes},
			wantErr: "consumer key is required",
		},
		{
			name:    "RSA without private key",
			config:  &OAuth1Config{ConsumerKey: "key"},
			wantErr: "private key is required",
		},
		{
			name:    "HMAC without consumer secret",
			config:  &OAuth1Config{ConsumerKey: "key", SignatureMethod: SignatureMethodHMACSHA1},
			wantErr: "consumer secret is required",
		},
		{
			name:    "unknown signature method",
			config:  &OAuth1Config{ConsumerKey: "key", SignatureMethod: "PLAINTEXT"},
			wantErr: "unsupported signature method",
		},
		{
			name:    "malformed private key",
			config:  &OAuth1Config{ConsumerKey: "key", PrivateKeyPEM: []byte("not a pem block")},
			wantErr: "failed to decode PEM block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewOAuth1Auth(tt.config)
			require.Error(t, err)
			assert.Nil(t, a)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestParseRSAPrivateKeyAcceptsPKCS8(t *testing.T) {
	key, _ := testRSAKey(t)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pkcs8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	parsed, err := ParseRSAPrivateKey(pkcs8)
	require.NoError(t, err)
	assert.True(t, key.Equal(parsed))
}

func TestOAuth1TwoLeggedSignature(t *testing.T) {
	key, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:   "my-consumer",
		PrivateKeyPEM: pemBytes,
	})
	require.NoError(t, err)
	assert.Equal(t, "oauth1_2lo", a.Type())

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/issue/TEST-1?expand=changelog", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	params := parseAuthorizationHeader(t, req.Header.Get("Authorization"))
	assert.Equal(t, "my-consumer", params["oauth_consumer_key"])
	assert.Equal(t, SignatureMethodRSASHA1, params["oauth_signature_method"])
	assert.Equal(t, "1.0", params["oauth_version"])
	assert.NotEmpty(t, params["oauth_nonce"])
	assert.NotEmpty(t, params["oauth_timestamp"])
	assert.NotContains(t, params, "oauth_token", "two-legged requests carry no token")

	verifyRSASignature(t, &key.PublicKey, req, params)
}

func TestOAuth1ThreeLeggedSignature(t *testing.T) {
	key, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:   "my-consumer",
		PrivateKeyPEM: pemBytes,
		Token:         "access-token",
		TokenSecret:   "access-token-secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "oauth1_3lo", a.Type())

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	params := parseAuthorizationHeader(t, req.Header.Get("Authorization"))
	assert.Equal(t, "access-token", params["oauth_token"])

	verifyRSASignature(t, &key.PublicKey, req, params)
}

func TestOAuth1ImpersonationAddsSignedUserID(t *testing.T) {
	key, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:     "my-consumer",
		PrivateKeyPEM:   pemBytes,
		ImpersonateUser: "jsmith",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	assert.Equal(t, "jsmith", req.URL.Query().Get("user_id"))

	// The signature must cover user_id, otherwise Jira rejects the request.
	verifyRSASignature(t, &key.PublicKey, req, parseAuthorizationHeader(t, req.Header.Get("Authorization")))
}

func TestOAuth1ImpersonationSkippedWhenThreeLegged(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:     "my-consumer",
		PrivateKeyPEM:   pemBytes,
		ImpersonateUser: "jsmith",
		Token:           "access-token",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	assert.Empty(t, req.URL.Query().Get("user_id"), "a user token already identifies the user")
}

func TestOAuth1HMACSignature(t *testing.T) {
	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:     "my-consumer",
		ConsumerSecret:  "consumer-secret",
		SignatureMethod: SignatureMethodHMACSHA1,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	params := parseAuthorizationHeader(t, req.Header.Get("Authorization"))
	assert.Equal(t, SignatureMethodHMACSHA1, params["oauth_signature_method"])
	assert.NotEmpty(t, params["oauth_signature"])
}

func TestOAuth1NoncesAreUnique(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{ConsumerKey: "my-consumer", PrivateKeyPEM: pemBytes})
	require.NoError(t, err)

	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
		require.NoError(t, err)
		require.NoError(t, a.Authenticate(req))

		nonce := parseAuthorizationHeader(t, req.Header.Get("Authorization"))["oauth_nonce"]
		require.False(t, seen[nonce], "nonce %q reused", nonce)
		seen[nonce] = true
	}
}

func TestSignatureBaseString(t *testing.T) {
	tests := []struct {
		name   string
		method string
		rawURL string
		want   string
	}{
		{
			name:   "query parameters are merged and sorted",
			method: "GET",
			rawURL: "https://jira.internal/rest/api/2/search?jql=order+by+key&maxResults=50",
			want:   "GET&https%3A%2F%2Fjira.internal%2Frest%2Fapi%2F2%2Fsearch&jql%3Dorder%2520by%2520key%26maxResults%3D50%26oauth_consumer_key%3Dck",
		},
		{
			name:   "default https port is dropped",
			method: "GET",
			rawURL: "https://jira.internal:443/rest/api/2/myself",
			want:   "GET&https%3A%2F%2Fjira.internal%2Frest%2Fapi%2F2%2Fmyself&oauth_consumer_key%3Dck",
		},
		{
			name:   "non-default port is kept",
			method: "GET",
			rawURL: "https://jira.internal:8443/rest/api/2/myself",
			want:   "GET&https%3A%2F%2Fjira.internal%3A8443%2Frest%2Fapi%2F2%2Fmyself&oauth_consumer_key%3Dck",
		},
		{
			name:   "host is lower-cased",
			method: "get",
			rawURL: "https://JIRA.Internal/rest/api/2/myself",
			want:   "GET&https%3A%2F%2Fjira.internal%2Frest%2Fapi%2F2%2Fmyself&oauth_consumer_key%3Dck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.rawURL, nil)
			require.NoError(t, err)

			got, err := SignatureBaseString(req, map[string]string{"oauth_consumer_key": "ck"})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSignatureBaseStringIncludesFormBodyAndRestoresIt(t *testing.T) {
	body := "field=value&other=thing"

	req, err := http.NewRequest(http.MethodPost, "https://jira.internal/rest/api/2/issue", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	base, err := SignatureBaseString(req, map[string]string{"oauth_consumer_key": "ck"})
	require.NoError(t, err)
	assert.Contains(t, base, "field%3Dvalue")
	assert.Contains(t, base, "other%3Dthing")

	// The body must still be readable by the actual request.
	rest, err := readAllString(req)
	require.NoError(t, err)
	assert.Equal(t, body, rest)
}

func TestSignatureBaseStringIgnoresJSONBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://jira.internal/rest/api/2/issue", strings.NewReader(`{"a":"b"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	base, err := SignatureBaseString(req, map[string]string{"oauth_consumer_key": "ck"})
	require.NoError(t, err)
	assert.NotContains(t, base, "a%3Db")
}

func TestOAuth1Handshake(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	var requestTokenCalls, accessTokenCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "OAuth "))

		switch r.URL.Path {
		case OAuth1RequestTokenPath:
			requestTokenCalls++
			_, _ = w.Write([]byte("oauth_token=req-token&oauth_token_secret=req-secret&oauth_callback_confirmed=true"))
		case OAuth1AccessTokenPath:
			accessTokenCalls++
			params := parseAuthorizationHeader(t, r.Header.Get("Authorization"))
			assert.Equal(t, "req-token", params["oauth_token"])
			assert.Equal(t, "the-verifier", params["oauth_verifier"])
			_, _ = w.Write([]byte("oauth_token=acc-token&oauth_token_secret=acc-secret"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:   "my-consumer",
		PrivateKeyPEM: pemBytes,
		BaseURL:       server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()

	reqToken, err := a.RequestToken(ctx, "https://app.example.com/callback")
	require.NoError(t, err)
	assert.Equal(t, "req-token", reqToken.Token)
	assert.Equal(t, "req-secret", reqToken.TokenSecret)
	assert.Equal(t, 1, requestTokenCalls)

	authURL, err := a.AuthorizationURL(reqToken.Token)
	require.NoError(t, err)
	assert.Equal(t, server.URL+OAuth1AuthorizePath+"?oauth_token=req-token", authURL)

	token, secret, err := a.AccessToken(ctx, reqToken, "the-verifier")
	require.NoError(t, err)
	assert.Equal(t, "acc-token", token)
	assert.Equal(t, "acc-secret", secret)
	assert.Equal(t, 1, accessTokenCalls)

	// The handshake should leave the authenticator ready for three-legged use.
	assert.Equal(t, "oauth1_3lo", a.Type())
	gotToken, gotSecret := a.Token()
	assert.Equal(t, "acc-token", gotToken)
	assert.Equal(t, "acc-secret", gotSecret)
}

func TestOAuth1HandshakeRequiresBaseURL(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	a, err := NewOAuth1Auth(&OAuth1Config{ConsumerKey: "my-consumer", PrivateKeyPEM: pemBytes})
	require.NoError(t, err)

	_, err = a.RequestToken(context.Background(), "oob")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")

	_, err = a.AuthorizationURL("req-token")
	require.Error(t, err)
}

func TestOAuth1HandshakePropagatesServerError(t *testing.T) {
	_, pemBytes := testRSAKey(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("signature_invalid"))
	}))
	defer server.Close()

	a, err := NewOAuth1Auth(&OAuth1Config{
		ConsumerKey:   "my-consumer",
		PrivateKeyPEM: pemBytes,
		BaseURL:       server.URL,
	})
	require.NoError(t, err)

	_, err = a.RequestToken(context.Background(), "oob")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature_invalid")
}

func readAllString(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
