package v1beta2

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryGRPCClient represents a gRPC client for the Kessel Inventory service
type InventoryGRPCClient struct {
	KesselInventoryService KesselInventoryServiceClient
	gRPCConn               *grpc.ClientConn
	tokenClient            *auth.TokenSource
	insecure               bool
}

// InventoryGRPCClientBuilder provides a fluent interface for building gRPC clients
type InventoryGRPCClientBuilder struct {
	endpoint              string
	insecure              bool
	tlsConfig             *tls.Config
	maxReceiveMessageSize int
	maxSendMessageSize    int
	enableOAuth           bool
	clientID              string
	clientSecret          string
	tokenURL              string
	issuerURL             string
	scopes                []string
	dialOptions           []grpc.DialOption
}

// NewInventoryGRPCClientBuilder creates a new builder with sensible defaults
func NewInventoryGRPCClientBuilder() *InventoryGRPCClientBuilder {
	return &InventoryGRPCClientBuilder{
		maxReceiveMessageSize: 4194304, // 4MB default
		maxSendMessageSize:    4194304, // 4MB default
		dialOptions:           []grpc.DialOption{},
	}
}

// WithEndpoint sets the server endpoint
func (b *InventoryGRPCClientBuilder) WithEndpoint(endpoint string) *InventoryGRPCClientBuilder {
	b.endpoint = endpoint
	return b
}

// WithInsecure enables insecure connections (no TLS)
func (b *InventoryGRPCClientBuilder) WithInsecure(insecure bool) *InventoryGRPCClientBuilder {
	b.insecure = insecure
	return b
}

// WithTLSConfig sets custom TLS configuration
func (b *InventoryGRPCClientBuilder) WithTLSConfig(tlsConfig *tls.Config) *InventoryGRPCClientBuilder {
	b.tlsConfig = tlsConfig
	b.insecure = false
	return b
}

// WithMaxReceiveMessageSize sets the maximum message size for receiving
func (b *InventoryGRPCClientBuilder) WithMaxReceiveMessageSize(size int) *InventoryGRPCClientBuilder {
	b.maxReceiveMessageSize = size
	return b
}

// WithMaxSendMessageSize sets the maximum message size for sending
func (b *InventoryGRPCClientBuilder) WithMaxSendMessageSize(size int) *InventoryGRPCClientBuilder {
	b.maxSendMessageSize = size
	return b
}

// WithOAuth2 configures OAuth2 authentication with direct token URL
func (b *InventoryGRPCClientBuilder) WithOAuth2(clientID, clientSecret, tokenURL string, scopes ...string) *InventoryGRPCClientBuilder {
	b.enableOAuth = true
	b.clientID = clientID
	b.clientSecret = clientSecret
	b.tokenURL = tokenURL
	b.scopes = scopes
	return b
}

// WithOAuth2Issuer configures OAuth2 authentication with issuer discovery
func (b *InventoryGRPCClientBuilder) WithOAuth2Issuer(clientID, clientSecret, issuerURL string, scopes ...string) *InventoryGRPCClientBuilder {
	b.enableOAuth = true
	b.clientID = clientID
	b.clientSecret = clientSecret
	b.issuerURL = issuerURL
	b.scopes = scopes
	return b
}

// WithDialOption adds a custom gRPC dial option
func (b *InventoryGRPCClientBuilder) WithDialOption(opt grpc.DialOption) *InventoryGRPCClientBuilder {
	b.dialOptions = append(b.dialOptions, opt)
	return b
}

// Build creates the final InventoryGRPCClient with the configured options
func (b *InventoryGRPCClientBuilder) Build() (*InventoryGRPCClient, error) {
	if b.endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	var opts []grpc.DialOption

	// Add custom dial options first
	opts = append(opts, b.dialOptions...)

	// Configure transport credentials
	if b.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		var tlsConfig *tls.Config
		if b.tlsConfig != nil {
			tlsConfig = b.tlsConfig
		} else {
			tlsConfig = &tls.Config{}
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Add default call options for message sizes
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(b.maxReceiveMessageSize),
		grpc.MaxCallSendMsgSize(b.maxSendMessageSize),
	))

	// Create OAuth token client if enabled
	var tokenClient *auth.TokenSource
	if b.enableOAuth {
		cfg := &config.GRPCConfig{
			BaseConfig: config.BaseConfig{
				Endpoint:    b.endpoint,
				Insecure:    b.insecure,
				TLSConfig:   b.tlsConfig,
				EnableOauth: true,
				Oauth2: config.Oauth2{
					ClientID:     b.clientID,
					ClientSecret: b.clientSecret,
					TokenURL:     b.tokenURL,
					IssuerURL:    b.issuerURL,
					Scopes:       b.scopes,
				},
			},
			MaxReceiveMessageSize: b.maxReceiveMessageSize,
			MaxSendMessageSize:    b.maxSendMessageSize,
		}

		var err error
		tokenClient, err = auth.NewTokenSource(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create token source: %w", err)
		}
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(b.endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return &InventoryGRPCClient{
		KesselInventoryService: NewKesselInventoryServiceClient(conn),
		gRPCConn:               conn,
		tokenClient:            tokenClient,
		insecure:               b.insecure,
	}, nil
}

// Close closes the gRPC connection
func (c *InventoryGRPCClient) Close() error {
	return c.gRPCConn.Close()
}

// GetCallOptions returns call options with OAuth2 credentials if configured
func (c *InventoryGRPCClient) GetCallOptions() []grpc.CallOption {
	if c.tokenClient == nil {
		return nil
	}
	return []grpc.CallOption{c.tokenClient.GetCallOption()}
}

// GetInsecureCallOptions returns call options with OAuth2 credentials for insecure connections
func (c *InventoryGRPCClient) GetInsecureCallOptions() []grpc.CallOption {
	if c.tokenClient == nil {
		return nil
	}
	return []grpc.CallOption{c.tokenClient.GetInsecureCallOption()}
}

// GetTokenCallOption returns call options with explicit token handling
func (c *InventoryGRPCClient) GetTokenCallOption() ([]grpc.CallOption, error) {
	if c.tokenClient == nil {
		return nil, nil
	}

	var opts []grpc.CallOption
	token, err := c.tokenClient.GetToken(context.Background())
	if err != nil {
		return nil, err
	}

	// Create per-RPC credentials with the token
	if c.insecure {
		opts = append(opts, grpc.PerRPCCredentials(&bearerToken{
			token:    token.AccessToken,
			insecure: true,
		}))
	} else {
		opts = append(opts, grpc.PerRPCCredentials(&bearerToken{
			token:    token.AccessToken,
			insecure: false,
		}))
	}

	return opts, nil
}

// bearerToken implements credentials.PerRPCCredentials for bearer token authentication
type bearerToken struct {
	token    string
	insecure bool
}

func (b *bearerToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", b.token),
	}, nil
}

func (b *bearerToken) RequireTransportSecurity() bool {
	return !b.insecure
}
