package config

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestNewCompatibilityConfig_DefaultValues(t *testing.T) {
	config := NewCompatibilityConfig()

	// Test default values
	if config.Insecure != false {
		t.Errorf("Expected Insecure to be false by default, got %v", config.Insecure)
	}

	expectedSize := 4 * 1024 * 1024 // 4MB
	if config.MaxReceiveMessageSize != expectedSize {
		t.Errorf("Expected MaxReceiveMessageSize to be %d, got %d", expectedSize, config.MaxReceiveMessageSize)
	}

	if config.MaxSendMessageSize != expectedSize {
		t.Errorf("Expected MaxSendMessageSize to be %d, got %d", expectedSize, config.MaxSendMessageSize)
	}

	if config.Url != "" {
		t.Errorf("Expected Url to be empty by default, got %s", config.Url)
	}

	if config.TLSConfig != nil {
		t.Errorf("Expected TLSConfig to be nil by default, got %v", config.TLSConfig)
	}

	if config.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0 by default, got %v", config.Timeout)
	}
}

func TestWithGRPCEndpoint(t *testing.T) {
	endpoint := "localhost:8080"
	config := NewCompatibilityConfig(WithGRPCEndpoint(endpoint))

	if config.Url != endpoint {
		t.Errorf("Expected Url to be %s, got %s", endpoint, config.Url)
	}
}

func TestWithGRPCInsecure(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
	}{
		{"set insecure true", true},
		{"set insecure false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewCompatibilityConfig(WithGRPCInsecure(tt.insecure))

			if config.Insecure != tt.insecure {
				t.Errorf("Expected Insecure to be %v, got %v", tt.insecure, config.Insecure)
			}
		})
	}
}

func TestWithGRPCTLSConfig(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "example.com",
		MinVersion: tls.VersionTLS12,
	}

	config := NewCompatibilityConfig(WithGRPCTLSConfig(tlsConfig))

	if config.TLSConfig != tlsConfig {
		t.Errorf("Expected TLSConfig to be the provided config")
	}

	// Should also set Insecure to false
	if config.Insecure != false {
		t.Errorf("Expected Insecure to be false when TLS config is set, got %v", config.Insecure)
	}
}

func TestWithGRPCMaxReceiveMessageSize(t *testing.T) {
	size := 8 * 1024 * 1024 // 8MB
	config := NewCompatibilityConfig(WithGRPCMaxReceiveMessageSize(size))

	if config.MaxReceiveMessageSize != size {
		t.Errorf("Expected MaxReceiveMessageSize to be %d, got %d", size, config.MaxReceiveMessageSize)
	}
}

