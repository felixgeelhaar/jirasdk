package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/felixgeelhaar/jirasdk"
)

func main() {
	fmt.Println("=== Jira SDK - Environment Variable Configuration ===")
	fmt.Println()

	// Example 1: Automatic configuration from environment
	fmt.Println("=== Example 1: Automatic Configuration ===")
	fmt.Println()
	fmt.Println("Environment variables detected:")
	printEnvVars()
	fmt.Println()

	// Method 1: Using WithEnv() option
	fmt.Println("Method 1: Using WithEnv() option")
	client1, err := jira.NewClient(jira.WithEnv())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("✓ Client created successfully with base URL: %s\n", client1.BaseURL.String())
	fmt.Println()

	// Method 2: Using LoadConfigFromEnv() convenience function
	fmt.Println("Method 2: Using LoadConfigFromEnv() convenience function")
	client2, err := jira.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config from environment: %v", err)
	}
	fmt.Printf("✓ Client created successfully with base URL: %s\n", client2.BaseURL.String())
	fmt.Println()

	// Example 2: Combining environment variables with explicit options
	fmt.Println("=== Example 2: Combining Environment with Options ===")
	fmt.Println()
	fmt.Println("Loading base config from environment, overriding timeout...")

	client3, err := jira.NewClient(
		jira.WithEnv(), // Load from environment
		// Override specific settings
		jira.WithTimeout(60),   // Override timeout to 60 seconds
		jira.WithMaxRetries(5), // Override max retries
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("✓ Client created with custom overrides (timeout: %v)\n", client3.HTTPClient.Timeout)
	fmt.Println()

	// Example 3: Make an actual API request
	fmt.Println("=== Example 3: Making API Request ===")
	fmt.Println()

	ctx := context.Background()

	// Get current user
	user, err := client1.User.GetMyself(ctx)
	if err != nil {
		fmt.Printf("⚠ API request failed (this is expected if credentials are demo): %v\n", err)
	} else {
		fmt.Printf("✓ Successfully authenticated as: %s\n", user.DisplayName)
		fmt.Printf("  Email: %s\n", user.EmailAddress)
		fmt.Printf("  Account ID: %s\n", user.AccountID)
	}
	fmt.Println()

	// Example 4: Environment variable reference
	fmt.Println("=== Example 4: Environment Variable Reference ===")
	fmt.Println()
	printEnvReference()
	fmt.Println()

	fmt.Println("=== Environment Configuration Complete ===")
}

// printEnvVars displays currently set jira environment variables
func printEnvVars() {
	envVars := map[string]string{
		"JIRA_BASE_URL":  os.Getenv(jira.EnvBaseURL),
		"JIRA_EMAIL":     os.Getenv(jira.EnvEmail),
		"JIRA_API_TOKEN": maskValue(os.Getenv(jira.EnvAPIToken)),
		"JIRA_PAT":       maskValue(os.Getenv(jira.EnvPAT)),
		"JIRA_USERNAME":  os.Getenv(jira.EnvUsername),
		"JIRA_PASSWORD":  maskValue(os.Getenv(jira.EnvPassword)),
		"JIRA_TIMEOUT":   os.Getenv(jira.EnvTimeout),
	}

	for key, value := range envVars {
		if value != "" {
			fmt.Printf("  %s=%s\n", key, value)
		}
	}
}

// maskValue masks sensitive values for display
func maskValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + "****"
}

// printEnvReference prints a reference guide for environment variables
func printEnvReference() {
	fmt.Println("Required Environment Variables:")
	fmt.Println("  JIRA_BASE_URL - Your Jira instance URL")
	fmt.Println("                  Example: https://your-domain.atlassian.net")
	fmt.Println()

	fmt.Println("Authentication (choose one):")
	fmt.Println()
	fmt.Println("  Option 1: API Token (Jira Cloud - Recommended)")
	fmt.Println("    JIRA_EMAIL     - Your Jira email address")
	fmt.Println("    JIRA_API_TOKEN - Your API token")
	fmt.Println("    Generate token: https://id.atlassian.com/manage-profile/security/api-tokens")
	fmt.Println()
	fmt.Println("  Option 2: Personal Access Token (Jira Server/Data Center)")
	fmt.Println("    JIRA_PAT - Your personal access token")
	fmt.Println()
	fmt.Println("  Option 3: Basic Auth (Legacy, not recommended)")
	fmt.Println("    JIRA_USERNAME - Your Jira username")
	fmt.Println("    JIRA_PASSWORD - Your Jira password")
	fmt.Println()
	fmt.Println("  Option 4: OAuth 2.0")
	fmt.Println("    JIRA_OAUTH_CLIENT_ID     - OAuth client ID")
	fmt.Println("    JIRA_OAUTH_CLIENT_SECRET - OAuth client secret")
	fmt.Println("    JIRA_OAUTH_REDIRECT_URL  - OAuth redirect URL")
	fmt.Println()

	fmt.Println("Optional Configuration:")
	fmt.Println("  JIRA_TIMEOUT            - HTTP timeout in seconds (default: 30)")
	fmt.Println("  JIRA_MAX_RETRIES        - Maximum retry attempts (default: 3)")
	fmt.Println("  JIRA_RATE_LIMIT_BUFFER  - Rate limit buffer in seconds (default: 5)")
	fmt.Println("  JIRA_USER_AGENT         - Custom user agent string")
	fmt.Println()

	fmt.Println("Example Setup (Jira Cloud):")
	fmt.Println("  export JIRA_BASE_URL=\"https://your-domain.atlassian.net\"")
	fmt.Println("  export JIRA_EMAIL=\"user@example.com\"")
	fmt.Println("  export JIRA_API_TOKEN=\"your-api-token\"")
	fmt.Println()

	fmt.Println("Example Setup (Jira Server):")
	fmt.Println("  export JIRA_BASE_URL=\"https://jira.company.com\"")
	fmt.Println("  export JIRA_PAT=\"your-personal-access-token\"")
}
