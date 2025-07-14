package auth

import (
	"context"
	stderrors "errors"
	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
)

// mockOAuthConfig implements OAuthConfig interface for testing
type mockOAuthConfig struct {
	enableOauth bool
	oauth2      config.Oauth2
}

func (m *mockOAuthConfig) GetEnableOauth() bool {
	return m.enableOauth
}

func (m *mockOAuthConfig) GetOauth2() config.Oauth2 {
	return m.oauth2
}

// mockTokenSource implements oauth2.TokenSource for testing
type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func TestNewTokenSource(t *testing.T) {
	tests := []struct {
		name        string
		config      OAuthConfig
		expectError bool
		errorCode   codes.Code
	}{
		{
			name: "valid config",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "https://auth.example.com/token",
					Scopes:       []string{"scope1", "scope2"},
				},
			},
			expectError: false,
		},
		{
			name: "missing client ID",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "",
					ClientSecret: "test-secret",
					TokenURL:     "https://auth.example.com/token",
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "missing client secret",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "",
					TokenURL:     "https://auth.example.com/token",
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "missing token URL and issuer URL",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "",
					IssuerURL:    "",
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "valid issuer URL (will fail with network error)",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "",
					IssuerURL:    "https://auth.example.com",
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name: "token URL takes precedence over issuer URL",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "https://auth.example.com/token",
					IssuerURL:    "https://auth.example.com",
					Scopes:       []string{"scope1", "scope2"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := NewTokenSource(tt.config)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}

				// Check if it's a custom error with the expected code
				if customErr, ok := err.(*errors.Error); ok {
					if customErr.Code != tt.errorCode {
						t.Errorf("expected error code %v, got %v", tt.errorCode, customErr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ts == nil {
				t.Fatal("expected TokenSource but got nil")
			}

			if ts.source == nil {
				t.Fatal("expected source to be set")
			}
		})
	}
}

func TestTokenSource_GetToken(t *testing.T) {
	tests := []struct {
		name        string
		mockSource  *mockTokenSource
		expectError bool
	}{
		{
			name: "successful token retrieval",
			mockSource: &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				},
				err: nil,
			},
			expectError: false,
		},
		{
			name: "token retrieval error",
			mockSource: &mockTokenSource{
				token: nil,
				err:   stderrors.New("network error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TokenSource{source: tt.mockSource}
			ctx := context.Background()

			token, err := ts.GetToken(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token == nil {
				t.Fatal("expected token but got nil")
			}

			if token.AccessToken != "test-access-token" {
				t.Errorf("expected access token 'test-access-token', got %q", token.AccessToken)
			}
		})
	}
}

func TestTokenSource_GetGRPCCredentials(t *testing.T) {
	mockSource := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
		err: nil,
	}

	ts := &TokenSource{source: mockSource}
	creds := ts.GetGRPCCredentials()

	if creds == nil {
		t.Fatal("expected credentials but got nil")
	}

	// Test that credentials require transport security
	if !creds.RequireTransportSecurity() {
		t.Error("expected credentials to require transport security")
	}
}

func TestTokenSource_GetCallOption(t *testing.T) {
	mockSource := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
		err: nil,
	}

	ts := &TokenSource{source: mockSource}
	opt := ts.GetCallOption()

	if opt == nil {
		t.Fatal("expected call option but got nil")
	}
}

func TestTokenSource_GetInsecureCallOption(t *testing.T) {
	mockSource := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
		err: nil,
	}

	ts := &TokenSource{source: mockSource}
	opt := ts.GetInsecureCallOption()

	if opt == nil {
		t.Fatal("expected call option but got nil")
	}
}

func TestInsecureOAuthCreds_GetRequestMetadata(t *testing.T) {
	tests := []struct {
		name        string
		mockSource  *mockTokenSource
		expectError bool
	}{
		{
			name: "successful metadata retrieval",
			mockSource: &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
				},
				err: nil,
			},
			expectError: false,
		},
		{
			name: "token retrieval error",
			mockSource: &mockTokenSource{
				token: nil,
				err:   stderrors.New("token error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &insecureOAuthCreds{tokenSource: tt.mockSource}
			ctx := context.Background()

			metadata, err := creds.GetRequestMetadata(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if metadata == nil {
				t.Fatal("expected metadata but got nil")
			}

			authHeader, exists := metadata["authorization"]
			if !exists {
				t.Fatal("expected authorization header but not found")
			}

			expectedHeader := "Bearer test-access-token"
			if authHeader != expectedHeader {
				t.Errorf("expected authorization header %q, got %q", expectedHeader, authHeader)
			}
		})
	}
}