func TestWithGRPCMaxSendMessageSize(t *testing.T) {
	size := 16 * 1024 * 1024 // 16MB
	config := NewCompatibilityConfig(WithGRPCMaxSendMessageSize(size))

	if config.MaxSendMessageSize != size {
		t.Errorf("Expected MaxSendMessageSize to be %d, got %d", size, config.MaxSendMessageSize)
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	config := NewCompatibilityConfig(WithTimeout(timeout))

	if config.Timeout != timeout {
		t.Errorf("Expected Timeout to be %v, got %v", timeout, config.Timeout)
	}
}

func TestMultipleOptions(t *testing.T) {
	endpoint := "grpc.example.com:443"
	timeout := 10 * time.Second
	receiveSize := 8 * 1024 * 1024
	sendSize := 2 * 1024 * 1024
	tlsConfig := &tls.Config{ServerName: "grpc.example.com"}

	config := NewCompatibilityConfig(
		WithGRPCEndpoint(endpoint),
		WithTimeout(timeout),
		WithGRPCMaxReceiveMessageSize(receiveSize),
		WithGRPCMaxSendMessageSize(sendSize),
		WithGRPCTLSConfig(tlsConfig),
	)

	if config.Url != endpoint {
		t.Errorf("Expected Url to be %s, got %s", endpoint, config.Url)
	}

	if config.Timeout != timeout {
		t.Errorf("Expected Timeout to be %v, got %v", timeout, config.Timeout)
	}

	if config.MaxReceiveMessageSize != receiveSize {
		t.Errorf("Expected MaxReceiveMessageSize to be %d, got %d", receiveSize, config.MaxReceiveMessageSize)
	}

	if config.MaxSendMessageSize != sendSize {
		t.Errorf("Expected MaxSendMessageSize to be %d, got %d", sendSize, config.MaxSendMessageSize)
	}

	if config.TLSConfig != tlsConfig {
		t.Errorf("Expected TLSConfig to be the provided config")
	}

	if config.Insecure != false {
		t.Errorf("Expected Insecure to be false when TLS config is set, got %v", config.Insecure)
	}
}

func TestOptionsOrder(t *testing.T) {
	// Test that options applied later override earlier ones
	config := NewCompatibilityConfig(
		WithGRPCInsecure(true),
		WithGRPCTLSConfig(&tls.Config{}), // This should set Insecure to false
	)

	if config.Insecure != false {
		t.Errorf("Expected final Insecure value to be false (TLS config should override), got %v", config.Insecure)
	}
}

func TestOptionsImmutability(t *testing.T) {
	// Test that options don't affect each other
	endpoint1 := "service1.example.com:443"
	endpoint2 := "service2.example.com:443"

	config1 := NewCompatibilityConfig(WithGRPCEndpoint(endpoint1))
	config2 := NewCompatibilityConfig(WithGRPCEndpoint(endpoint2))

	if config1.Url != endpoint1 {
		t.Errorf("Expected config1 Url to be %s, got %s", endpoint1, config1.Url)
	}

	if config2.Url != endpoint2 {
		t.Errorf("Expected config2 Url to be %s, got %s", endpoint2, config2.Url)
	}
}

func TestZeroValueOptions(t *testing.T) {
	// Test setting zero values
	config := NewCompatibilityConfig(
		WithTimeout(0),
		WithGRPCMaxReceiveMessageSize(0),
		WithGRPCMaxSendMessageSize(0),
		WithGRPCEndpoint(""),
	)

	if config.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0, got %v", config.Timeout)
	}

	if config.MaxReceiveMessageSize != 0 {
		t.Errorf("Expected MaxReceiveMessageSize to be 0, got %d", config.MaxReceiveMessageSize)
	}

	if config.MaxSendMessageSize != 0 {
		t.Errorf("Expected MaxSendMessageSize to be 0, got %d", config.MaxSendMessageSize)
	}

	if config.Url != "" {
		t.Errorf("Expected Url to be empty, got %s", config.Url)
	}
}

func TestCompatibilityClientOptionType(t *testing.T) {
	// Test that all option functions have the correct type
	var options []CompatibilityClientOption

	options = append(options,
		WithGRPCEndpoint("test"),
		WithGRPCInsecure(true),
		WithGRPCTLSConfig(nil),
		WithGRPCMaxReceiveMessageSize(1024),
		WithGRPCMaxSendMessageSize(1024),
		WithTimeout(time.Second),
	)

	// This test passes if it compiles - all options should be of type CompatibilityClientOption
	config := NewCompatibilityConfig(options...)

	if config == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestConfigurationFlow(t *testing.T) {
	// Test a realistic configuration flow
	config := NewCompatibilityConfig()

	// Simulate environment-based configuration
	if config.Url == "" {
		config = NewCompatibilityConfig(WithGRPCEndpoint("localhost:9090"))
	}

	// Add security configuration
	config = NewCompatibilityConfig(
		WithGRPCEndpoint(config.Url),
		WithGRPCInsecure(false),
		WithGRPCTLSConfig(&tls.Config{
			ServerName: "kessel-inventory",
			MinVersion: tls.VersionTLS12,
		}),
	)

	// Add performance tuning
	config = NewCompatibilityConfig(
		WithGRPCEndpoint(config.Url),
		WithGRPCTLSConfig(config.TLSConfig),
		WithTimeout(30*time.Second),
		WithGRPCMaxReceiveMessageSize(8*1024*1024),
		WithGRPCMaxSendMessageSize(4*1024*1024),
	)

	// Verify final configuration
	if config.Url != "localhost:9090" {
		t.Errorf("Expected final Url to be 'localhost:9090', got %s", config.Url)
	}

	if config.Insecure != false {
		t.Errorf("Expected Insecure to be false, got %v", config.Insecure)
	}

	if config.TLSConfig == nil || config.TLSConfig.ServerName != "kessel-inventory" {
		t.Errorf("Expected TLS config with ServerName 'kessel-inventory'")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", config.Timeout)
	}

	if config.MaxReceiveMessageSize != 8*1024*1024 {
		t.Errorf("Expected MaxReceiveMessageSize to be 8MB, got %d", config.MaxReceiveMessageSize)
	}

	if config.MaxSendMessageSize != 4*1024*1024 {
		t.Errorf("Expected MaxSendMessageSize to be 4MB, got %d", config.MaxSendMessageSize)
	}
}