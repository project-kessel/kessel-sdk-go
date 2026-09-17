package builder

import (
	"crypto/tls"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// noopStub is a minimal stub constructor for testing the generic ClientBuilder.
func noopStub(_ grpc.ClientConnInterface) struct{} { return struct{}{} }

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
	defer conn.Close()
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

func TestOAuth2PerRPCCredsRequireTransportSecurity(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		expected bool
	}{
		{"secure mode requires transport security", false, true},
		{"insecure mode does not require transport security", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &oauth2PerRPCCreds{insecure: tt.insecure}
			if got := creds.RequireTransportSecurity(); got != tt.expected {
				t.Errorf("RequireTransportSecurity() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
