package config

import (
	"crypto/tls"
	"time"
)

type Config struct {
	// Endpoint specifies the server address (host:port)
	Endpoint string `json:"endpoint" env:"KESSEL_ENDPOINT"`

	// Timeout specifies the default request timeout
	Timeout time.Duration `json:"timeout" env:"KESSEL_TIMEOUT" default:"10s"`

	// UserAgent specifies the client user agent
	UserAgent string `json:"user_agent" env:"KESSEL_USER_AGENT" default:"kessel-go-sdk"`

	// UseHTTP forces HTTP instead of gRPC
	UseHTTP bool `json:"use_http" env:"KESSEL_USE_HTTP"`

	// Insecure disables transport security
	Insecure bool `json:"insecure" env:"KESSEL_INSECURE"`

	// TLSConfig specifies custom TLS configuration
	TLSConfig *tls.Config `json:"-"`

	EnableOIDCAuth bool `json:"enableOIDCAuth" env:"KESSEL_ENABLE_OAUTH"`
	// OAuthConfig specifies OAuth2 configuration
	ClientID     string   `json:"client_id" env:"KESSEL_OAUTH_CLIENT_ID"`
	ClientSecret string   `json:"client_secret" env:"KESSEL_OAUTH_CLIENT_SECRET"`
	TokenURL     string   `json:"token_url" env:"KESSEL_OAUTH_TOKEN_URL"`
	Scopes       []string `json:"scopes" env:"KESSEL_OAUTH_SCOPES" envSeparator:","`
}

type ClientOptions func(*Config)

func WithAuthEnabled(clientId string, clientSecret string, authServerTokenUrl string) ClientOptions {
	return func(c *Config) {
		c.EnableOIDCAuth = true
		c.ClientID = clientId
		c.ClientSecret = clientSecret
		c.TokenURL = authServerTokenUrl
	}
}

func WithHttp(useHttp bool) ClientOptions {
	return func(c *Config) {
		c.UseHTTP = useHttp
	}
}

func WithTLSInsecure(insecure bool) ClientOptions {
	return func(c *Config) {
		c.Insecure = insecure
	}
}

func WithHTTPTLSConfig(tlsConfig *tls.Config) ClientOptions {
	return func(c *Config) {
		c.Insecure = false
		c.TLSConfig = tlsConfig
	}
}

func WithTimeout(timeout time.Duration) ClientOptions {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

func NewConfig(options ...func(*Config)) *Config {
	svr := &Config{}
	for _, o := range options {
		o(svr)
	}
	return svr
}

func WithEndpoint(endpoint string) func(*Config) {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}
