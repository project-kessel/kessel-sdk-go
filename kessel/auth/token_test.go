package auth

import (
	"strings"
	"testing"
)

func TestInsecureOAuthCreds_RequireTransportSecurity(t *testing.T) {
	// Test our custom insecure credentials logic
	creds := &insecureOAuthCreds{}

	if creds.RequireTransportSecurity() {
		t.Error("expected insecure credentials to not require transport security")
	}
}

func TestNewOAuth2ClientCredentials_MissingCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		tokenURL     string
		expectedErr  string
	}{
		{
			name:         "missing client ID",
			clientID:     "",
			clientSecret: "secret",
			tokenURL:     "https://auth.example.com/token",
			expectedErr:  "client_id and client_secret are required",
		},
		{
			name:         "missing client secret",
			clientID:     "client",
			clientSecret: "",
			tokenURL:     "https://auth.example.com/token",
			expectedErr:  "client_id and client_secret are required",
		},
		{
			name:         "missing token URL",
			clientID:     "client",
			clientSecret: "secret",
			tokenURL:     "",
			expectedErr:  "token_url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOAuth2ClientCredentials(tt.clientID, tt.clientSecret, tt.tokenURL)
			if err == nil {
				t.Fatal("expected error but got none")
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error to contain %v, got %v", tt.expectedErr, err.Error())
			}
		})
	}
}
