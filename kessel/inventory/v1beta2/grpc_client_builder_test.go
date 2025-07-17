package v1beta2

import (
	"crypto/tls"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func TestNewInventoryGRPCClientBuilder(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()

	// Test default values
	if builder.maxReceiveMessageSize != 4194304 {
		t.Errorf("expected default MaxReceiveMessageSize to be 4194304, got %d", builder.maxReceiveMessageSize)
	}
	if builder.maxSendMessageSize != 4194304 {
		t.Errorf("expected default MaxSendMessageSize to be 4194304, got %d", builder.maxSendMessageSize)
	}
	if builder.dialOptions == nil {
		t.Error("expected dialOptions to be initialized")
	}
	if len(builder.dialOptions) != 0 {
		t.Errorf("expected dialOptions to be empty initially, got %d options", len(builder.dialOptions))
	}
}

func TestInventoryGRPCClientBuilder_WithEndpoint(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	endpoint := "localhost:9000"

	result := builder.WithEndpoint(endpoint)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithEndpoint to return the same builder instance")
	}

	// Test value is set
	if builder.endpoint != endpoint {
		t.Errorf("expected endpoint to be %q, got %q", endpoint, builder.endpoint)
	}
}

func TestInventoryGRPCClientBuilder_WithInsecure(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()

	// Test setting insecure to true
	result := builder.WithInsecure(true)
	if result != builder {
		t.Error("expected WithInsecure to return the same builder instance")
	}
	if !builder.insecure {
		t.Error("expected insecure to be true")
	}

	// Test setting insecure to false
	builder.WithInsecure(false)
	if builder.insecure {
		t.Error("expected insecure to be false")
	}
}

func TestInventoryGRPCClientBuilder_WithTLSConfig(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	tlsConfig := &tls.Config{ServerName: "test.example.com"}

	result := builder.WithTLSConfig(tlsConfig)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithTLSConfig to return the same builder instance")
	}

	// Test values are set
	if builder.tlsConfig != tlsConfig {
		t.Error("expected tlsConfig to be set")
	}
	if builder.insecure {
		t.Error("expected insecure to be false when TLS config is set")
	}
}

func TestInventoryGRPCClientBuilder_WithMaxReceiveMessageSize(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	size := 8 * 1024 * 1024

	result := builder.WithMaxReceiveMessageSize(size)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithMaxReceiveMessageSize to return the same builder instance")
	}

	// Test value is set
	if builder.maxReceiveMessageSize != size {
		t.Errorf("expected maxReceiveMessageSize to be %d, got %d", size, builder.maxReceiveMessageSize)
	}
}

func TestInventoryGRPCClientBuilder_WithMaxSendMessageSize(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	size := 8 * 1024 * 1024

	result := builder.WithMaxSendMessageSize(size)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithMaxSendMessageSize to return the same builder instance")
	}

	// Test value is set
	if builder.maxSendMessageSize != size {
		t.Errorf("expected maxSendMessageSize to be %d, got %d", size, builder.maxSendMessageSize)
	}
}

func TestInventoryGRPCClientBuilder_WithOAuth2(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	clientID := "test-client"
	clientSecret := "test-secret"
	tokenURL := "https://auth.example.com/token"
	scopes := []string{"read", "write"}

	result := builder.WithOAuth2(clientID, clientSecret, tokenURL, scopes...)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithOAuth2 to return the same builder instance")
	}

	// Test values are set
	if !builder.enableOAuth {
		t.Error("expected enableOAuth to be true")
	}
	if builder.clientID != clientID {
		t.Errorf("expected clientID to be %q, got %q", clientID, builder.clientID)
	}
	if builder.clientSecret != clientSecret {
		t.Errorf("expected clientSecret to be %q, got %q", clientSecret, builder.clientSecret)
	}
	if builder.tokenURL != tokenURL {
		t.Errorf("expected tokenURL to be %q, got %q", tokenURL, builder.tokenURL)
	}
	if len(builder.scopes) != len(scopes) {
		t.Errorf("expected %d scopes, got %d", len(scopes), len(builder.scopes))
	}
	for i, scope := range scopes {
		if i < len(builder.scopes) && builder.scopes[i] != scope {
			t.Errorf("expected scope[%d] to be %q, got %q", i, scope, builder.scopes[i])
		}
	}
}

func TestInventoryGRPCClientBuilder_WithOAuth2Issuer(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	clientID := "test-client"
	clientSecret := "test-secret"
	issuerURL := "https://auth.example.com"
	scopes := []string{"read", "write"}

	result := builder.WithOAuth2Issuer(clientID, clientSecret, issuerURL, scopes...)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithOAuth2Issuer to return the same builder instance")
	}

	// Test values are set
	if !builder.enableOAuth {
		t.Error("expected enableOAuth to be true")
	}
	if builder.clientID != clientID {
		t.Errorf("expected clientID to be %q, got %q", clientID, builder.clientID)
	}
	if builder.clientSecret != clientSecret {
		t.Errorf("expected clientSecret to be %q, got %q", clientSecret, builder.clientSecret)
	}
	if builder.issuerURL != issuerURL {
		t.Errorf("expected issuerURL to be %q, got %q", issuerURL, builder.issuerURL)
	}
	if len(builder.scopes) != len(scopes) {
		t.Errorf("expected %d scopes, got %d", len(scopes), len(builder.scopes))
	}
	for i, scope := range scopes {
		if i < len(builder.scopes) && builder.scopes[i] != scope {
			t.Errorf("expected scope[%d] to be %q, got %q", i, scope, builder.scopes[i])
		}
	}
}

