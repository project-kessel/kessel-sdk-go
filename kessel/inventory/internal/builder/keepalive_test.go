package builder

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

type keepaliveTestClient struct{}

func keepaliveTestStub(grpc.ClientConnInterface) *keepaliveTestClient {
	return &keepaliveTestClient{}
}

type perRPCCredentialStub struct{}

func (perRPCCredentialStub) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return nil, nil
}

func (perRPCCredentialStub) RequireTransportSecurity() bool {
	return true
}

type clientConnCapture struct {
	calls           int
	target          string
	params          keepalive.ClientParameters
	dialOptionCount int
}

func captureClientConnConstruction[C any](b *ClientBuilder[C]) *clientConnCapture {
	capture := &clientConnCapture{}
	realConstructor := b.newClientConn
	b.newClientConn = func(target string, params keepalive.ClientParameters, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
		capture.calls++
		capture.target = target
		capture.params = params
		capture.dialOptionCount = len(dialOptions)
		return realConstructor(target, params, dialOptions...)
	}
	return capture
}

func TestBuildPassesKeepaliveDefaultsForEveryAuthMode(t *testing.T) {
	tests := []struct {
		name                string
		configure           func(*ClientBuilder[*keepaliveTestClient])
		wantDialOptionCount int
	}{
		{
			name:                "default tls",
			wantDialOptionCount: 1,
		},
		{
			name: "insecure",
			configure: func(b *ClientBuilder[*keepaliveTestClient]) {
				b.Insecure()
			},
			wantDialOptionCount: 1,
		},
		{
			name: "unauthenticated",
			configure: func(b *ClientBuilder[*keepaliveTestClient]) {
				b.Unauthenticated(nil)
			},
			wantDialOptionCount: 1,
		},
		{
			name: "authenticated",
			configure: func(b *ClientBuilder[*keepaliveTestClient]) {
				b.Authenticated(perRPCCredentialStub{}, nil)
			},
			wantDialOptionCount: 2,
		},
		{
			name: "oauth2 client authenticated",
			configure: func(b *ClientBuilder[*keepaliveTestClient]) {
				creds := auth.NewOAuth2ClientCredentials("client", "secret", "https://issuer.example/token", auth.WithMaxRetries(0))
				b.OAuth2ClientAuthenticated(&creds, nil)
			},
			wantDialOptionCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const target = "127.0.0.1:8080"
			b := NewClientBuilder(target, keepaliveTestStub)
			if tt.configure != nil {
				tt.configure(b)
			}
			capture := captureClientConnConstruction(b)

			client, conn, err := b.Build()
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if client == nil {
				t.Error("Build() returned a nil client")
			}
			if conn == nil {
				t.Fatal("Build() returned a nil connection")
			}
			defer func() {
				if closeErr := conn.Close(); closeErr != nil {
					t.Errorf("failed to close gRPC connection: %v", closeErr)
				}
			}()

			wantParams := keepalive.ClientParameters{
				Time:                45 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}
			if capture.calls != 1 {
				t.Errorf("client connection constructor calls = %d, want 1", capture.calls)
			}
			if capture.target != target || conn.Target() != target {
				t.Errorf("target = %q (connection target %q), want %q", capture.target, conn.Target(), target)
			}
			if capture.params != wantParams {
				t.Errorf("keepalive parameters = %+v, want %+v", capture.params, wantParams)
			}
			if capture.dialOptionCount != tt.wantDialOptionCount {
				t.Errorf("transport/auth dial option count = %d, want %d", capture.dialOptionCount, tt.wantDialOptionCount)
			}
		})
	}
}

