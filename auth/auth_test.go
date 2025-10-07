package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPITokenAuth(t *testing.T) {
	tests := []struct {
		name  string
		email string
		token string
	}{
		{
			name:  "valid credentials",
			email: "user@example.com",
			token: "api-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAPITokenAuth(tt.email, tt.token)
			require.NotNil(t, auth)

			assert.Equal(t, "api_token", auth.Type())

			req, err := http.NewRequest("GET", "https://example.com", nil)
			require.NoError(t, err)

			err = auth.Authenticate(req)
			require.NoError(t, err)

			authHeader := req.Header.Get("Authorization")
			assert.NotEmpty(t, authHeader)
			assert.Contains(t, authHeader, "Basic ")
		})
	}
}

func TestPATAuth(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "valid token",
			token: "pat-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewPATAuth(tt.token)
			require.NotNil(t, auth)

			assert.Equal(t, "pat", auth.Type())

			req, err := http.NewRequest("GET", "https://example.com", nil)
			require.NoError(t, err)

			err = auth.Authenticate(req)
			require.NoError(t, err)

			authHeader := req.Header.Get("Authorization")
			assert.NotEmpty(t, authHeader)
			assert.Contains(t, authHeader, "Bearer ")
			assert.Contains(t, authHeader, tt.token)
		})
	}
}

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "valid credentials",
			username: "admin",
			password: "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewBasicAuth(tt.username, tt.password)
			require.NotNil(t, auth)

			assert.Equal(t, "basic", auth.Type())

			req, err := http.NewRequest("GET", "https://example.com", nil)
			require.NoError(t, err)

			err = auth.Authenticate(req)
			require.NoError(t, err)

			authHeader := req.Header.Get("Authorization")
			assert.NotEmpty(t, authHeader)
			assert.Contains(t, authHeader, "Basic ")
		})
	}
}