func TestInsecureOAuthCreds_RequireTransportSecurity(t *testing.T) {
	creds := &insecureOAuthCreds{}

	if creds.RequireTransportSecurity() {
		t.Error("expected insecure credentials to not require transport security")
	}
}

func TestTokenSource_Integration(t *testing.T) {
	// Test with real config structure
	cfg := &config.GRPCConfig{
		BaseConfig: config.BaseConfig{
			EnableOauth: true,
			Oauth2: config.Oauth2{
				ClientID:     "integration-test-client",
				ClientSecret: "integration-test-secret",
				TokenURL:     "https://auth.example.com/token",
				Scopes:       []string{"read", "write"},
			},
		},
	}

	ts, err := NewTokenSource(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating token source: %v", err)
	}

	if ts == nil {
		t.Fatal("expected token source but got nil")
	}

	// Test that we can get credentials without errors
	creds := ts.GetGRPCCredentials()
	if creds == nil {
		t.Fatal("expected credentials but got nil")
	}

	insecureCreds := ts.GetInsecureCallOption()
	if insecureCreds == nil {
		t.Fatal("expected insecure call option but got nil")
	}
}

func TestTokenSource_WithIssuerConfig(t *testing.T) {
	// Test with gRPC config using issuer
	cfg := &config.GRPCConfig{
		BaseConfig: config.BaseConfig{
			EnableOauth: true,
			Oauth2: config.Oauth2{
				ClientID:     "issuer-test-client",
				ClientSecret: "issuer-test-secret",
				IssuerURL:    "https://auth.example.com",
				Scopes:       []string{"read", "write"},
			},
		},
	}

	// This should fail with network error since the issuer URL is fake
	_, err := NewTokenSource(cfg)
	if err == nil {
		t.Fatal("expected error with fake issuer URL but got none")
	}

	// Error should mention discovery failure
	if !strings.Contains(err.Error(), "failed to discover token endpoint") {
		t.Errorf("expected error message to mention discovery failure, got %q", err.Error())
	}
}

func TestTokenSource_WithBothTokenURLAndIssuer(t *testing.T) {
	// Test that TokenURL takes precedence over IssuerURL
	cfg := &config.GRPCConfig{
		BaseConfig: config.BaseConfig{
			EnableOauth: true,
			Oauth2: config.Oauth2{
				ClientID:     "both-test-client",
				ClientSecret: "both-test-secret",
				TokenURL:     "https://auth.example.com/token",
				IssuerURL:    "https://auth.example.com",
				Scopes:       []string{"read", "write"},
			},
		},
	}

	ts, err := NewTokenSource(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating token source: %v", err)
	}

	if ts == nil {
		t.Fatal("expected token source but got nil")
	}

	// Verify that we can get credentials without errors
	creds := ts.GetGRPCCredentials()
	if creds == nil {
		t.Fatal("expected credentials but got nil")
	}
}

func TestDiscoverTokenEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		issuerURL   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid issuer URL",
			issuerURL:   "https://invalid-issuer.example.com",
			expectError: true,
			errorMsg:    "failed to fetch discovery document",
		},
		{
			name:        "empty issuer URL",
			issuerURL:   "",
			expectError: true,
			errorMsg:    "failed to fetch discovery document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := discoverTokenEndpoint(ctx, tt.issuerURL)

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



func TestError_CreationWithNewTokenSource(t *testing.T) {
	// Test error creation from NewTokenSource
	cfg := &mockOAuthConfig{
		enableOauth: true,
		oauth2: config.Oauth2{
			ClientID:     "", // Missing client ID
			ClientSecret: "test-secret",
			TokenURL:     "https://auth.example.com/token",
		},
	}

	_, err := NewTokenSource(cfg)
	if err == nil {
		t.Fatal("expected error but got none")
	}

	// Test error message contains relevant information
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}

	// Test that the error message contains helpful information
	errorMsg := err.Error()
	if errorMsg != "" {
		// Just check that we have some error message
		t.Logf("Error message: %s", errorMsg)
	}
}
