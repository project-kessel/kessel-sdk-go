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

    "github.com/project-kessel/kessel-sdk-go/kessel/config"
    "github.com/project-kessel/kessel-sdk-go/kessel/errors"
    "github.com/project-kessel/kessel-sdk-go/kessel/grpc/auth"
    v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    ctx := context.Background()
    grpcConfig := config.NewCompatibilityConfig(
        config.WithGRPCEndpoint("your-kessel-server:9000"),
        config.WithGRPCInsecure(true),
    )

    issuerURL := "http://localhost:8085/realms/redhat-external"
	discovery, err := auth.FetchOIDCDiscovery(ctx, issuerURL)
	if err != nil {
		log.Fatal("Failed to fetch OIDC discovery: ", err)
	}

    authCredentials, err := auth.NewOAuth2ClientCredentials(
        "your-client-id",
        "your-client-secret",
        discovery.TokenEndpoint(),
    )
    if err != nil {
		log.Fatal("OAuth2 credential creation failed: ", err)
    }

    // Using insecure credentials for local development
    var dialOpts []grpc.DialOption
    dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(tokenSource.GetInsecureGRPCCredentials()))
    dialOpts = append(dialOpts,
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcConfig.MaxReceiveMessageSize)),
        grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcConfig.MaxSendMessageSize)),
    )

    conn, err := grpc.NewClient(grpcConfig.Url, dialOpts...)
    if err != nil {
        // Example of checking for specific error types using sentinel errors
        if errors.IsConnectionError(err) {
            log.Fatal("Failed to establish connection:", err)
        } else if errors.IsTokenError(err) {
            log.Fatal("OAuth2 token configuration failed:", err)
        } else {
            log.Fatal("Unknown error:", err)
        }
    }
    defer func() {
        if closeErr := conn.Close(); closeErr != nil {
            log.Printf("Failed to close gRPC client: %v", closeErr)
        }
    }()

    inventoryClient := v1beta2.NewKesselInventoryServiceClient(conn)

    // Example request using the external API types
    checkRequest := &v1beta2.CheckRequest{
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
authCredentials, err := auth.NewOAuth2ClientCredentials(
    "your-client-id",
    "your-client-secret",
    "https://auth.server.com/oauth/token",
)
```

#### 2. Issuer-Based Discovery

If you need to discover the token endpoint from an issuer:

```go
// Step 1: Discover endpoints
discovery, err := auth.FetchOIDCDiscovery(ctx, "https://your-issuer.com")
if err != nil {
    log.Fatal(err)
}

// Step 2: Create credentials with discovered endpoint
authCredentials, err := auth.NewOAuth2ClientCredentials(
    "your-client-id",
    "your-client-secret",
    discovery.TokenEndpoint(),
)
```

## API Reference

### Automatic Authentication

OAuth2 tokens are automatically injected into all requests when configured. No manual token management is required:

```go
// Configure OAuth2 once during client creation

// All subsequent calls automatically include OAuth2 tokens
response, err := client.Check(ctx, request)
response, err := client.ReportResource(ctx, reportRequest)
response, err := client.DeleteResource(ctx, deleteRequest)
```

### Configuration Options

## Error Handling

The SDK provides rich error types for different failure scenarios:

```go
import "github.com/project-kessel/kessel-sdk-go/kessel/errors"

tokenSource, err := auth.NewTokenSource(grpcConfig)
if err != nil {
    if errors.IsTokenError(err) {
        // Handle OAuth2 authentication errors
        log.Fatal("OAuth2 authentication failed:", err)
    } else {
        // Handle other errors
        log.Fatal("Unknown error:", err)
    }
}

conn, err := grpc.NewClient(endpoint, dialOpts...)
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
