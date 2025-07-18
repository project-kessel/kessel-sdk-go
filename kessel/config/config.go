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

// BaseConfig contains common configuration for gRPC client
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

// Oauth2 contains OAuth2 configuration
type Oauth2 struct {
	ClientID     string   `json:"client_id" env:"KESSEL_OAUTH2_CLIENT_ID"`
	ClientSecret string   `json:"client_secret" env:"KESSEL_OAUTH2_CLIENT_SECRET"`
	TokenURL     string   `json:"token_url" env:"KESSEL_OAUTH2_TOKEN_URL"`
	IssuerURL    string   `json:"issuer_url" env:"KESSEL_OAUTH2_ISSUER_URL"`
	Scopes       []string `json:"scopes" env:"KESSEL_OAUTH2_SCOPES" envSeparator:","`
}

// DiscoverTokenEndpoint discovers and sets the token endpoint from the issuer URL
func (o *Oauth2) DiscoverTokenEndpoint(ctx context.Context) error {
	if o.IssuerURL == "" {
		return fmt.Errorf("issuer_url is required for token endpoint discovery")
	}

	// Ensure issuerURL doesn't end with a slash
	issuerURL := strings.TrimSuffix(o.IssuerURL, "/")

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
	o.TokenURL = doc.TokenEndpoint

	return nil
}

// GRPCClientOption defines a function type for gRPC client configuration
type GRPCClientOption func(*GRPCConfig)

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

// GetEnableOauth returns the OAuth enable flag for GRPCConfig
func (c *GRPCConfig) GetEnableOauth() bool {
	return c.EnableOauth
}

// GetOauth2 returns the OAuth2 configuration for GRPCConfig
func (c *GRPCConfig) GetOauth2() Oauth2 {
	return c.Oauth2
}

// NewGRPCConfig creates a new gRPC configuration with default values
func NewGRPCConfig(options ...GRPCClientOption) *GRPCConfig {
	config := &GRPCConfig{
		BaseConfig: BaseConfig{
			Insecure: false,
		},
		MaxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
		MaxSendMessageSize:    4 * 1024 * 1024, // 4MB
	}

	for _, option := range options {
		option(config)
	}

	return config
}
