package http

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
)

func TestNewClientBuilder(t *testing.T) {
	endpoint := "https://api.example.com"
	builder := NewClientBuilder(endpoint)

	if builder.endpoint != endpoint {
		t.Errorf("expected endpoint %q, got %q", endpoint, builder.endpoint)
	}

	if builder.timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", builder.timeout)
	}

	if builder.userAgent != "kessel-go-sdk" {
		t.Errorf("expected userAgent 'kessel-go-sdk', got %q", builder.userAgent)
	}

	if builder.insecure {
		t.Error("expected insecure to be false by default")
	}

	if builder.maxIdleConns != 100 {
		t.Errorf("expected maxIdleConns 100, got %d", builder.maxIdleConns)
	}

	if builder.idleConnTimeout != 90*time.Second {
		t.Errorf("expected idleConnTimeout 90s, got %v", builder.idleConnTimeout)
	}

	if builder.interceptors == nil {
		t.Error("expected interceptors to be initialized")
	}

	if len(builder.interceptors) != 0 {
		t.Errorf("expected empty interceptors, got %d", len(builder.interceptors))
	}

	if builder.headers == nil {
		t.Error("expected headers to be initialized")
	}

	if len(builder.headers) != 0 {
		t.Errorf("expected empty headers, got %d", len(builder.headers))
	}
}

func TestNewClientBuilderFromConfig(t *testing.T) {
	cfg := &config.HTTPConfig{
		BaseConfig: config.BaseConfig{
			Endpoint: "https://api.example.com",
			Insecure: true,
			TLSConfig: &tls.Config{
				ServerName: "example.com",
			},
		},
		Timeout:         30 * time.Second,
		UserAgent:       "custom-agent/1.0",
		MaxIdleConns:    200,
		IdleConnTimeout: 120 * time.Second,
	}

	builder := NewClientBuilderFromConfig(cfg)

	if builder.endpoint != cfg.Endpoint {
		t.Errorf("expected endpoint %q, got %q", cfg.Endpoint, builder.endpoint)
	}

	if builder.timeout != cfg.Timeout {
		t.Errorf("expected timeout %v, got %v", cfg.Timeout, builder.timeout)
	}

	if builder.userAgent != cfg.UserAgent {
		t.Errorf("expected userAgent %q, got %q", cfg.UserAgent, builder.userAgent)
	}

	if builder.insecure != cfg.Insecure {
		t.Errorf("expected insecure %v, got %v", cfg.Insecure, builder.insecure)
	}

	if builder.tlsConfig != cfg.TLSConfig {
		t.Error("expected tlsConfig to match config")
	}

	if builder.maxIdleConns != cfg.MaxIdleConns {
		t.Errorf("expected maxIdleConns %d, got %d", cfg.MaxIdleConns, builder.maxIdleConns)
	}

	if builder.idleConnTimeout != cfg.IdleConnTimeout {
		t.Errorf("expected idleConnTimeout %v, got %v", cfg.IdleConnTimeout, builder.idleConnTimeout)
	}

	if builder.interceptors == nil {
		t.Error("expected interceptors to be initialized")
	}

	if builder.headers == nil {
		t.Error("expected headers to be initialized")
	}
}

func TestClientBuilder_WithTimeout(t *testing.T) {
	builder := NewClientBuilder("https://api.example.com")
	expectedTimeout := 60 * time.Second

	result := builder.WithTimeout(expectedTimeout)

	if result != builder {
		t.Error("expected WithTimeout to return the same builder instance")
	}

	if builder.timeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, builder.timeout)
	}
}

func TestClientBuilder_WithUserAgent(t *testing.T) {
	builder := NewClientBuilder("https://api.example.com")
	expectedUserAgent := "custom-agent/2.0"

	result := builder.WithUserAgent(expectedUserAgent)

	if result != builder {
		t.Error("expected WithUserAgent to return the same builder instance")
	}

	if builder.userAgent != expectedUserAgent {
		t.Errorf("expected userAgent %q, got %q", expectedUserAgent, builder.userAgent)
	}
}

func TestClientBuilder_WithTLSConfig(t *testing.T) {
	builder := NewClientBuilder("https://api.example.com")
	builder.insecure = true // Set to true initially

	tlsConfig := &tls.Config{
		ServerName: "example.com",
	}

	result := builder.WithTLSConfig(tlsConfig)

	if result != builder {
		t.Error("expected WithTLSConfig to return the same builder instance")
	}

	if builder.insecure {
		t.Error("expected insecure to be false after WithTLSConfig")
	}

	if builder.tlsConfig != tlsConfig {
		t.Error("expected tlsConfig to be set")
	}
}

func TestClientBuilder_WithInsecure(t *testing.T) {
	builder := NewClientBuilder("https://api.example.com")

	if builder.insecure {
		t.Error("expected insecure to be false initially")
	}

	result := builder.WithInsecure()

	if result != builder {
		t.Error("expected WithInsecure to return the same builder instance")
	}

	if !builder.insecure {
		t.Error("expected insecure to be true after WithInsecure")
	}

	if builder.tlsConfig == nil {
		t.Error("expected tlsConfig to be set after WithInsecure")
	}

	if !builder.tlsConfig.InsecureSkipVerify {
		t.Error("expected tlsConfig.InsecureSkipVerify to be true")
	}
}

func TestClientBuilder_Build(t *testing.T) {
	builder := NewClientBuilder("https://httpbin.org/get")
	ctx := context.Background()

	client, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if client == nil {
		t.Fatal("expected client but got nil")
	}

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Fatalf("failed to close client: %v", closeErr)
		}
	}()
}
