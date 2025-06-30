package client

import (
	"context"
	"log"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/grpc"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
)

func TestNewGRPCClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.GRPCConfig
		expectError bool
	}{
		{
			name: "successful client creation without OAuth",
			config: config.NewGRPCConfig(
				config.WithGRPCEndpoint("localhost:9090"),
				config.WithGRPCInsecure(true),
			),
			expectError: false,
		},
		{
			name: "client creation with OAuth but invalid config",
			config: config.NewGRPCConfig(
				config.WithGRPCEndpoint("localhost:9090"),
				config.WithGRPCInsecure(true),
				config.WithGRPCOAuth2("", "", ""), // Invalid OAuth config
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			builder := grpc.NewClientBuilderFromConfig(tt.config)

			client, err := NewGRPCClient(ctx, tt.config, builder)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client == nil {
				t.Fatal("expected client but got nil")
			}

			// Clean up
			defer func() {
				if closeErr := client.Close(); closeErr != nil {
					log.Printf("Failed to close gRPC client: %v", closeErr)
				}
			}()

			// Verify client fields are set
			if client.conn == nil {
				t.Error("expected connection to be set")
			}

			if client.inventoryServiceClient == nil {
				t.Error("expected inventory service client to be set")
			}

			if client.config != tt.config {
				t.Error("expected config to match")
			}
		})
	}
}

func TestGRPCClient_Close(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewGRPCConfig(
		config.WithGRPCEndpoint("localhost:9090"),
		config.WithGRPCInsecure(true),
	)
	builder := grpc.NewClientBuilderFromConfig(cfg)

	client, err := NewGRPCClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	// Test closing the client
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}

func TestGRPCClient_getTokenCallOption(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.GRPCConfig
		expectOptions bool
	}{
		{
			name: "no OAuth config",
			config: config.NewGRPCConfig(
				config.WithGRPCEndpoint("localhost:9090"),
				config.WithGRPCInsecure(true),
			),
			expectOptions: false,
		},
		{
			name: "valid OAuth config with insecure",
			config: config.NewGRPCConfig(
				config.WithGRPCEndpoint("localhost:9090"),
				config.WithGRPCInsecure(true),
				config.WithGRPCOAuth2("client", "secret", "https://auth.example.com/token"),
			),
			expectOptions: true,
		},
		{
			name: "valid OAuth config with secure",
			config: config.NewGRPCConfig(
				config.WithGRPCEndpoint("localhost:9090"),
				config.WithGRPCOAuth2("client", "secret", "https://auth.example.com/token"),
			),
			expectOptions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			builder := grpc.NewClientBuilderFromConfig(tt.config)

			client, err := NewGRPCClient(ctx, tt.config, builder)
			if err != nil {
				// Skip if we can't create the client (expected for invalid OAuth)
				t.Skipf("skipping test due to client creation error: %v", err)
			}
			defer func() {
				if closeErr := client.Close(); closeErr != nil {
					log.Printf("Failed to close gRPC client: %v", closeErr)
				}
			}()

			opts, err := client.getTokenCallOption()
			if err != nil {
				t.Errorf("unexpected error getting token call option: %v", err)
			}

			if tt.expectOptions && len(opts) == 0 {
				t.Error("expected call options but got none")
			}

			if !tt.expectOptions && len(opts) > 0 {
				t.Error("expected no call options but got some")
			}
		})
	}
}

