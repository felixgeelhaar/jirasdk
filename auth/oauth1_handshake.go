package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// OAuth1RequestToken is the temporary credential issued at the start of the
// three-legged handshake.
type OAuth1RequestToken struct {
	// Token is the temporary request token.
	Token string

	// TokenSecret signs the access-token exchange.
	TokenSecret string
}

// RequestToken performs step one of the OAuth 1.0a three-legged flow, asking
// Jira Server for a temporary credential.
//
// Pass callbackURL to have Jira redirect back with an oauth_verifier, or "oob"
// for out-of-band verification where the user copies the verifier manually.
func (a *OAuth1Auth) RequestToken(ctx context.Context, callbackURL string) (*OAuth1RequestToken, error) {
	if a.baseURL == "" {
		return nil, fmt.Errorf("base URL is required for the OAuth 1.0a handshake")
	}
	if callbackURL == "" {
		callbackURL = "oob"
	}

	params := a.oauthParams()
	params["oauth_callback"] = callbackURL

	values, err := a.handshakeCall(ctx, a.baseURL+OAuth1RequestTokenPath, params, "")
	if err != nil {
		return nil, fmt.Errorf("failed to obtain request token: %w", err)
	}

	token := values.Get("oauth_token")
	if token == "" {
		return nil, fmt.Errorf("request token response contained no oauth_token")
	}

	return &OAuth1RequestToken{
		Token:       token,
		TokenSecret: values.Get("oauth_token_secret"),
	}, nil
}

// AuthorizationURL returns the URL the user must visit to approve the request
// token. This is step two of the three-legged flow.
func (a *OAuth1Auth) AuthorizationURL(requestToken string) (string, error) {
	if a.baseURL == "" {
		return "", fmt.Errorf("base URL is required for the OAuth 1.0a handshake")
	}
	if requestToken == "" {
		return "", fmt.Errorf("request token is required")
	}

	u, err := url.Parse(a.baseURL + OAuth1AuthorizePath)
	if err != nil {
		return "", fmt.Errorf("invalid authorization URL: %w", err)
	}

	q := u.Query()
	q.Set("oauth_token", requestToken)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// AccessToken performs step three, exchanging the authorized request token and
// its verifier for a long-lived access token. On success the token is installed
// on the authenticator, so subsequent requests are signed three-legged.
func (a *OAuth1Auth) AccessToken(ctx context.Context, requestToken *OAuth1RequestToken, verifier string) (token, tokenSecret string, err error) {
	if a.baseURL == "" {
		return "", "", fmt.Errorf("base URL is required for the OAuth 1.0a handshake")
	}
	if requestToken == nil || requestToken.Token == "" {
		return "", "", fmt.Errorf("request token is required")
	}

	params := a.oauthParams()
	params["oauth_token"] = requestToken.Token
	if verifier != "" {
		params["oauth_verifier"] = verifier
	}

	values, err := a.handshakeCall(ctx, a.baseURL+OAuth1AccessTokenPath, params, requestToken.TokenSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange request token: %w", err)
	}

	token = values.Get("oauth_token")
	if token == "" {
		return "", "", fmt.Errorf("access token response contained no oauth_token")
	}
	tokenSecret = values.Get("oauth_token_secret")

	a.SetToken(token, tokenSecret)
	return token, tokenSecret, nil
}

// handshakeCall signs and executes one handshake request, returning the
// form-encoded response body. tokenSecret is only consulted by HMAC-SHA1, where
// the temporary credential's secret forms part of the signing key.
func (a *OAuth1Auth) handshakeCall(ctx context.Context, endpoint string, params map[string]string, tokenSecret string) (url.Values, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	signer := a
	if tokenSecret != "" && a.signatureMethod == SignatureMethodHMACSHA1 {
		clone := *a
		clone.tokenSecret = tokenSecret
		signer = &clone
	}

	signature, err := signer.sign(req, params)
	if err != nil {
		return nil, err
	}

	signed := make(map[string]string, len(params)+1)
	for k, v := range params {
		signed[k] = v
	}
	signed["oauth_signature"] = signature
	req.Header.Set("Authorization", authorizationHeader(signed))

	resp, err := a.httpClient.Do(req) //nolint:gosec // G704: the URL is the operator-configured Jira base URL, not tainted input
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("endpoint returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return values, nil
}
