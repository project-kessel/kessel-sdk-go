package grpc

import (
	"context"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
)

func TestOAuth2CallCredentials(t *testing.T) {
	// Create real OAuth2ClientCredentials for basic tests
	authCreds := auth.NewOAuth2ClientCredentials("test-client", "test-secret", "https://example.com/token")
	credentials := OAuth2CallCredentials(&authCreds)

	// Test RequireTransportSecurity
	if !credentials.RequireTransportSecurity() {
		t.Error("Expected RequireTransportSecurity to return true")
	}
}

func TestCallCredentials_RequireTransportSecurity(t *testing.T) {
	// Create OAuth2ClientCredentials (we don't need a real one for this test)
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "https://example.com/token")
	credentials := OAuth2CallCredentials(&authCreds)

	if !credentials.RequireTransportSecurity() {
		t.Error("Expected RequireTransportSecurity to return true")
	}
}

func TestCallCredentials_GetRequestMetadata_ErrorHandling(t *testing.T) {
	// Test error handling when credentials return an error
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "invalid-url")
	credentials := OAuth2CallCredentials(&authCreds)

	callCreds, ok := credentials.(callCredentials)
	if !ok {
		t.Fatal("Expected callCredentials type")
	}

	// This should fail because of invalid URL
	metadata, err := callCreds.GetRequestMetadata(context.Background(), "https://example.com")

	if err == nil {
		t.Error("Expected error due to invalid token endpoint URL")
	}

	if metadata != nil {
		t.Errorf("Expected metadata to be nil on error, got %v", metadata)
	}
}

func TestCallCredentials_GetRequestMetadata_WithMultipleURIs(t *testing.T) {
	// Test that GetRequestMetadata can be called with multiple URIs
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "invalid-url")
	credentials := OAuth2CallCredentials(&authCreds)

	callCreds, ok := credentials.(callCredentials)
	if !ok {
		t.Fatal("Expected callCredentials type")
	}

	// Test with multiple URIs (variadic parameter) - this will fail due to invalid URL
	// but we're testing that the interface accepts variadic parameters
	_, err := callCreds.GetRequestMetadata(
		context.Background(),
		"https://service1.example.com",
		"https://service2.example.com",
		"https://service3.example.com",
	)

	// Should still get an error due to invalid token endpoint, but no panic
	if err == nil {
		t.Error("Expected error due to invalid token endpoint")
	}
}

func TestCallCredentials_GetRequestMetadata_WithContext(t *testing.T) {
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "invalid-url")
	credentials := OAuth2CallCredentials(&authCreds)

	callCreds, ok := credentials.(callCredentials)
	if !ok {
		t.Fatal("Expected callCredentials type")
	}

	// Test with a context that has a timeout
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := callCreds.GetRequestMetadata(ctx, "https://example.com")

	// Should get an error due to invalid token endpoint
	if err == nil {
		t.Error("Expected error due to invalid token endpoint")
	}
}

func TestCallCredentials_GetRequestMetadata_ContextCancellation(t *testing.T) {
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "invalid-url")
	credentials := OAuth2CallCredentials(&authCreds)

	callCreds, ok := credentials.(callCredentials)
	if !ok {
		t.Fatal("Expected callCredentials type")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := callCreds.GetRequestMetadata(ctx, "https://example.com")

	if err == nil {
		t.Error("Expected error due to context cancellation or invalid URL")
	}
}

// Integration test that shows how the OAuth2CallCredentials would be used
func TestOAuth2CallCredentials_Integration(t *testing.T) {
	// Create real OAuth2ClientCredentials
	authCreds := auth.NewOAuth2ClientCredentials(
		"test-client-id",
		"test-client-secret",
		"https://example.com/token",
	)

	// Create gRPC call credentials
	grpcCreds := OAuth2CallCredentials(&authCreds)

	// Verify it implements the expected interface
	if grpcCreds == nil {
		t.Error("Expected OAuth2CallCredentials to return non-nil credentials")
	}

	// Verify RequireTransportSecurity works
	if !grpcCreds.RequireTransportSecurity() {
		t.Error("Expected RequireTransportSecurity to return true")
	}

	// Note: We can't easily test GetRequestMetadata without mocking the HTTP client
	// as it would try to make real HTTP requests. The other tests cover that functionality.
}

func TestCallCredentials_TypeAssertion(t *testing.T) {
	authCreds := auth.NewOAuth2ClientCredentials("client", "secret", "https://example.com/token")
	grpcCreds := OAuth2CallCredentials(&authCreds)

	// Verify that the returned credentials can be cast back to callCredentials
	if callCreds, ok := grpcCreds.(callCredentials); !ok {
		t.Error("Expected OAuth2CallCredentials to return callCredentials type")
	} else {
		// Verify the internal structure
		if callCreds.credentials != &authCreds {
			t.Error("Expected callCredentials to contain the original auth credentials")
		}
	}
}
