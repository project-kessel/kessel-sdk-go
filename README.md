# Kessel SDK for Go

The official Go SDK for the Kessel inventory and authorization service. This SDK provides gRPC client implementation for secure communication.

## Features

- **gRPC client support** - High-performance gRPC communication
- **Type-safe API** - Generated from protobuf definitions
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
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
    ctx := context.Background()
    
    grpcConfig := config.NewCompatibilityConfig(
        config.WithGRPCEndpoint("your-kessel-server:9000"),
        config.WithGRPCInsecure(true),
    )

    // Using insecure credentials for local development
    var dialOpts []grpc.DialOption
    dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
    dialOpts = append(dialOpts,
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcConfig.MaxReceiveMessageSize)),
        grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcConfig.MaxSendMessageSize)),
    )

    conn, err := grpc.NewClient(grpcConfig.Url, dialOpts...)
    if err != nil {
        log.Fatal("Failed to create gRPC client:", err)
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
            Reporter: &v1beta2.ReporterReference{
                Type: "HBI",
            },
        },
        Relation: "member",
        Subject: &v1beta2.SubjectReference{
            Resource: &v1beta2.ResourceReference{
                ResourceType: "user",
                ResourceId:   "alice",
            },
        },
    }

    response, err := inventoryClient.Check(ctx, checkRequest)
    if err != nil {
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.Unavailable:
                log.Fatal("Service unavailable: ", err)
            case codes.PermissionDenied:
                log.Fatal("Permission denied: ", err)
            default:
                log.Fatal("gRPC connection error: ", err)
            }
        } else {
            log.Fatal("Unknown error: ", err)
        }
    }

    log.Printf("Check result: %v", response.Allowed)
}
```

## Configuration

### gRPC Endpoint

Specify the gRPC endpoint:

```go
grpcConfig := config.NewCompatibilityConfig(
    config.WithGRPCEndpoint("your-kessel-server:9000"),
    config.WithGRPCInsecure(true),
)
```

## Error Handling

The SDK uses standard gRPC status codes:

```go
response, err := inventoryClient.Check(ctx, checkRequest)
if err != nil {
    if st, ok := status.FromError(err); ok {
        switch st.Code() {
        case codes.Unavailable:
            log.Fatal("Service unavailable:", err)
        case codes.PermissionDenied:
            log.Fatal("Permission denied:", err)
        default:
            log.Fatal("gRPC connection error:", err)
        }
    } else {
        log.Fatal("Unknown error:", err)
    }
}
```

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
