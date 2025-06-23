package client

import (
	"context"
	"log"

	"github.com/project-kessel/kessel-sdk-go/_proto/kessel/inventory/v1beta2"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
	kesselgrpc "github.com/project-kessel/kessel-sdk-go/pkg/kessel/grpc"
	grpc "google.golang.org/grpc"
)

type GRPCClient struct {
	conn                   *grpc.ClientConn
	inventoryServiceClient v1beta2.KesselInventoryServiceClient
	tokenClient            *auth.TokenClient
	config                 *config.Config
}

func (c *GRPCClient) getTokenCallOption() ([]grpc.CallOption, error) {
	var opts []grpc.CallOption
	if c.config.EnableOIDCAuth {
		opts = append(opts, grpc.EmptyCallOption{})
		token, err := c.tokenClient.GetToken()
		if err != nil {
			return nil, err
		}
		if c.config.Insecure {
			opts = append(opts, auth.WithInsecureBearerToken(token.AccessToken))
		} else {
			opts = append(opts, auth.WithBearerToken(token.AccessToken))
		}
	}
	return opts, nil
}

// NewGRPCClient creates a new Inventory inventoryServiceClient client
func NewGRPCClient(ctx context.Context, cfg *config.Config, builder *kesselgrpc.ClientBuilder) (*GRPCClient, error) {
	conn, err := builder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection")
	}

	var tokenClient *auth.TokenClient
	if cfg.EnableOIDCAuth {
		tokenClient = auth.NewTokenClient(cfg)
	}

	return &GRPCClient{
		config:                 cfg,
		tokenClient:            tokenClient,
		conn:                   conn,
		inventoryServiceClient: v1beta2.NewKesselInventoryServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *GRPCClient) Check(ctx context.Context, in *v1beta2.CheckRequest) (*v1beta2.CheckResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.Wrap(err, "failed to get token"))
	}
	return c.inventoryServiceClient.Check(ctx, in, opts...)
}

func (c GRPCClient) CheckForUpdate(ctx context.Context, in *v1beta2.CheckForUpdateRequest) (*v1beta2.CheckForUpdateResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.Wrap(err, "failed to get token"))
	}
	return c.inventoryServiceClient.CheckForUpdate(ctx, in, opts...)
}

func (c GRPCClient) ReportResource(ctx context.Context, in *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.Wrap(err, "failed to get token"))
	}
	return c.inventoryServiceClient.ReportResource(ctx, in, opts...)
}

func (c GRPCClient) DeleteResource(ctx context.Context, in *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.Wrap(err, "failed to get token"))
	}
	return c.inventoryServiceClient.DeleteResource(ctx, in, opts...)
}

func (c GRPCClient) StreamedListObjects(ctx context.Context, in *v1beta2.StreamedListObjectsRequest) (grpc.ServerStreamingClient[v1beta2.StreamedListObjectsResponse], error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.Wrap(err, "failed to get token"))
	}
	return c.inventoryServiceClient.StreamedListObjects(ctx, in, opts...)
}
