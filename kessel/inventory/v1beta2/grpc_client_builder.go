package v1beta2

import (
	"crypto/tls"
	"fmt"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryClient represents a simple wrapper around the gRPC client with connection lifecycle management
type InventoryClient struct {
	KesselInventoryServiceClient
	conn *grpc.ClientConn
}

// Close closes the underlying gRPC connection
func (c *InventoryClient) Close() error {
	return c.conn.Close()
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

// Build creates the final InventoryClient with the configured options
func (b *InventoryGRPCClientBuilder) Build() (*InventoryClient, error) {
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

	// Add OAuth2 credentials directly to connection if enabled
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

		tokenClient, err := auth.NewTokenSource(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create token source: %w", err)
		}

		// Add per-RPC credentials for automatic token injection
		if b.insecure {
			opts = append(opts, grpc.WithPerRPCCredentials(tokenClient.GetInsecureGRPCCredentials()))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(tokenClient.GetGRPCCredentials()))
		}
	}

	// Add default call options for message sizes
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(b.maxReceiveMessageSize),
		grpc.MaxCallSendMsgSize(b.maxSendMessageSize),
	))

	// Create gRPC connection
	conn, err := grpc.NewClient(b.endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return &InventoryClient{
		KesselInventoryServiceClient: NewKesselInventoryServiceClient(conn),
		conn:                         conn,
	}, nil
}
