package v1beta2

import (
	"time"

	genericBuilder "github.com/project-kessel/kessel-sdk-go/kessel/inventory/internal/builder"
)

type ClientBuilder = genericBuilder.ClientBuilder[KesselInventoryServiceClient]

// KeepaliveOption configures a gRPC client keepalive setting.
// Use the WithKeepalive functions to create options.
type KeepaliveOption = genericBuilder.KeepaliveOption

func NewClientBuilder(target string) *ClientBuilder {
	return genericBuilder.NewClientBuilder[KesselInventoryServiceClient](target, NewKesselInventoryServiceClient)
}

// WithKeepaliveInterval sets how often the client sends keepalive pings.
// The value must be positive. grpc-go clamps intervals below 10 seconds to
// 10 seconds when applying client keepalive parameters.
func WithKeepaliveInterval(interval time.Duration) KeepaliveOption {
	return genericBuilder.WithKeepaliveInterval(interval)
}

// WithKeepaliveTimeout sets how long the client waits for a keepalive ping
// acknowledgement before closing the connection. The value must be positive.
func WithKeepaliveTimeout(timeout time.Duration) KeepaliveOption {
	return genericBuilder.WithKeepaliveTimeout(timeout)
}

// WithKeepalivePermitWithoutCalls controls whether the client may send
// keepalive pings when no RPC is active. Its default is true.
func WithKeepalivePermitWithoutCalls(permit bool) KeepaliveOption {
	return genericBuilder.WithKeepalivePermitWithoutCalls(permit)
}
