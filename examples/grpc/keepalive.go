package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
	if err := keepalive(); err != nil {
		log.Printf("Keepalive example failed: %v", err)
		os.Exit(1)
	}
}

func keepalive() error {
	endpoint := os.Getenv("KESSEL_ENDPOINT")
	if endpoint == "" {
		return fmt.Errorf("KESSEL_ENDPOINT is required")
	}

	if err := buildWithDefaultKeepalive(endpoint); err != nil {
		return err
	}

	return buildWithKeepaliveOverrides(endpoint)
}

func buildWithDefaultKeepalive(endpoint string) error {
	_, conn, err := v1beta2.NewClientBuilder(endpoint).
		Insecure().
		Build()
	if err != nil {
		return fmt.Errorf("failed to create gRPC client with default keepalive: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	fmt.Println("Built a client with the default keepalive settings: 45s interval, 10s acknowledgment timeout, and pings permitted without active calls.")
	fmt.Println("No RPC was sent.")
	return nil
}

func buildWithKeepaliveOverrides(endpoint string) error {
	_, conn, err := v1beta2.NewClientBuilder(endpoint).
		Keepalive(
			v1beta2.WithKeepaliveInterval(60*time.Second),
			v1beta2.WithKeepaliveTimeout(15*time.Second),
			v1beta2.WithKeepalivePermitWithoutCalls(false),
		).
		Insecure().
		Build()
	if err != nil {
		return fmt.Errorf("failed to create gRPC client with keepalive overrides: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	fmt.Println("Built a client with a 60s interval, 15s acknowledgment timeout, and no pings without active calls.")
	fmt.Println("No RPC was sent.")
	return nil
}
