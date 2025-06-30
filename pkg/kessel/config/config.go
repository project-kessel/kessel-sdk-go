package config

import (
	"crypto/tls"
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

	// Timeout specifies the default request timeout for gRPC calls
	Timeout time.Duration `json:"timeout" env:"KESSEL_GRPC_TIMEOUT" default:"30s"`
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
	Scopes       []string `json:"scopes" env:"KESSEL_OAUTH_SCOPES" envSeparator:","`
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

func WithGRPCTimeout(timeout time.Duration) GRPCClientOption {
	return func(c *GRPCConfig) {
		c.Timeout = timeout
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
		Timeout: 30 * time.Second,
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
