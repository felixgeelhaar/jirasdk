package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
)

// JWTClaimQSH is the name of the Atlassian query string hash claim.
const JWTClaimQSH = "qsh"

// ContextQSH is the wildcard query string hash accepted by Atlassian for
// context JWTs, which are not bound to a single request.
const ContextQSH = "context-qsh"

// CanonicalRequest builds the Atlassian canonical request string used to derive
// the "qsh" claim of a Connect JWT.
//
// The format is METHOD&canonical-path&canonical-query-string, where the path has
// the product's context path removed and the query string is sorted and
// percent-encoded according to Atlassian's rules.
//
// See https://developer.atlassian.com/cloud/jira/platform/understanding-jwt-for-connect-apps/
func CanonicalRequest(req *http.Request, contextPath string) (string, error) {
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("request and URL are required")
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = http.MethodGet
	}

	path := canonicalPath(req.URL.EscapedPath(), contextPath)
	query := canonicalQuery(req.URL.Query())

	return strings.Join([]string{method, path, query}, "&"), nil
}

// QueryStringHash returns the hex-encoded SHA-256 digest of the canonical
// request, suitable for use as the "qsh" claim.
func QueryStringHash(req *http.Request, contextPath string) (string, error) {
	canonical, err := CanonicalRequest(req, contextPath)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:]), nil
}

// canonicalPath strips the context path, normalises the trailing slash and
// escapes ampersands so they cannot be confused with the canonical separator.
func canonicalPath(escapedPath, contextPath string) string {
	contextPath = strings.TrimSuffix(contextPath, "/")
	if contextPath != "" && contextPath != "/" {
		if escapedPath == contextPath {
			escapedPath = "/"
		} else {
			escapedPath = strings.TrimPrefix(escapedPath, contextPath)
		}
	}

	if escapedPath == "" {
		return "/"
	}

	// A trailing slash is insignificant unless the path is the root itself.
	if len(escapedPath) > 1 {
		escapedPath = strings.TrimSuffix(escapedPath, "/")
	}
	if escapedPath == "" {
		return "/"
	}

	// "&" is the canonical request separator, so it must never appear raw.
	return strings.ReplaceAll(escapedPath, "&", "%26")
}

// canonicalQuery sorts and encodes the query parameters. The "jwt" parameter is
// excluded because it cannot be part of the hash it is carried in.
//
// Names and values are sorted before encoding, not after. The two orders differ
// whenever a name or value contains a character that percent-encodes: "%" is
// 0x25, below every unreserved character, so encoding first would sort ":"
// ahead of "0" where Atlassian sorts it after.
func canonicalQuery(values map[string][]string) string {
	names := make([]string, 0, len(values))
	for name := range values {
		if name == "jwt" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	pairs := make([]string, 0, len(names))
	for _, name := range names {
		raw := slices.Clone(values[name])
		sort.Strings(raw)

		encoded := make([]string, 0, len(raw))
		for _, v := range raw {
			encoded = append(encoded, percentEncode(v))
		}

		// Repeated parameters collapse into one value separated by a literal
		// comma. A comma inside a value encodes to %2C, which is what keeps the
		// separator and the content distinguishable.
		pairs = append(pairs, percentEncode(name)+"="+strings.Join(encoded, ","))
	}

	return strings.Join(pairs, "&")
}

// percentEncode applies RFC 3986 unreserved-character encoding. Unlike
// url.QueryEscape it encodes spaces as %20 rather than "+" and leaves "~" alone,
// both of which Atlassian requires.
func percentEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "%%%02X", c)
	}

	return b.String()
}

// isUnreserved reports whether c is an RFC 3986 unreserved character.
func isUnreserved(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z',
		c >= 'a' && c <= 'z',
		c >= '0' && c <= '9':
		return true
	case c == '-', c == '.', c == '_', c == '~':
		return true
	default:
		return false
	}
}
