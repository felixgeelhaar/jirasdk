package jirasdk

import (
	"fmt"

	"github.com/felixgeelhaar/jirasdk/auth"
	"github.com/felixgeelhaar/jirasdk/transport"
)

// WithAuthenticator configures an arbitrary authenticator.
//
// Use this when you have built an authenticator directly, or implemented
// auth.Authenticator yourself. The dedicated options below are more convenient
// for the built-in schemes.
func WithAuthenticator(authenticator auth.Authenticator) Option {
	return func(cfg *Config) error {
		if authenticator == nil {
			return fmt.Errorf("authenticator is required")
		}
		cfg.authenticator = authenticator
		return nil
	}
}

// WithConnectJWT configures Atlassian Connect JWT authentication for Jira Cloud.
//
// This is the two-legged flow for Connect apps: the app authenticates as itself
// with the shared secret exchanged during installation, with no user
// authorization step. Set Subject to act in the context of a specific user.
//
// Example:
//
//	WithConnectJWT(&auth.ConnectJWTConfig{
//	    AppKey:       "com.example.my-app",
//	    SharedSecret: install.SharedSecret,
//	})
func WithConnectJWT(config *auth.ConnectJWTConfig) Option {
	return func(cfg *Config) error {
		authenticator, err := auth.NewConnectJWTAuth(config)
		if err != nil {
			return err
		}
		cfg.authenticator = authenticator
		return nil
	}
}

// WithJWTBearer configures the OAuth 2.0 JWT bearer grant for Jira Cloud.
//
// This is the two-legged flow that lets an installed Connect app act as a
// specific user without that user completing an authorization redirect. Tokens
// are fetched and renewed automatically.
//
// Example:
//
//	WithJWTBearer(&auth.JWTBearerConfig{
//	    OAuthClientID: install.OAuthClientID,
//	    SharedSecret:  install.SharedSecret,
//	    AccountID:     "5b10ac8d82e05b22cc7d4ef5",
//	    SiteURL:       "https://your-domain.atlassian.net",
//	})
func WithJWTBearer(config *auth.JWTBearerConfig) Option {
	return func(cfg *Config) error {
		authenticator, err := auth.NewJWTBearerAuth(config)
		if err != nil {
			return err
		}
		cfg.authenticator = authenticator
		return nil
	}
}

// WithOAuth1 configures OAuth 1.0a authentication for Jira Server and Data
// Center.
//
// With no Token in the config this is the two-legged flow, where the consumer
// signs requests with its own RSA key and acts as itself; set ImpersonateUser
// to act on behalf of a user. Supplying Token and TokenSecret — directly, or
// from the handshake helpers on auth.OAuth1Auth — switches to the three-legged
// flow.
//
// Example (two-legged):
//
//	WithOAuth1(&auth.OAuth1Config{
//	    ConsumerKey:   "my-consumer-key",
//	    PrivateKeyPEM: privateKey,
//	})
func WithOAuth1(config *auth.OAuth1Config) Option {
	return func(cfg *Config) error {
		authenticator, err := auth.NewOAuth1Auth(config)
		if err != nil {
			return err
		}
		cfg.authenticator = authenticator
		return nil
	}
}

// WithClientCredentials configures the OAuth 2.0 client credentials grant.
//
// Atlassian's own Jira REST API does not offer this grant. Use it when Jira
// sits behind an API gateway or an in-house identity provider that issues
// tokens Jira accepts; for Jira Cloud use WithConnectJWT or WithJWTBearer, and
// for Server/Data Center use WithOAuth1.
//
// Example:
//
//	WithClientCredentials(&auth.ClientCredentialsConfig{
//	    ClientID:     "service-account",
//	    ClientSecret: os.Getenv("CLIENT_SECRET"),
//	    TokenURL:     "https://idp.example.com/oauth2/token",
//	})
func WithClientCredentials(config *auth.ClientCredentialsConfig) Option {
	return func(cfg *Config) error {
		authenticator, err := auth.NewClientCredentialsAuth(config)
		if err != nil {
			return err
		}
		cfg.authenticator = authenticator
		return nil
	}
}

// WithCloudID pins the Jira Cloud site for OAuth 2.0 (3LO) requests, sending
// them to https://api.atlassian.com/ex/jira/{cloudID}.
//
// This skips the accessible-resources lookup that would otherwise run on the
// first request. Set it when you already know the cloud ID, or when a token
// can reach several sites and you must choose between them.
//
// Example:
//
//	WithCloudID("1324a887-45db-1bf4-1e99-ef0ff456d421")
func WithCloudID(cloudID string) Option {
	return func(cfg *Config) error {
		if cloudID == "" {
			return fmt.Errorf("cloud ID is required")
		}
		cfg.cloudID = cloudID
		return nil
	}
}

// WithoutCloudIDResolution disables the automatic cloud ID lookup, sending
// OAuth 2.0 (3LO) requests to the configured base URL unchanged.
//
// Use it when a proxy or gateway in front of Jira accepts 3LO tokens directly.
// Note that Atlassian site URLs such as https://your-domain.atlassian.net do
// not, so disabling resolution against one will fail with 401.
func WithoutCloudIDResolution() Option {
	return func(cfg *Config) error {
		cfg.disableCloudResolve = true
		return nil
	}
}

// WithBaseURLResolver sets a custom resolver for the base URL, consulted on
// each request. It takes precedence over WithBaseURL.
func WithBaseURLResolver(resolver transport.BaseURLResolver) Option {
	return func(cfg *Config) error {
		if resolver == nil {
			return fmt.Errorf("base URL resolver is required")
		}
		cfg.baseURLResolver = resolver
		return nil
	}
}

// configureCloudResolver installs the cloud ID resolver for Jira Cloud OAuth
// 2.0 (3LO), which is the only scheme whose effective host cannot be known
// before a token exists.
//
// An explicitly supplied resolver always wins. Otherwise resolution applies
// when a cloud ID was pinned, or when the authenticator is a 3LO authenticator
// pointed at Atlassian's own authorization server.
func (cfg *Config) configureCloudResolver() error {
	if cfg.baseURLResolver != nil || cfg.disableCloudResolve {
		return nil
	}

	oauthAuth, isOAuth2 := cfg.authenticator.(*auth.OAuth2Authenticator)
	needsResolution := cfg.cloudID != "" || (isOAuth2 && oauthAuth.IsAtlassianCloud())
	if !needsResolution {
		return nil
	}

	// A configured base URL names the site to select among those the token can
	// reach; requests still travel through the API gateway.
	siteURL := ""
	if cfg.baseURL != nil {
		siteURL = cfg.baseURL.String()
	}

	resolverConfig := &auth.CloudIDResolverConfig{
		Authenticator: cfg.authenticator,
		CloudID:       cfg.cloudID,
		SiteURL:       siteURL,
		HTTPClient:    cfg.httpClient,
	}

	resolver, err := auth.NewCloudIDResolver(resolverConfig)
	if err != nil {
		return fmt.Errorf("failed to configure cloud ID resolution: %w", err)
	}

	cfg.baseURLResolver = resolver
	return nil
}
