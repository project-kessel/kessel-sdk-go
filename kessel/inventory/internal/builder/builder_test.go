package builder

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// noopStub is a minimal stub constructor for testing the generic ClientBuilder.
func noopStub(_ grpc.ClientConnInterface) struct{} { return struct{}{} }

// insecurePerRPCCreds is a test PerRPCCredentials that does not require transport security.
type insecurePerRPCCreds struct{}

func (insecurePerRPCCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer test-token"}, nil
}

func (insecurePerRPCCreds) RequireTransportSecurity() bool {
	return false
}

func TestDefaultTLSConfigMinVersion(t *testing.T) {
	cfg := defaultTLSConfig()
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion tls.VersionTLS12 (0x%04x), got 0x%04x", tls.VersionTLS12, cfg.MinVersion)
	}
}

func TestDefaultTLSConfigDoesNotSetInsecureSkipVerify(t *testing.T) {
	cfg := defaultTLSConfig()
	if cfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be false")
	}
}

func TestDefaultTLSConfigReturnsNewInstance(t *testing.T) {
	a := defaultTLSConfig()
	b := defaultTLSConfig()
	if a == b {
		t.Error("expected each call to return a new *tls.Config instance")
	}
}

func TestNewClientBuilderDefaultsToTLS(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)

	info := b.channelCredentials.Info()
	if info.SecurityProtocol != "tls" {
		t.Errorf("expected security protocol 'tls', got %q", info.SecurityProtocol)
	}
}

func TestSetChannelCredentialsOrDefaultNilFallsBackToTLS(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	// Clear credentials to verify the fallback path
	b.channelCredentials = nil
	b.setChannelCredentialsOrDefault(nil)

	if b.channelCredentials == nil {
		t.Fatal("expected channelCredentials to be set after nil fallback")
	}

	info := b.channelCredentials.Info()
	if info.SecurityProtocol != "tls" {
		t.Errorf("expected security protocol 'tls', got %q", info.SecurityProtocol)
	}

	if b.insecure {
		t.Error("expected insecure to be false after TLS fallback")
	}
}

func TestSetChannelCredentialsOrDefaultPreservesCallerConfig(t *testing.T) {
	callerCreds := credentials.NewTLS(&tls.Config{
		ServerName: "custom-server",
		MinVersion: tls.VersionTLS13,
	})

	b := NewClientBuilder("localhost:8080", noopStub)
	b.setChannelCredentialsOrDefault(callerCreds)

	if b.channelCredentials != callerCreds {
		t.Error("expected caller-supplied credentials to be preserved, got different instance")
	}
}

func TestInsecureOverridesTLS(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	b.Insecure()

	info := b.channelCredentials.Info()
	if info.SecurityProtocol == "tls" {
		t.Error("expected Insecure() to switch away from TLS")
	}

	if !b.insecure {
		t.Error("expected insecure flag to be true")
	}
}

func TestInsecureClearsPerRPCCredentials(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	b.perRPCCredentials = &oauth2PerRPCCreds{}
	b.Insecure()

	if b.perRPCCredentials != nil {
		t.Error("expected Insecure() to clear per-RPC credentials")
	}
}

func TestBuildEmptyTargetReturnsError(t *testing.T) {
	b := NewClientBuilder("", noopStub)
	_, _, err := b.Build()
	if err == nil {
		t.Error("expected error when target is empty")
	}
}

func TestBuildWithValidTarget(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	_, conn, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()
}

func TestAuthModeOverwriting(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	b.Insecure()

	if !b.insecure {
		t.Fatal("expected insecure to be true after Insecure()")
	}

	// Calling Unauthenticated with nil should override back to TLS
	b.Unauthenticated(nil)

	if b.insecure {
		t.Error("expected insecure to be false after Unauthenticated()")
	}

	info := b.channelCredentials.Info()
	if info.SecurityProtocol != "tls" {
		t.Errorf("expected security protocol 'tls' after Unauthenticated(), got %q", info.SecurityProtocol)
	}
}

func TestOAuth2ClientAuthenticatedNilClearsStalePerRPCCredentials(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)

	// Simulate a prior Authenticated() call that sets perRPCCredentials
	b.Authenticated(&oauth2PerRPCCreds{}, nil)

	if b.perRPCCredentials == nil {
		t.Fatal("expected perRPCCredentials to be set after Authenticated()")
	}

	// Switch to OAuth2 mode with nil credentials — must clear stale per-RPC creds
	b.OAuth2ClientAuthenticated(nil, nil)

	if b.perRPCCredentials != nil {
		t.Error("expected OAuth2ClientAuthenticated(nil, nil) to clear stale perRPCCredentials")
	}

	info := b.channelCredentials.Info()
	if info.SecurityProtocol != "tls" {
		t.Errorf("expected security protocol 'tls' after OAuth2ClientAuthenticated(nil, nil), got %q", info.SecurityProtocol)
	}
}

func TestOAuth2PerRPCCredsAlwaysRequireTransportSecurity(t *testing.T) {
	creds := &oauth2PerRPCCreds{}
	if !creds.RequireTransportSecurity() {
		t.Error("expected RequireTransportSecurity() to always return true")
	}
}

func TestSetChannelCredentialsOrDefaultDetectsInsecureTransport(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub)
	b.setChannelCredentialsOrDefault(insecure.NewCredentials())

	if !b.insecure {
		t.Error("expected insecure flag to be true when insecure credentials are provided")
	}

	info := b.channelCredentials.Info()
	if info.SecurityProtocol != "insecure" {
		t.Errorf("expected security protocol 'insecure', got %q", info.SecurityProtocol)
	}
}

func TestBuildRejectsOAuth2WithInsecureChannel(t *testing.T) {
	creds := auth.NewOAuth2ClientCredentials("test-id", "test-secret", "https://issuer.example.com")
	b := NewClientBuilder("localhost:8080", noopStub).
		OAuth2ClientAuthenticated(&creds, insecure.NewCredentials())

	_, _, err := b.Build()
	if err == nil {
		t.Fatal("expected error when combining OAuth2 credentials with insecure channel")
	}
	if !strings.Contains(err.Error(), "can not authenticate with insecure channel") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildRejectsAuthenticatedWithSecureCredsOnInsecureChannel(t *testing.T) {
	// The exported kesselgrpc.OAuth2CallCredentials always requires transport security.
	// Simulate that with oauth2PerRPCCreds (also always requires transport security).
	b := NewClientBuilder("localhost:8080", noopStub).
		Authenticated(&oauth2PerRPCCreds{}, insecure.NewCredentials())

	_, _, err := b.Build()
	if err == nil {
		t.Fatal("expected error when combining secure per-RPC credentials with insecure channel")
	}
	if !strings.Contains(err.Error(), "can not authenticate with insecure channel") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildAllowsCustomCredsNotRequiringTransportSecurityOnInsecureChannel(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub).
		Authenticated(insecurePerRPCCreds{}, insecure.NewCredentials())

	_, conn, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()
}

func TestBuildAllowsInsecureWithNoCreds(t *testing.T) {
	b := NewClientBuilder("localhost:8080", noopStub).
		Insecure()

	_, conn, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()
}

func TestBuildAllowsOAuth2WithTLS(t *testing.T) {
	creds := auth.NewOAuth2ClientCredentials("test-id", "test-secret", "https://issuer.example.com")
	b := NewClientBuilder("localhost:8080", noopStub).
		OAuth2ClientAuthenticated(&creds, nil)

	_, conn, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("failed to close gRPC connection: %v", closeErr)
		}
	}()
}
