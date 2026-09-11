// Package main demonstrates the OAuth 2.0 client credentials grant.
//
// Atlassian's own Jira REST API does not offer this grant. It applies when Jira
// sits behind an API gateway or an in-house identity provider that issues
// tokens Jira accepts — a common arrangement for service-to-service access in
// an enterprise network.
//
// For Jira Cloud, reach for the Connect flows instead (see examples/connectjwt);
// for Server and Data Center, OAuth 1.0a (see examples/oauth1).
//
// Run it with:
//
//	export JIRA_BASE_URL=https://jira.internal
//	export JIRA_OAUTH_CLIENT_ID=service-account
//	export JIRA_OAUTH_CLIENT_SECRET=...
//	export JIRA_OAUTH_TOKEN_URL=https://idp.internal/oauth2/token
//	go run ./examples/clientcredentials
package main

import (
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
	clientID := os.Getenv("JIRA_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_OAUTH_CLIENT_SECRET")
	tokenURL := os.Getenv("JIRA_OAUTH_TOKEN_URL")

	if baseURL == "" || clientID == "" || clientSecret == "" || tokenURL == "" {
		log.Fatal("JIRA_BASE_URL, JIRA_OAUTH_CLIENT_ID, JIRA_OAUTH_CLIENT_SECRET and JIRA_OAUTH_TOKEN_URL are required")
	}

	ctx := context.Background()

	authenticator, err := auth.NewClientCredentialsAuth(&auth.ClientCredentialsConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       strings.Fields(os.Getenv("JIRA_OAUTH_SCOPES")),

		// Several authorization servers need an audience to target a specific
		// API; omit it when yours does not.
		Audience: os.Getenv("JIRA_OAUTH_AUDIENCE"),
	})
	if err != nil {
		log.Fatalf("Failed to create authenticator: %v", err)
	}

	// Tokens are fetched on first use and renewed before expiry, so this is only
	// here to show the grant working on its own.
	token, err := authenticator.Token(ctx)
	if err != nil {
		log.Fatalf("Failed to obtain token: %v", err)
	}

	fmt.Printf("Token type: %s\n", token.Type())
	if !token.Expiry.IsZero() {
		fmt.Printf("Expires:    %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	}

	client, err := jira.NewClient(
		jira.WithBaseURL(baseURL),
		jira.WithAuthenticator(authenticator),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	info, err := client.ServerInfo.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get server info: %v", err)
	}

	fmt.Printf("\nConnected to Jira %s (%s)\n", info.Version, info.DeploymentType)
}
