package config

import (
	"testing"
)

func TestNewGRPCConfig_Basic(t *testing.T) {
	// Test basic configuration creation
	cfg := NewCompatibilityConfig()

	// Verify defaults are set correctly
	if cfg.Insecure {
		t.Error("expected Insecure to be false by default")
	}
	if cfg.MaxReceiveMessageSize != 4*1024*1024 {
		t.Errorf("expected MaxReceiveMessageSize to be 4MB, got %d", cfg.MaxReceiveMessageSize)
	}
	if cfg.MaxSendMessageSize != 4*1024*1024 {
		t.Errorf("expected MaxSendMessageSize to be 4MB, got %d", cfg.MaxSendMessageSize)
	}
}

func TestNewGRPCConfig_WithOptions(t *testing.T) {
	// Test configuration with multiple options
	cfg := NewCompatibilityConfig(
		WithGRPCEndpoint("localhost:8080"),
		WithGRPCInsecure(true),
		WithGRPCMaxReceiveMessageSize(8*1024*1024),
		WithGRPCMaxSendMessageSize(8*1024*1024),
	)

	// Verify all options were applied
	if cfg.Url != "localhost:8080" {
		t.Errorf("expected Endpoint to be 'localhost:8080', got %v", cfg.Url)
	}
	if !cfg.Insecure {
		t.Error("expected Insecure to be true")
	}
	if cfg.MaxReceiveMessageSize != 8*1024*1024 {
		t.Errorf("expected MaxReceiveMessageSize to be 8MB, got %d", cfg.MaxReceiveMessageSize)
	}
	if cfg.MaxSendMessageSize != 8*1024*1024 {
		t.Errorf("expected MaxSendMessageSize to be 8MB, got %d", cfg.MaxSendMessageSize)
	}
}
