package v1beta2

import (
	"testing"
	"time"
)

func TestKeepaliveOptionsAreAvailableFromV1beta2(t *testing.T) {
	tests := []struct {
		name   string
		option KeepaliveOption
	}{
		{
			name:   "interval",
			option: WithKeepaliveInterval(0),
		},
		{
			name:   "timeout",
			option: WithKeepaliveTimeout(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewClientBuilder("127.0.0.1:8080").Keepalive(
				WithKeepaliveInterval(15*time.Second),
				WithKeepaliveTimeout(5*time.Second),
				WithKeepalivePermitWithoutCalls(false),
				tt.option,
			)
			client, conn, err := builder.Build()
			if err == nil {
				t.Fatal("Build() succeeded with a non-positive keepalive duration")
			}
			if client != nil {
				t.Errorf("Build() client = %v, want nil zero value", client)
			}
			if conn != nil {
				t.Errorf("Build() connection = %v, want nil", conn)
			}
		})
	}
}
