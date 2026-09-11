// Package main demonstrates OAuth 1.0a authentication against Jira Server and
// Data Center, in both the two-legged and three-legged forms.
//
// Setup, once per instance, in Jira under Settings > Applications >
// Application links:
//
//	openssl genrsa -out jira.pem 2048
//	openssl rsa -in jira.pem -pubout -out jira.pub
//
// Create an incoming application link with a consumer key of your choosing and
// the contents of jira.pub as the public key. For two-legged impersonation,
// also tick "Allow 2-legged OAuth with user impersonation".
//
// Run it with:
//
//	export JIRA_BASE_URL=https://jira.example.com
//	export JIRA_OAUTH1_CONSUMER_KEY=my-consumer-key
//	export JIRA_OAUTH1_PRIVATE_KEY_FILE=jira.pem
//	go run ./examples/oauth1
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/auth"
)

func main() {
	baseURL := os.Getenv("JIRA_BASE_URL")
	consumerKey := os.Getenv("JIRA_OAUTH1_CONSUMER_KEY")
	keyFile := os.Getenv("JIRA_OAUTH1_PRIVATE_KEY_FILE")

	if baseURL == "" || consumerKey == "" || keyFile == "" {
		log.Fatal("JIRA_BASE_URL, JIRA_OAUTH1_CONSUMER_KEY and JIRA_OAUTH1_PRIVATE_KEY_FILE are required")
	}

	privateKey, err := os.ReadFile(keyFile) //nolint:gosec // G304: operator-supplied path
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}

	ctx := context.Background()

	twoLegged(ctx, baseURL, consumerKey, privateKey)

	if impersonate := os.Getenv("JIRA_OAUTH1_IMPERSONATE_USER"); impersonate != "" {
		twoLeggedImpersonating(ctx, baseURL, consumerKey, privateKey, impersonate)
	} else {
		fmt.Println("\nSet JIRA_OAUTH1_IMPERSONATE_USER to see two-legged impersonation.")
	}

	if os.Getenv("JIRA_OAUTH1_INTERACTIVE") != "" {
		threeLegged(ctx, baseURL, consumerKey, privateKey)
	} else {
		fmt.Println("Set JIRA_OAUTH1_INTERACTIVE=1 to run the three-legged handshake.")
	}
}

// twoLegged signs each request with the consumer's own key, with no user
// involved. The consumer acts as itself.
func twoLegged(ctx context.Context, baseURL, consumerKey string, privateKey []byte) {
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:   consumerKey,
			PrivateKeyPEM: privateKey,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== OAuth 1.0a two-legged ===")

	info, err := client.ServerInfo.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get server info: %v", err)
	}

	fmt.Printf("Connected to Jira %s (%s)\n", info.Version, info.DeploymentType)
}

// twoLeggedImpersonating adds the user_id parameter Jira Server reads to act on
// a user's behalf. The application link must permit it.
func twoLeggedImpersonating(ctx context.Context, baseURL, consumerKey string, privateKey []byte, username string) {
	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithOAuth1(&auth.OAuth1Config{
			ConsumerKey:     consumerKey,
			PrivateKeyPEM:   privateKey,
			ImpersonateUser: username,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("\n=== OAuth 1.0a two-legged with impersonation ===")

	user, err := client.Myself.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("Acting as %s (%s)\n", user.DisplayName, user.EmailAddress)
}

// threeLegged walks the full handshake: temporary credential, user approval,
// then a long-lived access token.
func threeLegged(ctx context.Context, baseURL, consumerKey string, privateKey []byte) {
	authenticator, err := auth.NewOAuth1Auth(&auth.OAuth1Config{
		ConsumerKey:   consumerKey,
		PrivateKeyPEM: privateKey,
		BaseURL:       baseURL,
	})
	if err != nil {
		log.Fatalf("Failed to create authenticator: %v", err)
	}

	fmt.Println("\n=== OAuth 1.0a three-legged ===")

	// Step 1: a temporary credential. "oob" means the user reads the verifier
	// off the screen rather than being redirected.
	requestToken, err := authenticator.RequestToken(ctx, "oob")
	if err != nil {
		log.Fatalf("Failed to get request token: %v", err)
	}

	// Step 2: the user approves it.
	authURL, err := authenticator.AuthorizationURL(requestToken.Token)
	if err != nil {
		log.Fatalf("Failed to build authorization URL: %v", err)
	}

	fmt.Println("Open this URL and approve the request:")
	fmt.Println(" ", authURL)
	fmt.Print("\nEnter the verification code: ")

	verifier, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read verifier: %v", err)
	}

	// Step 3: trade it for an access token. This also installs the token on the
	// authenticator, so it signs three-legged from here on.
	token, secret, err := authenticator.AccessToken(ctx, requestToken, strings.TrimSpace(verifier))
	if err != nil {
		log.Fatalf("Failed to exchange request token: %v", err)
	}

	fmt.Printf("\nAccess token obtained. Store these to skip the handshake next time:\n")
	fmt.Printf("  JIRA_OAUTH1_TOKEN=%s\n", token)
	fmt.Printf("  JIRA_OAUTH1_TOKEN_SECRET=%s\n", secret)

	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAuthenticator(authenticator),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	user, err := client.Myself.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("\nAuthenticated as %s\n", user.DisplayName)
}
