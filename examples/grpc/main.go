package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	inventoryClient, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("127.0.0.1:9000").
		WithInsecure(true).
		WithOAuth2Issuer("svc-test", "h91qw8bPiDj9R6VSORsI5TYbceGU5PMH", "http://localhost:8085/realms/redhat-external").
		WithMaxReceiveMessageSize(8 * 1024 * 1024). // 8MB
		WithMaxSendMessageSize(8 * 1024 * 1024). // 8MB
		Build()

	if err != nil {
		// Example of checking for specific error types using sentinel errors
		if errors.IsConnectionError(err) {
			log.Fatal("Failed to establish connection:", err)
		} else if errors.IsTokenError(err) {
			log.Fatal("OAuth2 token configuration failed:", err)
		} else {
			log.Fatal("Unknown error:", err)
		}
	}
	defer func() {
		if closeErr := inventoryClient.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC client: %v", closeErr)
		}
	}()

	// Example request using the external API types
	checkRequest := &v1beta2.CheckRequest{
		Object: &v1beta2.ResourceReference{
			ResourceType: "host",
			ResourceId:   "1213",
			Reporter: &v1beta2.ReporterReference{
				Type: "HBI",
			},
		},
		Relation: "member",
		Subject: &v1beta2.SubjectReference{
			Resource: &v1beta2.ResourceReference{
				ResourceType: "user",
				ResourceId:   "tim",
			},
		},
	}

	// OAuth2 tokens are automatically injected - no manual token management needed!
	fmt.Println("Making request with automatic OAuth2 token injection:")
	response, err := inventoryClient.Check(ctx, checkRequest)
	if err != nil {
		fmt.Printf("Request failed with error: %v\n", err)
		return
	}
	fmt.Printf("Check response: %+v\n", response)
}

// Example showing different builder configurations
func demonstrateBuilderOptions() {
	ctx := context.Background()

	// Example 1: Basic insecure client (no OAuth)
	client1, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("localhost:9000").
		WithInsecure(true).
		Build()
	if err != nil {
		log.Printf("Failed to create basic client: %v", err)
		return
	}
	defer client1.Close()

	// Example 2: Secure client with OAuth2 and custom message sizes
	client2, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("secure.example.com:9000").
		WithOAuth2("client-id", "client-secret", "https://auth.example.com/token").
		WithMaxReceiveMessageSize(16 * 1024 * 1024). // 16MB
		WithMaxSendMessageSize(4 * 1024 * 1024). // 4MB
		Build()
	if err != nil {
		log.Printf("Failed to create secure client: %v", err)
		return
	}
	defer client2.Close()

	// Example 3: Client with custom TLS and OAuth2 issuer
	client3, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("secure.example.com:9000").
		WithOAuth2Issuer("client-id", "client-secret", "https://auth.example.com", "read", "write").
		Build()
	if err != nil {
		log.Printf("Failed to create TLS client: %v", err)
		return
	}
	defer client3.Close()

	// Use the clients directly - tokens are automatically injected
	request := &v1beta2.CheckRequest{
		Object: &v1beta2.ResourceReference{
			ResourceType: "host",
			ResourceId:   "server-123",
		},
		Relation: "member",
		Subject: &v1beta2.SubjectReference{
			Resource: &v1beta2.ResourceReference{
				ResourceType: "user",
				ResourceId:   "alice",
			},
		},
	}

	// All clients work the same way - just call the methods directly
	response1, err := client1.Check(ctx, request)
	if err != nil {
		log.Printf("Client 1 failed: %v", err)
	} else {
		log.Printf("Client 1 response: %+v", response1)
	}

	response2, err := client2.Check(ctx, request)
	if err != nil {
		log.Printf("Client 2 failed: %v", err)
	} else {
		log.Printf("Client 2 response: %+v", response2)
	}

	response3, err := client3.Check(ctx, request)
	if err != nil {
		log.Printf("Client 3 failed: %v", err)
	} else {
		log.Printf("Client 3 response: %+v", response3)
	}
}
