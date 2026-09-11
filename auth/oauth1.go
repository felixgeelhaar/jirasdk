package auth

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" //nolint:gosec // G505: OAuth 1.0a mandates SHA-1; required by the Jira Server protocol
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// OAuth 1.0a signature methods supported by Jira Server and Data Center.
const (
	// SignatureMethodRSASHA1 is the method used by Jira application links.
	SignatureMethodRSASHA1 = "RSA-SHA1"

	// SignatureMethodHMACSHA1 is the shared-secret alternative.
	SignatureMethodHMACSHA1 = "HMAC-SHA1"
)

// Jira Server OAuth 1.0a endpoint paths, relative to the instance base URL.
const (
	OAuth1RequestTokenPath = "/plugins/servlet/oauth/request-token" //nolint:gosec // G101: endpoint path, not a credential
	OAuth1AuthorizePath    = "/plugins/servlet/oauth/authorize"
	OAuth1AccessTokenPath  = "/plugins/servlet/oauth/access-token" //nolint:gosec // G101: endpoint path, not a credential
)

// OAuth1Config configures OAuth 1.0a authentication for Jira Server/Data Center.
type OAuth1Config struct {
	// ConsumerKey is the consumer key registered on the Jira application link.
	ConsumerKey string

	// PrivateKeyPEM is the PEM-encoded RSA private key whose public half was
	// uploaded to the application link. Required for RSA-SHA1.
	PrivateKeyPEM []byte

	// ConsumerSecret is the shared secret. Required for HMAC-SHA1, and used as
	// the token-secret prefix during the three-legged handshake.
	ConsumerSecret string

	// SignatureMethod selects the signing algorithm. Defaults to RSA-SHA1.
	SignatureMethod string

	// Token is the OAuth access token for three-legged use. Leave empty for
	// two-legged requests, where the app authenticates as the consumer itself.
	Token string

	// TokenSecret accompanies Token in three-legged use.
	TokenSecret string

	// ImpersonateUser sets the "user_id" parameter that Jira Server accepts on
	// two-legged requests to act on behalf of a user. Requires the application
	// link to permit two-legged impersonation.
	ImpersonateUser string

	// BaseURL is the Jira instance URL. Only needed for the three-legged
	// handshake helpers.
	BaseURL string

	// HTTPClient is used by the handshake helpers. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// OAuth1Auth implements OAuth 1.0a request signing for Jira Server and Data
// Center.
//
// With no Token configured it performs two-legged authentication: the consumer
// signs each request with its own key and acts as itself, optionally
// impersonating a user via ImpersonateUser. Once a Token and TokenSecret are
// set — either supplied directly or obtained through the handshake helpers —
// the same authenticator performs three-legged requests on behalf of the user
// who granted access.
type OAuth1Auth struct {
	consumerKey     string
	consumerSecret  string
	signatureMethod string
	privateKey      *rsa.PrivateKey
	token           string
	tokenSecret     string
	impersonateUser string
	baseURL         string
	httpClient      *http.Client
}

// NewOAuth1Auth creates an OAuth 1.0a authenticator.
//
// Example (two-legged):
//
//	auth, err := auth.NewOAuth1Auth(&auth.OAuth1Config{
//	    ConsumerKey:   "my-consumer-key",
//	    PrivateKeyPEM: keyBytes,
//	})
func NewOAuth1Auth(config *OAuth1Config) (*OAuth1Auth, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth 1.0a config is required")
	}
	if config.ConsumerKey == "" {
		return nil, fmt.Errorf("consumer key is required")
	}

	method := config.SignatureMethod
	if method == "" {
		method = SignatureMethodRSASHA1
	}

	a := &OAuth1Auth{
		consumerKey:     config.ConsumerKey,
		consumerSecret:  config.ConsumerSecret,
		signatureMethod: method,
		token:           config.Token,
		tokenSecret:     config.TokenSecret,
		impersonateUser: config.ImpersonateUser,
		baseURL:         strings.TrimSuffix(config.BaseURL, "/"),
		httpClient:      config.HTTPClient,
	}
	if a.httpClient == nil {
		a.httpClient = http.DefaultClient
	}

	switch method {
	case SignatureMethodRSASHA1:
		if len(config.PrivateKeyPEM) == 0 {
			return nil, fmt.Errorf("private key is required for %s", SignatureMethodRSASHA1)
		}
		key, err := ParseRSAPrivateKey(config.PrivateKeyPEM)
		if err != nil {
			return nil, err
		}
		a.privateKey = key
	case SignatureMethodHMACSHA1:
		if config.ConsumerSecret == "" {
			return nil, fmt.Errorf("consumer secret is required for %s", SignatureMethodHMACSHA1)
		}
	default:
		return nil, fmt.Errorf("unsupported signature method: %s", method)
	}

	return a, nil
}

// ParseRSAPrivateKey decodes a PEM-encoded RSA private key in either PKCS#1 or
// PKCS#8 form.
func ParseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from private key")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is %T, want *rsa.PrivateKey", parsed)
	}

	return key, nil
}

// SetToken installs the access token obtained from the three-legged handshake,
// switching subsequent requests from two-legged to three-legged.
func (a *OAuth1Auth) SetToken(token, tokenSecret string) {
	a.token = token
	a.tokenSecret = tokenSecret
}

