package config

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaseConfig contains common configuration shared between HTTP and gRPC
type BaseConfig struct {
	// Endpoint specifies the server address (host:port)
	Endpoint string `json:"endpoint" env:"KESSEL_ENDPOINT"`

	// Insecure disables transport security
	Insecure bool `json:"insecure" env:"KESSEL_INSECURE"`

	// TLSConfig specifies custom TLS configuration
	TLSConfig *tls.Config `json:"-"`

	// EnableOauth enables OAuth2 authentication
	EnableOauth bool `json:"enable_oauth" env:"KESSEL_ENABLE_OAUTH"`

	// Oauth2 specifies OAuth2 configuration
	Oauth2 Oauth2 `json:"oauth2"`
}

// GRPCConfig contains gRPC-specific configuration
type GRPCConfig struct {
	BaseConfig

	// MaxReceiveMessageSize sets the maximum message size in bytes the client can receive
	MaxReceiveMessageSize int `json:"max_receive_message_size" env:"KESSEL_GRPC_MAX_RECEIVE_MESSAGE_SIZE" default:"4194304"`

	// MaxSendMessageSize sets the maximum message size in bytes the client can send
	MaxSendMessageSize int `json:"max_send_message_size" env:"KESSEL_GRPC_MAX_SEND_MESSAGE_SIZE" default:"4194304"`
}

// HTTPConfig contains HTTP-specific configuration
type HTTPConfig struct {
	BaseConfig

	// Timeout specifies the default request timeout for HTTP calls
	Timeout time.Duration `json:"timeout" env:"KESSEL_HTTP_TIMEOUT" default:"10s"`

	// UserAgent specifies the client user agent
	UserAgent string `json:"user_agent" env:"KESSEL_HTTP_USER_AGENT" default:"kessel-go-sdk"`

	// MaxIdleConns controls the maximum number of idle connections
	MaxIdleConns int `json:"max_idle_conns" env:"KESSEL_HTTP_MAX_IDLE_CONNS" default:"100"`

	// IdleConnTimeout is the maximum amount of time an idle connection will remain idle
	IdleConnTimeout time.Duration `json:"idle_conn_timeout" env:"KESSEL_HTTP_IDLE_CONN_TIMEOUT" default:"90s"`
}

// Oauth2 contains OAuth2 configuration
type Oauth2 struct {
	ClientID     string   `json:"client_id" env:"KESSEL_OAUTH_CLIENT_ID"`
	ClientSecret string   `json:"client_secret" env:"KESSEL_OAUTH_CLIENT_SECRET"`
	TokenURL     string   `json:"token_url" env:"KESSEL_OAUTH_TOKEN_URL"`
	IssuerURL    string   `json:"issuer_url" env:"KESSEL_OAUTH_ISSUER_URL"`
	Scopes       []string `json:"scopes" env:"KESSEL_OAUTH_SCOPES" envSeparator:","`
}

