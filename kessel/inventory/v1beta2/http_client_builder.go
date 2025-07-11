package v1beta2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/kessel/config"
)

// InventoryHTTPClient represents an HTTP client for the Kessel Inventory service
type InventoryHTTPClient struct {
	KesselInventoryService KesselInventoryServiceHTTPClient
	httpClient             *kratoshttp.Client
	tokenClient            *auth.TokenSource
}

// NewInventoryHTTPClient creates a new HTTP client for the Kessel Inventory service
func NewInventoryHTTPClient(cfg *config.HTTPConfig) (*InventoryHTTPClient, error) {
	ctx := context.Background()

	var tokenClient *auth.TokenSource
	if cfg.EnableOauth {
		// Create a GRPCConfig for the auth.NewTokenSource since it expects that type
		grpcConfig := &config.GRPCConfig{
			BaseConfig: config.BaseConfig{
				Insecure:    cfg.Insecure,
				EnableOauth: cfg.EnableOauth,
				Oauth2:      cfg.Oauth2,
			},
		}
		var err error
		tokenClient, err = auth.NewTokenSource(grpcConfig)
		if err != nil {
			return nil, err
		}
	}

	opts := []kratoshttp.ClientOption{
		kratoshttp.WithEndpoint(cfg.Endpoint),
		kratoshttp.WithTimeout(cfg.Timeout),
	}

	// Add TLS configuration if provided
	if cfg.TLSConfig != nil {
		opts = append(opts, kratoshttp.WithTLSConfig(cfg.TLSConfig))
	} else if cfg.Insecure {
		opts = append(opts, kratoshttp.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	client, err := kratoshttp.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &InventoryHTTPClient{
		KesselInventoryService: NewKesselInventoryServiceHTTPClient(client),
		httpClient:             client,
		tokenClient:            tokenClient,
	}, nil
}

// Close closes the HTTP client (placeholder for interface compatibility)
func (c *InventoryHTTPClient) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// GetAuthHeader returns the Authorization header value if OAuth2 is configured
func (c *InventoryHTTPClient) GetAuthHeader() (string, error) {
	if c.tokenClient == nil {
		return "", nil
	}

	token, err := c.tokenClient.GetToken(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get OAuth2 token: %w", err)
	}

	return fmt.Sprintf("Bearer %s", token.AccessToken), nil
}

// GetCallOptions returns HTTP call options with OAuth2 authorization header if configured
func (c *InventoryHTTPClient) GetCallOptions() ([]kratoshttp.CallOption, error) {
	if c.tokenClient == nil {
		return nil, nil
	}

	authHeader, err := c.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	if authHeader == "" {
		return nil, nil
	}

	// Note: This is a placeholder - the actual HTTP CallOption for headers
	// would need to be implemented based on the Kratos HTTP client API
	return nil, nil
}

// GetTokenHTTPOption returns HTTP call options with explicit token handling
func (c *InventoryHTTPClient) GetTokenHTTPOption() ([]kratoshttp.CallOption, error) {
	if c.tokenClient == nil {
		return nil, nil
	}

	var opts []kratoshttp.CallOption
	token, err := c.tokenClient.GetToken(context.Background())
	if err != nil {
		return nil, err
	}

	// Create authorization header
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	opts = append(opts, kratoshttp.Header(&header))

	return opts, nil
}
