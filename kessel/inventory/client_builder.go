package inventory

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientBuilder[C any] struct {
	target             string
	channelCredentials credentials.TransportCredentials
	perRPCCredentials  credentials.PerRPCCredentials
	defaultCallOptions []grpc.CallOption
	insecure           bool
	extraDialOptions   []grpc.DialOption
	extraCallOptions   []grpc.CallOption
	newStub            func(grpc.ClientConnInterface) C
}

func NewClientBuilder[C any](target string, newStub func(grpc.ClientConnInterface) C) *ClientBuilder[C] {
	return &ClientBuilder[C]{
		target:             target,
		channelCredentials: credentials.NewTLS(&tls.Config{}),
		defaultCallOptions: []grpc.CallOption{
			grpc.MaxCallRecvMsgSize(4 * 1024 * 1024),
			grpc.MaxCallSendMsgSize(4 * 1024 * 1024),
		},
		newStub: newStub,
	}
}

func (b *ClientBuilder[C]) OAuth2ClientAuthenticated(oAuth2ClientCredentials *auth.OAuth2ClientCredentials, channelCredentials credentials.TransportCredentials) *ClientBuilder[C] {
	if channelCredentials != nil {
		b.channelCredentials = channelCredentials
	}
	if oAuth2ClientCredentials != nil {
		b.perRPCCredentials = &oauth2PerRPCCreds{creds: oAuth2ClientCredentials, insecure: b.insecure}
	}
	return b
}

func (b *ClientBuilder[C]) Authenticated(callCredentials credentials.PerRPCCredentials, channelCredentials credentials.TransportCredentials) *ClientBuilder[C] {
	b.perRPCCredentials = callCredentials
	if channelCredentials != nil {
		b.channelCredentials = channelCredentials
	}
	return b
}

func (b *ClientBuilder[C]) Unauthenticated(channelCredentials credentials.TransportCredentials) *ClientBuilder[C] {
	b.perRPCCredentials = nil
	if channelCredentials != nil {
		b.channelCredentials = channelCredentials
	}
	return b
}

func (b *ClientBuilder[C]) Insecure() *ClientBuilder[C] {
	b.insecure = true
	b.channelCredentials = insecure.NewCredentials()
	b.perRPCCredentials = nil
	return b
}

func (b *ClientBuilder[C]) WithDialOption(opt grpc.DialOption) *ClientBuilder[C] {
	b.extraDialOptions = append(b.extraDialOptions, opt)
	return b
}

func (b *ClientBuilder[C]) WithCallOption(opt grpc.CallOption) *ClientBuilder[C] {
	b.extraCallOptions = append(b.extraCallOptions, opt)
	return b
}

func (b *ClientBuilder[C]) Build() (C, *grpc.ClientConn, error) {
	var zero C
	if b.target == "" {
		return zero, nil, fmt.Errorf("target URI is required")
	}

	// Disallow auth credentials over insecure transport to mirror Python builder semantics
	if b.insecure && b.perRPCCredentials != nil {
		return zero, nil, fmt.Errorf("invalid credential configuration: cannot authenticate with insecure channel")
	}

	var dialOpts []grpc.DialOption
	// Transport security (TLS or insecure)
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(b.channelCredentials))
	// Default call options
	callOpts := append([]grpc.CallOption{}, b.defaultCallOptions...)
	if b.perRPCCredentials != nil {
		callOpts = append(callOpts, grpc.PerRPCCredentials(b.perRPCCredentials))
	}
	callOpts = append(callOpts, b.extraCallOptions...)
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(callOpts...))
	// Extra dial options
	dialOpts = append(dialOpts, b.extraDialOptions...)

	conn, err := grpc.NewClient(b.target, dialOpts...)
	if err != nil {
		return zero, nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return b.newStub(conn), conn, nil
}

type oauth2PerRPCCreds struct {
	creds    *auth.OAuth2ClientCredentials
	insecure bool
}

func (o *oauth2PerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	tok, err := o.creds.GetToken(ctx, auth.GetTokenOptions{})
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", tok.AccessToken),
	}, nil
}

func (o *oauth2PerRPCCreds) RequireTransportSecurity() bool {
	return !o.insecure
}