// Token returns the current access token and secret.
func (a *OAuth1Auth) Token() (token, tokenSecret string) {
	return a.token, a.tokenSecret
}

// Authenticate signs the request and attaches the OAuth Authorization header.
func (a *OAuth1Auth) Authenticate(req *http.Request) error {
	if a.impersonateUser != "" && a.token == "" {
		// Jira Server reads the impersonated user from the query string, and it
		// must be part of the signature.
		q := req.URL.Query()
		if q.Get("user_id") == "" {
			q.Set("user_id", a.impersonateUser)
			req.URL.RawQuery = q.Encode()
		}
	}

	params := a.oauthParams()
	if a.token != "" {
		params["oauth_token"] = a.token
	}

	signature, err := a.sign(req, params)
	if err != nil {
		return err
	}
	params["oauth_signature"] = signature

	req.Header.Set("Authorization", authorizationHeader(params))
	return nil
}

// Type returns the authentication type. Two- and three-legged use are reported
// separately so logs make the distinction visible.
func (a *OAuth1Auth) Type() string {
	if a.token == "" {
		return "oauth1_2lo"
	}
	return "oauth1_3lo"
}

// oauthParams builds the protocol parameters common to every signed request.
func (a *OAuth1Auth) oauthParams() map[string]string {
	return map[string]string{
		"oauth_consumer_key":     a.consumerKey,
		"oauth_nonce":            newNonce(),
		"oauth_signature_method": a.signatureMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_version":          "1.0",
	}
}

// sign computes the RFC 5849 signature over the request.
func (a *OAuth1Auth) sign(req *http.Request, oauthParams map[string]string) (string, error) {
	base, err := SignatureBaseString(req, oauthParams)
	if err != nil {
		return "", err
	}

	switch a.signatureMethod {
	case SignatureMethodRSASHA1:
		digest := sha1.Sum([]byte(base)) //nolint:gosec // G401: SHA-1 is mandated by the RSA-SHA1 signature method
		sig, err := rsa.SignPKCS1v15(rand.Reader, a.privateKey, crypto.SHA1, digest[:])
		if err != nil {
			return "", fmt.Errorf("failed to sign request: %w", err)
		}
		return base64.StdEncoding.EncodeToString(sig), nil

	case SignatureMethodHMACSHA1:
		key := percentEncode(a.consumerSecret) + "&" + percentEncode(a.tokenSecret)
		mac := hmac.New(sha1.New, []byte(key)) //nolint:gosec // G401: SHA-1 is mandated by the HMAC-SHA1 signature method
		mac.Write([]byte(base))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil

	default:
		return "", fmt.Errorf("unsupported signature method: %s", a.signatureMethod)
	}
}

// SignatureBaseString builds the RFC 5849 section 3.4.1 signature base string:
// METHOD&base-URI&normalized-parameters, each component percent-encoded.
//
// Form-encoded request bodies participate in the signature, so the body is read
// and restored.
func SignatureBaseString(req *http.Request, oauthParams map[string]string) (string, error) {
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("request and URL are required")
	}

	collected := make(map[string][]string, len(oauthParams))
	for k, v := range oauthParams {
		// oauth_signature is excluded from its own input.
		if k == "oauth_signature" {
			continue
		}
		collected[k] = append(collected[k], v)
	}

	for k, vs := range req.URL.Query() {
		collected[k] = append(collected[k], vs...)
	}

	bodyParams, err := formBodyParams(req)
	if err != nil {
		return "", err
	}
	for k, vs := range bodyParams {
		collected[k] = append(collected[k], vs...)
	}

	return strings.Join([]string{
		percentEncode(strings.ToUpper(req.Method)),
		percentEncode(baseStringURI(req.URL)),
		percentEncode(normalizeParams(collected)),
	}, "&"), nil
}

// formBodyParams extracts parameters from a form-encoded body, leaving the body
// readable for the actual request.
func formBodyParams(req *http.Request) (url.Values, error) {
	if req.Body == nil {
		return nil, nil
	}

	mediaType, _, _ := strings.Cut(req.Header.Get("Content-Type"), ";")
	if strings.TrimSpace(mediaType) != "application/x-www-form-urlencoded" {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	_ = req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(body))

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse form body: %w", err)
	}

	return values, nil
}

// baseStringURI lower-cases the scheme and host and drops the default port,
// the query and the fragment.
func baseStringURI(u *url.URL) string {
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)

	if h, port, ok := strings.Cut(host, ":"); ok {
		if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
			host = h
		}
	}

	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}

	return scheme + "://" + host + path
}

// normalizeParams percent-encodes every name and value, then sorts by encoded
// name and, for repeats, by encoded value.
func normalizeParams(params map[string][]string) string {
	pairs := make([]string, 0, len(params))

	for name, values := range params {
		encodedName := percentEncode(name)
		for _, value := range values {
			pairs = append(pairs, encodedName+"="+percentEncode(value))
		}
	}

	sort.Strings(pairs)
	return strings.Join(pairs, "&")
}

// authorizationHeader renders the OAuth Authorization header value.
func authorizationHeader(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", percentEncode(k), percentEncode(params[k])))
	}

	return "OAuth " + strings.Join(parts, ", ")
}

// newNonce returns a random, single-use value for the oauth_nonce parameter.
func newNonce() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand failure is unrecoverable; fall back to the clock so the
		// request is still well-formed rather than silently unsigned.
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:])
}
