package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// AccessibleResourcesURL lists the Atlassian sites an access token can reach.
	AccessibleResourcesURL = "https://api.atlassian.com/oauth/token/accessible-resources" //nolint:gosec // G101: public endpoint URL, not a credential

	// APIGatewayURL is the host that serves Jira Cloud APIs for OAuth 2.0 (3LO)
	// tokens. Site URLs such as https://your-domain.atlassian.net do not accept
	// 3LO tokens.
	APIGatewayURL = "https://api.atlassian.com"
)

// AccessibleResource is an Atlassian site an access token is authorized for.
type AccessibleResource struct {
	// ID is the site's cloud ID, used to build the API base URL.
	ID string `json:"id"`

	// URL is the site URL, for example https://your-domain.atlassian.net.
	URL string `json:"url"`

	// Name is the site name.
	Name string `json:"name"`

	// Scopes are the scopes granted for this site.
	Scopes []string `json:"scopes"`

	// AvatarURL is the site avatar.
	AvatarURL string `json:"avatarUrl"`
}

// FetchAccessibleResources calls Atlassian's accessible-resources endpoint.
//
// Pass an httpClient that already injects the token, or a plain client together
// with a non-empty accessToken.
func FetchAccessibleResources(ctx context.Context, httpClient *http.Client, accessToken string) ([]AccessibleResource, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, AccessibleResourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accessible resources request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read accessible resources response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accessible resources request returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var resources []AccessibleResource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("failed to decode accessible resources: %w", err)
	}

	return resources, nil
}

// JiraCloudBaseURL builds the Jira API base URL for a cloud ID.
func JiraCloudBaseURL(cloudID string) (*url.URL, error) {
	if cloudID == "" {
		return nil, fmt.Errorf("cloud ID is required")
	}

	u, err := url.Parse(APIGatewayURL + "/ex/jira/" + cloudID)
	if err != nil {
		return nil, fmt.Errorf("invalid cloud ID: %w", err)
	}

	return u, nil
}

// CloudIDResolverConfig configures a CloudIDResolver.
type CloudIDResolverConfig struct {
	// Authenticator authorizes the accessible-resources lookup. Required unless
	// CloudID is set, which skips the lookup entirely.
	Authenticator Authenticator

	// CloudID pins the site, skipping the lookup. Takes precedence over SiteURL.
	CloudID string

	// SiteURL selects which site to use when the token can reach several, for
	// example https://your-domain.atlassian.net. When empty and exactly one
	// site is accessible, that site is used; when several are accessible the
	// resolver reports an error rather than guessing.
	SiteURL string

	// HTTPClient is used for the lookup. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// CloudIDResolver resolves the Jira Cloud API base URL for an OAuth 2.0 (3LO)
// token by looking up the site's cloud ID, then caches the result.
//
// A 3LO access token is not accepted at a site URL such as
// https://your-domain.atlassian.net; requests must go to
// https://api.atlassian.com/ex/jira/{cloudID}. Since the cloud ID is only
// discoverable once a token exists, resolution happens on the first request
// rather than at client construction.
//
// It satisfies transport.BaseURLResolver and is safe for concurrent use.
type CloudIDResolver struct {
	authenticator Authenticator
	cloudID       string
	siteURL       string
	httpClient    *http.Client

	mu       sync.Mutex
	resolved *url.URL
}

// NewCloudIDResolver creates a cloud ID resolver.
func NewCloudIDResolver(config *CloudIDResolverConfig) (*CloudIDResolver, error) {
	if config == nil {
		return nil, fmt.Errorf("cloud ID resolver config is required")
	}
	if config.CloudID == "" && config.Authenticator == nil {
		return nil, fmt.Errorf("authenticator is required unless a cloud ID is supplied")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &CloudIDResolver{
		authenticator: config.Authenticator,
		cloudID:       config.CloudID,
		siteURL:       strings.TrimSuffix(config.SiteURL, "/"),
		httpClient:    httpClient,
	}, nil
}

// ResolveBaseURL returns the API base URL, performing the lookup once and
// caching the result for subsequent calls.
func (r *CloudIDResolver) ResolveBaseURL(ctx context.Context) (*url.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resolved != nil {
		return r.resolved, nil
	}

	cloudID := r.cloudID
	if cloudID == "" {
		resource, err := r.lookup(ctx)
		if err != nil {
			return nil, err
		}
		cloudID = resource.ID
	}

	base, err := JiraCloudBaseURL(cloudID)
	if err != nil {
		return nil, err
	}

	r.resolved = base
	return base, nil
}

// CloudID returns the resolved cloud ID, performing the lookup if needed.
func (r *CloudIDResolver) CloudID(ctx context.Context) (string, error) {
	base, err := r.ResolveBaseURL(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(base.Path, "/ex/jira/"), nil
}

// lookup fetches the accessible sites and picks the one to use.
func (r *CloudIDResolver) lookup(ctx context.Context) (*AccessibleResource, error) {
	resources, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}

	return selectResource(resources, r.siteURL)
}

// fetch authorizes and performs the accessible-resources request.
func (r *CloudIDResolver) fetch(ctx context.Context) ([]AccessibleResource, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, AccessibleResourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	if err := r.authenticator.Authenticate(req); err != nil {
		return nil, fmt.Errorf("failed to authenticate accessible resources request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accessible resources request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read accessible resources response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accessible resources request returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var resources []AccessibleResource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("failed to decode accessible resources: %w", err)
	}

	return resources, nil
}

// selectResource picks the site matching siteURL, or the only site available.
func selectResource(resources []AccessibleResource, siteURL string) (*AccessibleResource, error) {
	if len(resources) == 0 {
		return nil, fmt.Errorf("token grants access to no Atlassian sites; check the requested scopes and that the app is installed")
	}

	if siteURL != "" {
		for i := range resources {
			if strings.EqualFold(strings.TrimSuffix(resources[i].URL, "/"), siteURL) {
				return &resources[i], nil
			}
		}
		return nil, fmt.Errorf("token grants no access to %s; accessible sites: %s", siteURL, strings.Join(resourceURLs(resources), ", "))
	}

	if len(resources) > 1 {
		return nil, fmt.Errorf("token grants access to %d sites, so the site cannot be inferred; set the site URL or cloud ID explicitly. Accessible sites: %s",
			len(resources), strings.Join(resourceURLs(resources), ", "))
	}

	return &resources[0], nil
}

// resourceURLs lists site URLs for error messages.
func resourceURLs(resources []AccessibleResource) []string {
	urls := make([]string, 0, len(resources))
	for _, r := range resources {
		urls = append(urls, r.URL)
	}
	return urls
}
