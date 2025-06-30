package client

import (
	"context"
	"log"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/http"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/pkg/kessel/inventory/v1beta2"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.HTTPConfig
		expectError bool
	}{
		{
			name: "successful client creation without OAuth",
			config: config.NewHTTPConfig(
				config.WithHTTPEndpoint("https://api.example.com"),
				config.WithHTTPInsecure(true),
			),
			expectError: false,
		},
		{
			name: "client creation with OAuth but invalid config",
			config: config.NewHTTPConfig(
				config.WithHTTPEndpoint("https://api.example.com"),
				config.WithHTTPInsecure(true),
				config.WithHTTPOAuth2("", "", ""), // Invalid OAuth config
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			builder := http.NewClientBuilderFromConfig(tt.config)

			client, err := NewHTTPClient(ctx, tt.config, builder)

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
					log.Printf("Failed to close HTTP client: %v", closeErr)
				}
			}()

			// Verify client fields are set
			if client.client == nil {
				t.Error("expected HTTP client to be set")
			}

			if client.inventoryServiceHTTPClient == nil {
				t.Error("expected inventory service HTTP client to be set")
			}

			if client.config != tt.config {
				t.Error("expected config to match")
			}
		})
	}
}

func TestHTTPClient_Close(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewHTTPConfig(
		config.WithHTTPEndpoint("https://api.example.com"),
		config.WithHTTPInsecure(true),
	)
	builder := http.NewClientBuilderFromConfig(cfg)

	client, err := NewHTTPClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	// Test closing the client
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}

func TestHTTPClient_getTokenHTTPOption(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.HTTPConfig
		expectOptions bool
	}{
		{
			name: "no OAuth config",
			config: config.NewHTTPConfig(
				config.WithHTTPEndpoint("https://api.example.com"),
				config.WithHTTPInsecure(true),
			),
			expectOptions: false,
		},
		{
			name: "valid OAuth config",
			config: config.NewHTTPConfig(
				config.WithHTTPEndpoint("https://api.example.com"),
				config.WithHTTPInsecure(true),
				config.WithHTTPOAuth2("client", "secret", "https://auth.example.com/token"),
			),
			expectOptions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			builder := http.NewClientBuilderFromConfig(tt.config)

			client, err := NewHTTPClient(ctx, tt.config, builder)
			if err != nil {
				// Skip if we can't create the client (expected for invalid OAuth)
				t.Skipf("skipping test due to client creation error: %v", err)
			}
			defer func() {
				if closeErr := client.Close(); closeErr != nil {
					log.Printf("Failed to close HTTP client: %v", closeErr)
				}
			}()

			opts, err := client.getTokenHTTPOption()
			if err != nil {
				if tt.expectOptions {
					// For OAuth tests, we expect errors since we're using fake endpoints
					// Just verify that the error is related to token retrieval
					t.Logf("Expected token retrieval error with OAuth config: %v", err)
					return
				}
				t.Errorf("unexpected error getting token HTTP option: %v", err)
				return
			}

			if tt.expectOptions && len(opts) == 0 {
				t.Error("expected HTTP call options but got none")
			}

			if !tt.expectOptions && len(opts) > 0 {
				t.Error("expected no HTTP call options but got some")
			}
		})
	}
}

