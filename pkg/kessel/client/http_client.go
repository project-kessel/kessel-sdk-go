package client

import (
	"context"
	"fmt"
	"log"
	nethttp "net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/auth"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/http"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
	"google.golang.org/grpc"
)

type HTTPClient struct {
	client                     *khttp.Client
	inventoryServiceHTTPClient v1beta2.KesselInventoryServiceHTTPClient
	tokenSource                *auth.TokenSource
	config                     *config.HTTPConfig
}

func (h HTTPClient) getTokenHTTPOption() ([]khttp.CallOption, error) {
	var opts []khttp.CallOption
	if h.config.EnableOauth && h.tokenSource != nil {
		token, err := h.tokenSource.GetToken(context.Background())
		if err != nil {
			return nil, errors.NewTokenError(err, "failed to get OAuth2 token")
		}
		header := nethttp.Header{}
		header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		opts = append(opts, khttp.Header(&header))
	}
	return opts, nil
}

func (h HTTPClient) Check(ctx context.Context, in *v1beta2.CheckRequest) (*v1beta2.CheckResponse, error) {
	opts, err := h.getTokenHTTPOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return h.inventoryServiceHTTPClient.Check(ctx, in, opts...)
}

func (h HTTPClient) CheckForUpdate(ctx context.Context, in *v1beta2.CheckForUpdateRequest) (*v1beta2.CheckForUpdateResponse, error) {
	opts, err := h.getTokenHTTPOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return h.inventoryServiceHTTPClient.CheckForUpdate(ctx, in, opts...)
}

func (h HTTPClient) ReportResource(ctx context.Context, in *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error) {
	opts, err := h.getTokenHTTPOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return h.inventoryServiceHTTPClient.ReportResource(ctx, in, opts...)
}

func (h HTTPClient) DeleteResource(ctx context.Context, in *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error) {
	opts, err := h.getTokenHTTPOption()
	if err != nil {
		log.Printf("error: %v", errors.NewTokenError(err, "failed to get token"))
	}
	return h.inventoryServiceHTTPClient.DeleteResource(ctx, in, opts...)
}

func (h HTTPClient) StreamedListObjects(ctx context.Context, in *v1beta2.StreamedListObjectsRequest) (grpc.ServerStreamingClient[v1beta2.StreamedListObjectsResponse], error) {
	panic("Use grpc for streamed list objects")
}

func (h HTTPClient) Close() error {
	return h.client.Close()
}

// NewHTTPClient creates a new HTTP client for the Kessel inventory service
func NewHTTPClient(ctx context.Context, cfg *config.HTTPConfig, builder *http.ClientBuilder) (*HTTPClient, error) {
	client, err := builder.Build(ctx)
	if err != nil {
		return nil, errors.NewHTTPClientError(err, "failed to create HTTP client")
	}

	var tokenSource *auth.TokenSource
	if cfg.EnableOauth {
		tokenSource, err = auth.NewTokenSource(cfg)
		if err != nil {
			return nil, errors.NewTokenError(err, "failed to create OAuth2 token source")
		}
	}

	return &HTTPClient{
		config:                     cfg,
		client:                     client,
		tokenSource:                tokenSource,
		inventoryServiceHTTPClient: v1beta2.NewKesselInventoryServiceHTTPClient(client),
	}, nil
}
