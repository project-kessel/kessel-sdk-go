package client

import (
	"context"
	"log"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
	kesselgrpc "github.com/project-kessel/kessel-sdk-go/pkg/kessel/grpc"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
	grpc "google.golang.org/grpc"
)

type GRPCClient struct {
	conn                   *grpc.ClientConn
	inventoryServiceClient v1beta2.KesselInventoryServiceClient
	tokenSource            *auth.TokenSource
	config                 *config.GRPCConfig
}

func (c *GRPCClient) getTokenCallOption() ([]grpc.CallOption, error) {
	var opts []grpc.CallOption
	if c.config.EnableOauth && c.tokenSource != nil {
		if c.config.Insecure {
			opts = append(opts, c.tokenSource.GetInsecureCallOption())
		} else {
			opts = append(opts, c.tokenSource.GetCallOption())
		}
	}
	return opts, nil
}

// NewGRPCClient creates a new Inventory inventoryServiceClient client
func NewGRPCClient(ctx context.Context, cfg *config.GRPCConfig, builder *kesselgrpc.ClientBuilder) (*GRPCClient, error) {
	conn, err := builder.Build()
	if err != nil {
		return nil, errors.NewConnectionError(err, "failed to create gRPC connection")
	}

	var tokenSource *auth.TokenSource
	if cfg.EnableOauth {
		tokenSource, err = auth.NewTokenSource(cfg)
		if err != nil {
			if closeErr := conn.Close(); closeErr != nil {
				log.Printf("Failed to close connection: %v", closeErr)
			}
			return nil, errors.NewTokenError(err, "failed to create OAuth2 token source")
		}
	}

	return &GRPCClient{
		config:                 cfg,
		tokenSource:            tokenSource,
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
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return c.inventoryServiceClient.Check(ctx, in, opts...)
}

func (c GRPCClient) CheckForUpdate(ctx context.Context, in *v1beta2.CheckForUpdateRequest) (*v1beta2.CheckForUpdateResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return c.inventoryServiceClient.CheckForUpdate(ctx, in, opts...)
}

func (c GRPCClient) ReportResource(ctx context.Context, in *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return c.inventoryServiceClient.ReportResource(ctx, in, opts...)
}

func (c GRPCClient) DeleteResource(ctx context.Context, in *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return c.inventoryServiceClient.DeleteResource(ctx, in, opts...)
}

func (c GRPCClient) StreamedListObjects(ctx context.Context, in *v1beta2.StreamedListObjectsRequest) (grpc.ServerStreamingClient[v1beta2.StreamedListObjectsResponse], error) {
	opts, err := c.getTokenCallOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return c.inventoryServiceClient.StreamedListObjects(ctx, in, opts...)
}
