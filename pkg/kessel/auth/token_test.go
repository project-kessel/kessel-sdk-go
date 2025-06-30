package auth

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
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
			name: "missing token URL",
			config: &mockOAuthConfig{
				enableOauth: true,
				oauth2: config.Oauth2{
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					TokenURL:     "",
				},
			},
			expectError: true,
			errorCode:   codes.InvalidArgument,
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

func TestTokenSource_WithHTTPConfig(t *testing.T) {
	// Test with HTTP config structure
	cfg := &config.HTTPConfig{
		BaseConfig: config.BaseConfig{
			EnableOauth: true,
			Oauth2: config.Oauth2{
				ClientID:     "http-test-client",
				ClientSecret: "http-test-secret",
				TokenURL:     "https://auth.example.com/token",
				Scopes:       []string{"api:read", "api:write"},
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