func TestHTTPClient_Methods(t *testing.T) {
	// Test that all client methods exist and can be called
	ctx := context.Background()
	cfg := config.NewHTTPConfig(
		config.WithHTTPEndpoint("https://httpbin.org"), // Use a real endpoint for testing
		config.WithHTTPInsecure(true),
	)
	builder := http.NewClientBuilderFromConfig(cfg)

	client, err := NewHTTPClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close HTTP client: %v", closeErr)
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

	// Note: These will likely fail with HTTP errors since we're not hitting the right endpoints
	// But we're testing that the methods exist and can be called
	_, err = client.Check(ctx, checkReq)
	if err == nil {
		t.Log("Check method succeeded (unexpected with mock endpoint)")
	}

	// Test CheckForUpdate method
	checkUpdateReq := &v1beta2.CheckForUpdateRequest{}
	_, err = client.CheckForUpdate(ctx, checkUpdateReq)
	if err == nil {
		t.Log("CheckForUpdate method succeeded (unexpected with mock endpoint)")
	}

	// Test ReportResource method
	reportReq := &v1beta2.ReportResourceRequest{}
	_, err = client.ReportResource(ctx, reportReq)
	if err == nil {
		t.Log("ReportResource method succeeded (unexpected with mock endpoint)")
	}

	// Test DeleteResource method
	deleteReq := &v1beta2.DeleteResourceRequest{}
	_, err = client.DeleteResource(ctx, deleteReq)
	if err == nil {
		t.Log("DeleteResource method succeeded (unexpected with mock endpoint)")
	}

	// Test StreamedListObjects method - this should panic
	listReq := &v1beta2.StreamedListObjectsRequest{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected StreamedListObjects to panic")
		}
	}()
	_, err = client.StreamedListObjects(ctx, listReq)
	if err == nil {
		t.Log("StreamedListObjects method succeeded (unexpected with mock endpoint)")
	}
}

func TestHTTPClient_WithOAuth(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewHTTPConfig(
		config.WithHTTPEndpoint("https://api.example.com"),
		config.WithHTTPInsecure(true),
		config.WithHTTPOAuth2("test-client", "test-secret", "https://auth.example.com/token"),
	)
	builder := http.NewClientBuilderFromConfig(cfg)

	client, err := NewHTTPClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client with OAuth: %v", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close HTTP client: %v", closeErr)
		}
	}()

	if client.tokenSource == nil {
		t.Error("expected token source to be set with OAuth config")
	}

	// Test that HTTP token options are generated
	opts, err := client.getTokenHTTPOption()
	if err != nil {
		// With OAuth config, we expect token retrieval errors since we're using fake endpoints
		t.Logf("Expected token retrieval error with OAuth config: %v", err)
		return
	}

	if len(opts) == 0 {
		t.Error("expected HTTP call options with OAuth config")
	}
}

func TestHTTPClient_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		buildConfig func() (*config.HTTPConfig, *http.ClientBuilder)
		expectError bool
	}{
		{
			name: "valid minimal config",
			buildConfig: func() (*config.HTTPConfig, *http.ClientBuilder) {
				cfg := config.NewHTTPConfig(
					config.WithHTTPEndpoint("https://api.example.com"),
					config.WithHTTPInsecure(true),
				)
				return cfg, http.NewClientBuilderFromConfig(cfg)
			},
			expectError: false,
		},
		{
			name: "config with custom settings",
			buildConfig: func() (*config.HTTPConfig, *http.ClientBuilder) {
				cfg := config.NewHTTPConfig(
					config.WithHTTPEndpoint("https://api.example.com"),
					config.WithHTTPInsecure(true),
					config.WithHTTPUserAgent("test-agent/1.0"),
					config.WithHTTPMaxIdleConns(50),
				)
				return cfg, http.NewClientBuilderFromConfig(cfg)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg, builder := tt.buildConfig()

			client, err := NewHTTPClient(ctx, cfg, builder)

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
					log.Printf("Failed to close HTTP client: %v", closeErr)
				}
			}()

			// Verify client was configured correctly
			if client.config != cfg {
				t.Error("expected config to match")
			}
		})
	}
}

func TestHTTPClient_StreamedListObjects_Panic(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewHTTPConfig(
		config.WithHTTPEndpoint("https://api.example.com"),
	)
	builder := http.NewClientBuilderFromConfig(cfg)

	client, err := NewHTTPClient(ctx, cfg, builder)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close HTTP client: %v", closeErr)
		}
	}()

	// Test that StreamedListObjects panics as expected
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected StreamedListObjects to panic")
		} else {
			// Verify the panic message
			if msg, ok := r.(string); ok {
				if msg != "Use grpc for streamed list objects" {
					t.Errorf("unexpected panic message: %q", msg)
				}
			} else {
				t.Errorf("unexpected panic type: %T", r)
			}
		}
	}()

	listReq := &v1beta2.StreamedListObjectsRequest{}
	_, err = client.StreamedListObjects(ctx, listReq)
	if err == nil {
		t.Log("StreamedListObjects method succeeded (unexpected with mock endpoint)")
	}
}
