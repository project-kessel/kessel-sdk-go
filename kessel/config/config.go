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

// CompatibilityConfig contains common configuration for gRPC client
type CompatibilityConfig struct {
	// Url specifies the server address (host:port)
	Url string `json:"endpoint" env:"KESSEL_ENDPOINT"`

	// Insecure disables transport security
	Insecure bool `json:"insecure" env:"KESSEL_INSECURE"`

	// TLSConfig specifies custom TLS configuration
	TLSConfig *tls.Config `json:"-"`

	// EnableOIDCAuth enables OAuth2 authentication
	EnableOIDCAuth bool `json:"enable_oauth" env:"KESSEL_ENABLE_OAUTH"`

	// OAuth2 fields
	ClientID           string   `json:"client_id" env:"KESSEL_OAUTH2_CLIENT_ID"`
	ClientSecret       string   `json:"client_secret" env:"KESSEL_OAUTH2_CLIENT_SECRET"`
	AuthServerTokenUrl string   `json:"authServerTokenUrl" env:"KESSEL_OAUTH2_TOKEN_URL"`
	IssuerURL          string   `json:"issuer_url" env:"KESSEL_OAUTH2_ISSUER_URL"`
	Scopes             []string `json:"scopes" env:"KESSEL_OAUTH2_SCOPES" envSeparator:","`

	// Timeout specifies the timeout for the client
	Timeout time.Duration `json:"timeout" env:"KESSEL_TIMEOUT"`

	// MaxReceiveMessageSize sets the maximum message size in bytes the client can receive
	MaxReceiveMessageSize int `json:"max_receive_message_size" env:"KESSEL_GRPC_MAX_RECEIVE_MESSAGE_SIZE" default:"4194304"`

	// MaxSendMessageSize sets the maximum message size in bytes the client can send
	MaxSendMessageSize int `json:"max_send_message_size" env:"KESSEL_GRPC_MAX_SEND_MESSAGE_SIZE" default:"4194304"`
}

// DiscoverTokenEndpoint discovers and sets the token endpoint from the issuer URL
func (c *CompatibilityConfig) DiscoverTokenEndpoint(ctx context.Context) error {
	if c.IssuerURL == "" {
		return fmt.Errorf("issuer_url is required for token endpoint discovery")
	}

	// Ensure issuerURL doesn't end with a slash
	issuerURL := strings.TrimSuffix(c.IssuerURL, "/")

	// Construct the well-known configuration URL
	discoveryURL := issuerURL + "/.well-known/openid-configuration"

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
	c.AuthServerTokenUrl = doc.TokenEndpoint

	return nil
}

// CompatibilityClientOption defines a function type for gRPC client configuration
type CompatibilityClientOption func(*CompatibilityConfig)

// GRPC Configuration Options

func WithGRPCEndpoint(endpoint string) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.Url = endpoint
	}
}

func WithGRPCInsecure(insecure bool) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.Insecure = insecure
	}
}

func WithGRPCTLSConfig(tlsConfig *tls.Config) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.Insecure = false
		c.TLSConfig = tlsConfig
	}
}

func WithGRPCMaxReceiveMessageSize(size int) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.MaxReceiveMessageSize = size
	}
}

func WithGRPCMaxSendMessageSize(size int) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.MaxSendMessageSize = size
	}
}

func WithTimeout(timeout time.Duration) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.Timeout = timeout
	}
}

func WithGRPCOAuth2(clientID, clientSecret, tokenURL string, scopes ...string) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.EnableOIDCAuth = true
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.AuthServerTokenUrl = tokenURL
		c.Scopes = scopes
	}
}

// WithGRPCOAuth2Issuer configures OAuth2 authentication using an issuer URL for token endpoint discovery
func WithGRPCOAuth2Issuer(clientID, clientSecret, issuerURL string, scopes ...string) CompatibilityClientOption {
	return func(c *CompatibilityConfig) {
		c.EnableOIDCAuth = true
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.IssuerURL = issuerURL
		c.Scopes = scopes
	}
}

// GetEnableOIDCAuth returns the OAuth enable flag for CompatibilityConfig
func (c *CompatibilityConfig) GetEnableOIDCAuth() bool {
	return c.EnableOIDCAuth
}

// GetClientID returns the OAuth client ID
func (c *CompatibilityConfig) GetClientID() string {
	return c.ClientID
}

// GetClientSecret returns the OAuth client secret
func (c *CompatibilityConfig) GetClientSecret() string {
	return c.ClientSecret
}

// GetTokenURL returns the OAuth token URL
func (c *CompatibilityConfig) GetTokenURL() string {
	return c.AuthServerTokenUrl
}

// GetIssuerURL returns the OAuth issuer URL
func (c *CompatibilityConfig) GetIssuerURL() string {
	return c.IssuerURL
}

// GetScopes returns the OAuth scopes
func (c *CompatibilityConfig) GetScopes() []string {
	return c.Scopes
}

// NewCompatibilityConfig creates a new gRPC configuration with default values
func NewCompatibilityConfig(options ...CompatibilityClientOption) *CompatibilityConfig {
	config := &CompatibilityConfig{
		Insecure:              false,
		MaxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
		MaxSendMessageSize:    4 * 1024 * 1024, // 4MB
	}

	for _, option := range options {
		option(config)
	}

	return config
}
