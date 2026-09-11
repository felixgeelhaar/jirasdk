# Authentication

Pick the row that matches your deployment and what the caller should act as.

| Deployment | Acts as | Method | Option |
| --- | --- | --- | --- |
| Cloud | A user (their token) | API token | `WithAPIToken` |
| Cloud | A user (consented) | OAuth 2.0 (3LO) | `WithOAuth2` |
| Cloud | The app itself | Connect JWT (2LO) | `WithConnectJWT` |
| Cloud | A user, no redirect | JWT bearer grant (2LO) | `WithJWTBearer` |
| Server / Data Center | A user (their token) | Personal access token | `WithPAT` |
| Server / Data Center | The consumer itself | OAuth 1.0a (2LO) | `WithOAuth1` |
| Server / Data Center | A user (consented) | OAuth 1.0a (3LO) | `WithOAuth1` + handshake |
| Behind a gateway / custom IdP | A service account | OAuth 2.0 client credentials | `WithClientCredentials` |

"2LO" (two-legged) means no user approves anything: the caller proves who it is
with a key or secret it already holds. "3LO" (three-legged) adds the user as a
third party who consents in a browser.

---

## Jira Cloud

### API token

The simplest option, and the right default for scripts and internal tools.

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithAPIToken("user@example.com", "your-api-token"),
)
```

Generate a token at <https://id.atlassian.com/manage-profile/security/api-tokens>.

### OAuth 2.0 (3LO)

For an app acting on behalf of users who grant it access. See
[`examples/oauth2`](../examples/oauth2) for a runnable end-to-end flow.

```go
oauth := auth.NewOAuth2Authenticator(&auth.OAuth2Config{
    ClientID:     os.Getenv("JIRA_OAUTH_CLIENT_ID"),
    ClientSecret: os.Getenv("JIRA_OAUTH_CLIENT_SECRET"),
    RedirectURL:  "https://app.example.com/callback",
    Scopes:       []string{"read:jira-work", "write:jira-work"},
    TokenStore:   myTokenStore, // optional, see "Persisting tokens"
})

// Send the user here. The URL carries audience=api.atlassian.com and
// prompt=consent, both of which Atlassian requires.
authURL := oauth.GetAuthURL(state)

// Then exchange the code your callback receives.
token, err := oauth.Exchange(ctx, code)

client, err := jira.NewClient(jira.WithOAuth2(oauth))
```

Two things about Jira Cloud 3LO are easy to get wrong, and the SDK handles both:

**`offline_access` is a scope, not a parameter.** Atlassian issues a refresh
token only when `offline_access` is among the requested scopes — unlike Google,
where `access_type=offline` does the job. The SDK appends it automatically; set
`DisableOfflineAccess: true` to opt out.

**A 3LO token does not work against your site URL.** Requests must go to
`https://api.atlassian.com/ex/jira/{cloudID}`, never
`https://your-domain.atlassian.net`. The cloud ID comes from Atlassian's
accessible-resources endpoint, which needs a token — so it cannot be known when
the client is built.

The client therefore looks it up on the first request and caches the result:

```go
// One accessible site: resolved automatically, no base URL needed.
client, err := jira.NewClient(jira.WithOAuth2(oauth))

// Several accessible sites: name the one you want.
client, err := jira.NewClient(
    jira.WithOAuth2(oauth),
    jira.WithBaseURL("https://your-domain.atlassian.net"),
)

// Or skip the lookup entirely.
client, err := jira.NewClient(
    jira.WithOAuth2(oauth),
    jira.WithCloudID("1324a887-45db-1bf4-1e99-ef0ff456d421"),
)
```

`WithBaseURL` here selects *which* site to use; requests still travel through
the API gateway. To inspect what a token can reach:

```go
resources, err := oauth.AccessibleResources(ctx)
for _, r := range resources {
    fmt.Printf("%s %s (cloud ID %s)\n", r.Name, r.URL, r.ID)
}
```

If a proxy in front of Jira accepts 3LO tokens directly, disable resolution with
`jira.WithoutCloudIDResolution()`.

