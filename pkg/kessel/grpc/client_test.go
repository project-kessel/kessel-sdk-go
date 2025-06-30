package grpc

import (
	"crypto/tls"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewClientBuilder(t *testing.T) {
	endpoint := "localhost:9090"
	builder := NewClientBuilder(endpoint)

	if builder.endpoint != endpoint {
		t.Errorf("expected endpoint %q, got %q", endpoint, builder.endpoint)
	}

	if builder.insecure {
		t.Error("expected insecure to be false by default")
	}

	expectedSize := 4 * 1024 * 1024
	if builder.maxReceiveMessageSize != expectedSize {
		t.Errorf("expected maxReceiveMessageSize %d, got %d", expectedSize, builder.maxReceiveMessageSize)
	}

	if builder.maxSendMessageSize != expectedSize {
		t.Errorf("expected maxSendMessageSize %d, got %d", expectedSize, builder.maxSendMessageSize)
	}

	if builder.dialOptions == nil {
		t.Error("expected dialOptions to be initialized")
	}

	if len(builder.dialOptions) != 0 {
		t.Errorf("expected empty dialOptions, got %d options", len(builder.dialOptions))
	}
}

func TestNewClientBuilderFromConfig(t *testing.T) {
	cfg := &config.GRPCConfig{
		BaseConfig: config.BaseConfig{
			Endpoint: "example.com:443",
			Insecure: true,
			TLSConfig: &tls.Config{
				ServerName: "example.com",
			},
		},
		MaxReceiveMessageSize: 8 * 1024 * 1024,
		MaxSendMessageSize:    16 * 1024 * 1024,
	}

	builder := NewClientBuilderFromConfig(cfg)

	if builder.endpoint != cfg.Endpoint {
		t.Errorf("expected endpoint %q, got %q", cfg.Endpoint, builder.endpoint)
	}

	if builder.insecure != cfg.Insecure {
		t.Errorf("expected insecure %v, got %v", cfg.Insecure, builder.insecure)
	}

	if builder.tlsConfig != cfg.TLSConfig {
		t.Error("expected tlsConfig to match config")
	}

	if builder.maxReceiveMessageSize != cfg.MaxReceiveMessageSize {
		t.Errorf("expected maxReceiveMessageSize %d, got %d", cfg.MaxReceiveMessageSize, builder.maxReceiveMessageSize)
	}

	if builder.maxSendMessageSize != cfg.MaxSendMessageSize {
		t.Errorf("expected maxSendMessageSize %d, got %d", cfg.MaxSendMessageSize, builder.maxSendMessageSize)
	}

	if builder.dialOptions == nil {
		t.Error("expected dialOptions to be initialized")
	}
}

func TestClientBuilder_WithInsecure(t *testing.T) {
	builder := NewClientBuilder("localhost:9090")

	if builder.insecure {
		t.Error("expected insecure to be false initially")
	}

	result := builder.WithInsecure()

	if result != builder {
		t.Error("expected WithInsecure to return the same builder instance")
	}

	if !builder.insecure {
		t.Error("expected insecure to be true after WithInsecure")
	}
}

func TestClientBuilder_WithTransportSecurity(t *testing.T) {
	builder := NewClientBuilder("localhost:9090")
	builder.insecure = true // Set to true initially

	tlsConfig := &tls.Config{
		ServerName: "example.com",
	}

	result := builder.WithTransportSecurity(tlsConfig)

	if result != builder {
		t.Error("expected WithTransportSecurity to return the same builder instance")
	}

	if builder.insecure {
		t.Error("expected insecure to be false after WithTransportSecurity")
	}

	if builder.tlsConfig != tlsConfig {
		t.Error("expected tlsConfig to be set")
	}
}

func TestClientBuilder_WithMaxReceiveMessageSize(t *testing.T) {
	builder := NewClientBuilder("localhost:9090")
	expectedSize := 16 * 1024 * 1024

	result := builder.WithMaxReceiveMessageSize(expectedSize)

	if result != builder {
		t.Error("expected WithMaxReceiveMessageSize to return the same builder instance")
	}

	if builder.maxReceiveMessageSize != expectedSize {
		t.Errorf("expected maxReceiveMessageSize %d, got %d", expectedSize, builder.maxReceiveMessageSize)
	}
}

func TestClientBuilder_WithMaxSendMessageSize(t *testing.T) {
	builder := NewClientBuilder("localhost:9090")
	expectedSize := 8 * 1024 * 1024

	result := builder.WithMaxSendMessageSize(expectedSize)

	if result != builder {
		t.Error("expected WithMaxSendMessageSize to return the same builder instance")
	}

	if builder.maxSendMessageSize != expectedSize {
		t.Errorf("expected maxSendMessageSize %d, got %d", expectedSize, builder.maxSendMessageSize)
	}
}

func TestClientBuilder_WithDialOption(t *testing.T) {
	builder := NewClientBuilder("localhost:9090")

	if len(builder.dialOptions) != 0 {
		t.Error("expected empty dialOptions initially")
	}

	option := grpc.WithTransportCredentials(insecure.NewCredentials())
	result := builder.WithDialOption(option)

	if result != builder {
		t.Error("expected WithDialOption to return the same builder instance")
	}

	if len(builder.dialOptions) != 1 {
		t.Errorf("expected 1 dial option, got %d", len(builder.dialOptions))
	}

	// Add another option
	builder.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))

	if len(builder.dialOptions) != 2 {
		t.Errorf("expected 2 dial options, got %d", len(builder.dialOptions))
	}
}

