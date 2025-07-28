package auth

import (
	"strings"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/kessel/config"
)

func TestTokenSource_IssuerDiscovery(t *testing.T) {
	// Test that issuer discovery fails with fake URL (this tests our custom logic)
	cfg := &config.CompatibilityConfig{
		EnableOIDCAuth: true,
		ClientID:       "test-client",
		ClientSecret:   "test-secret",
		IssuerURL:      "https://fake-issuer.example.com",
	}

	_, err := NewTokenSource(cfg)
	if err == nil {
		t.Fatal("expected error with fake issuer URL but got none")
	}

	if !strings.Contains(err.Error(), "failed to discover token endpoint") {
		t.Errorf("expected error message to mention discovery failure, got %q", err.Error())
	}
}

func TestTokenSource_TokenURLPrecedence(t *testing.T) {
	// Test that TokenURL takes precedence over IssuerURL
	cfg := &config.CompatibilityConfig{
		EnableOIDCAuth:     true,
		ClientID:           "test-client",
		ClientSecret:       "test-secret",
		AuthServerTokenUrl: "https://auth.example.com/token",
		IssuerURL:          "https://fake-issuer.example.com",
	}

	ts, err := NewTokenSource(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts == nil {
		t.Fatal("expected TokenSource but got nil")
	}
}

func TestInsecureOAuthCreds_RequireTransportSecurity(t *testing.T) {
	// Test our custom insecure credentials logic
	creds := &insecureOAuthCreds{}

	if creds.RequireTransportSecurity() {
		t.Error("expected insecure credentials to not require transport security")
	}
}