### Connect JWT (2LO)

For an Atlassian Connect app calling Jira as itself, under the scopes in its
descriptor. The shared secret arrives in the `installed` lifecycle callback.

```go
client, err := jira.NewClient(
    jira.WithBaseURL(install.BaseURL),
    jira.WithConnectJWT(&auth.ConnectJWTConfig{
        AppKey:       "com.example.my-app",
        SharedSecret: install.SharedSecret,
    }),
)
```

Every request carries a freshly minted JWT whose `qsh` claim is a SHA-256 hash
of the canonical request — method, path and sorted query string. That binds the
token to one specific call, so a captured token cannot be replayed against a
different endpoint. Set `Subject` to an account ID to associate requests with a
user, and `ContextPath` (for example `/jira`) when the instance is served under
a path, so it is stripped before the hash is computed.

`auth.CanonicalRequest` and `auth.QueryStringHash` are exported if you need to
compute a hash yourself, and `(*ConnectJWTAuth).Sign` returns a token for a
request without sending it.

### JWT bearer grant (2LO impersonation)

For an installed Connect app acting as a named user, with no redirect. The app
signs an assertion with its shared secret and exchanges it for a short-lived
access token; the SDK caches and renews tokens automatically.

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://your-domain.atlassian.net"),
    jira.WithJWTBearer(&auth.JWTBearerConfig{
        OAuthClientID: install.OAuthClientID, // not the app key
        SharedSecret:  install.SharedSecret,
        AccountID:     "5b10ac8d82e05b22cc7d4ef5",
        SiteURL:       "https://your-domain.atlassian.net",
    }),
)
```

`OAuthClientID` is the `oauthClientId` field of the install payload, which is a
different value from `AppKey`. Mixing them up is the usual cause of
`invalid_grant`.

See [`examples/connectjwt`](../examples/connectjwt) for both Connect flows.

---

## Jira Server and Data Center

### Personal access token

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://jira.example.com"),
    jira.WithPAT("your-personal-access-token"),
)
```

### OAuth 1.0a

Jira Server authenticates apps with RSA-SHA1-signed OAuth 1.0a over an
application link. Generate a key pair and register the public half:

```bash
openssl genrsa -out jira.pem 2048
openssl rsa -in jira.pem -pubout -out jira.pub
```

In Jira: **Settings → Applications → Application links**, create an incoming
link with a consumer key of your choosing and the contents of `jira.pub`.

**Two-legged** — the consumer signs with its own key and acts as itself:

```go
privateKey, err := os.ReadFile("jira.pem")

client, err := jira.NewClient(
    jira.WithBaseURL("https://jira.example.com"),
    jira.WithOAuth1(&auth.OAuth1Config{
        ConsumerKey:   "my-consumer-key",
        PrivateKeyPEM: privateKey,
    }),
)
```

Add `ImpersonateUser: "jsmith"` to act on a user's behalf. This sends a signed
`user_id` parameter, and requires "Allow 2-legged OAuth with user
impersonation" on the application link.

**Three-legged** — the user approves access in a browser:

```go
authenticator, err := auth.NewOAuth1Auth(&auth.OAuth1Config{
    ConsumerKey:   "my-consumer-key",
    PrivateKeyPEM: privateKey,
    BaseURL:       "https://jira.example.com",
})

requestToken, err := authenticator.RequestToken(ctx, "https://app.example.com/callback")
authURL, err := authenticator.AuthorizationURL(requestToken.Token)
// ... user approves, you receive oauth_verifier ...
token, secret, err := authenticator.AccessToken(ctx, requestToken, verifier)

// AccessToken installs the token, so the authenticator now signs three-legged.
client, err := jira.NewClient(
    jira.WithBaseURL("https://jira.example.com"),
    jira.WithAuthenticator(authenticator),
)
```

Store `token` and `secret` and pass them as `Token` / `TokenSecret` next time to
skip the handshake. `HMAC-SHA1` is available via `SignatureMethod` for setups
that use a shared secret instead of a key pair.

See [`examples/oauth1`](../examples/oauth1).

---

