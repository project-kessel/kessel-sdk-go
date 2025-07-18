package v1beta2

import (
	"crypto/tls"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestInventoryGRPCClientBuilder_Build_MissingEndpoint(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	// Don't set endpoint

	_, err := builder.Build()

	if err == nil {
		t.Error("expected error when endpoint is not set")
	}
	if !strings.Contains(err.Error(), "endpoint is required") {
		t.Errorf("expected error message to contain 'endpoint is required', got %q", err.Error())
	}
}

func TestInventoryGRPCClientBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(*InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder
		expectError bool
	}{
		{
			name: "basic insecure client",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("localhost:9000").WithInsecure(true)
			},
			expectError: false,
		},
		{
			name: "secure client with TLS",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("secure.example.com:9000").
					WithTLSConfig(&tls.Config{ServerName: "secure.example.com"})
			},
			expectError: false,
		},
		{
			name: "client with custom options",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("localhost:9000").
					WithInsecure(true).
					WithMaxReceiveMessageSize(8 * 1024 * 1024).
					WithMaxSendMessageSize(4 * 1024 * 1024).
					WithDialOption(grpc.WithUserAgent("test-agent"))
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewInventoryGRPCClientBuilder()
			configured := tt.configure(builder)

			client, err := configured.Build()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error building client: %v", err)
			}

			if client == nil {
				t.Fatal("expected client to be non-nil")
			}

			// Verify client has required components
			if client.KesselInventoryServiceClient == nil {
				t.Error("expected KesselInventoryServiceClient to be initialized")
			}
			if client.conn == nil {
				t.Error("expected conn to be initialized")
			}

			// Clean up
			err = client.Close()
			if err != nil {
				t.Errorf("unexpected error closing client: %v", err)
			}
		})
	}
}

func TestInventoryGRPCClientBuilder_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	client, err := NewInventoryGRPCClientBuilder().
		WithEndpoint("localhost:9000").
		WithInsecure(true).
		WithMaxReceiveMessageSize(8*1024*1024).
		WithMaxSendMessageSize(4*1024*1024).
		WithOAuth2("client-id", "client-secret", "https://auth.example.com/token", "read", "write").
		WithDialOption(grpc.WithUserAgent("test-agent")).
		Build()

	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if client == nil {
		t.Fatal("expected client to be non-nil")
	}

	// Clean up
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}

func TestInventoryClient_Close(t *testing.T) {
	client, err := NewInventoryGRPCClientBuilder().
		WithEndpoint("localhost:9000").
		WithInsecure(true).
		Build()

	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if client == nil {
		t.Fatal("expected client to be non-nil")
	}

	// Test that Close() works without error
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}
