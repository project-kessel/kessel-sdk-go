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

// OAuthConfig interface for extracting OAuth configuration from different config types
type OAuthConfig interface {
	GetEnableOIDCAuth() bool
	GetClientID() string
	GetClientSecret() string
	GetTokenURL() string
	GetIssuerURL() string
	GetScopes() []string
}

// TokenSource wraps oauth2.TokenSource for easier testing and management
type TokenSource struct {
	source oauth2.TokenSource
}

// OIDCDiscoveryDocument represents the OpenID Connect Discovery document
type OIDCDiscoveryDocument struct {
	TokenEndpoint string `json:"token_endpoint"`
	Issuer        string `json:"issuer"`
}

// discoverTokenEndpoint performs OpenID Connect Discovery to get the token endpoint
func discoverTokenEndpoint(ctx context.Context, issuerURL string) (string, error) {
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
		return "", fmt.Errorf("failed to create discovery request: %w", err)
	}

	// Add User-Agent header - some servers (like Keycloak) may reject requests without it
	req.Header.Set("User-Agent", "kessel-go-sdk/1.0")
	req.Header.Set("Accept", "*/*")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Failed to close connection: %v", closeErr)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discovery request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var doc OIDCDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("failed to decode discovery document: %w", err)
	}

	// Validate token endpoint
	if doc.TokenEndpoint == "" {
		return "", fmt.Errorf("token_endpoint not found in discovery document")
	}

	// Validate that token endpoint is a valid URL
	if _, err := url.Parse(doc.TokenEndpoint); err != nil {
		return "", fmt.Errorf("invalid token_endpoint URL: %w", err)
	}

	return doc.TokenEndpoint, nil
}

// NewTokenSource creates a new OAuth2 token source using client credentials flow
func NewTokenSource(cfg OAuthConfig) (*TokenSource, error) {
	if cfg.GetClientID() == "" || cfg.GetClientSecret() == "" {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			"OAuth2 configuration incomplete: client_id and client_secret are required")
	}

	// Determine token URL - either from direct config or issuer discovery
	var tokenURL string
	if cfg.GetTokenURL() != "" {
		// Use directly configured token URL
		tokenURL = cfg.GetTokenURL()
	} else if cfg.GetIssuerURL() != "" {
		// Discover token endpoint from issuer
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		discoveredTokenURL, err := discoverTokenEndpoint(ctx, cfg.GetIssuerURL())
		if err != nil {
			return nil, errors.New(errors.ErrTokenRetrieval,
				codes.InvalidArgument,
				fmt.Sprintf("failed to discover token endpoint from issuer %s: %v", cfg.GetIssuerURL(), err))
		}
		tokenURL = discoveredTokenURL
	} else {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			"OAuth2 configuration incomplete: either token_url or issuer_url must be provided")
	}

	clientCredConfig := &clientcredentials.Config{
		ClientID:     cfg.GetClientID(),
		ClientSecret: cfg.GetClientSecret(),
		TokenURL:     tokenURL,
		Scopes:       cfg.GetScopes(),
	}

	return &TokenSource{
		source: clientCredConfig.TokenSource(context.Background()),
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
