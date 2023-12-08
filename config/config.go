package config

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	internal "github.com/cloudfoundry-community/go-cfclient/v3/internal/http"
	"github.com/cloudfoundry-community/go-cfclient/v3/internal/ios"
	"github.com/cloudfoundry-community/go-cfclient/v3/internal/jwt"
	"github.com/cloudfoundry-community/go-cfclient/v3/internal/path"
	"github.com/cloudfoundry-community/go-cfclient/v3/resource"
)

const (
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeClientCredentials = "client_credentials"
	GrantTypeAuthorizationCode = "authorization_code"

	DefaultRequestTimeout = 30 * time.Second
	DefaultUserAgent      = "Go-CF-Client/3.0"
	DefaultClientID       = "cf"
	DefaultSSHClientID    = "ssh-proxy"
)

// Config is used to configure the creation of a client
type Config struct {
	apiEndpointURL   string
	loginEndpointURL string
	uaaEndpointURL   string
	sshOAuthClient   string

	username          string
	password          string
	clientID          string
	clientSecret      string
	grantType         string
	origin            string
	scopes            []string
	oAuthToken        *oauth2.Token
	httpClient        *http.Client
	httpAuthClient    *http.Client
	skipTLSValidation bool
	requestTimeout    time.Duration
	userAgent         string
}

// New creates a new Config with specified API root URL and options.
func New(apiRootURL string, options ...Option) (*Config, error) {
	u, err := url.Parse(apiRootURL)
	if err != nil {
		return nil, fmt.Errorf("expected an http(s) CF API root URI, but got %s: %w", apiRootURL, err)
	}
	cfg := &Config{
		apiEndpointURL: strings.TrimRight(u.String(), "/"),
		userAgent:      DefaultUserAgent,
		requestTimeout: DefaultRequestTimeout,
		clientID:       DefaultClientID,
		sshOAuthClient: DefaultSSHClientID,
	}
	err = initConfig(cfg, options...)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// NewFromCFHome creates a client config from the CF CLI config.
//
// This will use the currently configured CF_HOME env var if it exists, otherwise attempts to use the
// default CF_HOME directory.
//
// If CF_USERNAME and CF_PASSWORD env vars are set then those credentials will be used to get an oauth2 token. If
// those env vars are not set then the stored oauth2 token is used.
func NewFromCFHome(options ...Option) (*Config, error) {
	dir, err := findCFHomeDir()
	if err != nil {
		return nil, err
	}
	return NewFromCFHomeDir(dir, options...)
}

// NewFromCFHomeDir creates a client config from the CF CLI config using the specified directory.
//
// This will attempt to read the CF CLI config from the specified directory only.
//
// If CF_USERNAME and CF_PASSWORD env vars are set then those credentials will be used to get an oauth2 token. If
// those env vars are not set then the stored oauth2 token is used.
func NewFromCFHomeDir(cfHomeDir string, options ...Option) (*Config, error) {
	cfg, err := createConfigFromCFCLIConfig(cfHomeDir)
	if err != nil {
		return nil, err
	}
	err = initConfig(cfg, options...)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) CreateOAuth2TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	// use our http.Client instance for token acquisition
	oauthCtx := context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)

	twoLeggedAuthConfigFn := func() *clientcredentials.Config {
		return &clientcredentials.Config{
			ClientID:     c.clientID,
			ClientSecret: c.clientSecret,
			TokenURL:     c.uaaEndpointURL,
		}
	}

	threeLeggedAuthConfigFn := func() *oauth2.Config {
		return &oauth2.Config{
			ClientID:     c.clientID,
			ClientSecret: c.clientSecret,
			Scopes:       c.scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.loginEndpointURL + "/oauth/auth",
				TokenURL: c.uaaEndpointURL + "/oauth/token",
			},
		}
	}

	var tokenSource oauth2.TokenSource
	switch c.grantType {
	case GrantTypeClientCredentials:
		authConfig := twoLeggedAuthConfigFn()
		tokenSource = authConfig.TokenSource(oauthCtx)
	case GrantTypeAuthorizationCode:
		authConfig := threeLeggedAuthConfigFn()

		// Add optional login hint to the token URL
		if c.origin != "" {
			authConfig.Endpoint.TokenURL = addLoginHintToURL(authConfig.Endpoint.TokenURL, c.origin)
		}

		// Login using user/pass
		token, err := authConfig.PasswordCredentialsToken(oauthCtx, c.username, c.password)
		if err != nil {
			return nil, err
		}
		tokenSource = authConfig.TokenSource(oauthCtx, token)
	case GrantTypeRefreshToken:
		authConfig := threeLeggedAuthConfigFn()
		tokenSource = authConfig.TokenSource(oauthCtx, c.oAuthToken)
	default:
		return nil, fmt.Errorf("unsupported OAuth2 grant type '%s'", c.grantType)
	}
	return tokenSource, nil
}

// HTTPClient returns the un-authenticated http.Client.
func (c *Config) HTTPClient() *http.Client {
	return c.httpClient
}

// HTTPAuthClient returns the authenticated http.Client.
func (c *Config) HTTPAuthClient() *http.Client {
	return c.httpClient
}

// SSHOAuthClientID returns the clientID used to request an SSH code, typically 'ssh-proxy'.
func (c *Config) SSHOAuthClientID() string {
	return c.sshOAuthClient
}

// UserAgent returns the configured user agent header string.
func (c *Config) UserAgent() string {
	return c.userAgent
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Ensure at least one of clientID, username, or token is provided
	if c.clientID == "" && c.username == "" && c.oAuthToken == nil {
		return errors.New("either client credentials, user credentials, or tokens are required")
	}

	// If a non-default clientID is provided, check for clientSecret
	if c.clientID != DefaultClientID && c.clientSecret == "" {
		return errors.New("client secret is required when using client credentials")
	}

	// If username is provided, check for password
	if c.username != "" && c.password == "" {
		return errors.New("password is required when using user credentials")
	}

	return nil
}

