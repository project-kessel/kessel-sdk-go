package config

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"
)

func TestNewGRPCConfig(t *testing.T) {
	tests := []struct {
		name     string
		options  []GRPCClientOption
		validate func(*testing.T, *GRPCConfig)
	}{
		{
			name:    "default config",
			options: []GRPCClientOption{},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if cfg.Insecure {
					t.Error("expected Insecure to be false by default")
				}
			},
		},
		{
			name: "with endpoint",
			options: []GRPCClientOption{
				WithGRPCEndpoint("localhost:9090"),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if cfg.Endpoint != "localhost:9090" {
					t.Errorf("expected Endpoint to be 'localhost:9090', got %q", cfg.Endpoint)
				}
			},
		},
		{
			name: "with insecure",
			options: []GRPCClientOption{
				WithGRPCInsecure(true),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if !cfg.Insecure {
					t.Error("expected Insecure to be true")
				}
			},
		},
		{
			name: "with TLS config",
			options: []GRPCClientOption{
				WithGRPCTLSConfig(&tls.Config{ServerName: "example.com"}),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if cfg.Insecure {
					t.Error("expected Insecure to be false when TLS config is set")
				}
				if cfg.TLSConfig == nil {
					t.Error("expected TLSConfig to be set")
				}
				if cfg.TLSConfig.ServerName != "example.com" {
					t.Errorf("expected ServerName to be 'example.com', got %q", cfg.TLSConfig.ServerName)
				}
			},
		},

		{
			name: "with message sizes",
			options: []GRPCClientOption{
				WithGRPCMaxReceiveMessageSize(8 * 1024 * 1024),
				WithGRPCMaxSendMessageSize(8 * 1024 * 1024),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				expected := 8 * 1024 * 1024
				if cfg.MaxReceiveMessageSize != expected {
					t.Errorf("expected MaxReceiveMessageSize to be %d, got %d", expected, cfg.MaxReceiveMessageSize)
				}
				if cfg.MaxSendMessageSize != expected {
					t.Errorf("expected MaxSendMessageSize to be %d, got %d", expected, cfg.MaxSendMessageSize)
				}
			},
		},
		{
			name: "with OAuth2",
			options: []GRPCClientOption{
				WithGRPCOAuth2("client-id", "client-secret", "https://auth.example.com/token", "scope1", "scope2"),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if !cfg.EnableOauth {
					t.Error("expected EnableOauth to be true")
				}
				if cfg.Oauth2.ClientID != "client-id" {
					t.Errorf("expected ClientID to be 'client-id', got %q", cfg.Oauth2.ClientID)
				}
				if cfg.Oauth2.ClientSecret != "client-secret" {
					t.Errorf("expected ClientSecret to be 'client-secret', got %q", cfg.Oauth2.ClientSecret)
				}
				if cfg.Oauth2.TokenURL != "https://auth.example.com/token" {
					t.Errorf("expected TokenURL to be 'https://auth.example.com/token', got %q", cfg.Oauth2.TokenURL)
				}
				expectedScopes := []string{"scope1", "scope2"}
				if len(cfg.Oauth2.Scopes) != len(expectedScopes) {
					t.Errorf("expected %d scopes, got %d", len(expectedScopes), len(cfg.Oauth2.Scopes))
				}
				for i, scope := range expectedScopes {
					if i < len(cfg.Oauth2.Scopes) && cfg.Oauth2.Scopes[i] != scope {
						t.Errorf("expected scope[%d] to be %q, got %q", i, scope, cfg.Oauth2.Scopes[i])
					}
				}
			},
		},
		{
			name: "with OAuth2 Issuer",
			options: []GRPCClientOption{
				WithGRPCOAuth2Issuer("issuer-client-id", "issuer-client-secret", "https://auth.example.com", "scope1", "scope2"),
			},
			validate: func(t *testing.T, cfg *GRPCConfig) {
				if !cfg.EnableOauth {
					t.Error("expected EnableOauth to be true")
				}
				if cfg.Oauth2.ClientID != "issuer-client-id" {
					t.Errorf("expected ClientID to be 'issuer-client-id', got %q", cfg.Oauth2.ClientID)
				}
				if cfg.Oauth2.ClientSecret != "issuer-client-secret" {
					t.Errorf("expected ClientSecret to be 'issuer-client-secret', got %q", cfg.Oauth2.ClientSecret)
				}
				if cfg.Oauth2.IssuerURL != "https://auth.example.com" {
					t.Errorf("expected IssuerURL to be 'https://auth.example.com', got %q", cfg.Oauth2.IssuerURL)
				}
				expectedScopes := []string{"scope1", "scope2"}
				if len(cfg.Oauth2.Scopes) != len(expectedScopes) {
					t.Errorf("expected %d scopes, got %d", len(expectedScopes), len(cfg.Oauth2.Scopes))
				}
				for i, scope := range expectedScopes {
					if i < len(cfg.Oauth2.Scopes) && cfg.Oauth2.Scopes[i] != scope {
						t.Errorf("expected scope[%d] to be %q, got %q", i, scope, cfg.Oauth2.Scopes[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewGRPCConfig(tt.options...)
			tt.validate(t, cfg)
		})
	}
}



func TestGRPCConfig_OAuthInterface(t *testing.T) {
	cfg := NewGRPCConfig(
		WithGRPCOAuth2("test-client", "test-secret", "https://auth.example.com/token", "scope1"),
	)

	if !cfg.GetEnableOauth() {
		t.Error("expected GetEnableOauth to return true")
	}

	oauth2 := cfg.GetOauth2()
	if oauth2.ClientID != "test-client" {
		t.Errorf("expected ClientID to be 'test-client', got %q", oauth2.ClientID)
	}
	if oauth2.ClientSecret != "test-secret" {
		t.Errorf("expected ClientSecret to be 'test-secret', got %q", oauth2.ClientSecret)
	}
	if oauth2.TokenURL != "https://auth.example.com/token" {
		t.Errorf("expected TokenURL to be 'https://auth.example.com/token', got %q", oauth2.TokenURL)
	}
	if len(oauth2.Scopes) != 1 || oauth2.Scopes[0] != "scope1" {
		t.Errorf("expected Scopes to be ['scope1'], got %v", oauth2.Scopes)
	}
}



func TestOauth2_EmptyValues(t *testing.T) {
	cfg := NewGRPCConfig()

	if cfg.GetEnableOauth() {
		t.Error("expected GetEnableOauth to return false for default config")
	}

	oauth2 := cfg.GetOauth2()
	if oauth2.ClientID != "" {
		t.Errorf("expected empty ClientID, got %q", oauth2.ClientID)
	}
	if oauth2.ClientSecret != "" {
		t.Errorf("expected empty ClientSecret, got %q", oauth2.ClientSecret)
	}
	if oauth2.TokenURL != "" {
		t.Errorf("expected empty TokenURL, got %q", oauth2.TokenURL)
	}
	if len(oauth2.Scopes) != 0 {
		t.Errorf("expected empty Scopes, got %v", oauth2.Scopes)
	}
}

func TestOauth2_DiscoverTokenEndpoint(t *testing.T) {
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

func TestMultipleOptions(t *testing.T) {
	cfg := NewGRPCConfig(
		WithGRPCEndpoint("localhost:8080"),
		WithGRPCInsecure(true),
		WithGRPCMaxReceiveMessageSize(16*1024*1024),
		WithGRPCMaxSendMessageSize(16*1024*1024),
		WithGRPCOAuth2("multi-client", "multi-secret", "https://multi.example.com/token", "scope1", "scope2", "scope3"),
	)

	// Verify all options were applied
	if cfg.Endpoint != "localhost:8080" {
		t.Errorf("expected Endpoint to be 'localhost:8080', got %q", cfg.Endpoint)
	}
	if !cfg.Insecure {
		t.Error("expected Insecure to be true")
	}
	expectedSize := 16 * 1024 * 1024
	if cfg.MaxReceiveMessageSize != expectedSize {
		t.Errorf("expected MaxReceiveMessageSize to be %d, got %d", expectedSize, cfg.MaxReceiveMessageSize)
	}
	if cfg.MaxSendMessageSize != expectedSize {
		t.Errorf("expected MaxSendMessageSize to be %d, got %d", expectedSize, cfg.MaxSendMessageSize)
	}
	if !cfg.EnableOauth {
		t.Error("expected EnableOauth to be true")
	}
	if cfg.Oauth2.ClientID != "multi-client" {
		t.Errorf("expected ClientID to be 'multi-client', got %q", cfg.Oauth2.ClientID)
	}
	if len(cfg.Oauth2.Scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(cfg.Oauth2.Scopes))
	}
}
