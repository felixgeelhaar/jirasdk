// Package main demonstrates OAuth 2.0 (3LO) authentication against Jira Cloud.
//
// This is the three-legged flow: a user approves the requested scopes in their
// browser, and the resulting authorization code is exchanged for a token that
// acts on their behalf.
//
// Two details make Jira Cloud different from a textbook OAuth 2.0 provider, and
// the SDK handles both:
//
//  1. The authorization URL must carry audience=api.atlassian.com and
//     prompt=consent, and offline_access must be among the scopes for Atlassian
//     to issue a refresh token.
//
//  2. A 3LO token is not accepted at https://your-domain.atlassian.net.
//     Requests go to https://api.atlassian.com/ex/jira/{cloudID}, where the
//     cloud ID comes from the accessible-resources endpoint. The client looks
//     it up on the first request and caches the result.
//
// Run it with:
//
//	export JIRA_OAUTH_CLIENT_ID=...
//	export JIRA_OAUTH_CLIENT_SECRET=...
//	go run ./examples/oauth2
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/auth"
	"golang.org/x/oauth2"
)

const callbackAddr = "localhost:8080"

func main() {
	clientID := os.Getenv("JIRA_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_OAUTH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("JIRA_OAUTH_CLIENT_ID and JIRA_OAUTH_CLIENT_SECRET are required")
	}

	store := &fileTokenStore{path: filepath.Join(os.TempDir(), "jira-oauth-token.json")}

	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://" + callbackAddr + "/callback",
		Scopes:       []string{"read:jira-work", "read:jira-user"},

		// Tokens are written here after every exchange and refresh, so a
		// restart does not send the user through consent again.
		TokenStore: store,
	})

	ctx := context.Background()

	// Reuse a stored token when there is one; otherwise run the full flow.
	if _, err := oauth.LoadToken(); err != nil {
		if err := authorize(ctx, oauth); err != nil {
			log.Fatalf("Authorization failed: %v", err)
		}
	} else {
		fmt.Println("Reusing the stored token.")
	}

	// Every site this token can reach. The "id" of a site is its cloud ID.
	resources, err := oauth.AccessibleResources(ctx)
	if err != nil {
		log.Fatalf("Failed to list accessible resources: %v", err)
	}

	fmt.Println("\nAccessible sites:")
	for _, r := range resources {
		fmt.Printf("  %s  %s  (cloud ID %s)\n", r.Name, r.URL, r.ID)
	}

	// With one accessible site the client resolves the cloud ID itself. Pass
	// jira.WithCloudID(...) or jira.WithBaseURL(siteURL) to choose between
	// several.
	client, err := jira.NewClient(jira.WithOAuth2(oauth))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	user, err := client.Myself.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("\nAuthenticated as %s (%s)\n", user.DisplayName, user.EmailAddress)
}

// authorize runs the three-legged flow, serving the callback locally.
func authorize(ctx context.Context, oauth *auth.OAuth2Authenticator) error {
	state, err := randomState()
	if err != nil {
		return err
	}

	codes := make(chan string, 1)
	errs := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Comparing the state defends against CSRF: a code delivered without
		// the state we generated did not originate from our request.
		if query.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errs <- fmt.Errorf("state mismatch")
			return
		}

		if authErr := query.Get("error"); authErr != "" {
			http.Error(w, authErr, http.StatusBadRequest)
			errs <- fmt.Errorf("authorization denied: %s", authErr)
			return
		}

		fmt.Fprintln(w, "Authorization complete. You can close this tab.")
		codes <- query.Get("code")
	})

	listener, err := net.Listen("tcp", callbackAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", callbackAddr, err)
	}

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errs <- err
		}
	}()
	defer func() { _ = server.Shutdown(ctx) }()

	fmt.Println("Open this URL to authorize:")
	fmt.Println(" ", oauth.GetAuthURL(state))

	select {
	case code := <-codes:
		if _, err := oauth.Exchange(ctx, code); err != nil {
			return fmt.Errorf("failed to exchange code: %w", err)
		}
		fmt.Println("\nToken obtained and saved.")
		return nil

	case err := <-errs:
		return err

	case <-time.After(5 * time.Minute):
		return fmt.Errorf("timed out waiting for authorization")
	}
}

// randomState returns an unguessable state value for the authorization request.
func randomState() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}

// fileTokenStore persists tokens to disk, implementing auth.OAuth2TokenStore.
//
// A refresh token is a long-lived credential: store it somewhere encrypted in
// production, not a plain file.
type fileTokenStore struct {
	path string
}

func (s *fileTokenStore) SaveToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	return os.WriteFile(s.path, data, 0o600)
}

func (s *fileTokenStore) LoadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return &token, nil
}

func (s *fileTokenStore) DeleteToken() error {
	return os.Remove(s.path)
}
