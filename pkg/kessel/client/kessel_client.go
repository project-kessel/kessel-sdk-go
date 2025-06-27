package client

import (
	"context"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	kesselgrpc "github.com/project-kessel/kessel-sdk-go/pkg/kessel/grpc"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/http"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
	grpc "google.golang.org/grpc"
)

type KesselClient interface {
	Check(ctx context.Context, in *v1beta2.CheckRequest) (*v1beta2.CheckResponse, error)
	CheckForUpdate(ctx context.Context, in *v1beta2.CheckForUpdateRequest) (*v1beta2.CheckForUpdateResponse, error)
	ReportResource(ctx context.Context, in *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error)
	DeleteResource(ctx context.Context, in *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error)
	StreamedListObjects(ctx context.Context, in *v1beta2.StreamedListObjectsRequest) (grpc.ServerStreamingClient[v1beta2.StreamedListObjectsResponse], error)
	Close() error
}

func NewClient(ctx context.Context, cfg *config.Config) (KesselClient, error) {
	if cfg.UseHTTP {
		return NewHTTPClient(ctx, cfg, createHTTPBuilder(cfg))
	}
	return NewGRPCClient(ctx, cfg, createGRPCBuilder(cfg))
}

func createGRPCBuilder(cfg *config.Config) *kesselgrpc.ClientBuilder {
	builder := kesselgrpc.NewClientBuilder(cfg.Endpoint)

	if cfg.Insecure {
		builder.WithInsecure()
	} else if cfg.TLSConfig != nil {
		builder.WithTransportSecurity(cfg.TLSConfig)
	}

	return builder
}

func createHTTPBuilder(cfg *config.Config) *http.ClientBuilder {
	builder := http.NewClientBuilder(cfg.Endpoint).
		WithTimeout(cfg.Timeout).
		WithUserAgent(cfg.UserAgent)

	if cfg.Insecure {
		builder.WithInsecure()
	} else if cfg.TLSConfig != nil {
		builder.WithTLSConfig(cfg.TLSConfig)
	}

	return builder
}
