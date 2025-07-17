# Kessel SDK for Go

The official Go SDK for the Kessel inventory and authorization service. This SDK provides gRPC client implementation with built-in OAuth2 support for secure authentication.

## Features

- **gRPC client support** - High-performance gRPC communication
- **OAuth2 authentication** - Built-in support with automatic token injection
- **Flexible configuration** - Direct token URL or issuer-based discovery
- **Type-safe API** - Generated from protobuf definitions
- **Comprehensive error handling** - Rich error types with gRPC status codes
- **Production ready** - Built with security, performance, and reliability in mind

## Installation

```bash
go get github.com/project-kessel/kessel-sdk-go
```

## Quick Start

### gRPC Client

```go
package main

import (
    "context"
    "log"

    v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
    // Create client using fluent builder pattern
    client, err := v1beta2.NewInventoryGRPCClientBuilder().
        WithEndpoint("your-kessel-server:9000").
        WithOAuth2("your-client-id", "your-client-secret", "https://your-auth-server/token").
        WithMaxReceiveMessageSize(8 * 1024 * 1024). // 8MB
        WithMaxSendMessageSize(4 * 1024 * 1024).    // 4MB
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Make API call - OAuth2 tokens are automatically injected!
    request := &v1beta2.CheckRequest{
        Object: &v1beta2.ResourceReference{
            ResourceType: "host",
            ResourceId:   "server-123",
        },
        Relation: "member",
        Subject: &v1beta2.SubjectReference{
            Resource: &v1beta2.ResourceReference{
                ResourceType: "user",
                ResourceId:   "alice",
            },
        },
    }

    response, err := client.Check(context.Background(), request)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Check result: %v", response.Allowed)
}
```

## Configuration

### OAuth2 Authentication

The SDK supports two OAuth2 configuration approaches:

#### 1. Direct Token URL

Specify the exact OAuth2 token endpoint:

```go
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithOAuth2("client-id", "client-secret", "https://auth.example.com/oauth/token").
    Build()
```

#### 2. Issuer-Based Discovery

Provide the issuer URL for automatic endpoint discovery via OpenID Connect:

```go
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithOAuth2Issuer("client-id", "client-secret", "https://auth.example.com"). // Will discover token endpoint automatically
    Build()
```

### Keycloak Example

For Keycloak authentication:

```go
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithOAuth2Issuer("your-service-account", "your-client-secret", "http://localhost:8085/realms/your-realm").
    Build()
```

### TLS Configuration

```go
// Secure connection (default)
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    Build()

// Insecure connection (development only)
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithInsecure(true).
    Build()

// Custom TLS configuration
tlsConfig := &tls.Config{
    ServerName: "your-server.com",
}
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithTLSConfig(tlsConfig).
    Build()
```

## API Reference

### Client Types

- `InventoryClient` - Simple wrapper around the gRPC client with connection lifecycle management

### Automatic Authentication

OAuth2 tokens are automatically injected into all requests when configured. No manual token management is required:

```go
// Configure OAuth2 once during client creation
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithOAuth2("client-id", "client-secret", "https://auth.example.com/token").
    Build()

// All subsequent calls automatically include OAuth2 tokens
response, err := client.Check(ctx, request)
response, err := client.ReportResource(ctx, reportRequest)
response, err := client.DeleteResource(ctx, deleteRequest)
```

### Configuration Options

#### Fluent Builder Methods

- `WithEndpoint(endpoint)` - Set server endpoint
- `WithInsecure(bool)` - Enable/disable TLS
- `WithTLSConfig(*tls.Config)` - Custom TLS configuration
- `WithMaxReceiveMessageSize(int)` - Set max receive message size
- `WithMaxSendMessageSize(int)` - Set max send message size
- `WithOAuth2(clientID, secret, tokenURL, scopes...)` - OAuth2 with direct token URL
- `WithOAuth2Issuer(clientID, secret, issuerURL, scopes...)` - OAuth2 with issuer discovery
- `WithDialOption(grpc.DialOption)` - Add custom gRPC dial options

```go
// Example configurations using the fluent builder pattern:

// Basic insecure client
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("localhost:9000").
    WithInsecure(true).
    Build()

// Secure client with OAuth2 and custom message sizes
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("secure.example.com:9000").
    WithOAuth2("client-id", "client-secret", "https://auth.example.com/token").
    WithMaxReceiveMessageSize(16 * 1024 * 1024). // 16MB
    WithMaxSendMessageSize(4 * 1024 * 1024).     // 4MB
    Build()

// Client with OAuth2 issuer discovery
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("secure.example.com:9000").
    WithOAuth2Issuer("client-id", "client-secret", "https://auth.example.com", "read", "write").
    Build()

// Client with custom TLS and dial options
client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("secure.example.com:9000").
    WithTLSConfig(&tls.Config{ServerName: "secure.example.com"}).
    WithDialOption(grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             time.Second,
        PermitWithoutStream: true,
    })).
    Build()
```

## Error Handling

The SDK provides rich error types for different failure scenarios:

```go
import "github.com/project-kessel/kessel-sdk-go/kessel/errors"

client, err := v1beta2.NewInventoryGRPCClientBuilder().
    WithEndpoint("your-server:9000").
    WithOAuth2("client-id", "client-secret", "https://auth.example.com/token").
    Build()

if err != nil {
    if errors.IsConnectionError(err) {
        log.Fatal("Failed to connect to server:", err)
    } else if errors.IsTokenError(err) {
        log.Fatal("OAuth2 authentication failed:", err)
    } else {
        log.Fatal("Unknown error:", err)
    }
}
```

Available error checkers:
- `errors.IsConnectionError(err)` - Network/connection failures
- `errors.IsTokenError(err)` - OAuth2 authentication failures

## Examples

Complete examples are available in the [`examples/`](./examples/) directory:

- [`examples/grpc/main.go`](./examples/grpc/main.go) - gRPC client usage

To run the examples:

```bash
# Build examples
make build

# Run gRPC example
./bin/grpc-example
```

## Development

### Prerequisites

- Go 1.21 or later
- Docker or Podman (for linting)
- Protocol Buffers compiler (for code generation)

### Building

```bash
# Install dependencies
go mod download

# Run linting
make lint

# Run tests
make test

# Build examples
make build

# Run tests with coverage
make test-coverage
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make help` | Display all available targets |