func TestClientBuilder_Build_Insecure(t *testing.T) {
	builder := NewClientBuilder("localhost:9090").WithInsecure()

	conn, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if conn == nil {
		t.Fatal("expected connection but got nil")
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()

	// Verify the connection state
	state := conn.GetState()
	if state.String() == "" {
		t.Error("expected connection to have a state")
	}
}

func TestClientBuilder_Build_WithTLS(t *testing.T) {
	builder := NewClientBuilder("localhost:9090").
		WithTransportSecurity(&tls.Config{
			InsecureSkipVerify: true,
		})

	conn, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if conn == nil {
		t.Fatal("expected connection but got nil")
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()
}

func TestClientBuilder_Build_DefaultTLS(t *testing.T) {
	// Test building with default TLS (no custom config, not insecure)
	builder := NewClientBuilder("localhost:9090")
	// Don't set insecure or custom TLS config - should use default TLS

	conn, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if conn == nil {
		t.Fatal("expected connection but got nil")
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()
}

func TestClientBuilder_Build_WithCustomOptions(t *testing.T) {
	builder := NewClientBuilder("localhost:9090").
		WithInsecure().
		WithMaxReceiveMessageSize(8 * 1024 * 1024).
		WithMaxSendMessageSize(16 * 1024 * 1024).
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error building client: %v", err)
	}

	if conn == nil {
		t.Fatal("expected connection but got nil")
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()
}

func TestClientBuilder_ChainedCalls(t *testing.T) {
	// Test that all methods return the same builder instance for chaining
	builder := NewClientBuilder("localhost:9090")

	result := builder.
		WithInsecure().
		WithMaxReceiveMessageSize(8 * 1024 * 1024).
		WithMaxSendMessageSize(16 * 1024 * 1024).
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()))

	if result != builder {
		t.Error("expected chained calls to return the same builder instance")
	}
}

func TestClientBuilder_Build_WithInvalidEndpoint(t *testing.T) {
	// Test with various invalid endpoints
	invalidEndpoints := []string{
		"",
		"invalid-endpoint",
		":",
		"localhost:",
		":9090",
	}

	for _, endpoint := range invalidEndpoints {
		t.Run("endpoint_"+endpoint, func(t *testing.T) {
			builder := NewClientBuilder(endpoint).WithInsecure()

			conn, err := builder.Build()

			// We expect the connection to be created but may fail on actual use
			// gRPC client creation is lazy, so invalid endpoints don't fail immediately
			if err != nil {
				t.Logf("Expected behavior: build failed for invalid endpoint %q: %v", endpoint, err)
			} else {
				if conn != nil {
					err = conn.Close()
					if err != nil {
						t.Fatalf("failed to close connection: %v", err)
					}
				}
				t.Logf("Connection created for endpoint %q (will fail on actual use)", endpoint)
			}
		})
	}
}

func TestClientBuilder_MultipleBuilds(t *testing.T) {
	builder := NewClientBuilder("localhost:9090").WithInsecure()

	// Build multiple connections from the same builder
	conn1, err1 := builder.Build()
	if err1 != nil {
		t.Fatalf("unexpected error building first client: %v", err1)
	}
	defer func() {
		if closeErr := conn1.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()

	conn2, err2 := builder.Build()
	if err2 != nil {
		t.Fatalf("unexpected error building second client: %v", err2)
	}
	defer func() {
		if closeErr := conn2.Close(); closeErr != nil {
			t.Fatalf("failed to close connection: %v", closeErr)
		}
	}()

	// They should be different connection instances
	if conn1 == conn2 {
		t.Error("expected different connection instances")
	}
}

func TestClientBuilder_TLSConfigVariations(t *testing.T) {
	tests := []struct {
		name      string
		tlsConfig *tls.Config
	}{
		{
			name:      "nil TLS config",
			tlsConfig: nil,
		},
		{
			name: "TLS config with server name",
			tlsConfig: &tls.Config{
				ServerName: "example.com",
			},
		},
		{
			name: "TLS config with insecure skip verify",
			tlsConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		{
			name: "TLS config with min version",
			tlsConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewClientBuilder("localhost:9090")

			if tt.tlsConfig != nil {
				builder.WithTransportSecurity(tt.tlsConfig)
			}

			conn, err := builder.Build()
			if err != nil {
				t.Fatalf("unexpected error building client: %v", err)
			}

			if conn == nil {
				t.Fatal("expected connection but got nil")
			}

			defer func() {
				if closeErr := conn.Close(); closeErr != nil {
					t.Fatalf("failed to close connection: %v", closeErr)
				}
			}()
		})
	}
}
