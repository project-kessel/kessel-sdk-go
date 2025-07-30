package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/project-kessel/kessel-sdk-go/kessel/errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// TokenSource wraps oauth2.TokenSource for easier testing and management
type TokenSource struct {
	source oauth2.TokenSource
}

// OIDCDiscoveryMetadata represents OIDC discovery metadata
type OIDCDiscoveryMetadata struct {
	document map[string]interface{}
}

// TokenEndpoint returns the token endpoint from the discovery document
func (m *OIDCDiscoveryMetadata) TokenEndpoint() string {
	if endpoint, ok := m.document["token_endpoint"].(string); ok {
		return endpoint
	}
	return ""
}

// FetchOIDCDiscovery fetches OIDC discovery metadata from some provider
//
// This function makes a network request to the OIDC provider's discovery endpoint
// to retrieve the provider's metadata including the token endpoint.
func FetchOIDCDiscovery(ctx context.Context, issuerURL string) (*OIDCDiscoveryMetadata, error) {
	// Ensure issuerURL doesn't end with a slash
	issuerURL = strings.TrimSuffix(issuerURL, "/")

	// Construct the well-known configuration URL
	//http://localhost:8085/realms/redhat-external/.well-known/openid-configuration
	discoveryURL := issuerURL + "/.well-known/openid-configuration"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	// Add User-Agent header - some servers (like Keycloak) may reject requests without it
	req.Header.Set("User-Agent", "kessel-go-sdk/1.0")
	req.Header.Set("Accept", "*/*")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Failed to close connection: %v", closeErr)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var document map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	return &OIDCDiscoveryMetadata{document: document}, nil
}

// OAuth2ClientCredentials handles the OAuth 2.0 Client Credentials flow
//
// This class only accepts a direct token URL. For OIDC discovery, use the
// FetchOIDCDiscovery function to obtain the token endpoint first.
type OAuth2ClientCredentials struct {
	*TokenSource
}

// NewOAuth2ClientCredentials creates a new OAuth2ClientCredentials instance
func NewOAuth2ClientCredentials(clientID, clientSecret, tokenURL string) (*OAuth2ClientCredentials, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			"OAuth2 configuration incomplete: client_id and client_secret are required")
	}

	if tokenURL == "" {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			"OAuth2 configuration incomplete: token_url is required")
	}

	// Validate token URL format
	if _, err := url.Parse(tokenURL); err != nil {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			fmt.Sprintf("invalid token_url: %v", err))
	}

	clientCredConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}

	tokenSource := &TokenSource{
		source: clientCredConfig.TokenSource(context.Background()),
	}

	return &OAuth2ClientCredentials{
		TokenSource: tokenSource,
	}, nil
}

// GetToken retrieves a valid OAuth2 token
func (ts *TokenSource) GetToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := ts.source.Token()
	if err != nil {
		return nil, errors.NewTokenError(err, "failed to retrieve OAuth2 token")
	}
	return token, nil
}

// GetGRPCCredentials returns gRPC credentials for OAuth2 authentication
func (ts *TokenSource) GetGRPCCredentials() credentials.PerRPCCredentials {
	return oauth.TokenSource{TokenSource: ts.source}
}

// GetInsecureGRPCCredentials returns gRPC credentials for OAuth2 authentication that don't require transport security
func (ts *TokenSource) GetInsecureGRPCCredentials() credentials.PerRPCCredentials {
	return &insecureOAuthCreds{
		tokenSource: ts.source,
	}
}

// GetCallOption returns a gRPC call option with OAuth2 credentials
func (ts *TokenSource) GetCallOption() grpc.CallOption {
	return grpc.PerRPCCredentials(ts.GetGRPCCredentials())
}

// GetInsecureCallOption returns a gRPC call option with OAuth2 credentials for insecure connections
func (ts *TokenSource) GetInsecureCallOption() grpc.CallOption {
	// For insecure connections, we create a custom credentials that doesn't require transport security
	return grpc.PerRPCCredentials(ts.GetInsecureGRPCCredentials())
}

// insecureOAuthCreds implements credentials.PerRPCCredentials for insecure connections
type insecureOAuthCreds struct {
	tokenSource oauth2.TokenSource
}

func (c *insecureOAuthCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (c *insecureOAuthCreds) RequireTransportSecurity() bool {
	return false
}
