package config

import (
	"crypto/tls"
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

	// Timeout specifies the timeout for the client
	Timeout time.Duration `json:"timeout" env:"KESSEL_TIMEOUT"`

	// MaxReceiveMessageSize sets the maximum message size in bytes the client can receive
	MaxReceiveMessageSize int `json:"max_receive_message_size" env:"KESSEL_GRPC_MAX_RECEIVE_MESSAGE_SIZE" default:"4194304"`

	// MaxSendMessageSize sets the maximum message size in bytes the client can send
	MaxSendMessageSize int `json:"max_send_message_size" env:"KESSEL_GRPC_MAX_SEND_MESSAGE_SIZE" default:"4194304"`
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
