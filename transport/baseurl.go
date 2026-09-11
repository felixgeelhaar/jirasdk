package transport

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// BaseURLResolver supplies the base URL for requests, resolved lazily on first
// use. It exists for authentication schemes that cannot know the effective host
// up front — notably Jira Cloud OAuth 2.0 (3LO), where requests go to
// https://api.atlassian.com/ex/jira/{cloudID} and the cloud ID is only
// discoverable once an access token exists.
//
// Implementations must be safe for concurrent use and should cache their result.
type BaseURLResolver interface {
	ResolveBaseURL(ctx context.Context) (*url.URL, error)
}

// WithBaseURLResolver sets a resolver consulted on each request. When set it
// takes precedence over the base URL passed to New.
func WithBaseURLResolver(resolver BaseURLResolver) TransportOption {
	return func(cfg *Config) {
		cfg.baseURLResolver = resolver
	}
}

// resolveURL joins a request path onto a base URL.
//
// It deliberately does not use url.URL.Parse, whose RFC 3986 reference
// resolution discards the base path whenever the request path is absolute.
// Every service in this SDK builds absolute paths such as "/rest/api/3/myself",
// so reference resolution would drop the "/ex/jira/{cloudID}" prefix of a
// gateway base URL, and the context path of a Jira Server instance deployed
// under one. Joining preserves both.
func resolveURL(base *url.URL, path string) (*url.URL, error) {
	if base == nil {
		return nil, fmt.Errorf("base URL is required")
	}

	ref, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// An absolute URL overrides the base entirely, which callers rely on for
	// following server-supplied links such as pagination cursors.
	if ref.IsAbs() {
		return ref, nil
	}

	resolved := *base
	resolved.Path = joinPaths(base.Path, ref.Path)
	resolved.RawPath = ""
	resolved.RawQuery = ref.RawQuery
	resolved.Fragment = ref.Fragment

	return &resolved, nil
}

// joinPaths concatenates two URL paths with exactly one separating slash.
func joinPaths(basePath, refPath string) string {
	switch {
	case basePath == "" || basePath == "/":
		if refPath == "" {
			return "/"
		}
		if strings.HasPrefix(refPath, "/") {
			return refPath
		}
		return "/" + refPath
	case refPath == "":
		return basePath
	default:
		return strings.TrimSuffix(basePath, "/") + "/" + strings.TrimPrefix(refPath, "/")
	}
}
