// Package main demonstrates OAuth 2.0 authentication with the jira-connect library.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jira "github.com/felixgeelhaar/jirasdk"
	"github.com/felixgeelhaar/jirasdk/auth"
)

func main() {
	// OAuth 2.0 Configuration
	oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
		ClientID:     os.Getenv("JIRA_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("JIRA_OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work", "write:jira-work", "read:jira-user"},
	})

	// Step 1: Get authorization URL
	fmt.Println("=== OAuth 2.0 Authorization Flow ===")
	state := "random-state-string" // In production, generate a cryptographically secure random string
	authURL := oauth.GetAuthURL(state)

	fmt.Println("\n1. Visit this URL to authorize the application:")
	fmt.Println(authURL)
	fmt.Println("\n2. After authorization, you'll be redirected to your callback URL")
	fmt.Println("3. Extract the 'code' parameter from the callback URL")

	// Step 2: Exchange authorization code for access token
	// In a real application, you would:
	// 1. Start a local HTTP server to receive the callback
	// 2. Extract the code from the callback URL
	// 3. Exchange the code for an access token

	// Simulate receiving the authorization code
	var authCode string
	fmt.Print("\nEnter the authorization code: ")
	fmt.Scanln(&authCode)

	if authCode == "" {
		log.Fatal("Authorization code is required")
	}

	ctx := context.Background()

	// Exchange code for token
	fmt.Println("\n=== Exchanging Code for Token ===")
	token, err := oauth.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Failed to exchange code: %v", err)
	}

	fmt.Printf("Access Token: %s...\n", token.AccessToken[:20])
	fmt.Printf("Refresh Token: %s...\n", token.RefreshToken[:20])
	fmt.Printf("Token Type: %s\n", token.TokenType)
	fmt.Printf("Expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))

	// Step 3: Create Jira client with OAuth 2.0
	fmt.Println("\n=== Creating Jira Client ===")
	client, err := jira.NewClient(
		jira.WithBaseURL(os.Getenv("JIRA_BASE_URL")),
		jira.WithOAuth2(oauth),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Step 4: Make API requests
	fmt.Println("\n=== Making API Request ===")
	user, err := client.User.Get(ctx, "currentUser", nil)
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.DisplayName, user.EmailAddress)

	// Step 5: Token refresh (automatic)
	fmt.Println("\n=== Automatic Token Refresh ===")
	fmt.Println("The OAuth2 authenticator automatically refreshes the token when it expires")
	fmt.Println("You don't need to handle token refresh manually")

	// Example: Manual token refresh
	if token.RefreshToken != "" {
		fmt.Println("\n=== Manual Token Refresh ===")
		newToken, err := oauth.RefreshToken(ctx)
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
		} else {
			fmt.Printf("New Access Token: %s...\n", newToken.AccessToken[:20])
			fmt.Printf("Expires: %s\n", newToken.Expiry.Format("2006-01-02 15:04:05"))
		}
	}

	// Step 6: Token storage (for persistent authentication)
	fmt.Println("\n=== Token Storage ===")
	fmt.Println("In a production application, you should:")
	fmt.Println("1. Implement the OAuth2TokenStore interface")
	fmt.Println("2. Save tokens securely (encrypted)")
	fmt.Println("3. Load tokens on application startup")
	fmt.Println("4. Refresh tokens as needed")

	// Example token storage implementation
	exampleTokenStorage()

	fmt.Println("\n=== OAuth 2.0 Example Complete ===")
}

// Example implementation of a simple file-based token store
func exampleTokenStorage() {
	fmt.Println("\nExample: File-based Token Storage")
	fmt.Println("```go")
	fmt.Println(`type FileTokenStore struct {
    filepath string
}

func (s *FileTokenStore) SaveToken(token *oauth2.Token) error {
    data, err := json.Marshal(token)
    if err != nil {
        return err
    }
    return os.WriteFile(s.filepath, data, 0600)
}

func (s *FileTokenStore) LoadToken() (*oauth2.Token, error) {
    data, err := os.ReadFile(s.filepath)
    if err != nil {
        return nil, err
    }
    var token oauth2.Token
    err = json.Unmarshal(data, &token)
    return &token, err
}

func (s *FileTokenStore) DeleteToken() error {
    return os.Remove(s.filepath)
}`)
	fmt.Println("```")
}

// Example: OAuth 2.0 callback server
func startCallbackServer() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Extract code and state from query parameters
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		// Verify state to prevent CSRF attacks
		// expectedState := getStoredState() // Retrieve the state you generated earlier
		// if state != expectedState {
		//     http.Error(w, "Invalid state", http.StatusBadRequest)
		//     return
		// }

		fmt.Fprintf(w, "Authorization successful! Code: %s, State: %s", code, state)
		// Sanitize user input before logging to prevent log injection
		sanitizedCode := sanitizeForLog(code)
		fmt.Printf("Received code: %s\n", sanitizedCode)

		// Here you would exchange the code for a token
		// and store it for future use
	})

	fmt.Println("Starting callback server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// sanitizeForLog removes newline characters from strings to prevent log injection attacks.
func sanitizeForLog(s string) string {
	return strings.NewReplacer("\n", "", "\r", "").Replace(s)
}