## Behind a gateway or custom IdP

Atlassian's own Jira REST API does not offer the client credentials grant. Use
this when Jira sits behind an API gateway or in-house identity provider that
issues tokens Jira accepts.

```go
client, err := jira.NewClient(
    jira.WithBaseURL("https://jira.internal"),
    jira.WithClientCredentials(&auth.ClientCredentialsConfig{
        ClientID:     "service-account",
        ClientSecret: os.Getenv("CLIENT_SECRET"),
        TokenURL:     "https://idp.internal/oauth2/token",
        Scopes:       []string{"jira:read", "jira:write"},
        Audience:     "https://jira.internal", // when your IdP requires one
    }),
)
```

See [`examples/clientcredentials`](../examples/clientcredentials).

---

## Persisting tokens

Implement `auth.OAuth2TokenStore` and pass it as `TokenStore`. The SDK writes to
it after every exchange and refresh, so a restart does not send users through
consent again. Without a store, a refresh token is lost when the process exits.

```go
type OAuth2TokenStore interface {
    SaveToken(token *oauth2.Token) error
    LoadToken() (*oauth2.Token, error)
    DeleteToken() error
}
```

Restore on startup with `oauth.LoadToken()`. A refresh token is a long-lived
credential — encrypt it at rest.

---

## Environment variables

`jira.WithEnv()` configures a client from the environment, trying methods in
this order:

| Variables | Method |
| --- | --- |
| `JIRA_EMAIL` + `JIRA_API_TOKEN` | API token |
| `JIRA_PAT` | Personal access token |
| `JIRA_USERNAME` + `JIRA_PASSWORD` | Basic auth (legacy) |
| `JIRA_OAUTH_CLIENT_ID` + `JIRA_OAUTH_CLIENT_SECRET` + `JIRA_OAUTH_REDIRECT_URL` | OAuth 2.0 (3LO) |
| `JIRA_CONNECT_APP_KEY` + `JIRA_CONNECT_SHARED_SECRET` | Connect JWT |
| `JIRA_OAUTH1_CONSUMER_KEY` + `JIRA_OAUTH1_PRIVATE_KEY` | OAuth 1.0a |
| `JIRA_OAUTH_CLIENT_ID` + `JIRA_OAUTH_CLIENT_SECRET` + `JIRA_OAUTH_TOKEN_URL` | Client credentials |

Supporting variables:

| Variable | Purpose |
| --- | --- |
| `JIRA_BASE_URL` | Instance URL. Required unless `JIRA_CLOUD_ID` is set. |
| `JIRA_CLOUD_ID` | Pins the Cloud site, skipping the accessible-resources lookup. |
| `JIRA_OAUTH_SCOPES` | Space-separated scopes for 3LO and client credentials. |
| `JIRA_OAUTH_ACCESS_TOKEN`, `JIRA_OAUTH_REFRESH_TOKEN` | Restore a previously granted 3LO token. |
| `JIRA_OAUTH_AUDIENCE` | Audience for the client credentials grant. |
| `JIRA_CONNECT_OAUTH_CLIENT_ID`, `JIRA_CONNECT_ACCOUNT_ID` | Switch Connect to the JWT bearer grant, acting as that user. |
| `JIRA_OAUTH1_PRIVATE_KEY_FILE` | Path to the RSA key, instead of inline PEM. |
| `JIRA_OAUTH1_TOKEN`, `JIRA_OAUTH1_TOKEN_SECRET` | Use OAuth 1.0a three-legged. |
| `JIRA_OAUTH1_IMPERSONATE_USER` | Two-legged impersonation. |
| `JIRA_OAUTH1_CONSUMER_SECRET` | Selects HMAC-SHA1 when no private key is set. |

---

## Custom authenticators

Anything implementing `auth.Authenticator` works with `jira.WithAuthenticator`:

```go
type Authenticator interface {
    Authenticate(req *http.Request) error
    Type() string
}
```

To control the base URL per request — a sharded gateway, say — implement
`transport.BaseURLResolver` and pass it to `jira.WithBaseURLResolver`.
