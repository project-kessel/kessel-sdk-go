package auth

import (
	"context"
	"fmt"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// OAuthConfig interface for extracting OAuth configuration from different config types
type OAuthConfig interface {
	GetEnableOauth() bool
	GetOauth2() config.Oauth2
}

// TokenSource wraps oauth2.TokenSource for easier testing and management
type TokenSource struct {
	source oauth2.TokenSource
}

// NewTokenSource creates a new OAuth2 token source using client credentials flow
func NewTokenSource(cfg OAuthConfig) (*TokenSource, error) {
	oauth2Config := cfg.GetOauth2()
	if oauth2Config.ClientID == "" || oauth2Config.ClientSecret == "" || oauth2Config.TokenURL == "" {
		return nil, errors.New(errors.ErrTokenRetrieval,
			codes.InvalidArgument,
			"OAuth2 configuration incomplete: client_id, client_secret, and token_url are required")
	}

	clientCredConfig := &clientcredentials.Config{
		ClientID:     oauth2Config.ClientID,
		ClientSecret: oauth2Config.ClientSecret,
		TokenURL:     oauth2Config.TokenURL,
		Scopes:       oauth2Config.Scopes,
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

// GetCallOption returns a gRPC call option with OAuth2 credentials
func (ts *TokenSource) GetCallOption() grpc.CallOption {
	return grpc.PerRPCCredentials(ts.GetGRPCCredentials())
}

// GetInsecureCallOption returns a gRPC call option with OAuth2 credentials for insecure connections
func (ts *TokenSource) GetInsecureCallOption() grpc.CallOption {
	// For insecure connections, we create a custom credentials that doesn't require transport security
	return grpc.PerRPCCredentials(&insecureOAuthCreds{
		tokenSource: ts.source,
	})
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
