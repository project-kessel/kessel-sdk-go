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

// NewInventoryGRPCClient creates a new gRPC client for the Kessel Inventory service
func NewInventoryGRPCClient(cfg *config.GRPCConfig) (*InventoryGRPCClient, error) {
	var opts []grpc.DialOption

	var tokenClient *auth.TokenSource
	if cfg.EnableOauth {
		var err error
		tokenClient, err = auth.NewTokenSource(cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		var tlsConfig *tls.Config
		if cfg.TLSConfig != nil {
			tlsConfig = cfg.TLSConfig
		} else {
			tlsConfig = &tls.Config{}
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Add default call options for message sizes
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(cfg.MaxReceiveMessageSize),
		grpc.MaxCallSendMsgSize(cfg.MaxSendMessageSize),
	))

	conn, err := grpc.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, err
	}

	return &InventoryGRPCClient{
		KesselInventoryService: NewKesselInventoryServiceClient(conn),
		gRPCConn:               conn,
		tokenClient:            tokenClient,
		insecure:               cfg.Insecure,
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
