package v1beta2

import (
    "context"
    "crypto/tls"
    "fmt"

    "github.com/project-kessel/kessel-sdk-go/kessel/auth"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/credentials/insecure"
)

// NewClientBuilder starts a builder using the provided target URI.
// Defaults:
// - TLS channel credentials (runtime-default trust bundle)
// - Stock gRPC defaults otherwise (no keepalive)
func NewClientBuilder(target string) *ClientBuilder {
    return &ClientBuilder{
        target:             target,
        channelCredentials: credentials.NewTLS(&tls.Config{}),
        defaultCallOptions: []grpc.CallOption{
            grpc.MaxCallRecvMsgSize(4 * 1024 * 1024),
            grpc.MaxCallSendMsgSize(4 * 1024 * 1024),
        },
    }
}

// ClientBuilder implements the builder pattern for constructing a Kessel Inventory client.
type ClientBuilder struct {
    target             string
    channelCredentials credentials.TransportCredentials
    perRPCCredentials  credentials.PerRPCCredentials
    defaultCallOptions []grpc.CallOption
    insecure           bool
    extraDialOptions   []grpc.DialOption
    extraCallOptions   []grpc.CallOption
}

// OAuth2ClientAuthenticated configures CallCredentials using OAuth2 Client Credentials.
// If channelCredentials is nil, the default TLS channel credentials are used.
func (b *ClientBuilder) OAuth2ClientAuthenticated(oauth2Creds *auth.OAuth2ClientCredentials, channelCredentials credentials.TransportCredentials) *ClientBuilder {
    if channelCredentials != nil {
        b.channelCredentials = channelCredentials
    }
    if oauth2Creds != nil {
        b.perRPCCredentials = &oauth2PerRPCCreds{creds: oauth2Creds, insecure: b.insecure}
    }
    return b
}

// Authenticated configures explicit CallCredentials and (optionally) ChannelCredentials.
func (b *ClientBuilder) Authenticated(callCredentials credentials.PerRPCCredentials, channelCredentials credentials.TransportCredentials) *ClientBuilder {
    b.perRPCCredentials = callCredentials
    if channelCredentials != nil {
        b.channelCredentials = channelCredentials
    }
    return b
}

// Unauthenticated configures only ChannelCredentials (server auth via TLS/mTLS). Client is not authenticated.
func (b *ClientBuilder) Unauthenticated(channelCredentials credentials.TransportCredentials) *ClientBuilder {
    b.perRPCCredentials = nil
    if channelCredentials != nil {
        b.channelCredentials = channelCredentials
    }
    return b
}

// Insecure configures an insecure connection (testing only). Disables both channel and per-RPC credentials.
func (b *ClientBuilder) Insecure() *ClientBuilder {
    b.insecure = true
    b.channelCredentials = insecure.NewCredentials()
    b.perRPCCredentials = nil
    return b
}

// WithDialOption appends a custom grpc.DialOption.
func (b *ClientBuilder) WithDialOption(opt grpc.DialOption) *ClientBuilder {
    b.extraDialOptions = append(b.extraDialOptions, opt)
    return b
}

// WithCallOption appends a default grpc.CallOption applied to the stub.
func (b *ClientBuilder) WithCallOption(opt grpc.CallOption) *ClientBuilder {
    b.extraCallOptions = append(b.extraCallOptions, opt)
    return b
}

// Build constructs the generated stub for the inventory v1beta2 service following the builder's configuration.
// Returns a client wrapper that exposes the generated methods and a Close() for resource cleanup.
func (b *ClientBuilder) Build() (*InventoryClient, error) {
    if b.target == "" {
        return nil, fmt.Errorf("target URI is required")
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
        return nil, fmt.Errorf("failed to create gRPC client: %w", err)
    }

    return &InventoryClient{
        KesselInventoryServiceClient: NewKesselInventoryServiceClient(conn),
        conn:                         conn,
    }, nil
}

// InventoryClient wraps the generated client and manages the underlying connection lifecycle.
type InventoryClient struct {
    KesselInventoryServiceClient
    conn *grpc.ClientConn
}

// Close releases the underlying connection resources.
func (c *InventoryClient) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}

// oauth2PerRPCCreds implements credentials.PerRPCCredentials using OAuth2 Client Credentials.
type oauth2PerRPCCreds struct {
    creds   *auth.OAuth2ClientCredentials
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