func TestKeepaliveOptionsUpdateOnlyProvidedFields(t *testing.T) {
	b := NewClientBuilder("127.0.0.1:8080", keepaliveTestStub)
	capture := captureClientConnConstruction(b)

	if got := b.Keepalive(WithKeepaliveInterval(15 * time.Second)); got != b {
		t.Fatal("Keepalive() did not return the same builder")
	}
	b.Keepalive()
	b.Keepalive(WithKeepaliveTimeout(25*time.Second), WithKeepalivePermitWithoutCalls(false))
	b.Keepalive(WithKeepaliveInterval(21 * time.Second))

	client, conn, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if client == nil {
		t.Error("Build() returned a nil client")
	}
	if conn == nil {
		t.Fatal("Build() returned a nil connection")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()

	wantParams := keepalive.ClientParameters{
		Time:                21 * time.Second,
		Timeout:             25 * time.Second,
		PermitWithoutStream: false,
	}
	if capture.params != wantParams {
		t.Errorf("keepalive parameters = %+v, want %+v", capture.params, wantParams)
	}
}

func TestBuildRejectsNonPositiveKeepaliveDurations(t *testing.T) {
	tests := []struct {
		name        string
		option      KeepaliveOption
		wantErrText string
	}{
		{
			name:        "zero interval",
			option:      WithKeepaliveInterval(0),
			wantErrText: "keepalive interval must be greater than zero",
		},
		{
			name:        "negative interval",
			option:      WithKeepaliveInterval(-time.Second),
			wantErrText: "keepalive interval must be greater than zero",
		},
		{
			name:        "zero timeout",
			option:      WithKeepaliveTimeout(0),
			wantErrText: "keepalive timeout must be greater than zero",
		},
		{
			name:        "negative timeout",
			option:      WithKeepaliveTimeout(-time.Second),
			wantErrText: "keepalive timeout must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewClientBuilder("127.0.0.1:8080", keepaliveTestStub).Keepalive(tt.option)
			constructorCalls := 0
			b.newClientConn = func(target string, params keepalive.ClientParameters, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
				constructorCalls++
				return newGRPCClientConn(target, params, dialOptions...)
			}

			client, conn, err := b.Build()
			if err == nil || err.Error() != tt.wantErrText {
				t.Errorf("Build() error = %v, want %q", err, tt.wantErrText)
			}
			if client != nil {
				t.Errorf("Build() client = %v, want nil zero value", client)
			}
			if conn != nil {
				t.Errorf("Build() connection = %v, want nil", conn)
			}
			if constructorCalls != 0 {
				t.Errorf("client connection constructor calls = %d, want 0", constructorCalls)
			}
		})
	}
}

