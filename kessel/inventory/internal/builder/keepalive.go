package builder

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultKeepaliveInterval = 45 * time.Second
	defaultKeepaliveTimeout  = 10 * time.Second
)

type keepaliveConfig struct {
	interval            time.Duration
	timeout             time.Duration
	permitWithoutStream bool
}

func defaultKeepaliveConfig() keepaliveConfig {
	return keepaliveConfig{
		interval:            defaultKeepaliveInterval,
		timeout:             defaultKeepaliveTimeout,
		permitWithoutStream: true,
	}
}

// KeepaliveOption configures one keepalive setting on a ClientBuilder.
// Options only update the fields they represent.
type KeepaliveOption func(*keepaliveConfig)

// Keepalive applies the provided keepalive options and returns the builder.
// Repeated calls update only the fields provided by each option; calling
// Keepalive with no options leaves the current settings unchanged.
func (b *ClientBuilder[C]) Keepalive(options ...KeepaliveOption) *ClientBuilder[C] {
	for _, option := range options {
		if option != nil {
			option(&b.keepalive)
		}
	}
	return b
}

// WithKeepaliveInterval sets how often the client sends keepalive pings.
// The value must be positive. grpc-go clamps intervals below 10 seconds to
// 10 seconds when applying client keepalive parameters.
func WithKeepaliveInterval(interval time.Duration) KeepaliveOption {
	return func(config *keepaliveConfig) {
		config.interval = interval
	}
}

// WithKeepaliveTimeout sets how long the client waits for a keepalive ping
// acknowledgement before closing the connection. The value must be positive.
func WithKeepaliveTimeout(timeout time.Duration) KeepaliveOption {
	return func(config *keepaliveConfig) {
		config.timeout = timeout
	}
}

// WithKeepalivePermitWithoutCalls controls whether the client may send
// keepalive pings when no RPC is active. Its default is true.
func WithKeepalivePermitWithoutCalls(permit bool) KeepaliveOption {
	return func(config *keepaliveConfig) {
		config.permitWithoutStream = permit
	}
}

func (config keepaliveConfig) validate() error {
	if config.interval <= 0 {
		return fmt.Errorf("keepalive interval must be greater than zero")
	}
	if config.timeout <= 0 {
		return fmt.Errorf("keepalive timeout must be greater than zero")
	}
	return nil
}

func (config keepaliveConfig) clientParameters() keepalive.ClientParameters {
	return keepalive.ClientParameters{
		Time:                config.interval,
		Timeout:             config.timeout,
		PermitWithoutStream: config.permitWithoutStream,
	}
}

func newGRPCClientConn(target string, params keepalive.ClientParameters, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialOptions = append(dialOptions, grpc.WithKeepaliveParams(params))
	return grpc.NewClient(target, dialOptions...)
}
