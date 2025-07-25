package config

import (
	"context"
	"strings"
	"testing"
)

func TestNewGRPCConfig_Basic(t *testing.T) {
	// Test basic configuration creation
	cfg := NewGRPCConfig()

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
	cfg := NewGRPCConfig(
		WithGRPCEndpoint("localhost:8080"),
		WithGRPCInsecure(true),
		WithGRPCOAuth2("client-id", "client-secret", "https://auth.example.com/token", "scope1", "scope2"),
	)

	// Verify all options were applied
	if cfg.Endpoint != "localhost:8080" {
		t.Errorf("expected Endpoint to be 'localhost:8080', got %q", cfg.Endpoint)
	}
	if !cfg.Insecure {
		t.Error("expected Insecure to be true")
	}
	if !cfg.EnableOauth {
		t.Error("expected EnableOauth to be true")
	}
	if cfg.Oauth2.ClientID != "client-id" {
		t.Errorf("expected ClientID to be 'client-id', got %q", cfg.Oauth2.ClientID)
	}
	if len(cfg.Oauth2.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(cfg.Oauth2.Scopes))
	}
}

func TestOauth2_DiscoverTokenEndpoint_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		oauth2      *Oauth2
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing issuer URL",
			oauth2: &Oauth2{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			expectError: true,
			errorMsg:    "issuer_url is required",
		},
		{
			name: "invalid issuer URL",
			oauth2: &Oauth2{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				IssuerURL:    "https://invalid-issuer.example.com",
			},
			expectError: true,
			errorMsg:    "failed to fetch discovery document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.oauth2.DiscoverTokenEndpoint(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
