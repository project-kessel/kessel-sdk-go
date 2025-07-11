package main

import (
	"context"
	"fmt"
	"log"

	"github.com/project-kessel/kessel-sdk-go/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/kessel/errors"
	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func main() {
	ctx := context.Background()

	// The SDK supports two OAuth2 configuration approaches:
	// 1. Direct token URL: Specify the exact OAuth2 token endpoint URL
	// 2. Issuer-based discovery: Provide the issuer URL and let the SDK discover
	//    the token endpoint via OpenID Connect Discovery (/.well-known/openid_configuration)

	// Create gRPC configuration with OAuth2
	grpcConfig := config.NewGRPCConfig(
		config.WithGRPCEndpoint("127.0.0.1:9000"),
		config.WithGRPCInsecure(true),
		config.WithGRPCOAuth2Issuer("svc-test", "h91qw8bPiDj9R6VSORsI5TYbceGU5PMH", "http://localhost:8085/realms/redhat-external"),
		// Option 1: Direct token URL configuration
		//config.WithGRPCOAuth2(
		//	"svc-test",
		//	"h91qw8bPiDj9R6VSORsI5TYbceGU5PMH",
		//	"http://localhost:8085/realms/redhat-external/protocol/openid-connect/token",
		//),
		// Option 2: Issuer-based configuration with auto-discovery
		//config.WithGRPCOAuth2Issuer(
		//	"svc-test",
		//	"h91qw8bPiDj9R6VSORsI5TYbceGU5PMH",
		//	"http://localhost:8085/realms/redhat-external",
		//),
	)

	// Create the inventory gRPC client directly
	inventoryClient, err := v1beta2.NewInventoryGRPCClient(grpcConfig)
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

	// Method 1: Get call options using explicit token handling
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
