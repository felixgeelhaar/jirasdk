package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSharedSecret = "shared-secret-value"

func parseConnectToken(t *testing.T, raw string) jwt.MapClaims {
	t.Helper()

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(tok *jwt.Token) (interface{}, error) {
		assert.Equal(t, "HS256", tok.Method.Alg())
		return []byte(testSharedSecret), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	return claims
}

func TestNewConnectJWTAuthValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConnectJWTConfig
		wantErr string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: "config is required",
		},
		{
			name:    "missing app key",
			config:  &ConnectJWTConfig{SharedSecret: testSharedSecret},
			wantErr: "app key is required",
		},
		{
			name:    "missing shared secret",
			config:  &ConnectJWTConfig{AppKey: "com.example.app"},
			wantErr: "shared secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewConnectJWTAuth(tt.config)
			require.Error(t, err)
			assert.Nil(t, a)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConnectJWTAuthenticate(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
	})
	require.NoError(t, err)
	assert.Equal(t, "connect_jwt", a.Type())

	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)

	require.NoError(t, a.Authenticate(req))

	header := req.Header.Get("Authorization")
	require.True(t, strings.HasPrefix(header, "JWT "), "expected JWT scheme, got %q", header)

	claims := parseConnectToken(t, strings.TrimPrefix(header, "JWT "))
	assert.Equal(t, "com.example.app", claims["iss"])
	assert.NotContains(t, claims, "sub")

	wantQSH, err := QueryStringHash(req, "")
	require.NoError(t, err)
	assert.Equal(t, wantQSH, claims[JWTClaimQSH])
}

func TestConnectJWTIncludesSubjectWhenImpersonating(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
		Subject:      "5b10ac8d82e05b22cc7d4ef5",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)
	require.NoError(t, a.Authenticate(req))

	claims := parseConnectToken(t, strings.TrimPrefix(req.Header.Get("Authorization"), "JWT "))
	assert.Equal(t, "5b10ac8d82e05b22cc7d4ef5", claims["sub"])
}

func TestConnectJWTQSHBindsToTheRequest(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
	})
	require.NoError(t, err)

	first, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)
	second, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/project", nil)
	require.NoError(t, err)

	firstToken, err := a.Sign(first)
	require.NoError(t, err)
	secondToken, err := a.Sign(second)
	require.NoError(t, err)

	firstQSH := parseConnectToken(t, firstToken)[JWTClaimQSH]
	secondQSH := parseConnectToken(t, secondToken)[JWTClaimQSH]
	assert.NotEqual(t, firstQSH, secondQSH, "qsh must differ for different paths")
}

func TestConnectJWTExpiry(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:        "com.example.app",
		SharedSecret:  testSharedSecret,
		TokenLifetime: 60 * time.Second,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)

	claims := parseConnectToken(t, mustSign(t, a, req))

	iat, err := claims.GetIssuedAt()
	require.NoError(t, err)
	exp, err := claims.GetExpirationTime()
	require.NoError(t, err)

	assert.Equal(t, 60*time.Second, exp.Sub(iat.Time))
}

func TestConnectJWTDefaultLifetime(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)

	claims := parseConnectToken(t, mustSign(t, a, req))
	iat, err := claims.GetIssuedAt()
	require.NoError(t, err)
	exp, err := claims.GetExpirationTime()
	require.NoError(t, err)

	assert.Equal(t, DefaultConnectJWTLifetime, exp.Sub(iat.Time))
}

func TestConnectJWTRejectsWrongSecret(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	require.NoError(t, err)

	_, err = jwt.Parse(mustSign(t, a, req), func(*jwt.Token) (interface{}, error) {
		return []byte("the-wrong-secret"), nil
	})
	assert.Error(t, err)
}

func TestConnectJWTContextPath(t *testing.T) {
	a, err := NewConnectJWTAuth(&ConnectJWTConfig{
		AppKey:       "com.example.app",
		SharedSecret: testSharedSecret,
		ContextPath:  "/jira",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://jira.internal/jira/rest/api/2/myself", nil)
	require.NoError(t, err)

	claims := parseConnectToken(t, mustSign(t, a, req))

	// The hash must match the path with the context path removed.
	bare, err := http.NewRequest(http.MethodGet, "https://jira.internal/rest/api/2/myself", nil)
	require.NoError(t, err)
	wantQSH, err := QueryStringHash(bare, "")
	require.NoError(t, err)

	assert.Equal(t, wantQSH, claims[JWTClaimQSH])
}

func mustSign(t *testing.T, a *ConnectJWTAuth, req *http.Request) string {
	t.Helper()
	token, err := a.Sign(req)
	require.NoError(t, err)
	return token
}
