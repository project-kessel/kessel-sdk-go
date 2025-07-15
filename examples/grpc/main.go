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
		WithMaxSendMessageSize(8 * 1024 * 1024).    // 8MB
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

	// Get call options using explicit token handling
	tokenOpts, err := inventoryClient.GetTokenCallOption()
	if err != nil {
		fmt.Printf("Failed to get token call options: %v\n", err)
		return
	}
	fmt.Println("\nUsing explicit token handling:")
	response2, err := inventoryClient.KesselInventoryService.Check(ctx, checkRequest, tokenOpts...)
	if err != nil {
		fmt.Printf("Request failed with error: %v\n", err)
		return
	}
	fmt.Printf("Check response: %+v\n", response2)

	// Method : Get call options using built-in token source integration
	// Doesn't work with Keycloak
	// HTTP Request to /kessel.inventory.v1beta2.KesselInventoryService/Check failed, error id: 398c029e-cf41-4b6e-844a-0e411a11df79-1: io.vertx.core.http.StreamResetException: Stream reset: 8
	//callOpts := inventoryClient.GetCallOptions()
	//fmt.Println("Using built-in token source integration:")
	//response, err := inventoryClient.KesselInventoryService.Check(ctx, checkRequest, callOpts...)
	//if err != nil {
	//	fmt.Printf("Request failed with error: %v\n", err)
	//	return
	//}
	//fmt.Printf("Check response: %+v\n", response)
}

// Example showing different builder configurations
func demonstrateBuilderOptions() {
	ctx := context.Background()

	// Example 1: Basic insecure client
	client1, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("localhost:9000").
		WithInsecure(true).
		Build()
	if err != nil {
		log.Printf("Failed to create basic client: %v", err)
	}
	defer client1.Close()

	// Example 2: Secure client with OAuth2 and custom message sizes
	client2, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("secure.example.com:9000").
		WithOAuth2("client-id", "client-secret", "https://auth.example.com/token").
		WithMaxReceiveMessageSize(16 * 1024 * 1024). // 16MB
		WithMaxSendMessageSize(4 * 1024 * 1024).     // 4MB
		Build()
	if err != nil {
		log.Printf("Failed to create secure client: %v", err)
	}
	defer client2.Close()

	// Example 3: Client with custom TLS and OAuth2 issuer
	client3, err := v1beta2.NewInventoryGRPCClientBuilder().
		WithEndpoint("secure.example.com:9000").
		WithOAuth2Issuer("client-id", "client-secret", "https://auth.example.com", "read", "write").
		Build()
	if err != nil {
		log.Printf("Failed to create TLS client: %v", err)
	}
	defer client3.Close()

	// Use the clients...
	_ = ctx
}