// DiscoverTokenEndpoint discovers and sets the token endpoint from the issuer URL
func (o *Oauth2) DiscoverTokenEndpoint(ctx context.Context) error {
	if o.IssuerURL == "" {
		return fmt.Errorf("issuer_url is required for token endpoint discovery")
	}

	// Ensure issuerURL doesn't end with a slash
	issuerURL := strings.TrimSuffix(o.IssuerURL, "/")

	// Construct the well-known configuration URL
	discoveryURL := issuerURL + "/.well-known/openid_configuration"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create discovery request: %w", err)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Failed to close connection: %v", closeErr)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var doc struct {
		TokenEndpoint string `json:"token_endpoint"`
		Issuer        string `json:"issuer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("failed to decode discovery document: %w", err)
	}

	// Validate token endpoint
	if doc.TokenEndpoint == "" {
		return fmt.Errorf("token_endpoint not found in discovery document")
	}

	// Validate that token endpoint is a valid URL
	if _, err := url.Parse(doc.TokenEndpoint); err != nil {
		return fmt.Errorf("invalid token_endpoint URL: %w", err)
	}

	// Update the TokenURL field
	o.TokenURL = doc.TokenEndpoint

	return nil
}

// GRPCClientOption defines a function type for gRPC client configuration
type GRPCClientOption func(*GRPCConfig)

// HTTPClientOption defines a function type for HTTP client configuration
type HTTPClientOption func(*HTTPConfig)

// GRPC Configuration Options

func WithGRPCEndpoint(endpoint string) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.Endpoint = endpoint
	}
}

func WithGRPCInsecure(insecure bool) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.Insecure = insecure
	}
}

func WithGRPCTLSConfig(tlsConfig *tls.Config) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.Insecure = false
		c.TLSConfig = tlsConfig
	}
}

func WithGRPCMaxReceiveMessageSize(size int) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.MaxReceiveMessageSize = size
	}
}

func WithGRPCMaxSendMessageSize(size int) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.MaxSendMessageSize = size
	}
}

func WithGRPCOAuth2(clientID, clientSecret, tokenURL string, scopes ...string) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.EnableOauth = true
		c.Oauth2.ClientID = clientID
		c.Oauth2.ClientSecret = clientSecret
		c.Oauth2.TokenURL = tokenURL
		c.Oauth2.Scopes = scopes
	}
}

// WithGRPCOAuth2Issuer configures OAuth2 authentication using an issuer URL for token endpoint discovery
func WithGRPCOAuth2Issuer(clientID, clientSecret, issuerURL string, scopes ...string) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.EnableOauth = true
		c.Oauth2.ClientID = clientID
		c.Oauth2.ClientSecret = clientSecret
		c.Oauth2.IssuerURL = issuerURL
		c.Oauth2.Scopes = scopes
	}
}

// HTTP Configuration Options

func WithHTTPEndpoint(endpoint string) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.Endpoint = endpoint
	}
}

func WithHTTPInsecure(insecure bool) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.Insecure = insecure
	}
}

func WithHTTPTLSConfig(tlsConfig *tls.Config) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.Insecure = false
		c.TLSConfig = tlsConfig
	}
}

func WithHTTPTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.Timeout = timeout
	}
}

func WithHTTPUserAgent(userAgent string) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.UserAgent = userAgent
	}
}

func WithHTTPMaxIdleConns(maxIdleConns int) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.MaxIdleConns = maxIdleConns
	}
}

func WithHTTPIdleConnTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.IdleConnTimeout = timeout
	}
}

func WithHTTPOAuth2(clientID, clientSecret, tokenURL string, scopes ...string) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.EnableOauth = true
		c.Oauth2.ClientID = clientID
		c.Oauth2.ClientSecret = clientSecret
		c.Oauth2.TokenURL = tokenURL
		c.Oauth2.Scopes = scopes
	}
}

// WithHTTPOAuth2Issuer configures OAuth2 authentication using an issuer URL for token endpoint discovery
func WithHTTPOAuth2Issuer(clientID, clientSecret, issuerURL string, scopes ...string) HTTPClientOption {
	return func(c *HTTPConfig) {
		c.EnableOauth = true
		c.Oauth2.ClientID = clientID
		c.Oauth2.ClientSecret = clientSecret
		c.Oauth2.IssuerURL = issuerURL
		c.Oauth2.Scopes = scopes
	}
}

// GetEnableOauth returns the OAuth enable flag for GRPCConfig
func (c *GRPCConfig) GetEnableOauth() bool {
	return c.EnableOauth
}

// GetOauth2 returns the OAuth2 configuration for GRPCConfig
func (c *GRPCConfig) GetOauth2() Oauth2 {
	return c.Oauth2
}

// GetEnableOauth returns the OAuth enable flag for HTTPConfig
func (c *HTTPConfig) GetEnableOauth() bool {
	return c.EnableOauth
}

// GetOauth2 returns the OAuth2 configuration for HTTPConfig
func (c *HTTPConfig) GetOauth2() Oauth2 {
	return c.Oauth2
}

// NewGRPCConfig creates a new gRPC configuration with default values
func NewGRPCConfig(options ...GRPCClientOption) *GRPCConfig {
	config := &GRPCConfig{
		BaseConfig: BaseConfig{
			Insecure: false,
		},
		//MaxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
		//MaxSendMessageSize:    4 * 1024 * 1024, // 4MB
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// NewHTTPConfig creates a new HTTP configuration with default values
func NewHTTPConfig(options ...HTTPClientOption) *HTTPConfig {
	config := &HTTPConfig{
		BaseConfig: BaseConfig{
			Insecure: false,
		},
		Timeout:         10 * time.Second,
		UserAgent:       "kessel-go-sdk",
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	for _, option := range options {
		option(config)
	}

	return config
}
