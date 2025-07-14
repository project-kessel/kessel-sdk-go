# Kessel SDK for Go


The official Go SDK for the Kessel inventory and authorization service. This SDK provides gRPC client implementation with built-in OAuth2 support for secure authentication.

## Features

- **gRPC client support** - High-performance gRPC communication
- **OAuth2 authentication** - Built-in support with token refresh and caching
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

    inventoryapi "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
    "github.com/project-kessel/kessel-sdk-go/kessel/config"
    v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
    // Configure gRPC client with OAuth2
    cfg := config.NewGRPCConfig(
        config.WithGRPCEndpoint("your-kessel-server:9000"),
        config.WithGRPCOAuth2(
            "your-client-id",
            "your-client-secret", 
            "https://your-auth-server/token",
        ),
    )

    // Create client
    client, err := v1beta2.NewInventoryGRPCClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Make API call
    request := &inventoryapi.CheckRequest{
        Object: &inventoryapi.ResourceReference{
            ResourceType: "host",
            ResourceId:   "server-123",
        },
        Relation: "member",
        Subject: &inventoryapi.SubjectReference{
            Resource: &inventoryapi.ResourceReference{
                ResourceType: "user",
                ResourceId:   "alice",
            },
        },
    }

    // Get call options with OAuth2 token
    callOpts := client.GetCallOptions()
    response, err := client.KesselInventoryService.Check(context.Background(), request, callOpts...)
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
config.WithGRPCOAuth2(
    "client-id",
    "client-secret", 
    "https://auth.example.com/oauth/token",
)
```

#### 2. Issuer-Based Discovery

Provide the issuer URL for automatic endpoint discovery via OpenID Connect:

```go
config.WithGRPCOAuth2Issuer(
    "client-id",
    "client-secret",
    "https://auth.example.com", // Will discover token endpoint automatically
)
```

### Keycloak Example

For Keycloak authentication:

```go
config.WithGRPCOAuth2Issuer(
    "your-service-account",
    "your-client-secret",
    "http://localhost:8085/realms/your-realm",
)
```

### TLS Configuration

```go
// Secure connection (default)
config.WithGRPCEndpoint("your-server:9000")

// Insecure connection (development only)
config.WithGRPCInsecure(true)

// Custom TLS configuration
tlsConfig := &tls.Config{
    ServerName: "your-server.com",
}
config.WithGRPCTLSConfig(tlsConfig)
```

## API Reference

### Client Types

- `InventoryGRPCClient` - gRPC client for the Kessel Inventory service

### Authentication Methods

The client provides multiple ways to handle OAuth2 tokens:

#### Explicit Token Handling

```go
// Manual token retrieval and injection
tokenOpts, err := client.GetTokenCallOption()
if err != nil {
    return err
}

response, err := client.KesselInventoryService.Check(ctx, request, tokenOpts...)
```

### Configuration Options

#### gRPC Options

- `WithGRPCEndpoint(endpoint)` - Set server endpoint
- `WithGRPCInsecure(bool)` - Enable/disable TLS
- `WithGRPCTLSConfig(*tls.Config)` - Custom TLS configuration
- `WithGRPCMaxReceiveMessageSize(int)` - Set max receive message size
- `WithGRPCMaxSendMessageSize(int)` - Set max send message size
- `WithGRPCOAuth2(clientID, secret, tokenURL, scopes...)` - OAuth2 with direct token URL
- `WithGRPCOAuth2Issuer(clientID, secret, issuerURL, scopes...)` - OAuth2 with issuer discovery

## Error Handling

The SDK provides rich error types for different failure scenarios:

```go
import "github.com/project-kessel/kessel-sdk-go/kessel/errors"

client, err := v1beta2.NewInventoryGRPCClient(cfg)
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
| `make lint` | Run golangci-lint |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make build` | Build example binaries |
| `make clean` | Clean build artifacts |
| `make fmt` | Format Go code |
| `make mod-tidy` | Run go mod tidy |
| `make generate` | Generate protobuf files |

### Code Generation

The SDK uses Protocol Buffers for type definitions. To regenerate code:

```bash
make generate
```

This requires:
- [buf](https://buf.build/docs/installation) CLI tool
- Protocol Buffer compiler plugins

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test lint`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Run `make fmt` to format code
- Ensure `make lint` passes without errors
- Add tests for new functionality
- Update documentation as needed

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.