func TestInventoryGRPCClientBuilder_WithDialOption(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()
	dialOpt := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		PermitWithoutStream: true,
	})

	result := builder.WithDialOption(dialOpt)

	// Test fluent interface
	if result != builder {
		t.Error("expected WithDialOption to return the same builder instance")
	}

	// Test dial option is added
	if len(builder.dialOptions) != 1 {
		t.Errorf("expected 1 dial option, got %d", len(builder.dialOptions))
	}

	// Test multiple dial options can be added
	dialOpt2 := grpc.WithUserAgent("test-agent")
	builder.WithDialOption(dialOpt2)
	if len(builder.dialOptions) != 2 {
		t.Errorf("expected 2 dial options, got %d", len(builder.dialOptions))
	}
}

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
		validate    func(*testing.T, *InventoryClient)
		expectError bool
	}{
		{
			name: "basic insecure client",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("localhost:9000").WithInsecure(true)
			},
			validate: func(t *testing.T, client *InventoryClient) {
				if client.KesselInventoryServiceClient == nil {
					t.Error("expected KesselInventoryServiceClient to be initialized")
				}
				if client.conn == nil {
					t.Error("expected conn to be initialized")
				}
			},
			expectError: false,
		},
		{
			name: "client with custom TLS config",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("secure.example.com:9000").
					WithTLSConfig(&tls.Config{ServerName: "secure.example.com"})
			},
			validate: func(t *testing.T, client *InventoryClient) {
				if client.KesselInventoryServiceClient == nil {
					t.Error("expected KesselInventoryServiceClient to be initialized")
				}
				if client.conn == nil {
					t.Error("expected conn to be initialized")
				}
			},
			expectError: false,
		},
		{
			name: "client with custom dial options",
			configure: func(b *InventoryGRPCClientBuilder) *InventoryGRPCClientBuilder {
				return b.WithEndpoint("localhost:9000").
					WithInsecure(true).
					WithDialOption(grpc.WithUserAgent("test-agent"))
			},
			validate: func(t *testing.T, client *InventoryClient) {
				if client.KesselInventoryServiceClient == nil {
					t.Error("expected KesselInventoryServiceClient to be initialized")
				}
				if client.conn == nil {
					t.Error("expected conn to be initialized")
				}
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
				// Don't validate the client when we expect an error
				return
			}

			if err != nil {
				t.Fatalf("unexpected error building client: %v", err)
			}

			if client == nil {
				t.Fatal("expected client to be non-nil")
			}

			tt.validate(t, client)

			// Clean up
			if client.conn != nil {
				client.Close()
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

	// Verify some key properties
	if client.KesselInventoryServiceClient == nil {
		t.Error("expected KesselInventoryServiceClient to be initialized")
	}
	if client.conn == nil {
		t.Error("expected conn to be initialized")
	}

	// Clean up
	client.Close()
}

func TestInventoryGRPCClientBuilder_DefaultValues(t *testing.T) {
	builder := NewInventoryGRPCClientBuilder()

	// Test default values
	if builder.endpoint != "" {
		t.Errorf("expected default endpoint to be empty, got %q", builder.endpoint)
	}
	if builder.insecure {
		t.Error("expected default insecure to be false")
	}
	if builder.tlsConfig != nil {
		t.Error("expected default tlsConfig to be nil")
	}
	if builder.maxReceiveMessageSize != 4194304 {
		t.Errorf("expected default maxReceiveMessageSize to be 4194304, got %d", builder.maxReceiveMessageSize)
	}
	if builder.maxSendMessageSize != 4194304 {
		t.Errorf("expected default maxSendMessageSize to be 4194304, got %d", builder.maxSendMessageSize)
	}
	if builder.enableOAuth {
		t.Error("expected default enableOAuth to be false")
	}
	if builder.clientID != "" {
		t.Errorf("expected default clientID to be empty, got %q", builder.clientID)
	}
	if builder.clientSecret != "" {
		t.Errorf("expected default clientSecret to be empty, got %q", builder.clientSecret)
	}
	if builder.tokenURL != "" {
		t.Errorf("expected default tokenURL to be empty, got %q", builder.tokenURL)
	}
	if builder.issuerURL != "" {
		t.Errorf("expected default issuerURL to be empty, got %q", builder.issuerURL)
	}
	if len(builder.scopes) != 0 {
		t.Errorf("expected default scopes to be empty, got %d scopes", len(builder.scopes))
	}
	if len(builder.dialOptions) != 0 {
		t.Errorf("expected default dialOptions to be empty, got %d options", len(builder.dialOptions))
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