// initConfig fully populates and then validates the provided base config
func initConfig(cfg *Config, options ...Option) error {
	// Apply any user provided config overrides
	err := applyOptions(cfg, options...)
	if err != nil {
		return err
	}

	// Validate the config object is ready to use
	err = cfg.Validate()
	if err != nil {
		return err
	}

	// Ensure a http.Client is available and properly configured
	configureHTTPClient(cfg)

	// Query the CF API for UAA/Login endpoints
	err = discoverAuthConfig(context.Background(), cfg)
	if err != nil {
		return err
	}

	// Finally create a http.Client for making API calls that require authentication
	return createHTTPAuthClient(context.Background(), cfg)
}

// applyOptions executes each option function to create the config.
func applyOptions(cfg *Config, options ...Option) error {
	for _, option := range options {
		if err := option(cfg); err != nil {
			return err
		}
	}
	return nil
}

// configureHTTPClient creates a default http.Client if one wasn't supplied in the config and then
// configures the base http.Client from the config.
func configureHTTPClient(c *Config) {
	// Only configure the client if it has not been configured before
	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		}
	}
	if transport := getHTTPTransport(c.httpClient); transport != nil {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = c.skipTLSValidation
	}
	c.httpClient.CheckRedirect = internal.CheckRedirect
	c.httpClient.Timeout = c.requestTimeout
}

// createHTTPAuthClient creates the http.Client used for any API calls that require authentication.
func createHTTPAuthClient(ctx context.Context, c *Config) (err error) {
	c.httpAuthClient, err = internal.NewAuthenticatedClient(ctx, c.httpClient, c)
	return err
}

// discoverAuthConfig configures the UAA and Login config properties from the CF API if none were supplied in the
// config.
func discoverAuthConfig(ctx context.Context, c *Config) error {
	// Return immediately if URLs have already been configured
	if c.loginEndpointURL != "" && c.uaaEndpointURL != "" {
		return nil
	}

	// Query the CF API root for the service locator records
	root, err := globalAPIRoot(ctx, c.httpClient, c.ToURL("/"))
	if err != nil {
		return fmt.Errorf("error while discovering token service URL: %w", err)
	}
	c.loginEndpointURL = root.Links.Login.Href
	c.uaaEndpointURL = root.Links.Uaa.Href
	c.sshOAuthClient = root.Links.AppSSH.Meta.OauthClient
	return nil
}

// globalAPIRoot queries the CF API service discovery root endpoint
func globalAPIRoot(ctx context.Context, httpClient *http.Client, url string) (*resource.Root, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error occurred while generating the request for the global API root: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request, failed during HTTP request send: %w", err)
	}
	if !internal.IsStatusSuccess(resp.StatusCode) {
		return nil, internal.DecodeError(resp)
	}
	defer ios.Close(resp.Body)

	var root resource.Root
	if err := internal.DecodeBody(resp, &root); err != nil {
		return nil, fmt.Errorf("failed to decode API root response: %w", err)
	}
	return &root, nil
}

// createConfigFromCFCLIConfig generates a config object from the CF CLI config found in the specified CF home
// directory.
func createConfigFromCFCLIConfig(cfHomeDir string) (*Config, error) {
	cf, err := loadCFCLIConfig(cfHomeDir)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		apiEndpointURL:    cf.Target,
		loginEndpointURL:  cf.AuthorizationEndpoint,
		uaaEndpointURL:    cf.UaaEndpoint,
		clientID:          cf.UAAOAuthClient,
		clientSecret:      cf.UAAOAuthClientSecret,
		sshOAuthClient:    cf.SSHOAuthClient,
		skipTLSValidation: cf.SSLDisabled,
		userAgent:         DefaultUserAgent,
		requestTimeout:    DefaultRequestTimeout,
	}

	// if the username and password are specified via env vars use password based auth
	if os.Getenv("CF_USERNAME") != "" && os.Getenv("CF_PASSWORD") != "" {
		cfg.username = os.Getenv("CF_USERNAME")
		cfg.password = os.Getenv("CF_PASSWORD")
		cfg.grantType = GrantTypeAuthorizationCode
	} else {
		oAuthToken, err := jwt.ToOAuth2Token(cf.AccessToken, cf.RefreshToken)
		if err != nil {
			return nil, err
		}
		cfg.oAuthToken = oAuthToken
		cfg.grantType = GrantTypeRefreshToken
	}

	return cfg, nil
}

func getHTTPTransport(client *http.Client) *http.Transport {
	switch t := client.Transport.(type) {
	case *http.Transport:
		return t
	case *oauth2.Transport:
		if httpTransport, ok := t.Base.(*http.Transport); ok {
			return httpTransport
		}
	}
	return nil
}

func addLoginHintToURL(tokenURL, origin string) string {
	u, err := url.Parse(tokenURL)
	if err != nil {
		// Handle the error, or return the original URL
		return tokenURL
	}

	q := u.Query()
	q.Add("login_hint", fmt.Sprintf(`{"origin":"%s"}`, origin))
	u.RawQuery = q.Encode()

	return u.String()
}

//////////////////////// OAuth2/HTTP ///////////////////////////////

func (c *Config) ToURL(urlPath string) string {
	return path.Join(c.apiEndpointURL, urlPath)
}

func (c *Config) ToAuthenticateURL(urlPath string) string {
	return path.Join(c.uaaEndpointURL, urlPath)
}
