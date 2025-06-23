package grpc

import (
	"crypto/tls"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type ClientBuilder struct {
	address     string
	dialOptions []grpc.DialOption
}

func NewClientBuilder(address string) *ClientBuilder {
	return &ClientBuilder{
		address: address,
		dialOptions: []grpc.DialOption{
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             20 * time.Second,
				PermitWithoutStream: true,
			}),
		},
	}
}

func (b *ClientBuilder) WithTransportSecurity(tlsConfig *tls.Config) *ClientBuilder {
	b.dialOptions = append(b.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	return b
}

func (b *ClientBuilder) WithInsecure() *ClientBuilder {
	b.dialOptions = append(b.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return b
}

func (b *ClientBuilder) Build() (*grpc.ClientConn, error) {
	return grpc.NewClient(b.address, b.dialOptions...)
}
