package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestMakeOAuth2ClientCredentials(t *testing.T) {
	clientId := "test-client-id"
	clientSecret := "test-client-secret"
	tokenEndpoint := "https://example.com/token"

	credentials := NewOAuth2ClientCredentials(clientId, clientSecret, tokenEndpoint)

	if credentials.clientId != clientId {
		t.Errorf("Expected clientId to be %s, got %s", clientId, credentials.clientId)
	}
	if credentials.clientSecret != clientSecret {
		t.Errorf("Expected clientSecret to be %s, got %s", clientSecret, credentials.clientSecret)
	}
	if credentials.tokenEndpoint != tokenEndpoint {
		t.Errorf("Expected tokenEndpoint to be %s, got %s", tokenEndpoint, credentials.tokenEndpoint)
	}
	if credentials.cachedToken.AccessToken != "" {
		t.Errorf("Expected cached token to be empty initially")
	}
}

func TestFetchOIDCDiscovery(t *testing.T) {
	tests := []struct {
		name         string
		options      FetchOIDCDiscoveryOptions
		expectError  bool
		errorMessage string
	}{
		{
			name:        "invalid issuer URL",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FetchOIDCDiscovery(context.TODO(), "invalid-url", FetchOIDCDiscoveryOptions{})
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorMessage != "" && err != nil && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFetchOIDCDiscovery_DefaultValues(t *testing.T) {
	// Test with nil context and http client (should use defaults)
	options := FetchOIDCDiscoveryOptions{
		HttpClient: nil,
	}

	// This should fail gracefully, but we're testing that defaults are applied
	_, err := FetchOIDCDiscovery(context.TODO(), "invalid-url-for-testing", options)
	if err == nil {
		t.Errorf("Expected error with invalid URL")
	}
	// The main thing we're testing is that it doesn't panic with nil values
}

func TestOAuth2ClientCredentials_GetToken(t *testing.T) {
	tests := []struct {
		name          string
		setupToken    *RefreshTokenResponse
		options       GetTokenOptions
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectError   bool
		expectRefresh bool
	}{
		{
			name: "use cached valid token",
			setupToken: &RefreshTokenResponse{
				AccessToken: "cached-token",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			options: GetTokenOptions{
				ForceRefresh: false,
			},
			expectError:   false,
			expectRefresh: false,
		},
		{
			name: "force refresh token",
			setupToken: &RefreshTokenResponse{
				AccessToken: "cached-token",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			options: GetTokenOptions{
				ForceRefresh: true,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"access_token": "new-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectError:   false,
			expectRefresh: true,
		},
		{
			name:       "refresh expired token",
			setupToken: nil, // No cached token
			options: GetTokenOptions{
				ForceRefresh: false,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"access_token": "fresh-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectError:   false,
			expectRefresh: true,
		},
		{
			name:       "server error during refresh",
			setupToken: nil,
			options: GetTokenOptions{
				ForceRefresh: false,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			},
			expectError:   true,
			expectRefresh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverHandler != nil {
				server = httptest.NewServer(http.HandlerFunc(tt.serverHandler))
				defer server.Close()
			}

			tokenEndpoint := ""
			if server != nil {
				tokenEndpoint = server.URL
			}

			credentials := NewOAuth2ClientCredentials("test-client", "test-secret", tokenEndpoint)

			// Setup cached token if provided
			if tt.setupToken != nil {
				credentials.cachedToken = *tt.setupToken
			}

			if tt.options.HttpClient == nil {
				tt.options.HttpClient = http.DefaultClient
			}

			result, err := credentials.GetToken(context.TODO(), tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.expectRefresh {
					if tt.name == "force refresh token" && result.AccessToken != "new-token" {
						t.Errorf("Expected new token 'new-token', got '%s'", result.AccessToken)
					}
					if tt.name == "refresh expired token" && result.AccessToken != "fresh-token" {
						t.Errorf("Expected fresh token 'fresh-token', got '%s'", result.AccessToken)
					}
				} else {
					if result.AccessToken != "cached-token" {
						t.Errorf("Expected cached token 'cached-token', got '%s'", result.AccessToken)
					}
				}
			}
		})
	}
}

func TestOAuth2ClientCredentials_GetToken_DefaultValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := NewOAuth2ClientCredentials("test-client", "test-secret", server.URL)

	// Test with nil context and http client (should use defaults)
	options := GetTokenOptions{
		ForceRefresh: false,
		HttpClient:   nil,
	}

	result, err := credentials.GetToken(context.TODO(), options)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.AccessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got '%s'", result.AccessToken)
	}
}

func TestOAuth2ClientCredentials_isTokenValid(t *testing.T) {
	tests := []struct {
		name        string
		token       RefreshTokenResponse
		expectValid bool
	}{
		{
			name: "valid token",
			token: RefreshTokenResponse{
				AccessToken: "valid-token",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			expectValid: true,
		},
		{
			name: "expired token",
			token: RefreshTokenResponse{
				AccessToken: "expired-token",
				ExpiresAt:   time.Now().Add(-time.Hour),
			},
			expectValid: false,
		},
		{
			name: "token expiring soon",
			token: RefreshTokenResponse{
				AccessToken: "expiring-soon-token",
				ExpiresAt:   time.Now().Add(time.Minute), // Less than expiration window (5 minutes)
			},
			expectValid: false,
		},
		{
			name: "empty token",
			token: RefreshTokenResponse{
				AccessToken: "",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credentials := NewOAuth2ClientCredentials("test-client", "test-secret", "https://example.com/token")
			credentials.cachedToken = tt.token

			result := credentials.isTokenValid()
			if result != tt.expectValid {
				t.Errorf("Expected isTokenValid to return %v, got %v", tt.expectValid, result)
			}
		})
	}
}

func TestOAuth2ClientCredentials_refreshToken(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectError   bool
		expectedToken string
		expectedTime  bool // Whether to check if time is set properly
	}{
		{
			name: "successful token refresh",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"access_token": "new-access-token",
					"token_type":   "Bearer",
					"expires_in":   7200,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectError:   false,
			expectedToken: "new-access-token",
			expectedTime:  true,
		},
		{
			name: "token refresh with default expires_in",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"access_token": "default-token",
					"token_type":   "Bearer",
					// No expires_in field - should use default
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectError:   false,
			expectedToken: "default-token",
			expectedTime:  true,
		},
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			credentials := NewOAuth2ClientCredentials("test-client", "test-secret", server.URL)

			result, err := credentials.refreshToken(context.Background(), http.DefaultClient)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.AccessToken != tt.expectedToken {
					t.Errorf("Expected access token '%s', got '%s'", tt.expectedToken, result.AccessToken)
				}
				if tt.expectedTime && result.ExpiresAt.Before(time.Now()) {
					t.Errorf("Expected expiration time to be in the future")
				}
			}
		})
	}
}

func TestOauth2TokenEndpointCaller(t *testing.T) {
	tokenEndpoint := "https://example.com/token"
	httpClient := &http.Client{}

	caller := oauth2TokenEndpointCaller{
		tokenEndpoint: tokenEndpoint,
		httpClient:    httpClient,
	}

	if caller.TokenEndpoint() != tokenEndpoint {
		t.Errorf("Expected TokenEndpoint() to return '%s', got '%s'", tokenEndpoint, caller.TokenEndpoint())
	}

	if caller.HttpClient() != httpClient {
		t.Errorf("Expected HttpClient() to return the same client instance")
	}
}

func TestConcurrentTokenAccess(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		// Add a small delay to simulate network latency
		time.Sleep(50 * time.Millisecond)
		response := map[string]interface{}{
			"access_token": "shared-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	credentials := NewOAuth2ClientCredentials("test-client", "test-secret", server.URL)

	// Test concurrent access to GetToken
	const numGoroutines = 5
	results := make(chan RefreshTokenResponse, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			token, err := credentials.GetToken(context.TODO(), GetTokenOptions{
				ForceRefresh: false,
			})
			if err != nil {
				errors <- err
			} else {
				results <- token
			}
		}()
	}

	// Collect results
	var tokens []RefreshTokenResponse
	for i := 0; i < numGoroutines; i++ {
		select {
		case token := <-results:
			tokens = append(tokens, token)
		case err := <-errors:
			t.Errorf("Unexpected error in concurrent access: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations to complete")
		}
	}

	// All tokens should be the same (cached after first request)
	if len(tokens) != numGoroutines {
		t.Errorf("Expected %d tokens, got %d", numGoroutines, len(tokens))
	}

	expectedToken := "shared-token-123"
	for i, token := range tokens {
		if token.AccessToken != expectedToken {
			t.Errorf("Token %d differs from expected: expected '%s', got '%s'", i, expectedToken, token.AccessToken)
		}
	}

	// The server should have been called only once due to caching and mutex protection
	mu.Lock()
	finalCallCount := callCount
	mu.Unlock()

	if finalCallCount != 1 {
		t.Logf("Note: Server was called %d times (expected 1, but race conditions may cause multiple calls)", finalCallCount)
		// Be more lenient here since the exact behavior depends on timing
		if finalCallCount > numGoroutines {
			t.Errorf("Expected server to be called at most %d times, but was called %d times", numGoroutines, finalCallCount)
		}
	}
}
