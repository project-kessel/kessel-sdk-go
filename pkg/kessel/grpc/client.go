package grpc

import (
	"crypto/tls"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientBuilder struct {
	endpoint              string
	insecure              bool
	tlsConfig             *tls.Config
	maxReceiveMessageSize int
	maxSendMessageSize    int
	dialOptions           []grpc.DialOption
}

func NewClientBuilder(endpoint string) *ClientBuilder {
	return &ClientBuilder{
		endpoint:              endpoint,
		insecure:              false,
		maxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
		maxSendMessageSize:    4 * 1024 * 1024, // 4MB
		dialOptions:           []grpc.DialOption{},
	}
}

// NewClientBuilderFromConfig creates a ClientBuilder from GRPCConfig
func NewClientBuilderFromConfig(cfg *config.GRPCConfig) *ClientBuilder {
	builder := &ClientBuilder{
		endpoint:              cfg.Endpoint,
		insecure:              cfg.Insecure,
		tlsConfig:             cfg.TLSConfig,
		maxReceiveMessageSize: cfg.MaxReceiveMessageSize,
		maxSendMessageSize:    cfg.MaxSendMessageSize,
		dialOptions:           []grpc.DialOption{},
	}
	return builder
}

func (b *ClientBuilder) WithInsecure() *ClientBuilder {
	b.insecure = true
	return b
}

func (b *ClientBuilder) WithTransportSecurity(tlsConfig *tls.Config) *ClientBuilder {
	b.insecure = false
	b.tlsConfig = tlsConfig
	return b
}

func (b *ClientBuilder) WithMaxReceiveMessageSize(size int) *ClientBuilder {
	b.maxReceiveMessageSize = size
	return b
}

func (b *ClientBuilder) WithMaxSendMessageSize(size int) *ClientBuilder {
	b.maxSendMessageSize = size
	return b
}

func (b *ClientBuilder) WithDialOption(opt grpc.DialOption) *ClientBuilder {
	b.dialOptions = append(b.dialOptions, opt)
	return b
}

func (b *ClientBuilder) Build() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(b.maxReceiveMessageSize),
			grpc.MaxCallSendMsgSize(b.maxSendMessageSize),
		),
	}

	// Add custom dial options
	opts = append(opts, b.dialOptions...)

	// Configure transport security
	if b.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else if b.tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(b.tlsConfig)))
	} else {
		// Use default TLS credentials
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	return grpc.NewClient(b.endpoint, opts...)
}
