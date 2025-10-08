# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The jirasdk team takes security seriously. We appreciate your efforts to responsibly disclose your findings.

### Where to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by emailing:

**security@felixgeelhaar.com**

### What to Include

To help us better understand the nature and scope of the potential issue, please include as much of the following information as possible:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- **Initial Response**: We will acknowledge receipt of your vulnerability report within 48 hours.
- **Communication**: We will keep you informed of the progress towards a fix and full announcement.
- **Timeline**: We aim to address critical vulnerabilities within 7 days and less critical ones within 30 days.
- **Credit**: We will credit you in the security advisory unless you prefer to remain anonymous.

## Security Best Practices

When using jirasdk in your applications, we recommend following these security best practices:

### 1. Credential Management

**Never hardcode credentials:**
```go
// ❌ DON'T: Hardcode credentials
client, _ := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithAPIToken("user@example.com", "hardcoded-token"), // DON'T DO THIS
)

// ✅ DO: Use environment variables
client, _ := jira.LoadConfigFromEnv()
```

**Use environment variables or secure secret management:**
```bash
export JIRA_BASE_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="user@example.com"
export JIRA_API_TOKEN="your-api-token"
```

For production systems, use dedicated secret management solutions:
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Google Cloud Secret Manager

### 2. API Token Security

**Recommended Authentication Methods (in order of preference):**

1. **API Tokens** (Jira Cloud) - Generate from [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. **Personal Access Tokens** (Jira Server/Data Center) - More secure than username/password
3. **OAuth 2.0** - Best for third-party integrations
4. **Basic Authentication** - Legacy, not recommended for production

**Token Rotation:**
- Rotate API tokens regularly (every 90 days recommended)
- Revoke tokens immediately when compromised
- Use different tokens for different environments

### 3. Transport Security

**Always use HTTPS:**
```go
// ✅ DO: Use HTTPS
client, _ := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"), // HTTPS
    jira.WithAPIToken(email, token),
)

// ❌ DON'T: Use HTTP for production
// HTTP should only be used for local development/testing
```

**Configure appropriate timeouts:**
```go
client, _ := jira.NewClient(
    jira.WithBaseURL(baseURL),
    jira.WithAPIToken(email, token),
    jira.WithTimeout(30*time.Second), // Prevent hanging connections
)
```

### 4. Input Validation

Always validate and sanitize user input before passing to API methods:

```go
// Validate issue keys
if !isValidIssueKey(issueKey) {
    return errors.New("invalid issue key format")
}

// Validate JQL queries to prevent injection
query := sanitizeJQL(userInput)
```

### 5. Error Handling

**Don't expose sensitive information in errors:**
```go
// ❌ DON'T: Expose credentials in errors
if err != nil {
    log.Printf("Failed with token %s: %v", token, err) // DON'T DO THIS
}

// ✅ DO: Log errors without credentials
if err != nil {
    log.Printf("API request failed: %v", err)
}
```

### 6. Rate Limiting

Use built-in rate limiting to prevent abuse:
```go
client, _ := jira.NewClient(
    jira.WithBaseURL(baseURL),
    jira.WithAPIToken(email, token),
    jira.WithRateLimitBuffer(5*time.Second), // Built-in rate limiting
)
```

### 7. Dependency Security

Keep dependencies up to date:
```bash
# Check for vulnerabilities
go list -json -m all | docker run --rm -i sonatypeoss/nancy:latest sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

### 8. Minimal Permissions

**Grant minimum necessary permissions:**
- Use API tokens with limited scopes when possible
- Create service accounts with only required project access
- Regularly audit and revoke unnecessary permissions

### 9. Logging Security

**Configure secure logging:**
```go
import "github.com/felixgeelhaar/jirasdk/logger/bolt"

logger := bolt.NewLogger()
client, _ := jira.NewClient(
    jira.WithBaseURL(baseURL),
    jira.WithAPIToken(email, token),
    jira.WithLogger(logger), // Structured logging without credentials
)
```

**Never log credentials or sensitive data:**
- API tokens
- Passwords
- OAuth secrets
- Personal user information

### 10. Context and Cancellation

Use context for timeout and cancellation to prevent resource exhaustion:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

issue, err := client.Issue.Get(ctx, "PROJ-123")
```

## Known Security Considerations

### 1. Third-Party Dependencies

jirasdk depends on the following security-critical packages:
- `golang.org/x/oauth2` - OAuth 2.0 implementation
- `github.com/felixgeelhaar/fortify` - Resilience patterns

We monitor these dependencies for security updates and will release patches as needed.

### 2. TLS/SSL Verification

By default, the library enforces TLS certificate verification. Do not disable this in production:

```go
// ❌ NEVER do this in production
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // DANGEROUS
}
httpClient := &http.Client{Transport: tr}

client, _ := jira.NewClient(
    jira.WithHTTPClient(httpClient), // Don't use unverified TLS
    // ...
)
```

### 3. Memory Safety

All sensitive data (tokens, passwords) should be cleared from memory when no longer needed. While Go doesn't provide direct memory zeroing, we recommend:

```go
// Clear sensitive strings after use where possible
defer func() {
    token = ""
}()
```

## Security Updates

Subscribe to security updates:
- Watch this repository for security advisories
- Enable GitHub security alerts for your projects using jirasdk
- Review the [CHANGELOG](CHANGELOG.md) for security-related updates

## Acknowledgments

We thank the security researchers and contributors who help keep jirasdk secure:

- [Contributors will be listed here]

## Contact

For any security questions or concerns, please contact: security@felixgeelhaar.com

---

*Last updated: 2025-01-08*