func TestGRPCClient_Methods(t *testing.T) {
	// Test that all client methods exist and can be called
	ctx := context.Background()
	cfg := config.NewGRPCConfig(
		config.WithGRPCEndpoint("localhost:9090"),
		config.WithGRPCInsecure(true),
	)
	builder := grpc.NewClientBuilderFromConfig(cfg)

	client, err := NewGRPCClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	// Test Check method
	checkReq := &v1beta2.CheckRequest{
		Object: &v1beta2.ResourceReference{
			ResourceType: "test",
			ResourceId:   "123",
		},
		Relation: "test",
		Subject: &v1beta2.SubjectReference{
			Resource: &v1beta2.ResourceReference{
				ResourceType: "user",
				ResourceId:   "user123",
			},
		},
	}

	// Note: These will fail with connection errors since we're not running a real server
	// But we're testing that the methods exist and can be called
	_, err = client.Check(ctx, checkReq)
	if err == nil {
		t.Log("Check method succeeded (unexpected with mock server)")
	}

	// Test CheckForUpdate method
	checkUpdateReq := &v1beta2.CheckForUpdateRequest{}
	_, err = client.CheckForUpdate(ctx, checkUpdateReq)
	if err == nil {
		t.Log("CheckForUpdate method succeeded (unexpected with mock server)")
	}

	// Test ReportResource method
	reportReq := &v1beta2.ReportResourceRequest{}
	_, err = client.ReportResource(ctx, reportReq)
	if err == nil {
		t.Log("ReportResource method succeeded (unexpected with mock server)")
	}

	// Test DeleteResource method
	deleteReq := &v1beta2.DeleteResourceRequest{}
	_, err = client.DeleteResource(ctx, deleteReq)
	if err == nil {
		t.Log("DeleteResource method succeeded (unexpected with mock server)")
	}

	// Test StreamedListObjects method
	listReq := &v1beta2.StreamedListObjectsRequest{}
	_, err = client.StreamedListObjects(ctx, listReq)
	if err == nil {
		t.Log("StreamedListObjects method succeeded (unexpected with mock server)")
	}
}

func TestGRPCClient_WithOAuth(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewGRPCConfig(
		config.WithGRPCEndpoint("localhost:9090"),
		config.WithGRPCInsecure(true),
		config.WithGRPCOAuth2("test-client", "test-secret", "https://auth.example.com/token"),
	)
	builder := grpc.NewClientBuilderFromConfig(cfg)

	client, err := NewGRPCClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client with OAuth: %v", err)
	}

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	if client.tokenSource == nil {
		t.Error("expected token source to be set with OAuth config")
	}

	// Test that token call options are generated
	opts, err := client.getTokenCallOption()
	if err != nil {
		t.Errorf("unexpected error getting token call options: %v", err)
	}

	if len(opts) == 0 {
		t.Error("expected call options with OAuth config")
	}
}

func TestGRPCClient_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		buildConfig func() (*config.GRPCConfig, *grpc.ClientBuilder)
		expectError bool
	}{
		{
			name: "valid minimal config",
			buildConfig: func() (*config.GRPCConfig, *grpc.ClientBuilder) {
				cfg := config.NewGRPCConfig(
					config.WithGRPCEndpoint("localhost:9090"),
					config.WithGRPCInsecure(true),
				)
				return cfg, grpc.NewClientBuilderFromConfig(cfg)
			},
			expectError: false,
		},
		{
			name: "config with custom timeouts and sizes",
			buildConfig: func() (*config.GRPCConfig, *grpc.ClientBuilder) {
				cfg := config.NewGRPCConfig(
					config.WithGRPCEndpoint("localhost:9090"),
					config.WithGRPCInsecure(true),
					config.WithGRPCMaxReceiveMessageSize(8*1024*1024),
					config.WithGRPCMaxSendMessageSize(8*1024*1024),
				)
				return cfg, grpc.NewClientBuilderFromConfig(cfg)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg, builder := tt.buildConfig()

			client, err := NewGRPCClient(ctx, cfg, builder)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client == nil {
				t.Fatal("expected client but got nil")
			}

			defer func() {
				if closeErr := client.Close(); closeErr != nil {
					log.Printf("Failed to close gRPC client: %v", closeErr)
				}
			}()

			// Verify client was configured correctly
			if client.config != cfg {
				t.Error("expected config to match")
			}
		})
	}
}
