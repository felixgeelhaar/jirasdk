// Package main demonstrates the two-legged flows available to an Atlassian
// Connect app on Jira Cloud.
//
// A Connect app receives a shared secret in the "installed" lifecycle callback.
// That secret authenticates the app with no user authorization step, in two
// ways:
//
//   - Connect JWT: the app acts as itself, under the scopes declared in its
//     descriptor. Each request carries a JWT whose "qsh" claim binds it to that
//     exact method, path and query string.
//
//   - JWT bearer grant: the app exchanges a signed assertion for an access
//     token that acts as a named user, without that user completing a redirect.
//
// Run it with:
//
//	export JIRA_BASE_URL=https://your-domain.atlassian.net
//	export JIRA_CONNECT_APP_KEY=com.example.my-app
//	export JIRA_CONNECT_SHARED_SECRET=...
//	go run ./examples/connectjwt
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/auth"
)

func main() {
	baseURL := os.Getenv("JIRA_BASE_URL")
	appKey := os.Getenv("JIRA_CONNECT_APP_KEY")
	sharedSecret := os.Getenv("JIRA_CONNECT_SHARED_SECRET")

	if baseURL == "" || appKey == "" || sharedSecret == "" {
		log.Fatal("JIRA_BASE_URL, JIRA_CONNECT_APP_KEY and JIRA_CONNECT_SHARED_SECRET are required")
	}

	ctx := context.Background()

	asTheApp(ctx, baseURL, appKey, sharedSecret)
	showQueryStringHash(appKey, sharedSecret)

	// Impersonation needs the oauthClientId from the install payload, which is
	// a different value from the app key.
	if accountID, oauthClientID := os.Getenv("JIRA_CONNECT_ACCOUNT_ID"), os.Getenv("JIRA_CONNECT_OAUTH_CLIENT_ID"); accountID != "" && oauthClientID != "" {
		asAUser(ctx, baseURL, oauthClientID, sharedSecret, accountID)
	} else {
		fmt.Println("\nSet JIRA_CONNECT_OAUTH_CLIENT_ID and JIRA_CONNECT_ACCOUNT_ID to see the JWT bearer grant.")
	}
}

// asTheApp calls Jira with the app's own identity and scopes.
func asTheApp(ctx context.Context, baseURL, appKey, sharedSecret string) {
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithConnectJWT(&auth.ConnectJWTConfig{
			AppKey:       appKey,
			SharedSecret: sharedSecret,

			// Set ContextPath when the instance is served under a path, for
			// example "/jira". It is stripped before the qsh is computed.
			ContextPath: os.Getenv("JIRA_CONTEXT_PATH"),
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create Connect JWT client: %v", err)
	}

	fmt.Println("=== Connect JWT: acting as the app ===")

	projects, err := client.Project.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list projects: %v", err)
	}

	fmt.Printf("The app can see %d project(s)\n", len(projects))
	for _, p := range projects {
		fmt.Printf("  %s  %s\n", p.Key, p.Name)
	}
}

// asAUser exchanges a signed assertion for a token scoped to one user.
func asAUser(ctx context.Context, baseURL, oauthClientID, sharedSecret, accountID string) {
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithJWTBearer(&auth.JWTBearerConfig{
			OAuthClientID: oauthClientID,
			SharedSecret:  sharedSecret,
			AccountID:     accountID,
			SiteURL:       baseURL,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create JWT bearer client: %v", err)
	}

	fmt.Println("\n=== JWT bearer grant: acting as a user ===")

	user, err := client.Myself.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	// The token acts as the impersonated user, so this reports them rather
	// than the app.
	fmt.Printf("Acting as %s (%s)\n", user.DisplayName, user.AccountID)
}

// showQueryStringHash illustrates what the qsh claim commits to. Two requests
// that differ in any way produce different hashes, so a token minted for one
// cannot be replayed against the other.
func showQueryStringHash(appKey, sharedSecret string) {
	authenticator, err := auth.NewConnectJWTAuth(&auth.ConnectJWTConfig{
		AppKey:       appKey,
		SharedSecret: sharedSecret,
	})
	if err != nil {
		log.Fatalf("Failed to create authenticator: %v", err)
	}

	fmt.Println("\n=== Query string hash ===")

	for _, target := range []string{
		"https://example.atlassian.net/rest/api/3/myself",
		"https://example.atlassian.net/rest/api/3/search?jql=project%3DTEST",
		"https://example.atlassian.net/rest/api/3/search?jql=project%3DPROD",
	} {
		req, err := http.NewRequest(http.MethodGet, target, nil)
		if err != nil {
			log.Fatalf("Failed to build request: %v", err)
		}

		canonical, err := auth.CanonicalRequest(req, "")
		if err != nil {
			log.Fatalf("Failed to build canonical request: %v", err)
		}

		qsh, err := auth.QueryStringHash(req, "")
		if err != nil {
			log.Fatalf("Failed to compute qsh: %v", err)
		}

		fmt.Printf("  %-52s -> %s\n", canonical, qsh[:16]+"...")
	}

	// Sign returns the token without sending anything, which is useful for
	// requests this SDK does not model.
	req, err := http.NewRequest(http.MethodGet, "https://example.atlassian.net/rest/api/3/myself", nil)
	if err != nil {
		log.Fatalf("Failed to build request: %v", err)
	}

	token, err := authenticator.Sign(req)
	if err != nil {
		log.Fatalf("Failed to sign: %v", err)
	}

	fmt.Printf("\nAuthorization: JWT %s...\n", token[:32])
}