func TestValidKeepaliveOverwriteRecoversInvalidConfiguration(t *testing.T) {
	b := NewClientBuilder("127.0.0.1:8080", keepaliveTestStub).
		Keepalive(WithKeepaliveInterval(0)).
		Keepalive(WithKeepaliveInterval(5 * time.Second))
	capture := captureClientConnConstruction(b)

	client, conn, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if client == nil {
		t.Error("Build() returned a nil client")
	}
	if conn == nil {
		t.Fatal("Build() returned a nil connection")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()

	wantParams := keepalive.ClientParameters{
		Time:                5 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}
	if capture.params != wantParams {
		t.Errorf("keepalive parameters = %+v, want %+v", capture.params, wantParams)
	}
}

type peerClosure struct {
	time time.Time
	err  error
}

type localHTTP2Peer struct {
	listener net.Listener
	pings    chan time.Time
	closed   chan peerClosure
	done     chan struct{}

	mu      sync.Mutex
	conn    net.Conn
	stopped bool
}

func newLocalHTTP2Peer(listener net.Listener) *localHTTP2Peer {
	peer := &localHTTP2Peer{
		listener: listener,
		pings:    make(chan time.Time, 1),
		closed:   make(chan peerClosure, 1),
		done:     make(chan struct{}),
	}
	go peer.serve()
	return peer
}

func (peer *localHTTP2Peer) close() {
	_ = peer.listener.Close()
	peer.mu.Lock()
	peer.stopped = true
	if peer.conn != nil {
		_ = peer.conn.Close()
	}
	peer.mu.Unlock()
	<-peer.done
}

func (peer *localHTTP2Peer) setConn(conn net.Conn) bool {
	peer.mu.Lock()
	defer peer.mu.Unlock()
	if peer.stopped {
		_ = conn.Close()
		return false
	}
	peer.conn = conn
	return true
}

func (peer *localHTTP2Peer) reportClosed(err error) {
	peer.mu.Lock()
	stopped := peer.stopped
	peer.mu.Unlock()
	if stopped {
		return
	}
	select {
	case peer.closed <- peerClosure{time: time.Now(), err: err}:
	default:
	}
}

func (peer *localHTTP2Peer) serve() {
	defer close(peer.done)

	conn, err := peer.listener.Accept()
	if err != nil {
		peer.reportClosed(err)
		return
	}
	if !peer.setConn(conn) {
		return
	}
	defer func() {
		_ = conn.Close() // Cleanup may already have closed the peer connection.
	}()
	if err := conn.SetDeadline(time.Now().Add(45 * time.Second)); err != nil {
		peer.reportClosed(err)
		return
	}

	const clientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
	preface := make([]byte, len(clientPreface))
	if _, err := io.ReadFull(conn, preface); err != nil {
		peer.reportClosed(err)
		return
	}
	if string(preface) != clientPreface {
		peer.reportClosed(errors.New("unexpected HTTP/2 client preface"))
		return
	}

	framer := http2.NewFramer(conn, conn)
	if err := framer.WriteSettings(); err != nil {
		peer.reportClosed(err)
		return
	}

	for {
		frame, err := framer.ReadFrame()
		if err != nil {
			peer.reportClosed(err)
			return
		}
		switch frame := frame.(type) {
		case *http2.SettingsFrame:
			if !frame.IsAck() {
				if err := framer.WriteSettingsAck(); err != nil {
					peer.reportClosed(err)
					return
				}
			}
		case *http2.PingFrame:
			if !frame.IsAck() {
				select {
				case peer.pings <- time.Now():
				default:
				}
				// Deliberately do not acknowledge the client's keepalive ping.
			}
		}
	}
}

func TestBuildKeepaliveClosesIdleConnectionAfterUnacknowledgedPing(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	peer := newLocalHTTP2Peer(listener)
	defer peer.close()

	target := "passthrough:///" + listener.Addr().String()
	b := NewClientBuilder(target, keepaliveTestStub).
		Insecure().
		Keepalive(WithKeepaliveInterval(10*time.Second), WithKeepaliveTimeout(200*time.Millisecond))
	client, conn, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if client == nil {
		t.Error("Build() returned a nil client")
	}
	if conn == nil {
		t.Fatal("Build() returned a nil connection")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()

	conn.Connect()
	readyContext, cancelReady := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelReady()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if state == connectivity.Shutdown {
			t.Fatalf("connection entered %s before becoming ready", state)
		}
		if !conn.WaitForStateChange(readyContext, state) {
			t.Fatalf("connection did not become ready within the deadline; state = %s", conn.GetState())
		}
	}

	var pingTime time.Time
	pingTimer := time.NewTimer(13 * time.Second)
	defer pingTimer.Stop()
	select {
	case pingTime = <-peer.pings:
	case closure := <-peer.closed:
		t.Fatalf("HTTP/2 peer connection closed before receiving a keepalive ping: %v", closure.err)
	case <-pingTimer.C:
		t.Fatal("no idle keepalive ping received within 13 seconds")
	}

	closeTimer := time.NewTimer(3 * time.Second)
	defer closeTimer.Stop()
	select {
	case closure := <-peer.closed:
		if closure.err == nil {
			t.Fatal("HTTP/2 peer observed a connection close without a read error")
		}
		transportClosed := errors.Is(closure.err, io.EOF) ||
			errors.Is(closure.err, io.ErrUnexpectedEOF) ||
			errors.Is(closure.err, net.ErrClosed)
		var networkError net.Error
		if errors.As(closure.err, &networkError) {
			if networkError.Timeout() {
				t.Fatalf("HTTP/2 peer read timed out instead of observing client closure: %v", closure.err)
			}
			transportClosed = true
		}
		if !transportClosed {
			t.Fatalf("HTTP/2 peer encountered an error instead of observing client closure: %v", closure.err)
		}
		if elapsed := closure.time.Sub(pingTime); elapsed > 3*time.Second {
			t.Errorf("client closed %s after the unacknowledged keepalive ping, want within 3 seconds", elapsed)
		}
	case <-closeTimer.C:
		t.Fatal("client did not close the HTTP/2 transport after the unacknowledged keepalive ping")
	}
}
